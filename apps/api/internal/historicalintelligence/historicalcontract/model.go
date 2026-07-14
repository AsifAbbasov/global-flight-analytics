package historicalcontract

import (
	"sort"
	"time"
)

const Version = "historical-intelligence-contract-v1"

type SchemaVersion string

const SchemaVersionV1 SchemaVersion = "historical-intelligence-v1"

type SeriesStatus string

const (
	SeriesStatusUnavailable SeriesStatus = "unavailable"
	SeriesStatusPartial     SeriesStatus = "partial"
	SeriesStatusComplete    SeriesStatus = "complete"
)

type BucketStatus string

const (
	BucketStatusUnavailable BucketStatus = "unavailable"
	BucketStatusPartial     BucketStatus = "partial"
	BucketStatusComplete    BucketStatus = "complete"
)

type Granularity string

const (
	GranularityHour   Granularity = "hour"
	GranularityDay    Granularity = "day"
	GranularityWeek   Granularity = "week"
	GranularityCustom Granularity = "custom"
)

type ScopeType string

const (
	ScopeTypeGlobal  ScopeType = "global"
	ScopeTypeRegion  ScopeType = "region"
	ScopeTypeAirport ScopeType = "airport"
	ScopeTypeRoute   ScopeType = "route"
)

type MetricName string

const (
	MetricNameActiveAircraft        MetricName = "active_aircraft"
	MetricNameFlightCount           MetricName = "flight_count"
	MetricNameTrajectoryCount       MetricName = "trajectory_count"
	MetricNameObservationCount      MetricName = "observation_count"
	MetricNamePeakActivity          MetricName = "peak_activity"
	MetricNameAverageActivity       MetricName = "average_activity"
	MetricNameTrafficDensity        MetricName = "traffic_density"
	MetricNameDataFreshness         MetricName = "data_freshness"
	MetricNameCoverageScore         MetricName = "coverage_score"
	MetricNameAirportDepartures     MetricName = "airport_departures"
	MetricNameAirportArrivals       MetricName = "airport_arrivals"
	MetricNameAirportOperations     MetricName = "airport_operations"
	MetricNameUniqueAircraft        MetricName = "unique_aircraft"
	MetricNameActiveRoutes          MetricName = "active_routes"
	MetricNameRouteObservations     MetricName = "route_observations"
	MetricNameRouteConfidence       MetricName = "route_confidence"
	MetricNameCompleteRouteRatio    MetricName = "complete_route_ratio"
	MetricNamePartialRouteRatio     MetricName = "partial_route_ratio"
	MetricNameUnavailableRouteRatio MetricName = "unavailable_route_ratio"
	MetricNameGreatCircleDistanceKM MetricName = "great_circle_distance_km"
)

type Aggregation string

const (
	AggregationCount   Aggregation = "count"
	AggregationSum     Aggregation = "sum"
	AggregationMinimum Aggregation = "minimum"
	AggregationMaximum Aggregation = "maximum"
	AggregationAverage Aggregation = "average"
	AggregationMedian  Aggregation = "median"
	AggregationRatio   Aggregation = "ratio"
)

type ConfidenceLevel string

const (
	ConfidenceLevelNone   ConfidenceLevel = "none"
	ConfidenceLevelLow    ConfidenceLevel = "low"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelHigh   ConfidenceLevel = "high"
)

type TrendDirection string

const (
	TrendDirectionUnavailable TrendDirection = "unavailable"
	TrendDirectionDown        TrendDirection = "down"
	TrendDirectionFlat        TrendDirection = "flat"
	TrendDirectionUp          TrendDirection = "up"
)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationSeverity string

const (
	ValidationSeverityError   ValidationSeverity = "error"
	ValidationSeverityWarning ValidationSeverity = "warning"
)

type TimeWindow struct {
	StartTime time.Time
	EndTime   time.Time
	AsOfTime  time.Time
}

func (window TimeWindow) Duration() time.Duration {
	if window.StartTime.IsZero() ||
		window.EndTime.IsZero() {
		return 0
	}

	return window.EndTime.Sub(window.StartTime)
}

type Scope struct {
	Type                ScopeType
	RegionCode          string
	AirportICAOCode     string
	OriginICAOCode      string
	DestinationICAOCode string
}

type Metric struct {
	Name        MetricName
	Unit        string
	Aggregation Aggregation
}

type ConfidenceReason struct {
	Code         string
	Message      string
	Contribution float64
}

type Confidence struct {
	Score       float64
	Level       ConfidenceLevel
	SampleCount int
	Reasons     []ConfidenceReason
}

type Limitation struct {
	Code    string
	Message string
	Scope   string
}

type Point struct {
	StartTime     time.Time
	EndTime       time.Time
	Status        BucketStatus
	Value         float64
	SampleCount   int
	CoverageRatio float64
	Confidence    Confidence
	Limitations   []Limitation
}

type Summary struct {
	PointCount int
	Total      float64
	Minimum    float64
	Maximum    float64
	Average    float64
	Median     float64
}

type PeriodComparison struct {
	PreviousWindow   TimeWindow
	CurrentValue     float64
	PreviousValue    float64
	AbsoluteChange   float64
	PercentageChange *float64
	Direction        TrendDirection
}

type Provenance struct {
	BuilderVersion        string
	InputFingerprint      string
	SourceNames           []string
	LatestSourceUpdatedAt time.Time
}

type Result struct {
	SchemaVersion SchemaVersion
	Status        SeriesStatus
	Metric        Metric
	Scope         Scope
	Window        TimeWindow
	Granularity   Granularity
	Points        []Point
	Summary       Summary
	Comparison    *PeriodComparison
	Confidence    Confidence
	Limitations   []Limitation
	Provenance    Provenance
	GeneratedAt   time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Points = clonePoints(result.Points)
	cloned.Comparison = cloneComparison(
		result.Comparison,
	)
	cloned.Confidence = cloneConfidence(
		result.Confidence,
	)
	cloned.Limitations = append(
		[]Limitation(nil),
		result.Limitations...,
	)
	cloned.Provenance.SourceNames = append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)

	return cloned
}

func clonePoints(items []Point) []Point {
	cloned := make([]Point, 0, len(items))
	for _, item := range items {
		copied := item
		copied.Confidence = cloneConfidence(
			item.Confidence,
		)
		copied.Limitations = append(
			[]Limitation(nil),
			item.Limitations...,
		)
		cloned = append(cloned, copied)
	}

	return cloned
}

func cloneComparison(
	comparison *PeriodComparison,
) *PeriodComparison {
	if comparison == nil {
		return nil
	}

	cloned := *comparison
	if comparison.PercentageChange != nil {
		value := *comparison.PercentageChange
		cloned.PercentageChange = &value
	}

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

func ConfidenceLevelForScore(
	score float64,
) ConfidenceLevel {
	switch {
	case score >= 0.8:
		return ConfidenceLevelHigh
	case score >= 0.6:
		return ConfidenceLevelMedium
	case score > 0:
		return ConfidenceLevelLow
	default:
		return ConfidenceLevelNone
	}
}

func TrendDirectionForChange(
	absoluteChange float64,
) TrendDirection {
	const flatTolerance = 1e-9

	switch {
	case absoluteChange > flatTolerance:
		return TrendDirectionUp
	case absoluteChange < -flatTolerance:
		return TrendDirectionDown
	default:
		return TrendDirectionFlat
	}
}

func SupportedMetricNames() []MetricName {
	result := append(
		[]MetricName(nil),
		supportedMetricNames...,
	)
	sort.Slice(
		result,
		func(left int, right int) bool {
			return result[left] < result[right]
		},
	)

	return result
}

var supportedMetricNames = []MetricName{
	MetricNameActiveAircraft,
	MetricNameFlightCount,
	MetricNameTrajectoryCount,
	MetricNameObservationCount,
	MetricNamePeakActivity,
	MetricNameAverageActivity,
	MetricNameTrafficDensity,
	MetricNameDataFreshness,
	MetricNameCoverageScore,
	MetricNameAirportDepartures,
	MetricNameAirportArrivals,
	MetricNameAirportOperations,
	MetricNameUniqueAircraft,
	MetricNameActiveRoutes,
	MetricNameRouteObservations,
	MetricNameRouteConfidence,
	MetricNameCompleteRouteRatio,
	MetricNamePartialRouteRatio,
	MetricNameUnavailableRouteRatio,
	MetricNameGreatCircleDistanceKM,
}
