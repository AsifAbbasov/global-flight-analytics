package metrics

import (
	"fmt"
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

const TrafficDensityMetricID = "traffic_density"

type TrafficDensityMetric struct{}

func (TrafficDensityMetric) ID() string {
	return TrafficDensityMetricID
}

func (TrafficDensityMetric) Name() string {
	return "Traffic Density"
}

func (TrafficDensityMetric) Calculate(data snapshot.Snapshot) (float64, error) {
	if data.ActiveAircraft < 0 {
		return 0, fmt.Errorf("active aircraft count cannot be negative")
	}

	if math.IsNaN(data.AreaSquareKilometers) ||
		math.IsInf(data.AreaSquareKilometers, 0) ||
		data.AreaSquareKilometers <= 0 {
		return 0, fmt.Errorf("area must be finite and greater than zero")
	}

	return float64(data.ActiveAircraft) / data.AreaSquareKilometers, nil
}
