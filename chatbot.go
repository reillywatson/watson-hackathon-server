package main

import (
	"fmt"
)

type Chatbot struct{}

func (c *Chatbot) GotMessage(s Socket, info map[string]interface{}) error {
	fmt.Println("Got message:", info)
	msg := "hello world!"
	response := map[string]interface{}{
		"message": msg,
	}
	return reply(s, "chatbot_receive", response)
}
