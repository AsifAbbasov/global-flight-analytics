package metrics

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

const DataFreshnessMetricID = "data_freshness"

type DataFreshnessMetric struct {
	MaxAge time.Duration
}

func (DataFreshnessMetric) ID() string {
	return DataFreshnessMetricID
}

func (DataFreshnessMetric) Name() string {
	return "Data Freshness"
}

func (metric DataFreshnessMetric) Calculate(
	data snapshot.Snapshot,
	evaluatedAt time.Time,
) (float64, error) {
	if metric.MaxAge <= 0 {
		return 0, fmt.Errorf("maximum data age must be greater than zero")
	}

	if data.Time.IsZero() {
		return 0, fmt.Errorf("snapshot time is required")
	}

	observedAt := data.Time.UTC()
	referenceTime := evaluatedAt.UTC()

	if referenceTime.Before(observedAt) {
		return 0, fmt.Errorf("evaluation time cannot be before snapshot time")
	}

	age := referenceTime.Sub(observedAt)
	if age >= metric.MaxAge {
		return 0, nil
	}

	return 1 - float64(age)/float64(metric.MaxAge), nil
}
