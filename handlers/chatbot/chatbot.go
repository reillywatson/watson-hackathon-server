package chatbot

import (
	"fmt"
	"github.com/reillywatson/watson-hackathon-server/handlers"
)

type Chatbot struct{}

func (c *Chatbot) GotMessage(s handlers.Socket, info map[string]interface{}) error {
	fmt.Println("Got message:", info)
	msg := "hello world!"
	response := map[string]interface{}{
		"message": msg,
		"last":    info["message"],
	}
	return handlers.Reply(s, "chatbot_receive", response)
}
