package reconciliation

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPendingDerivationBuildsStableDeduplicationKey(
	t *testing.T,
) {
	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	task := PendingDerivation{
		IngestionRunID: "  550e8400-e29b-41d4-a716-446655440000  ",
		ICAO24:         " ABC123 ",
		DerivationType: DerivationTypeTrajectory,
		ObservedFrom:   observedAt,
		ObservedTo:     observedAt.Add(time.Minute),
	}

	key := task.DeduplicationKey()

	if key != "trajectory|abc123|550e8400-e29b-41d4-a716-446655440000|2026-07-11T10:00:00Z|2026-07-11T10:01:00Z" {
		t.Fatalf(
			"unexpected deduplication key: %s",
			key,
		)
	}
}

func TestPendingDerivationValidationRejectsBlankICAO24(
	t *testing.T,
) {
	observedAt := time.Now().UTC()

	task := PendingDerivation{
		DerivationType: DerivationTypeTrajectory,
		ObservedFrom:   observedAt,
		ObservedTo:     observedAt,
	}

	err := task.Validate()
	if !errors.Is(err, ErrICAO24Required) {
		t.Fatalf(
			"expected ErrICAO24Required, got %v",
			err,
		)
	}
}

func TestPendingDerivationValidationRejectsUnknownDerivationType(
	t *testing.T,
) {
	observedAt := time.Now().UTC()

	task := PendingDerivation{
		ICAO24:         "abc123",
		DerivationType: DerivationType("unknown"),
		ObservedFrom:   observedAt,
		ObservedTo:     observedAt,
	}

	err := task.Validate()
	if !errors.Is(err, ErrDerivationTypeInvalid) {
		t.Fatalf(
			"expected ErrDerivationTypeInvalid, got %v",
			err,
		)
	}
}

func TestPendingDerivationValidationRequiresObservedRange(
	t *testing.T,
) {
	task := PendingDerivation{
		ICAO24:         "abc123",
		DerivationType: DerivationTypeTrajectory,
	}

	err := task.Validate()
	if !errors.Is(err, ErrObservedRangeRequired) {
		t.Fatalf(
			"expected ErrObservedRangeRequired, got %v",
			err,
		)
	}
}

func TestPendingDerivationValidationRejectsReversedObservedRange(
	t *testing.T,
) {
	observedAt := time.Now().UTC()

	task := PendingDerivation{
		ICAO24:         "abc123",
		DerivationType: DerivationTypeTrajectory,
		ObservedFrom:   observedAt,
		ObservedTo:     observedAt.Add(-time.Minute),
	}

	err := task.Validate()
	if !errors.Is(err, ErrObservedRangeInvalid) {
		t.Fatalf(
			"expected ErrObservedRangeInvalid, got %v",
			err,
		)
	}
}

func TestPendingDerivationTruncatesLongLastError(
	t *testing.T,
) {
	task := PendingDerivation{
		ICAO24:         "abc123",
		DerivationType: DerivationTypeTrajectory,
		LastError: strings.Repeat(
			"x",
			maximumLastErrorLength+10,
		),
	}

	normalized := task.Normalize()

	if len(normalized.LastError) != maximumLastErrorLength {
		t.Fatalf(
			"expected truncated last error length %d, got %d",
			maximumLastErrorLength,
			len(normalized.LastError),
		)
	}
}
