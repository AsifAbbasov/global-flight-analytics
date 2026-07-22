package flight

import (
	"errors"
	"testing"
	"time"
)

func TestFlightValidateRejectsReversedObservationRange(t *testing.T) {
	now := time.Now().UTC()
	err := (Flight{FirstSeenAt: now, LastSeenAt: now.Add(-time.Second)}).Validate()
	if !errors.Is(err, ErrFlightObservedRangeInvalid) {
		t.Fatalf("Validate() error = %v", err)
	}
}
