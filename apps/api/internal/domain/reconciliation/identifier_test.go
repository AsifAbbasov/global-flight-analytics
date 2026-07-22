package reconciliation

import (
	"errors"
	"testing"
	"time"
)

func TestPendingDerivationRejectsAmbiguousDeduplicationIdentifier(t *testing.T) {
	now := time.Now().UTC()
	value := PendingDerivation{IngestionRunID: "run|other", ICAO24: "abc123", DerivationType: DerivationTypeTrajectory, ObservedFrom: now, ObservedTo: now}
	if err := value.Validate(); !errors.Is(err, ErrIngestionRunIDInvalid) {
		t.Fatalf("Validate() error = %v", err)
	}
}
