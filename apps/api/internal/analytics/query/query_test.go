package query

import (
	"testing"
	"time"
)

func TestRequestIsValid(t *testing.T) {
	now := time.Now().UTC()

	request := New(
		"traffic.active_aircraft",
		now,
		now.Add(time.Minute),
	)

	if !request.IsValid() {
		t.Fatal("request must be valid")
	}
}

func TestRequestInvalidRange(t *testing.T) {
	now := time.Now().UTC()

	request := New(
		"traffic.active_aircraft",
		now,
		now.Add(-time.Minute),
	)

	if request.IsValid() {
		t.Fatal("request must be invalid")
	}
}
