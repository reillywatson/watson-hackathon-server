package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strings"
)

func main() {
	host := "localhost:8080"
	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	dialer := websocket.Dialer{}
	url := fmt.Sprintf("ws://%s/ws", host)
	if strings.Contains(host, "localhost") {
		url = strings.Replace(url, "wss://", "ws://", 1)
	}
	fmt.Println(url)
	ws, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("Err:", err)
				os.Exit(0)
			}
			fmt.Println(string(msg))
		}
	}()
	for scanner.Scan() {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"message_type":"chatbot_send","info":{"message":"%s"}}`, scanner.Text())))
	}
}
