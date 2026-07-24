package metrics

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

func TestTrafficDensityMetric(t *testing.T) {
	metric := TrafficDensityMetric{}

	value, err := metric.Calculate(snapshot.Snapshot{
		ActiveAircraft:       250,
		AreaSquareKilometers: 5000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := 0.05

	if value != expected {
		t.Fatalf("expected %v, got %v", expected, value)
	}
}

func TestTrafficDensityMetricRejectsZeroArea(t *testing.T) {
	metric := TrafficDensityMetric{}

	_, err := metric.Calculate(snapshot.Snapshot{})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTrafficDensityMetricRejectsNonFiniteArea(
	t *testing.T,
) {
	testCases := []struct {
		name string
		area float64
	}{
		{
			name: "not a number",
			area: math.NaN(),
		},
		{
			name: "positive infinity",
			area: math.Inf(1),
		},
		{
			name: "negative infinity",
			area: math.Inf(-1),
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				_, err := (TrafficDensityMetric{}).Calculate(
					snapshot.Snapshot{
						ActiveAircraft:       1,
						AreaSquareKilometers: testCase.area,
					},
				)
				if err == nil {
					t.Fatal("expected non-finite area error")
				}
			},
		)
	}
}

func TestTrafficDensityMetricRejectsNegativeAircraftCount(
	t *testing.T,
) {
	_, err := (TrafficDensityMetric{}).Calculate(
		snapshot.Snapshot{
			ActiveAircraft:       -1,
			AreaSquareKilometers: 100,
		},
	)
	if err == nil {
		t.Fatal("expected negative aircraft count error")
	}
}
