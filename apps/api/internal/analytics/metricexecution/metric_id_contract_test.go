package metricexecution

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
)

func TestMetricIDsUseOneCanonicalNamespace(
	t *testing.T,
) {
	testCases := []struct {
		name      string
		execution string
		metric    string
		expected  string
	}{
		{
			name:      "active aircraft",
			execution: MetricIDActiveAircraft,
			metric:    metrics.ActiveAircraftMetricID,
			expected:  "traffic.active_aircraft",
		},
		{
			name:      "traffic density",
			execution: MetricIDTrafficDensity,
			metric:    metrics.TrafficDensityMetricID,
			expected:  "traffic.traffic_density",
		},
		{
			name:      "airport activity",
			execution: MetricIDAirportActivity,
			metric:    metrics.AirportActivityMetricID,
			expected:  "traffic.airport_activity",
		},
		{
			name:      "coverage score",
			execution: MetricIDCoverageScore,
			metric:    metrics.CoverageScoreMetricID,
			expected:  "traffic.coverage_score",
		},
		{
			name:      "data freshness",
			execution: MetricIDDataFreshness,
			metric:    metrics.DataFreshnessMetricID,
			expected:  "traffic.data_freshness",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				if testCase.execution !=
					testCase.expected ||
					testCase.metric !=
						testCase.expected {
					t.Fatalf(
						"expected canonical id %q, got execution=%q metric=%q",
						testCase.expected,
						testCase.execution,
						testCase.metric,
					)
				}
			},
		)
	}

	if (metrics.ActiveAircraft{}).ID() !=
		MetricIDActiveAircraft {
		t.Fatal(
			"active aircraft metric ID method is inconsistent",
		)
	}

	if (metrics.AirportActivity{}).ID() !=
		MetricIDAirportActivity {
		t.Fatal(
			"airport activity metric ID method is inconsistent",
		)
	}
}
