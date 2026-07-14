package validator

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const Version = "flight-feature-validator-v1"

type IssueSeverity string

const (
	IssueSeverityWarning IssueSeverity = "warning"
	IssueSeverityError   IssueSeverity = "error"
)

type Issue struct {
	Code     string
	Message  string
	Path     string
	Group    flightfeatures.FeatureGroup
	Severity IssueSeverity
}

type Report struct {
	ValidatorVersion string
	Status           flightfeatures.ValidationStatus
	ErrorCount       int
	WarningCount     int
	Issues           []Issue
	ValidatedAt      time.Time
}

func (report Report) Clone() Report {
	cloned := report
	cloned.Issues = append([]Issue(nil), report.Issues...)

	return cloned
}

type Policy struct {
	MinimumValidCompletenessScore float64
	MinimumValidInputQualityScore float64
	NumericTolerance              float64
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
