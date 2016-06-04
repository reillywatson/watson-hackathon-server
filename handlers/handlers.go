package handlers

import (
	"errors"
)

var NotFound = errors.New("No handler found")

type Socket interface {
	Send(msg interface{}) error
	Reply(msgType string, info map[string]interface{}) error
	Error(msg string) error
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
