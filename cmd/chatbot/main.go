package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/reillywatson/watson-hackathon-server/handlers"
	_ "github.com/reillywatson/watson-hackathon-server/handlers/chatbot"
	"log"
	"net/http"
	"os"
	"time"
	//for extracting service credentials from VCAP_SERVICES
	//"github.com/cloudfoundry-community/go-cfenv"
)

const (
	DEFAULT_PORT = "8080"
)

const pongWait = time.Second * 60

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSSocket websocket.Conn

func (s *WSSocket) Send(msg interface{}) error {
	return (*websocket.Conn)(s).WriteJSON(msg)
}

func (s *WSSocket) Reply(msgType string, info map[string]interface{}) error {
	return s.Send(map[string]interface{}{"message_type": msgType, "info": info})
}

func (s *WSSocket) Error(msg string) error {
	return s.Reply("error", map[string]interface{}{"msg": msg})
}

type request struct {
	Info        map[string]interface{} `json:"info"`
	MessageType string                 `json:"message_type"`
}

func invalidMessage(s handlers.Socket, r request) {
	fmt.Println("Invalid message:", r)
}

func helloworld(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Hello world!\n"))
}

func handleWs(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic: %v\n", r)
		}
	}()
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			fmt.Println("Couldn't handshake!")
		}
		fmt.Println("Other error:", err)
		return
	}
	socket := (*WSSocket)(ws)
	defer ws.Close()
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var message request
		err = ws.ReadJSON(&message)
		fmt.Println(message)
		if err != nil {
			fmt.Println("socket err:", err)
			return
		}
		handlers.CallHandler(socket, message.MessageType, message.Info)
	}
}

func main() {
	log.Printf("App starting!\n")
	defer func() {
		if r := recover(); r != nil {
			log.Panicf("Panic: %v\n", r)
		}
	}()
	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		if port = os.Getenv("VCAP_APP_PORT"); len(port) == 0 {
			port = DEFAULT_PORT
		}
	}
	http.HandleFunc("/ws", handleWs)
	http.HandleFunc("/", helloworld)
	log.Printf("Starting app on port %+v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
