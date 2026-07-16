package interactiongraph

import "time"

const Version = "airborne-interaction-graph-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "airborne-interaction-graph-v1"

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

type InteractionKind string

const (
	InteractionKindUnknown    InteractionKind = "unknown"
	InteractionKindNearby     InteractionKind = "nearby"
	InteractionKindConverging InteractionKind = "converging"
	InteractionKindParallel   InteractionKind = "parallel"
	InteractionKindDiverging  InteractionKind = "diverging"
)

func (kind InteractionKind) IsKnown() bool {
	switch kind {
	case InteractionKindUnknown,
		InteractionKindNearby,
		InteractionKindConverging,
		InteractionKindParallel,
		InteractionKindDiverging:
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

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type Explanation struct {
	Code    string
	Message string
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

type NodeInput struct {
	ID           string
	TrajectoryID string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Latitude  float64
	Longitude float64

	AltitudeMeters    *float64
	AltitudeReference AltitudeReference

	VelocityMetersPerSecond     float64
	HeadingDegrees              float64
	VerticalRateMetersPerSecond float64
	OnGround                    bool

	ObservedAt   time.Time
	SourceName   string
	QualityScore float64
}

type EdgeInput struct {
	SourceNodeID string
	TargetNodeID string
	Kind         InteractionKind

	HorizontalDistanceKilometers float64
	VerticalSeparationMeters     *float64

	EvaluatedAt     time.Time
	SourceName      string
	ConfidenceScore float64
	Limitations     []Limitation
}

type Request struct {
	RegionCode  string
	AsOfTime    time.Time
	GeneratedAt time.Time
	Nodes       []NodeInput
	Edges       []EdgeInput
}

type Node struct {
	ID           string
	TrajectoryID string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Latitude  float64
	Longitude float64

	AltitudeMeters    *float64
	AltitudeReference AltitudeReference

	VelocityMetersPerSecond     float64
	HeadingDegrees              float64
	VerticalRateMetersPerSecond float64

	ObservedAt   time.Time
	SourceName   string
	QualityScore float64
	Degree       int
}

type Edge struct {
	ID           string
	SourceNodeID string
	TargetNodeID string
	Kind         InteractionKind

	HorizontalDistanceKilometers float64
	VerticalSeparationMeters     *float64
	ObservationTimeDifference    time.Duration

	EvaluatedAt     time.Time
	SourceName      string
	ConfidenceScore float64
	Limitations     []Limitation
}

type GraphMetrics struct {
	NodeCount               int
	EdgeCount               int
	IsolatedNodeCount       int
	ConnectedComponentCount int
	Density                 float64
}

type Provenance struct {
	InputFingerprint string
	SourceNames      []string
	LatestObservedAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus
	RegionCode    string
	AsOfTime      time.Time

	Nodes   []Node
	Edges   []Edge
	Metrics GraphMetrics

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	GeneratedAt  time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Nodes = make([]Node, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		cloned.Nodes = append(cloned.Nodes, cloneNode(node))
	}
	cloned.Edges = make([]Edge, 0, len(result.Edges))
	for _, edge := range result.Edges {
		cloned.Edges = append(cloned.Edges, cloneEdge(edge))
	}
	cloned.Confidence.Reasons = append(
		[]ConfidenceReason(nil),
		result.Confidence.Reasons...,
	)
	cloned.Limitations = append(
		[]Limitation(nil),
		result.Limitations...,
	)
	cloned.Explanations = append(
		[]Explanation(nil),
		result.Explanations...,
	)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)
	return cloned
}

func cloneNode(node Node) Node {
	cloned := node
	cloned.AltitudeMeters = cloneFloat64(node.AltitudeMeters)
	return cloned
}

func cloneEdge(edge Edge) Edge {
	cloned := edge
	cloned.VerticalSeparationMeters = cloneFloat64(
		edge.VerticalSeparationMeters,
	)
	cloned.Limitations = append(
		[]Limitation(nil),
		edge.Limitations...,
	)
	return cloned
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
