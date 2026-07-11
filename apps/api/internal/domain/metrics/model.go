package metrics

import "time"

const (
	ActiveAircraftMetricName = MetricName("active_aircraft")

	DefaultActiveAircraftWindowMinutes = 15
	MinimumActiveAircraftWindowMinutes = 1
	MaximumActiveAircraftWindowMinutes = 180
)

type MetricName string

type MetricScopeType string

const (
	MetricScopeGlobal MetricScopeType = "global"
	MetricScopeRegion MetricScopeType = "region"
)

type ConfidenceLevel string

const (
	ConfidenceLevelHigh   ConfidenceLevel = "high"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelNone   ConfidenceLevel = "none"
)

type MetricScope struct {
	Type MetricScopeType
	Code string
}

type MetricConfidence struct {
	Level   ConfidenceLevel
	Score   float64
	Reasons []string
}

type MetricSource struct {
	Name string
	Role string
}

type Bounds struct {
	MinLatitude  float64
	MaxLatitude  float64
	MinLongitude float64
	MaxLongitude float64
}

type ActiveAircraftRequest struct {
	RegionCode    string
	WindowMinutes int
}

type ActiveAircraftQuery struct {
	ObservedFrom time.Time
	ObservedTo   time.Time
	UseBounds    bool
	Bounds       Bounds
}

type ActiveAircraftObservationSummary struct {
	Count            int
	FirstObservedAt  time.Time
	LatestObservedAt time.Time
	SourceNames      []string
	HasObservations  bool
}

type ActiveAircraftMetric struct {
	Metric        MetricName
	Value         int
	WindowMinutes int
	Scope         MetricScope
	ObservedFrom  time.Time
	ObservedTo    time.Time
	CalculatedAt  time.Time
	Confidence    MetricConfidence
	Sources       []MetricSource
	Limitations   []string
}
