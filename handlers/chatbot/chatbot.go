package chatbot

import (
	"fmt"
	"github.com/reillywatson/watson-hackathon-server/handlers"
	"github.com/reillywatson/watson-hackathon-server/util"
	"time"
)

type Message struct {
	Id        string    `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type Conversation struct {
	Id       string    `json:"id"`
	Messages []Message `json:"messages"`
}

type Chatbot struct {
	Conversations map[string]*Conversation
}

func init() {
	chatbot := Chatbot{Conversations: make(map[string]*Conversation)}
	handlers.AddHandler("chatbot_init", chatbot.Init)
	handlers.AddHandler("chatbot_send", chatbot.GotMessage)
	handlers.AddHandler("chatbot_end", chatbot.End)
	handlers.AddHandler("sensor", chatbot.Sensor)
}

func (c *Chatbot) GetConversation(id string) (*Conversation, error) {
	conversation, ok := c.Conversations[id]
	if !ok {
		return nil, fmt.Errorf("Conversation not found!")
	}
	return conversation, nil
}

func (c *Chatbot) Sensor(s handlers.Socket, req map[string]interface{}) error {
	return nil
}

func (c *Chatbot) Init(s handlers.Socket, req map[string]interface{}) error {
	conv := &Conversation{
		Id: util.NewId(),
	}
	c.Conversations[conv.Id] = conv
	return s.Reply("chatbot_initialized", util.ToMap(conv))
}

func (c *Chatbot) End(s handlers.Socket, req map[string]interface{}) error {
	var info struct {
		ConversationId string `json:"conversation_id"`
	}
	util.ToStruct(req, &info)
	delete(c.Conversations, info.ConversationId)
	return s.Reply("chatbot_ended", req)
}

func (c *Chatbot) GotMessage(s handlers.Socket, req map[string]interface{}) error {
	var info struct {
		ConversationId string `json:"conversation_id"`
		Message        string `json:"message"`
	}
	util.ToStruct(req, &info)
	conversation, err := c.GetConversation(info.ConversationId)
	if err != nil {
		return err
	}

	msg := Message{
		Id:        util.NewId(),
		Text:      info.Message,
		Timestamp: time.Now(),
	}
	conversation.Messages = append(conversation.Messages, msg)
	fmt.Println("Got message:", info)
	return s.Reply("chatbot_receive", map[string]interface{}{
		"message": "hello world!",
		"last":    msg,
	})
}
