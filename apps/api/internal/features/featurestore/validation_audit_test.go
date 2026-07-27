package featurestore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestMemoryStoreRejectsLegacyAuditOnNewWrite(
	t *testing.T,
) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"66196dc0-56f2-5a06-ab2c-6e0ac07316e3",
		time.Date(
			2026,
			time.July,
			25,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)
	features.ValidationReport = flightfeatures.ValidationReport{}

	_, err := store.Put(
		context.Background(),
		features,
	)
	if !errors.Is(err, ErrValidationProofRequired) {
		t.Fatalf(
			"Put() error = %v, want %v",
			err,
			ErrValidationProofRequired,
		)
	}
}

func TestMemoryStoreRejectsCorruptCompleteValidationAudit(
	t *testing.T,
) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"66196dc0-56f2-5a06-ab2c-6e0ac07316e3",
		time.Date(
			2026,
			time.July,
			25,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)
	features.ValidationReport =
		flightfeatures.ValidationReport{
			AuditState: flightfeatures.
				ValidationAuditStateComplete,
			ValidatorVersion: "wrong-validator",
			Status: flightfeatures.
				ValidationStatusValid,
			ValidatedAt: time.Now().UTC(),
		}

	_, err := store.Put(
		context.Background(),
		features,
	)
	if !errors.Is(err, validator.ErrInvalidReport) {
		t.Fatalf(
			"Put() error = %v, want %v",
			err,
			validator.ErrInvalidReport,
		)
	}
}
