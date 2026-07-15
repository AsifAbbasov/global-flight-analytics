package projectioncontract

import "time"

const Version = "projection-intelligence-contract-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "projection-intelligence-v1"

type ResultStatus string

const (
	ResultStatusUnavailable ResultStatus = "unavailable"
	ResultStatusLimited     ResultStatus = "limited"
	ResultStatusComplete    ResultStatus = "complete"
)

func (status ResultStatus) IsKnown() bool {
	switch status {
	case ResultStatusUnavailable,
		ResultStatusLimited,
		ResultStatusComplete:
		return true
	default:
		return false
	}
}

type DecisionClass string

const (
	DecisionClassSourceBacked    DecisionClass = "source_backed"
	DecisionClassResearchAdapted DecisionClass = "research_adapted"
	DecisionClassPhysicsDerived  DecisionClass = "physics_derived"
	DecisionClassProjectDerived  DecisionClass = "project_derived"
	DecisionClassExperimental    DecisionClass = "experimental"
)

func (class DecisionClass) IsKnown() bool {
	switch class {
	case DecisionClassSourceBacked,
		DecisionClassResearchAdapted,
		DecisionClassPhysicsDerived,
		DecisionClassProjectDerived,
		DecisionClassExperimental:
		return true
	default:
		return false
	}
}

type InputClassification string

const (
	InputClassificationObserved      InputClassification = "observed"
	InputClassificationOpenlySourced InputClassification = "openly_sourced"
	InputClassificationDerived       InputClassification = "derived"
	InputClassificationEstimated     InputClassification = "estimated"
	InputClassificationUnknown       InputClassification = "unknown"
)

func (classification InputClassification) IsKnown() bool {
	switch classification {
	case InputClassificationObserved,
		InputClassificationOpenlySourced,
		InputClassificationDerived,
		InputClassificationEstimated,
		InputClassificationUnknown:
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

const (
	ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_use"
)

type Method struct {
	Name          string
	Version       string
	DecisionClass DecisionClass
}

type Horizon struct {
	AsOfTime time.Time
	EndTime  time.Time
	Step     time.Duration
}

func (horizon Horizon) Duration() time.Duration {
	if horizon.AsOfTime.IsZero() ||
		horizon.EndTime.IsZero() {
		return 0
	}

	return horizon.EndTime.Sub(
		horizon.AsOfTime,
	)
}

type Position struct {
	Latitude  float64
	Longitude float64
	AltitudeM *float64
}

type Uncertainty struct {
	HorizontalRadiusM float64
	VerticalRadiusM   *float64
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

type ProjectionPoint struct {
	Sequence     int
	ForecastTime time.Time
	Position     Position
	Uncertainty  Uncertainty
	Confidence   Confidence
}

type ArrivalEstimate struct {
	AirportICAOCode string

	EarliestTime  time.Time
	EstimatedTime time.Time
	LatestTime    time.Time

	Confidence  Confidence
	Limitations []Limitation
}

type InputReference struct {
	Name           string
	Classification InputClassification
	SourceName     string
	ObservedAt     time.Time
	RetrievedAt    time.Time
	Limitation     string
}

type Provenance struct {
	InputFingerprint      string
	Inputs                []InputReference
	LatestInputObservedAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus

	TrajectoryID string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Method  Method
	Horizon Horizon
	Points  []ProjectionPoint
	Arrival *ArrivalEstimate

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	GeneratedAt  time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Points = cloneProjectionPoints(
		result.Points,
	)
	cloned.Arrival = cloneArrivalEstimate(
		result.Arrival,
	)
	cloned.Confidence = cloneConfidence(
		result.Confidence,
	)
	cloned.Limitations = append(
		[]Limitation(nil),
		result.Limitations...,
	)
	cloned.Explanations = append(
		[]Explanation(nil),
		result.Explanations...,
	)
	cloned.Provenance.Inputs = append(
		[]InputReference(nil),
		result.Provenance.Inputs...,
	)

	return cloned
}

func cloneProjectionPoints(
	items []ProjectionPoint,
) []ProjectionPoint {
	cloned := make(
		[]ProjectionPoint,
		0,
		len(items),
	)
	for _, item := range items {
		copied := item
		copied.Position.AltitudeM = cloneFloat64(
			item.Position.AltitudeM,
		)
		copied.Uncertainty.VerticalRadiusM =
			cloneFloat64(
				item.Uncertainty.
					VerticalRadiusM,
			)
		copied.Confidence = cloneConfidence(
			item.Confidence,
		)
		cloned = append(
			cloned,
			copied,
		)
	}

	return cloned
}

func cloneArrivalEstimate(
	estimate *ArrivalEstimate,
) *ArrivalEstimate {
	if estimate == nil {
		return nil
	}

	cloned := *estimate
	cloned.Confidence = cloneConfidence(
		estimate.Confidence,
	)
	cloned.Limitations = append(
		[]Limitation(nil),
		estimate.Limitations...,
	)

	return &cloned
}

func cloneConfidence(
	confidence Confidence,
) Confidence {
	cloned := confidence
	cloned.Reasons = append(
		[]ConfidenceReason(nil),
		confidence.Reasons...,
	)

	return cloned
}

func cloneFloat64(
	value *float64,
) *float64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
