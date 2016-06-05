package chatbot

import (
	"fmt"
	"github.com/reillywatson/watson-hackathon-server/handlers"
	"github.com/reillywatson/watson-hackathon-server/util"
	"github.com/reillywatson/watson-hackathon-server/watson"
	"log"
	"math"
	"sort"
	"strings"
	"time"
)

const SenderUser = "user"
const SenderBot = "bot"
const MaxDatapoints = 50000

type Message struct {
	Id        string    `json:"id"`
	Sender    string    `json:"sender"`
	Text      string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Conversation struct {
	Id            string                 `json:"id"`
	Messages      []Message              `json:"messages"`
	Sensors       SensorData             `json:"sensors"`
	LearnerStates []LearnerState         `json:"learner_states"`
	CustomData    map[string]interface{} `json:"custom_data"`
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

func (s SensorData) Avg() float64 {
	if len(s) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, d := range s {
		sum += d.Value
	}
	return sum / float64(len(s))
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

func (s SensorData) ByTypes(types []string) SensorData {
	var result SensorData
	for _, d := range s {
		for _, t := range types {
			if d.Type == t {
				result = append(result, d)
				break
			}
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
	Learn(c *Conversation, data SensorData) LearnerState
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

func (h HeartbeatLearner) Learn(c *Conversation, data SensorData) LearnerState {
	data = data.Since(time.Minute).RemoveOutliers()
	max := data.Max()
	min := data.Min()
	if data.Max() == 0 || len(data) == 0 {
		return NoState
	}
	current := data[len(data)-1].Value
	c.CustomData["current_pulse"] = current
	allTimeMax, _ := c.CustomData["max_pulse"].(float64)
	if current > allTimeMax {
		c.CustomData["max_pulse"] = current
	}
	if max-min > 5 {
		if current == max {
			return HeartrateIncreasing
		}
		if current == min {
			return HeartrateDecreasing
		}
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

func (a AccelerometerLearner) Learn(c *Conversation, data SensorData) LearnerState {
	return NoState
}

func (c *Chatbot) Sensor(s handlers.Socket, req map[string]interface{}) error {
	var info struct {
		Datapoint      `json:",inline"`
		ConversationId string `json:"conversation_id"`
		Echo           bool
	}
	util.ToStruct(req, &info)
	info.Datapoint.Timestamp = time.Now()
	conv, err := c.GetConversation(info.ConversationId)
	if err != nil {
		return err
	}
	conv.Sensors = append(conv.Sensors, info.Datapoint)
	if len(conv.Sensors) > MaxDatapoints {
		conv.Sensors = conv.Sensors[MaxDatapoints:]
	}
	for i, learner := range learners {
		state := learner.Learn(conv, conv.Sensors.ByTypes(learner.SensorTypes()))
		if len(conv.LearnerStates) <= i {
			conv.LearnerStates = append(conv.LearnerStates, state)
		}
		if state != conv.LearnerStates[i] {
			conv.LearnerStates[i] = state
			switch state {
			case HeartrateIncreasing:
				c.SendMessage(s, conv, "Your heart rate is going through the roof!", map[string]interface{}{"vibrate": true})
			case HeartrateDecreasing:
				c.SendMessage(s, conv, "Your heart rate is dropping like a rock, better get to work!", map[string]interface{}{"vibrate": true})
			}
		}
	}
	if info.Echo {
		return s.Reply("got_sensor", req)
	}
	return nil
}

func (c *Chatbot) Init(s handlers.Socket, req map[string]interface{}) error {
	conv := &Conversation{
		Id:         util.NewId(),
		CustomData: map[string]interface{}{},
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
	return c.ProcessMessage(s, conversation, msg)
}

func (c *Chatbot) SendMessage(s handlers.Socket, conversation *Conversation, text string, payload map[string]interface{}) error {
	msg := Message{
		Id:        util.NewId(),
		Text:      text,
		Sender:    SenderBot,
		Timestamp: time.Now(),
	}
	conversation.Messages = append(conversation.Messages, msg)
	reply := util.ToMap(msg)
	for k, v := range payload {
		reply[k] = v
	}
	return s.Reply("chatbot_receive", reply)
}

func (c *Chatbot) ProcessMessage(s handlers.Socket, conv *Conversation, msg Message) error {
	fmt.Println("CUSTOM DATA:", util.ToJson(conv.CustomData))
	conv.Messages = append(conv.Messages, msg)
	classes, err := watson.Classify(msg.Text)
	if err != nil {
		log.Printf("Error classifying: %v", err)
	}
	class := "confused"
	if len(classes) > 0 && classes[0].Confidence > 0.8 {
		class = classes[0].Name
	}
	if class == "gender" {
		text := strings.ToLower(msg.Text)
		if text == "female" {
			conv.CustomData["gender"] = "female"
		} else if text == "male" {
			conv.CustomData["gender"] = "male"
		} else {
			return c.SendMessage(s, conv, "I didn't understand that. Are you male or female?", nil)
		}
	}
	return c.SendMessage(s, conv, messageForClass(class, conv.CustomData), nil)
}
