package ingestionrun

import (
	"errors"
	"testing"
	"time"
)

func TestRunValidateRejectsNegativeCountersAndTerminalRunWithoutFinish(t *testing.T) {
	now := time.Now().UTC()
	run := Run{SourceName: "opensky", StartedAt: now, Status: StatusRunning, RecordsReceived: -1}
	if err := run.Validate(); !errors.Is(err, ErrIngestionCountersInvalid) {
		t.Fatalf("counter error = %v", err)
	}
	run = Run{SourceName: "opensky", StartedAt: now, Status: StatusSuccess}
	if err := run.Validate(); !errors.Is(err, ErrIngestionFinishedAtInvalid) {
		t.Fatalf("finish error = %v", err)
	}
}
