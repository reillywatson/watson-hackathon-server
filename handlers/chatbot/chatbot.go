package chatbot

import (
	"fmt"
	"github.com/reillywatson/watson-hackathon-server/handlers"
	"github.com/reillywatson/watson-hackathon-server/util"
	"math"
	"sort"
	"time"
)

const SenderUser = "user"
const SenderBot = "bot"
const MaxDatapoints = 50000

type Message struct {
	Id        string    `json:"id"`
	Sender    string    `json:"sender"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type Conversation struct {
	Id            string         `json:"id"`
	Messages      []Message      `json:"messages"`
	Sensors       SensorData     `json:"sensors"`
	LearnerStates []LearnerState `json:"learner_states"`
}

type SensorData []Datapoint

func (s SensorData) Max() float64 {
	max := 0.0
	for _, d := range s {
		if d.Value > max {
			max = d.Value
		}
	}
	return max
}

func (s SensorData) Min() float64 {
	min := math.MaxFloat64
	if len(s) == 0 {
		return 0.0
	}
	for _, d := range s {
		if d.Value < min {
			min = d.Value
		}
	}
	return min
}

type byValue SensorData

func (b byValue) Len() int           { return len(b) }
func (b byValue) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byValue) Less(i, j int) bool { return b[i].Value < b[j].Value }

func (s SensorData) SortedByValue() SensorData {
	var sorted SensorData
	for _, a := range s {
		sorted = append(sorted, a)
	}
	sort.Sort(byValue(sorted))
	return sorted
}

func (s SensorData) Quartile(quartile int) float64 {
	if len(s) == 0 {
		return 0.0
	}
	divisionPoint := (len(s) - 1) * quartile / 4
	if len(s)%2 == 1 {
		return s[divisionPoint].Value
	}
	return (s[divisionPoint].Value + s[divisionPoint+1].Value) / 2
}

// Needs to be sorted first!
func (s SensorData) RemoveOutliers() SensorData {
	if len(s) == 0 {
		return s
	}
	lowerQuartile := s.Quartile(1)
	upperQuartile := s.Quartile(3)
	iqr := upperQuartile - lowerQuartile
	lowerBound := lowerQuartile - (iqr * 1.5)
	upperBound := upperQuartile + (iqr * 1.5)
	var result SensorData
	for _, x := range s {
		if x.Value >= lowerBound || x.Value <= upperBound {
			result = append(result, x)
		}
	}
	return result
}

func (s SensorData) Since(dur time.Duration) SensorData {
	var result SensorData
	for _, d := range s {
		if time.Since(d.Timestamp) < dur {
			result = append(result, d)
		}
	}
	return result
}

type Datapoint struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type Chatbot struct {
	Conversations map[string]*Conversation
}

func init() {
	chatbot := Chatbot{Conversations: make(map[string]*Conversation)}
	handlers.AddNewSocketHandler(chatbot.Init)
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

type Learner interface {
	SensorTypes() []string
	Learn(data SensorData) LearnerState
}

type LearnerState int

const (
	NoState                          = iota
	HeartrateIncreasing LearnerState = iota
	HeartrateDecreasing LearnerState = iota
	HeartrateStable     LearnerState = iota
)

type HeartbeatLearner struct{}

func (h HeartbeatLearner) SensorTypes() []string {
	return []string{"heartbeat"}
}

func (h HeartbeatLearner) Learn(data SensorData) LearnerState {
	data = data.Since(time.Minute).RemoveOutliers()
	min := data.Min()
	max := data.Max()
	if max == 0 || len(data) < 5 {
		return NoState
	}
	if max-min > 5 {
		return HeartrateIncreasing
	}
	if max-min < -5 {
		return HeartrateDecreasing
	}
	return NoState
}

var learners = []Learner{
	HeartbeatLearner{},
	AccelerometerLearner{},
}

type AccelerometerLearner struct{}

func (a AccelerometerLearner) SensorTypes() []string {
	return []string{"accelerometer", "linear_acceleration"}
}

func (a AccelerometerLearner) Learn(data SensorData) LearnerState {
	return NoState
}

func (c *Chatbot) Sensor(s handlers.Socket, req map[string]interface{}) error {
	var info struct {
		Datapoint      `json:",inline"`
		ConversationId string `json:"conversation_id"`
		Echo           bool
	}
	util.ToStruct(req, &info)
	conv, err := c.GetConversation(info.ConversationId)
	if err != nil {
		return err
	}
	conv.Sensors = append(conv.Sensors, info.Datapoint)
	if info.Echo {
		return s.Reply("got_sensor", req)
	}
	if len(conv.Sensors) > MaxDatapoints {
		conv.Sensors = conv.Sensors[MaxDatapoints:]
	}
	for i, learner := range learners {
		state := learner.Learn(conv.Sensors)
		if len(conv.LearnerStates) <= i {
			conv.LearnerStates = append(conv.LearnerStates, state)
		}
		if state != conv.LearnerStates[i] {
			switch state {
			case HeartrateIncreasing:
				c.SendMessage(s, conv, "Your heart rate is going through the roof!")
			case HeartrateDecreasing:
				c.SendMessage(s, conv, "Your heart rate is dropping like a rock, better get to work!")
			}
		}
	}
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
		Sender:    SenderUser,
		Timestamp: time.Now(),
	}
	conversation.Messages = append(conversation.Messages, msg)
	return c.SendMessage(s, conversation, "hello world!")
}

func (c *Chatbot) SendMessage(s handlers.Socket, conversation *Conversation, text string) error {
	msg := Message{
		Id:        util.NewId(),
		Text:      text,
		Sender:    SenderBot,
		Timestamp: time.Now(),
	}
	conversation.Messages = append(conversation.Messages, msg)
	return s.Reply("chatbot_receive", map[string]interface{}{
		"message": msg,
	})
}
