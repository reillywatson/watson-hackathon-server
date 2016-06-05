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
	"Duration":           {"?"},
	"Improvement":        {"?"},
	"Daily Calories":     {"?"},
}

/*
"Bad Calories":	{"Don't eat it! you dummy"}
Calories	"""<if {{.gender}} == """"female"""" then <655 + (4.35 x {{.weight_in_pounds}}) + (4.7 x {{.height_in _inches}}) - (4.7 x {{.age}}) else <66 + (6.23 x {{.weight_in_pounds}}) + (12.7 x {{.height_in _inches}}) - (6.8 x {{.age}})"""
//Calories	"<if {{.gender}} == ""female"" then <655 + (4.35 x {{.weight_in_pounds}}) + (4.7 x {{.height_in _inches}}) - (4.7 x {{.age}}) else <66 + (6.23 x {{.weight_in_pounds}}) + (12.7 x {{.height_in _inches}}) - (6.8 x {{.age}})"
Daily Calories	Some calorie sources are better than others. Calories from sugar without fiber are empty calories. While 250 calories from a candy bar are utilized the same as the combined 250 calories from a banana, an apple, and a pear, the fruit is obviously much better for you.
Daily Calories	The number of calories you need depends on your age, body size, and activity levels -- most teens and adults need somewhere around 1,500 to 2,500 calories per day.
Duration	You have been doing it for {{.duration}}
Exercise Time	You'll need one minute. Ready?
Improvement	You doing better and better, keep it up!
Jumping Jack	https://www.youtube.com/watch?v=p64YlMRIDVM
Max heart rate	Your max heart rate is {{.max_pulse}}
Overall heart rate	You recomended maximum heart rate is <220 - {{.age}}>
Overall heart rate	Your target heart rate is {{.max_pulse}}
Pushups	https://www.youtube.com/watch?v=Eh00_rniF8E
Situp	https://www.youtube.com/watch?v=1fbU_MkV7NE
Target heart rate	Your target heart rate is  {{.max_pulse}}
Time	Time is not important, the fact that you doing it is. Keep it up
*/

func messageForClass(class string, vars map[string]interface{}) string {
	if class == "Calories" {
		if _, ok := vars["gender"]; !ok {
			return "Are you male or female?"
		}
	}
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
