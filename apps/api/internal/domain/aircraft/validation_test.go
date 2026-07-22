package aircraft

import (
	"errors"
	"testing"
)

func TestAircraftValidateRejectsInvalidICAO24(t *testing.T) {
	if err := (Aircraft{ICAO24: "xyz"}).Validate(); !errors.Is(err, ErrAircraftICAO24Invalid) {
		t.Fatalf("Validate() error = %v", err)
	}
}
