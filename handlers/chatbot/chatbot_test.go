package chatbot

import (
	"testing"
)

type reply struct {
	msgType string
	info    map[string]interface{}
}

type TestSocket struct {
	got []reply
}

func (s *TestSocket) Send(msg interface{}) error {
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

func TestSend(t *testing.T) {
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

	c.GotMessage(s, map[string]interface{}{
		"conversation_id": conversationId,
		"message":         "What's my heart rate?",
	})
	if len(s.got) != 2 {
		t.Fatal("Got nothing!")
	}
	resp, _ := s.got[1].info["message"].(string)
	if resp != "Your current heart rate is 70 BPM" {
		t.Errorf("Got response: %v", s.got[1])
	}
}