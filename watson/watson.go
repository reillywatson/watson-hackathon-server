package watson

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var userName = "f6793223-7d28-446f-849f-373d81b2e504"
var password = "pEntn2Vu6frX"
var classifierId = "3a84dfx64-nlc-21515"

type ClassificationResult struct {
	Name       string  `json:"class_name"`
	Confidence float64 `json:"confidence"`
}

type regexOverride struct {
	Regex *regexp.Regexp
	Class string
}

var regexOverrides = []regexOverride{
	{regexp.MustCompile(`(fe)?male`), "gender"},
}

func Classify(text string) ([]ClassificationResult, error) {
	for _, override := range regexOverrides {
		if override.Regex.MatchString(text) {
			return []ClassificationResult{{Name: override.Class, Confidence: 1.0}}, nil
		}
	}
	client := http.Client{}
	url, err := url.Parse(fmt.Sprintf("https://gateway.watsonplatform.net/natural-language-classifier/api/v1/classifiers/%s/classify", classifierId))
	if err != nil {
		panic(err)
	}
	values := url.Query()
	values.Add("text", text)
	url.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", url.String(), nil)
	req.SetBasicAuth(userName, password)
	if err != nil {
		fmt.Println("ERR:", err)
		return nil, err
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var resp struct {
		Classes []ClassificationResult `json:"classes"`
	}
	err = json.NewDecoder(r.Body).Decode(&resp)
	return resp.Classes, err
}
