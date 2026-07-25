package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var ErrInvalidReport = errors.New(
	"feature validation report is internally inconsistent",
)

type ReportIntegrityError struct {
	Field  string
	Detail string
}

func (err *ReportIntegrityError) Error() string {
	if err == nil {
		return ErrInvalidReport.Error()
	}

	return fmt.Sprintf(
		"%s: field=%q detail=%s",
		ErrInvalidReport,
		err.Field,
		err.Detail,
	)
}

func (err *ReportIntegrityError) Unwrap() error {
	return ErrInvalidReport
}

func NormalizeReport(report Report) Report {
	normalized := report.Clone()
	if normalized.AuditState == "" {
		normalized.AuditState = AuditStateComplete
	}
	normalized.ValidatorVersion =
		strings.TrimSpace(normalized.ValidatorVersion)
	if !normalized.ValidatedAt.IsZero() {
		normalized.ValidatedAt =
			normalized.ValidatedAt.UTC()
	}

	return normalized
}

func NormalizeStoredReport(
	report Report,
	expectedStatus flightfeatures.ValidationStatus,
) Report {
	if reportIsZero(report) {
		return Report{
			AuditState: AuditStateLegacyUnavailable,
			Status:     expectedStatus,
		}
	}

	normalized := report.Clone()
	if normalized.AuditState == "" {
		normalized.AuditState = AuditStateComplete
	}
	normalized.ValidatorVersion =
		strings.TrimSpace(normalized.ValidatorVersion)
	if !normalized.ValidatedAt.IsZero() {
		normalized.ValidatedAt =
			normalized.ValidatedAt.UTC()
	}

	return normalized
}

func ValidateStoredReport(
	report Report,
	expectedStatus flightfeatures.ValidationStatus,
) error {
	report = NormalizeStoredReport(report, expectedStatus)

	switch report.AuditState {
	case AuditStateComplete:
		return ValidateReport(report, expectedStatus)
	case AuditStateLegacyUnavailable:
		if report.Status != expectedStatus {
			return invalidReport(
				"status",
				"legacy audit status must match stored feature status",
			)
		}
		if strings.TrimSpace(report.ValidatorVersion) != "" {
			return invalidReport(
				"validator_version",
				"legacy unavailable audit must not invent a validator version",
			)
		}
		if !report.ValidatedAt.IsZero() {
			return invalidReport(
				"validated_at",
				"legacy unavailable audit must not invent validation time",
			)
		}
		if report.ErrorCount != 0 ||
			report.WarningCount != 0 ||
			len(report.Issues) != 0 {
			return invalidReport(
				"issues",
				"legacy unavailable audit must not invent issues or counts",
			)
		}

		return nil
	default:
		return invalidReport(
			"audit_state",
			"must be complete or legacy_unavailable",
		)
	}
}

func reportIsZero(report Report) bool {
	return report.AuditState == "" &&
		strings.TrimSpace(report.ValidatorVersion) == "" &&
		report.Status == "" &&
		report.ErrorCount == 0 &&
		report.WarningCount == 0 &&
		len(report.Issues) == 0 &&
		report.ValidatedAt.IsZero()
}

func ValidateReport(
	report Report,
	expectedStatus flightfeatures.ValidationStatus,
) error {
	report = NormalizeReport(report)
	if report.AuditState != AuditStateComplete {
		return invalidReport(
			"audit_state",
			"current validation report must be complete",
		)
	}
	if strings.TrimSpace(report.ValidatorVersion) != Version {
		return invalidReport(
			"validator_version",
			"must equal the current validator version",
		)
	}
	if report.ValidatedAt.IsZero() {
		return invalidReport(
			"validated_at",
			"must be present",
		)
	}
	if report.Status != expectedStatus {
		return invalidReport(
			"status",
			"must match validated feature status",
		)
	}
	if report.ErrorCount < 0 {
		return invalidReport(
			"error_count",
			"must not be negative",
		)
	}
	if report.WarningCount < 0 {
		return invalidReport(
			"warning_count",
			"must not be negative",
		)
	}

	errorCount := 0
	warningCount := 0
	seenIssues := make(map[string]struct{}, len(report.Issues))

	for _, issue := range report.Issues {
		code := strings.TrimSpace(issue.Code)
		if code == "" {
			return invalidReport(
				"issues.code",
				"must not be empty",
			)
		}
		identity := strings.Join(
			[]string{
				code,
				strings.TrimSpace(issue.Path),
				string(issue.Group),
				string(issue.Severity),
			},
			"\x00",
		)
		if _, exists := seenIssues[identity]; exists {
			return invalidReport(
				"issues",
				"entries must be unique",
			)
		}
		seenIssues[identity] = struct{}{}

		switch issue.Severity {
		case IssueSeverityError:
			errorCount++
		case IssueSeverityWarning:
			warningCount++
		default:
			return invalidReport(
				"issues.severity",
				"must be warning or error",
			)
		}
	}

	if report.ErrorCount != errorCount {
		return invalidReport(
			"error_count",
			"does not match error issues",
		)
	}
	if report.WarningCount != warningCount {
		return invalidReport(
			"warning_count",
			"does not match warning issues",
		)
	}

	derivedStatus := flightfeatures.ValidationStatusValid
	if errorCount > 0 {
		derivedStatus = flightfeatures.ValidationStatusInvalid
	} else if warningCount > 0 {
		derivedStatus = flightfeatures.ValidationStatusLimited
	}
	if report.Status != derivedStatus {
		return invalidReport(
			"status",
			"does not match issue severities",
		)
	}

	return nil
}

func invalidReport(
	field string,
	detail string,
) error {
	return &ReportIntegrityError{
		Field:  field,
		Detail: detail,
	}
}
