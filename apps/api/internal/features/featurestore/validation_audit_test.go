package featurestore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

func TestMemoryStorePersistsExplicitLegacyAuditForCompatibility(
	t *testing.T,
) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"trajectory-validation-audit",
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

	record, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	report := record.Features.ValidationReport
	if report.AuditState !=
		flightfeatures.
			ValidationAuditStateLegacyUnavailable ||
		report.Status != record.Features.Quality.Status {
		t.Fatalf("stored legacy audit = %#v", report)
	}

	loaded, err := store.Get(
		context.Background(),
		record.Key,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.Features.ValidationReport.AuditState !=
		flightfeatures.
			ValidationAuditStateLegacyUnavailable {
		t.Fatalf(
			"loaded legacy audit = %#v",
			loaded.Features.ValidationReport,
		)
	}
}

func TestMemoryStoreRejectsCorruptCompleteValidationAudit(
	t *testing.T,
) {
	store := NewMemory(MemoryConfig{})
	features := validStoredFeatures(
		"trajectory-validation-audit",
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
