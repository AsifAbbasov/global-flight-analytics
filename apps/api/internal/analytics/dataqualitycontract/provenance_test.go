package dataqualitycontract

import (
	"errors"
	"testing"
	"time"
)

func TestProvenanceValidation(t *testing.T) {
	sourceTime := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)
	value := validProvenance(sourceTime)
	if err := value.Validate(); err != nil {
		t.Fatalf("expected valid provenance, got %v", err)
	}

	value.ReceivedAt = sourceTime.Add(-time.Second)
	if err := value.Validate(); !errors.Is(err, ErrReceivedBeforeSourceRecord) {
		t.Fatalf("expected received-before-source error, got %v", err)
	}
}
