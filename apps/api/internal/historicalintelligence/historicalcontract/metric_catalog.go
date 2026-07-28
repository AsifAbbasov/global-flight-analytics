package historicalcontract

import "math"

// MetricFamily identifies the production builder family that owns a metric.
type MetricFamily string

const (
	MetricFamilyTraffic MetricFamily = "traffic"
	MetricFamilyAirport MetricFamily = "airport"
	MetricFamilyRoute   MetricFamily = "route"
)

// MetricValueKind identifies semantic validation for one metric value.
type MetricValueKind string

const (
	MetricValueKindCount      MetricValueKind = "count"
	MetricValueKindRatio      MetricValueKind = "ratio"
	MetricValueKindContinuous MetricValueKind = "continuous"
)

// MaximumExactCountValue is the largest integer accepted through the float64
// transport field without relying on non-portable integer rounding semantics.
const MaximumExactCountValue = 9_007_199_254_740_991

const (
	continuousRelativeTolerance = 1e-9
	ratioAbsoluteTolerance      = 1e-12
)

// MetricSpec is the single production catalog for metric identity, unit,
// aggregation, value semantics, builder ownership, and supported scopes.
type MetricSpec struct {
	Name          MetricName
	Unit          string
	Aggregation   Aggregation
	ValueKind     MetricValueKind
	Family        MetricFamily
	AllowedScopes []ScopeType
}

func (specification MetricSpec) AllowsScope(scopeType ScopeType) bool {
	for _, allowed := range specification.AllowedScopes {
		if allowed == scopeType {
			return true
		}
	}
	return false
}

func MetricSpecFor(name MetricName) (MetricSpec, bool) {
	for _, specification := range metricCatalog {
		if specification.Name != name {
			continue
		}
		cloned := specification
		cloned.AllowedScopes = append([]ScopeType(nil), specification.AllowedScopes...)
		return cloned, true
	}
	return MetricSpec{}, false
}

func comparisonValueForMetric(metric Metric, summary Summary) (float64, bool) {
	switch metric.Aggregation {
	case AggregationCount, AggregationSum:
		return summary.Total, true
	case AggregationMinimum:
		return summary.Minimum, true
	case AggregationMaximum:
		return summary.Maximum, true
	case AggregationAverage, AggregationRatio:
		return summary.Average, true
	case AggregationMedian:
		return summary.Median, true
	default:
		return 0, false
	}
}

func metricValuesEqual(specification MetricSpec, left float64, right float64) bool {
	if !isFinite(left) || !isFinite(right) {
		return false
	}
	switch specification.ValueKind {
	case MetricValueKindCount:
		return left == right
	case MetricValueKindRatio:
		return math.Abs(left-right) <= ratioAbsoluteTolerance
	default:
		difference := math.Abs(left - right)
		scale := math.Max(1, math.Max(math.Abs(left), math.Abs(right)))
		return difference <= continuousRelativeTolerance*scale
	}
}

func validateMetricPointValue(
	point Point,
	metric Metric,
	fieldPrefix string,
	collector *validationCollector,
) {
	specification, exists := MetricSpecFor(metric.Name)
	if !exists || !isFinite(point.Value) || point.Value < 0 {
		return
	}
	switch specification.ValueKind {
	case MetricValueKindCount:
		if math.Trunc(point.Value) != point.Value {
			collector.add(
				ValidationSeverityError,
				"bucket_count_value_not_integral",
				fieldPrefix+".value",
				"Count metric values must be exact non-negative integers.",
			)
		}
		if point.Value > MaximumExactCountValue {
			collector.add(
				ValidationSeverityError,
				"bucket_count_value_not_exact",
				fieldPrefix+".value",
				"Count metric value exceeds the exact float64 integer boundary.",
			)
		}
	case MetricValueKindRatio:
		if point.Value > 1 {
			collector.add(
				ValidationSeverityError,
				"bucket_ratio_value_invalid",
				fieldPrefix+".value",
				"Ratio metric values must be between zero and one.",
			)
		}
	}
}

var metricCatalog = []MetricSpec{
	{Name: MetricNameActiveAircraft, Unit: "aircraft", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyTraffic, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameFlightCount, Unit: "flights", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyTraffic, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameTrajectoryCount, Unit: "trajectories", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyTraffic, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameObservationCount, Unit: "observations", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyTraffic, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameTrafficDensity, Unit: "observations_per_hour", Aggregation: AggregationAverage, ValueKind: MetricValueKindContinuous, Family: MetricFamilyTraffic, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameAirportDepartures, Unit: "departures", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyAirport, AllowedScopes: []ScopeType{ScopeTypeAirport}},
	{Name: MetricNameAirportArrivals, Unit: "arrivals", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyAirport, AllowedScopes: []ScopeType{ScopeTypeAirport}},
	{Name: MetricNameAirportOperations, Unit: "operations", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyAirport, AllowedScopes: []ScopeType{ScopeTypeAirport}},
	{Name: MetricNameUniqueAircraft, Unit: "aircraft", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyAirport, AllowedScopes: []ScopeType{ScopeTypeAirport}},
	{Name: MetricNameActiveRoutes, Unit: "route_pairs", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal, ScopeTypeRoute}},
	{Name: MetricNameRouteObservations, Unit: "route_results", Aggregation: AggregationCount, ValueKind: MetricValueKindCount, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal, ScopeTypeRoute}},
	{Name: MetricNameRouteConfidence, Unit: "ratio", Aggregation: AggregationAverage, ValueKind: MetricValueKindRatio, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal, ScopeTypeRoute}},
	{Name: MetricNameCompleteRouteRatio, Unit: "ratio", Aggregation: AggregationRatio, ValueKind: MetricValueKindRatio, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNamePartialRouteRatio, Unit: "ratio", Aggregation: AggregationRatio, ValueKind: MetricValueKindRatio, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameUnavailableRouteRatio, Unit: "ratio", Aggregation: AggregationRatio, ValueKind: MetricValueKindRatio, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal}},
	{Name: MetricNameGreatCircleDistanceKM, Unit: "kilometres", Aggregation: AggregationAverage, ValueKind: MetricValueKindContinuous, Family: MetricFamilyRoute, AllowedScopes: []ScopeType{ScopeTypeGlobal, ScopeTypeRoute}},
}
