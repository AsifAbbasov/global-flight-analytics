package validator

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestValidateReportAcceptsConsistentReports(t *testing.T) {
	now := time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		report Report
		status flightfeatures.ValidationStatus
	}{
		{
			name: "valid",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusValid,
				ValidatedAt:      now,
			},
			status: flightfeatures.ValidationStatusValid,
		},
		{
			name: "limited",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusLimited,
				WarningCount:     1,
				Issues: []Issue{
					{
						Code:     "validator.warning",
						Severity: IssueSeverityWarning,
					},
				},
				ValidatedAt: now,
			},
			status: flightfeatures.ValidationStatusLimited,
		},
		{
			name: "invalid",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusInvalid,
				ErrorCount:       1,
				Issues: []Issue{
					{
						Code:     "validator.error",
						Severity: IssueSeverityError,
					},
				},
				ValidatedAt: now,
			},
			status: flightfeatures.ValidationStatusInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateReport(test.report, test.status); err != nil {
				t.Fatalf("ValidateReport() error = %v", err)
			}
		})
	}
}

func TestValidateReportAllowsSameCodeOnDifferentPaths(t *testing.T) {
	now := time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)
	report := Report{
		ValidatorVersion: Version,
		Status:           flightfeatures.ValidationStatusLimited,
		WarningCount:     2,
		Issues: []Issue{
			{
				Code:     "validator.range",
				Path:     "temporal.start",
				Group:    flightfeatures.FeatureGroupTemporal,
				Severity: IssueSeverityWarning,
			},
			{
				Code:     "validator.range",
				Path:     "temporal.end",
				Group:    flightfeatures.FeatureGroupTemporal,
				Severity: IssueSeverityWarning,
			},
		},
		ValidatedAt: now,
	}

	if err := ValidateReport(
		report,
		flightfeatures.ValidationStatusLimited,
	); err != nil {
		t.Fatalf("ValidateReport() error = %v", err)
	}
}

func TestValidateReportRejectsInconsistentEvidence(t *testing.T) {
	now := time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		report Report
		status flightfeatures.ValidationStatus
	}{
		{
			name: "missing validator version",
			report: Report{
				Status:      flightfeatures.ValidationStatusValid,
				ValidatedAt: now,
			},
			status: flightfeatures.ValidationStatusValid,
		},
		{
			name: "missing validation time",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusValid,
			},
			status: flightfeatures.ValidationStatusValid,
		},
		{
			name: "count mismatch",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusLimited,
				WarningCount:     2,
				Issues: []Issue{
					{
						Code:     "validator.warning",
						Severity: IssueSeverityWarning,
					},
				},
				ValidatedAt: now,
			},
			status: flightfeatures.ValidationStatusLimited,
		},
		{
			name: "duplicate issue identity",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusLimited,
				WarningCount:     2,
				Issues: []Issue{
					{
						Code:     "validator.warning",
						Severity: IssueSeverityWarning,
					},
					{
						Code:     "validator.warning",
						Severity: IssueSeverityWarning,
					},
				},
				ValidatedAt: now,
			},
			status: flightfeatures.ValidationStatusLimited,
		},
		{
			name: "valid status with warning",
			report: Report{
				ValidatorVersion: Version,
				Status:           flightfeatures.ValidationStatusValid,
				WarningCount:     1,
				Issues: []Issue{
					{
						Code:     "validator.warning",
						Severity: IssueSeverityWarning,
					},
				},
				ValidatedAt: now,
			},
			status: flightfeatures.ValidationStatusValid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateReport(test.report, test.status)
			if !errors.Is(err, ErrInvalidReport) {
				t.Fatalf(
					"ValidateReport() error = %v, want %v",
					err,
					ErrInvalidReport,
				)
			}
		})
	}
}
