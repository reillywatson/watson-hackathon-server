package handlers

import (
	"errors"
)

var NotFound = errors.New("No handler found")

type Socket interface {
	Send(msg interface{}) error
}

type handlerFn func(Socket, map[string]interface{}) error

var handlers = map[string]handlerFn{}

func AddHandler(msgType string, handler handlerFn) {
	handlers[msgType] = handler
}

func CallHandler(s Socket, msgType string, info map[string]interface{}) error {
	h, ok := handlers[msgType]
	if ok {
		return h(s, info)
	}
	return NotFound
}

func Reply(s Socket, msgType string, info map[string]interface{}) error {
	return s.Send(map[string]interface{}{"message_type": msgType, "info": info})
}
