package interactionradius

import "time"

const Version = "interaction-radius-decision-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "interaction-radius-decision-v1"

type DecisionStatus string

const (
	DecisionStatusBlocked DecisionStatus = "blocked"
	DecisionStatusLimited DecisionStatus = "limited"
	DecisionStatusAllowed DecisionStatus = "allowed"
)

func (status DecisionStatus) IsKnown() bool {
	switch status {
	case DecisionStatusBlocked,
		DecisionStatusLimited,
		DecisionStatusAllowed:
		return true
	default:
		return false
	}
}

type AltitudeReference string

const (
	AltitudeReferenceUnknown    AltitudeReference = "unknown"
	AltitudeReferenceBarometric AltitudeReference = "barometric"
	AltitudeReferenceGeometric  AltitudeReference = "geometric"
)

func (reference AltitudeReference) IsKnown() bool {
	switch reference {
	case AltitudeReferenceUnknown,
		AltitudeReferenceBarometric,
		AltitudeReferenceGeometric:
		return true
	default:
		return false
	}
}

type MotionClass string

const (
	MotionClassUnknown        MotionClass = "unknown"
	MotionClassLowSpeed       MotionClass = "low_speed"
	MotionClassTransit        MotionClass = "transit"
	MotionClassHighSpeed      MotionClass = "high_speed"
	MotionClassVerticalChange MotionClass = "vertical_change"
)

func (class MotionClass) IsKnown() bool {
	switch class {
	case MotionClassUnknown,
		MotionClassLowSpeed,
		MotionClassTransit,
		MotionClassHighSpeed,
		MotionClassVerticalChange:
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

type ScopeGuard string

const ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_separation_use"

type Request struct {
	RegionCode string
	NodeID     string
	ICAO24     string
	Callsign   string

	VelocityMetersPerSecond     float64
	VerticalRateMetersPerSecond float64
	AltitudeMeters              *float64
	AltitudeReference           AltitudeReference
	OnGround                    bool

	ObservedAt   time.Time
	AsOfTime     time.Time
	GeneratedAt  time.Time
	SourceName   string
	QualityScore float64
}

type Component struct {
	Name   string
	Score  float64
	Weight float64
}

type ConfidenceReason struct {
	Code         string
	Message      string
	Contribution float64
}

type Confidence struct {
	Score   float64
	Level   ConfidenceLevel
	Reasons []ConfidenceReason
}

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type Explanation struct {
	Code    string
	Message string
}

type Provenance struct {
	InputFingerprint string
	SourceNames      []string
	ObservedAt       time.Time
}

type Decision struct {
	SchemaVersion SchemaVersion
	Status        DecisionStatus
	RegionCode    string
	NodeID        string
	ICAO24        string
	Callsign      string

	MotionClass MotionClass

	HorizontalRadiusKilometers float64
	VerticalRadiusMeters       float64
	MaximumObservationAge      time.Duration
	MaximumPairTimeDifference  time.Duration
	LookaheadDuration          time.Duration
	VerticalFilteringPermitted bool

	Components   []Component
	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	AsOfTime     time.Time
	GeneratedAt  time.Time
}

func (decision Decision) Clone() Decision {
	cloned := decision
	cloned.Components = append([]Component(nil), decision.Components...)
	cloned.Confidence.Reasons = append(
		[]ConfidenceReason(nil),
		decision.Confidence.Reasons...,
	)
	cloned.Limitations = append([]Limitation(nil), decision.Limitations...)
	cloned.Explanations = append([]Explanation(nil), decision.Explanations...)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		decision.Provenance.SourceNames...,
	)
	return cloned
}
