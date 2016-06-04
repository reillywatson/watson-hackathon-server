package main

import (
	"fmt"
	"github.com/gorilla/websocket"
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

type handlerFn func(Socket, map[string]interface{}) error

var handlers = map[string]handlerFn{}

func addHandler(msgType string, handler handlerFn) {
	handlers[msgType] = handler
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Socket interface {
	Send(msg interface{}) error
}

type WSSocket websocket.Conn

func (s *WSSocket) Send(msg interface{}) error {
	return (*websocket.Conn)(s).WriteJSON(msg)
}

type request struct {
	Info        map[string]interface{} `json:"info"`
	MessageType string                 `json:"message_type"`
}

func invalidMessage(s Socket, r request) {
	fmt.Println("Invalid message:", r)
}

func reply(s Socket, msgType string, info map[string]interface{}) error {
	return s.Send(map[string]interface{}{"message_type": msgType, "info": info})
}

func helloworld(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Hello world!\n"))
}

func handleWs(w http.ResponseWriter, req *http.Request) {
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
		if err != nil {
			fmt.Println("socket err:", err)
			break
		}
		handler, ok := handlers[message.MessageType]
		if !ok {
			invalidMessage(socket, message)
			continue
		}
		handler(socket, message.Info)
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
	bot := &Chatbot{}
	addHandler("chatbot_send", bot.GotMessage)
	http.HandleFunc("/ws", handleWs)
	http.HandleFunc("/", helloworld)
	log.Printf("Starting app on port %+v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
