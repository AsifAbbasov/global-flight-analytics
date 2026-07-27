package validator

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const Version = "flight-feature-validator-v6"

type AuditState = flightfeatures.ValidationAuditState

const (
	AuditStateComplete          = flightfeatures.ValidationAuditStateComplete
	AuditStateLegacyUnavailable = flightfeatures.ValidationAuditStateLegacyUnavailable
)

type IssueSeverity = flightfeatures.ValidationIssueSeverity

const (
	IssueSeverityWarning = flightfeatures.ValidationIssueSeverityWarning
	IssueSeverityError   = flightfeatures.ValidationIssueSeverityError
)

type Issue = flightfeatures.ValidationIssue
type Report = flightfeatures.ValidationReport

type Policy struct {
	MinimumValidCompletenessScore float64
	MinimumValidInputQualityScore float64
	// NumericTolerance is a dimensionless relative tolerance. Unit-bearing
	// comparisons must use relative comparison helpers rather than adding this
	// value directly to metres, kilometres, degrees, or seconds.
	NumericTolerance float64
}

func DefaultPolicy() Policy {
	return Policy{
		MinimumValidCompletenessScore: 1,
		MinimumValidInputQualityScore: 0.8,
		NumericTolerance:              1e-6,
	}
}

type Config struct {
	Policy *Policy
	Now    func() time.Time
}
