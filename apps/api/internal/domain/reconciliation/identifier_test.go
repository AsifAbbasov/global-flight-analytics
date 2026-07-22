package reconciliation

import (
	"errors"
	"testing"
	"time"
)

func TestPendingDerivationDeduplicationKeyValidatesIdentifier(t *testing.T) {
	now := time.Now().UTC()
	value := PendingDerivation{
		IngestionRunID: "run|other",
		ICAO24:         "abc123",
		DerivationType: DerivationTypeTrajectory,
		ObservedFrom:   now,
		ObservedTo:     now,
	}
	key, err := value.DeduplicationKey()
	if key != "" || !errors.Is(err, ErrIngestionRunIDInvalid) {
		t.Fatalf("DeduplicationKey() = %q, %v", key, err)
	}
}
