package historicalcontract

import "testing"

func TestHistoricalRouteMetricScopePolicy(t *testing.T) {
	statusRatios := []MetricName{
		MetricNameCompleteRouteRatio,
		MetricNamePartialRouteRatio,
		MetricNameUnavailableRouteRatio,
	}
	for _, metricName := range statusRatios {
		specification, exists := MetricSpecFor(metricName)
		if !exists {
			t.Fatalf("metric %q is missing", metricName)
		}
		if !specification.AllowsScope(ScopeTypeGlobal) {
			t.Fatalf("metric %q must support global scope", metricName)
		}
		if specification.AllowsScope(ScopeTypeRoute) {
			t.Fatalf("metric %q must reject route-pair scope", metricName)
		}
	}
}

func TestActiveRoutesDefinesDirectionalRoutePairs(t *testing.T) {
	specification, exists := MetricSpecFor(MetricNameActiveRoutes)
	if !exists {
		t.Fatal("active routes metric is missing")
	}
	if specification.Unit != "route_pairs" {
		t.Fatalf("active routes unit = %q, want route_pairs", specification.Unit)
	}
}
