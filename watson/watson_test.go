package watson

import (
	"testing"
)

func TestWatson(t *testing.T) {
	classes, err := Classify("What's my heart rate?")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(classes) == 0 {
		t.Fatal("got no classes!")
	}
	if classes[0].Name != "Current heart rate" {
		t.Errorf("Got unexpected class: %s", classes[0].Name)
	}
}
