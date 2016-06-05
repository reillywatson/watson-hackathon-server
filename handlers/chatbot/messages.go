package chatbot

import (
	"bytes"
	"math/rand"
	"text/template"
)

var messagesByClass = map[string][]string{
	"confused":           {"I don't understand!", "I'm confused!", "I don't get it"},
	"Current heart rate": {"Your current heart rate is {{.current_pulse}} BPM"},
	"Overall heart rate": {"?"},
	"Calories":           {"?"},
	"Duration":           {"?"},
	"Improvement":        {"?"},
	"Daily Calories":     {"?"},
}

func messageForClass(class string, vars map[string]interface{}) string {
	msgs, ok := messagesByClass[class]
	if !ok {
		msgs = messagesByClass["confused"]
	}
	msg := msgs[rand.Intn(len(msgs))]
	t, err := template.New("letter").Parse(msg)
	if err != nil {
		return msg
	}
	var buf bytes.Buffer
	t.Execute(&buf, vars)
	return buf.String()
}
