package forecastanalysis

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
)

const (
	Version                = "forecast-stability-analysis-v1"
	SchemaVersionV1        = "forecast-stability-analysis-v1"
	ScopeGuardResearchOnly = "research_only_not_for_operational_forecast_or_decision_use"
)

type ResultStatus string

const (
	ResultStatusLimited  ResultStatus = "limited"
	ResultStatusComplete ResultStatus = "complete"
)

type Trend string

const (
	TrendInsufficient Trend = "insufficient_history"
	TrendSteady       Trend = "steady"
	TrendImproving    Trend = "improving"
	TrendDegrading    Trend = "degrading"
	TrendVolatile     Trend = "volatile"
)

type Health string

const (
	HealthInsufficient Health = "insufficient_evidence"
	HealthStable       Health = "stable"
	HealthWatch        Health = "watch"
	HealthUnstable     Health = "unstable"
)

type Request struct {
	Versions    []forecaststability.ForecastVersionRecord
	EvaluatedAt time.Time
}

type Metrics struct {
	VersionCount                     int
	TransitionCount                  int
	ComparableTransitionCount        int
	UnchangedCount                   int
	StableCount                      int
	ChangedCount                     int
	MaterialChangeCount              int
	IndeterminateCount               int
	StableTransitionShare            float64
	ComparableTransitionShare        float64
	MaterialChangeShare              float64
	MeanStabilityScore               float64
	MinimumStabilityScore            float64
	ScoreStandardDeviation           float64
	LongestStableRun                 int
	MethodChangeCount                int
	PolicyChangeCount                int
	ImplementationChangeCount        int
	InputChangeCount                 int
	OutputChangeCount                int
	MeanHorizontalShiftKilometers    float64
	MaximumHorizontalShiftKilometers float64
	LatestLevel                      forecaststability.StabilityLevel
}

type Reason struct {
	Code    string
	Message string
	Impact  float64
}

type Confidence struct {
	Score   float64
	Level   string
	Reasons []Reason
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
	InputFingerprint   string
	VersionIDs         []string
	OutputFingerprints []string
	PolicyVersion      string
}

type Result struct {
	SchemaVersion string
	Status        ResultStatus
	TrajectoryID  string
	Trend         Trend
	Health        Health
	Metrics       Metrics
	Transitions   []forecaststability.StabilityResult
	Confidence    Confidence
	Limitations   []Limitation
	Explanations  []Explanation
	ScopeGuard    string
	Provenance    Provenance
	EvaluatedAt   time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Transitions = make([]forecaststability.StabilityResult, 0, len(result.Transitions))
	for _, item := range result.Transitions {
		cloned.Transitions = append(cloned.Transitions, item.Clone())
	}
	cloned.Confidence.Reasons = append([]Reason(nil), result.Confidence.Reasons...)
	cloned.Limitations = append([]Limitation(nil), result.Limitations...)
	cloned.Explanations = append([]Explanation(nil), result.Explanations...)
	cloned.Provenance.VersionIDs = append([]string(nil), result.Provenance.VersionIDs...)
	cloned.Provenance.OutputFingerprints = append([]string(nil), result.Provenance.OutputFingerprints...)
	return cloned
}
