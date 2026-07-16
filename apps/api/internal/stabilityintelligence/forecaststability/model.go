package forecaststability

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const Version = "forecast-versioning-and-decision-stability-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "forecast-versioning-and-decision-stability-v1"

type ResultStatus string

const (
	ResultStatusLimited  ResultStatus = "limited"
	ResultStatusComplete ResultStatus = "complete"
)

func (status ResultStatus) IsKnown() bool {
	switch status {
	case ResultStatusLimited, ResultStatusComplete:
		return true
	default:
		return false
	}
}

type ScopeGuard string

const ScopeGuardResearchOnly ScopeGuard = "research_only_not_for_operational_forecast_or_decision_use"

type RegistrationDecision string

const (
	RegistrationDecisionInitial RegistrationDecision = "initial_version_created"
	RegistrationDecisionReused  RegistrationDecision = "identical_version_reused"
	RegistrationDecisionCreated RegistrationDecision = "successor_version_created"
)

func (decision RegistrationDecision) IsKnown() bool {
	switch decision {
	case RegistrationDecisionInitial, RegistrationDecisionReused, RegistrationDecisionCreated:
		return true
	default:
		return false
	}
}

type VersionChangeKind string

const (
	VersionChangeProjectionSchema VersionChangeKind = "projection_schema_changed"
	VersionChangeMethod           VersionChangeKind = "method_changed"
	VersionChangePolicy           VersionChangeKind = "policy_changed"
	VersionChangeImplementation   VersionChangeKind = "implementation_changed"
	VersionChangeInput            VersionChangeKind = "input_changed"
	VersionChangeOutput           VersionChangeKind = "output_changed"
	VersionChangeHorizon          VersionChangeKind = "horizon_changed"
)

func (kind VersionChangeKind) IsKnown() bool {
	switch kind {
	case VersionChangeProjectionSchema,
		VersionChangeMethod,
		VersionChangePolicy,
		VersionChangeImplementation,
		VersionChangeInput,
		VersionChangeOutput,
		VersionChangeHorizon:
		return true
	default:
		return false
	}
}

type VersionChange struct {
	Kind     VersionChangeKind
	Previous string
	Current  string
}

type ForecastVersionRequest struct {
	Projection            projectioncontract.Result
	PolicyVersion         string
	ImplementationVersion string
	Previous              *ForecastVersionRecord
	RegisteredAt          time.Time
}

type ForecastVersionRecord struct {
	SchemaVersion SchemaVersion
	VersionID     string
	Ordinal       int
	TrajectoryID  string

	ProjectionSchemaVersion projectioncontract.SchemaVersion
	Method                  projectioncontract.Method
	PolicyVersion           string
	ImplementationVersion   string

	InputFingerprint    string
	OutputFingerprint   string
	DecisionFingerprint string
	ParentVersionID     string

	Projection projectioncontract.Result
	CreatedAt  time.Time
	ScopeGuard ScopeGuard
}

func (record ForecastVersionRecord) Clone() ForecastVersionRecord {
	cloned := record
	cloned.Projection = record.Projection.Clone()
	return cloned
}

type RegistrationResult struct {
	SchemaVersion SchemaVersion
	Status        ResultStatus
	Decision      RegistrationDecision
	Record        ForecastVersionRecord
	Changes       []VersionChange
	Limitations   []Limitation
	Explanations  []Explanation
	ScopeGuard    ScopeGuard
	GeneratedAt   time.Time
}

func (result RegistrationResult) Clone() RegistrationResult {
	cloned := result
	cloned.Record = result.Record.Clone()
	cloned.Changes = append([]VersionChange(nil), result.Changes...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	return cloned
}

type StabilityLevel string

const (
	StabilityLevelUnchanged      StabilityLevel = "unchanged"
	StabilityLevelStable         StabilityLevel = "stable"
	StabilityLevelChanged        StabilityLevel = "changed"
	StabilityLevelMaterialChange StabilityLevel = "material_change"
	StabilityLevelIndeterminate  StabilityLevel = "indeterminate"
)

func (level StabilityLevel) IsKnown() bool {
	switch level {
	case StabilityLevelUnchanged,
		StabilityLevelStable,
		StabilityLevelChanged,
		StabilityLevelMaterialChange,
		StabilityLevelIndeterminate:
		return true
	default:
		return false
	}
}

type StabilityRequest struct {
	Baseline    ForecastVersionRecord
	Candidate   ForecastVersionRecord
	EvaluatedAt time.Time
}

type StabilityComponent struct {
	Name        string
	Stability   float64
	Weight      float64
	Comparable  bool
	Explanation string
}

type StabilityReason struct {
	Code    string
	Message string
	Impact  float64
}

type StabilityMetrics struct {
	BaselinePointCount  int
	CandidatePointCount int
	AlignedPointCount   int
	AlignedPointShare   float64

	MeanHorizontalShiftKilometers           float64
	MaximumHorizontalShiftKilometers        float64
	MeanAbsolutePointConfidenceDelta        float64
	AggregateConfidenceDelta                float64
	MeanRelativeHorizontalUncertaintyChange float64

	ArrivalComparable   bool
	ArrivalShiftSeconds float64

	ProjectionStatusChanged bool
	MethodChanged           bool
	PolicyChanged           bool
	ImplementationChanged   bool
	InputChanged            bool
	OutputChanged           bool
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

type StabilityProvenance struct {
	InputFingerprint           string
	BaselineVersionID          string
	CandidateVersionID         string
	BaselineOutputFingerprint  string
	CandidateOutputFingerprint string
}

type StabilityResult struct {
	SchemaVersion      SchemaVersion
	Status             ResultStatus
	TrajectoryID       string
	BaselineVersionID  string
	CandidateVersionID string
	Level              StabilityLevel
	Score              float64
	Metrics            StabilityMetrics
	Components         []StabilityComponent
	Reasons            []StabilityReason
	Limitations        []Limitation
	Explanations       []Explanation
	ScopeGuard         ScopeGuard
	Provenance         StabilityProvenance
	EvaluatedAt        time.Time
}

func (result StabilityResult) Clone() StabilityResult {
	cloned := result
	cloned.Components = append([]StabilityComponent(nil), result.Components...)
	cloned.Reasons = append([]StabilityReason(nil), result.Reasons...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	return cloned
}
