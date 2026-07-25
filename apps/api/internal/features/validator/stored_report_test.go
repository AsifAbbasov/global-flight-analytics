package validator

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestNormalizeStoredReportMarksMissingAuditUnavailable(
	t *testing.T,
) {
	report := NormalizeStoredReport(
		Report{},
		flightfeatures.ValidationStatusValid,
	)

	if report.AuditState !=
		AuditStateLegacyUnavailable ||
		report.Status !=
			flightfeatures.ValidationStatusValid {
		t.Fatalf("normalized report = %#v", report)
	}
}

func TestValidateStoredReportAcceptsHonestLegacyAudit(
	t *testing.T,
) {
	err := ValidateStoredReport(
		Report{
			AuditState: AuditStateLegacyUnavailable,
			Status: flightfeatures.
				ValidationStatusLimited,
		},
		flightfeatures.ValidationStatusLimited,
	)
	if err != nil {
		t.Fatalf("ValidateStoredReport() error = %v", err)
	}
}

func TestValidateStoredReportRejectsInventedLegacyEvidence(
	t *testing.T,
) {
	err := ValidateStoredReport(
		Report{
			AuditState:       AuditStateLegacyUnavailable,
			ValidatorVersion: Version,
			Status: flightfeatures.
				ValidationStatusValid,
			ValidatedAt: time.Now().UTC(),
		},
		flightfeatures.ValidationStatusValid,
	)
	if !errors.Is(err, ErrInvalidReport) {
		t.Fatalf(
			"ValidateStoredReport() error = %v, want %v",
			err,
			ErrInvalidReport,
		)
	}
}
