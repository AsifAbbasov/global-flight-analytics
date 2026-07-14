package analyticalresult

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

type Status string

const (
	StatusComplete Status = "complete"
	StatusLimited  Status = "limited"
	StatusDenied   Status = "denied"
	StatusFailed   Status = "failed"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusComplete,
		StatusLimited,
		StatusDenied,
		StatusFailed:
		return true
	default:
		return false
	}
}

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

func (level ConfidenceLevel) IsKnown() bool {
	switch level {
	case ConfidenceLevelNone,
		ConfidenceLevelLow,
		ConfidenceLevelMedium,
		ConfidenceLevelHigh:
		return true
	default:
		return false
	}
}

type Notice struct {
	Code    string
	Message string
}

type Confidence struct {
	Level   ConfidenceLevel
	Score   float64
	Reasons []Notice
}

func NoneConfidence() Confidence {
	return Confidence{
		Level: ConfidenceLevelNone,
		Score: 0,
	}
}

type SourceRole string

const (
	SourceRoleObservation SourceRole = "observation"
	SourceRoleReference   SourceRole = "reference"
	SourceRoleDerived     SourceRole = "derived"
	SourceRoleFallback    SourceRole = "fallback"
)

func (role SourceRole) IsKnown() bool {
	switch role {
	case SourceRoleObservation,
		SourceRoleReference,
		SourceRoleDerived,
		SourceRoleFallback:
		return true
	default:
		return false
	}
}

type Source struct {
	Name         string
	Role         SourceRole
	ObservedFrom time.Time
	ObservedTo   time.Time
	RetrievedAt  time.Time
	Limitations  []Notice
}

type Eligibility struct {
	Capability  trajectoryeligibility.Capability
	Allowed     bool
	Reasons     []trajectoryeligibility.ReasonCode
	EvaluatedAt time.Time
}

type Failure struct {
	Code      string
	Message   string
	Retriable bool
}

type Result[T any] struct {
	Status       Status
	Value        T
	HasValue     bool
	Confidence   Confidence
	DataQuality  *dataqualitycontract.Report
	Eligibility  *Eligibility
	Sources      []Source
	Warnings     []Notice
	Limitations  []Notice
	CalculatedAt time.Time
	Failure      *Failure
}

func (result Result[T]) IsUsable() bool {
	return result.HasValue &&
		(result.Status == StatusComplete ||
			result.Status == StatusLimited)
}

func (result Result[T]) ValueOrZero() T {
	if result.HasValue {
		return result.Value
	}

	var zero T
	return zero
}
