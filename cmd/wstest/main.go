package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strings"
)

type req struct {
	MessageType string                 `json:"message_type"`
	Info        map[string]interface{} `json:"info"`
}

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
	var conv req
	ws.ReadJSON(&conv)
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for {
			var resp map[string]interface{}
			err := ws.ReadJSON(&resp)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(resp)
		}
	}()
	convId := conv.Info["id"].(string)
	fmt.Println("ID:", convId)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		var info map[string]interface{}
		err := json.Unmarshal([]byte(parts[1]), &info)
		if err != nil {
			fmt.Println(err)
			continue
		}
		info["conversation_id"] = convId
		msg := req{
			MessageType: parts[0],
			Info:        info,
		}
		fmt.Println("Writing:", msg)
		ws.WriteJSON(&msg)
	}
}
