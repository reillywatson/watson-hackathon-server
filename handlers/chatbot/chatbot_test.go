package chatbot

import (
	"testing"
)

type reply struct {
	msgType string
	info    map[string]interface{}
}

type TestSocket struct {
	got  []reply
	sent []interface{}
}

func (s *TestSocket) Send(msg interface{}) error {
	s.sent = append(s.sent, msg)
	return nil
}
func (s *TestSocket) Reply(msgType string, info map[string]interface{}) error {
	s.got = append(s.got, reply{msgType, info})
	return nil
}
func (s *TestSocket) Error(msg string) error {
	return s.Reply("error", map[string]interface{}{"msg": msg})
}

func TestSensor(t *testing.T) {
	c := Chatbot{Conversations: map[string]*Conversation{}}
	s := &TestSocket{}
	c.Init(s, nil)
	conversationId := ""
	for k := range c.Conversations {
		conversationId = k
		break
	}
	c.Sensor(s, map[string]interface{}{
		"conversation_id": conversationId,
		"type":            "heartbeat",
		"value":           70.0,
	})
	c.Sensor(s, map[string]interface{}{
		"conversation_id": conversationId,
		"type":            "heartbeat",
		"value":           75.0,
	})
}
