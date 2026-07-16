package localtrafficscene

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
)

const Version = "local-traffic-scene-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "local-traffic-scene-v1"

type ResultStatus string

const (
	ResultStatusUnavailable ResultStatus = "unavailable"
	ResultStatusLimited     ResultStatus = "limited"
	ResultStatusComplete    ResultStatus = "complete"
)

func (status ResultStatus) IsKnown() bool {
	switch status {
	case ResultStatusUnavailable, ResultStatusLimited, ResultStatusComplete:
		return true
	default:
		return false
	}
}

type ExclusionReason string

const (
	ExclusionReasonOnGround            ExclusionReason = "on_ground"
	ExclusionReasonOutsideRegion       ExclusionReason = "outside_region"
	ExclusionReasonFutureEvidence      ExclusionReason = "future_evidence"
	ExclusionReasonSupersededDuplicate ExclusionReason = "superseded_duplicate"
	ExclusionReasonRadiusPolicyBlocked ExclusionReason = "radius_policy_blocked"
)

func (reason ExclusionReason) IsKnown() bool {
	switch reason {
	case ExclusionReasonOnGround,
		ExclusionReasonOutsideRegion,
		ExclusionReasonFutureEvidence,
		ExclusionReasonSupersededDuplicate,
		ExclusionReasonRadiusPolicyBlocked:
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
	case ConfidenceLevelNone, ConfidenceLevelLow, ConfidenceLevelMedium, ConfidenceLevelHigh:
		return true
	default:
		return false
	}
}

type ScopeGuard string

const ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_separation_use"

type Bounds struct {
	MinimumLatitude  float64
	MaximumLatitude  float64
	MinimumLongitude float64
	MaximumLongitude float64
}

type ObservationInput struct {
	ID           string
	TrajectoryID string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Latitude  float64
	Longitude float64

	AltitudeMeters    *float64
	AltitudeReference interactiongraph.AltitudeReference

	VelocityMetersPerSecond     float64
	HeadingDegrees              float64
	VerticalRateMetersPerSecond float64
	OnGround                    bool

	ObservedAt   time.Time
	SourceName   string
	QualityScore float64
}

type Request struct {
	RegionCode   string
	RegionBounds Bounds
	AsOfTime     time.Time
	GeneratedAt  time.Time
	Observations []ObservationInput
}

type Aircraft struct {
	NodeID       string
	TrajectoryID string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Latitude  float64
	Longitude float64

	AltitudeMeters    *float64
	AltitudeReference interactiongraph.AltitudeReference

	VelocityMetersPerSecond     float64
	HeadingDegrees              float64
	VerticalRateMetersPerSecond float64

	ObservedAt     time.Time
	ObservationAge time.Duration
	SourceName     string
	QualityScore   float64
	RadiusDecision interactionradius.Decision
}

type ExcludedObservation struct {
	NodeID     string
	ICAO24     string
	Callsign   string
	ObservedAt time.Time
	Reason     ExclusionReason
	Message    string
	SourceName string
}

type SceneMetrics struct {
	InputObservationCount         int
	CandidateObservationCount     int
	IncludedAircraftCount         int
	AllowedAircraftCount          int
	LimitedAircraftCount          int
	ExcludedObservationCount      int
	GroundExcludedCount           int
	OutsideRegionExcludedCount    int
	FutureEvidenceExcludedCount   int
	DuplicateExcludedCount        int
	RadiusPolicyBlockedCount      int
	MaterialEvidenceRejectedCount int
	SceneCoverage                 float64
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
	LatestObservedAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus
	RegionCode    string
	RegionBounds  Bounds
	AsOfTime      time.Time

	Aircraft             []Aircraft
	ExcludedObservations []ExcludedObservation
	Metrics              SceneMetrics

	Confidence   Confidence
	Limitations  []Limitation
	Explanations []Explanation
	ScopeGuard   ScopeGuard
	Provenance   Provenance
	GeneratedAt  time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Aircraft = make([]Aircraft, 0, len(result.Aircraft))
	for _, aircraft := range result.Aircraft {
		cloned.Aircraft = append(cloned.Aircraft, cloneAircraft(aircraft))
	}
	cloned.ExcludedObservations = append(
		[]ExcludedObservation(nil),
		result.ExcludedObservations...,
	)
	cloned.Confidence.Reasons = append(
		[]ConfidenceReason(nil),
		result.Confidence.Reasons...,
	)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)
	return cloned
}

func (result Result) GraphNodeInputs() []interactiongraph.NodeInput {
	nodes := make([]interactiongraph.NodeInput, 0, len(result.Aircraft))
	for _, aircraft := range result.Aircraft {
		nodes = append(nodes, interactiongraph.NodeInput{
			ID:                          aircraft.NodeID,
			TrajectoryID:                aircraft.TrajectoryID,
			FlightID:                    aircraft.FlightID,
			AircraftID:                  aircraft.AircraftID,
			ICAO24:                      aircraft.ICAO24,
			Callsign:                    aircraft.Callsign,
			Latitude:                    aircraft.Latitude,
			Longitude:                   aircraft.Longitude,
			AltitudeMeters:              cloneFloat64(aircraft.AltitudeMeters),
			AltitudeReference:           aircraft.AltitudeReference,
			VelocityMetersPerSecond:     aircraft.VelocityMetersPerSecond,
			HeadingDegrees:              aircraft.HeadingDegrees,
			VerticalRateMetersPerSecond: aircraft.VerticalRateMetersPerSecond,
			OnGround:                    false,
			ObservedAt:                  aircraft.ObservedAt,
			SourceName:                  aircraft.SourceName,
			QualityScore:                aircraft.QualityScore,
		})
	}
	return nodes
}

func cloneAircraft(aircraft Aircraft) Aircraft {
	cloned := aircraft
	cloned.AltitudeMeters = cloneFloat64(aircraft.AltitudeMeters)
	cloned.RadiusDecision = aircraft.RadiusDecision.Clone()
	return cloned
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
