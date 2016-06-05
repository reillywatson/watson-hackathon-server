package chatbot

import (
	"bytes"
	"fmt"
	"github.com/reillywatson/watson-hackathon-server/util"
	"math/rand"
	"regexp"
	"text/template"
)

var messagesByClass = map[string][]string{
	"confused":           {"I don't understand!", "I'm confused!", "I don't get it"},
	"Current heart rate": {"Your current heart rate is {{.current_pulse}} BPM"},
	"Bad Calories":       {"Don't eat it! you dummy"},
	"Daily Calories": {
		"Some calorie sources are better than others. Calories from sugar without fiber are empty calories. While 250 calories from a candy bar are utilized the same as the combined 250 calories from a banana, an apple, and a pear, the fruit is obviously much better for you.",
		"The number of calories you need depends on your age, body size, and activity levels -- most teens and adults need somewhere around 1,500 to 2,500 calories per day.",
	},
	"Calories":       {"{{if false}}{{.gender}}{{.age}}{{.weight}}{{.height}}{{end}}You should be burning around {{.daily_calories}} calories per day."},
	"Jumping Jack":   {"https://www.youtube.com/watch?v=p64YlMRIDVM"},
	"Pushups":        {"https://www.youtube.com/watch?v=Eh00_rniF8E"},
	"Situp":          {"https://www.youtube.com/watch?v=1fbU_MkV7NE"},
	"Max heart rate": {"Your recommended maximum heart rate is {{.target_pulse}}. Your max heart rate so far is {{.max_pulse}}"},
	"Exercise Time":  {"You'll need one minute. Ready?"},
	"Time":           {"Time is not important, the fact that you're doing it is. Keep it up!"},
	"Improvement":    {"You're doing better and better, keep it up!"},
}

/*
"Duration":       {"You have been doing it for {{.duration}}"},
Calories	"""<if {{.gender}} == """"female"""" then <655 + (4.35 x {{.weight_in_pounds}}) + (4.7 x {{.height_in _inches}}) - (4.7 x {{.age}}) else <66 + (6.23 x {{.weight_in_pounds}}) + (12.7 x {{.height_in _inches}}) - (6.8 x {{.age}})"""
//Calories	"<if {{.gender}} == ""female"" then <655 + (4.35 x {{.weight_in_pounds}}) + (4.7 x {{.height_in _inches}}) - (4.7 x {{.age}}) else <66 + (6.23 x {{.weight_in_pounds}}) + (12.7 x {{.height_in _inches}}) - (6.8 x {{.age}})"

Overall heart rate	You recomended maximum heart rate is <220 - {{.age}}>
Overall heart rate	Your target heart rate is {{.max_pulse}}
Target heart rate	Your target heart rate is  {{.max_pulse}}
*/

type BotState struct {
	DataNeeded []string
	EndClass   string
}

func messageForClass(class string, state *BotState, vars map[string]interface{}) (string, *BotState) {
	msgs, ok := messagesByClass[class]
	if !ok {
		return fmt.Sprintf("No message found. Class: %v", class), nil
		msgs = messagesByClass["confused"]
	}
	msg := msgs[rand.Intn(len(msgs))]
	params := regexp.MustCompile(`{{\.(\w+)}}`).FindAllString(msg, -1)
	dataNeeded := []string{}
	for _, param := range params {
		param = param[3 : len(param)-2]
		fmt.Println("PARAM:", param, "VARS:", util.ToJson(vars))
		if _, ok := vars[param]; !ok {
			fmt.Println("Needed!")
			dataNeeded = append(dataNeeded, param)
		}
	}
	if len(dataNeeded) > 0 {
		if state == nil {
			state = &BotState{
				DataNeeded: dataNeeded,
				EndClass:   class,
			}
		}
		switch dataNeeded[0] {
		case "gender":
			return "We need some information first. Are you male or female?", state
		case "age", "target_pulse":
			state.DataNeeded[0] = "age"
			return "What is your age?", state
		case "height":
			return "How tall are you, in inches?", state
		case "weight":
			return "How much do you weigh?", state
		case "max_pulse", "current_pulse":
			return "We don't have any pulse information yet!", nil
		}
	}
	if class == "Calories" {

	}
	t, err := template.New("letter").Parse(msg)
	if err != nil {
		return msg, nil
	}

	var buf bytes.Buffer
	t.Execute(&buf, vars)
	return buf.String(), nil
}
