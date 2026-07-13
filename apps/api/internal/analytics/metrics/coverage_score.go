package metrics

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

const CoverageScoreMetricID = "coverage_score"

type CoverageScoreMetric struct{}

func (CoverageScoreMetric) ID() string {
	return CoverageScoreMetricID
}

func (CoverageScoreMetric) Name() string {
	return "Coverage Score"
}

func (CoverageScoreMetric) Calculate(data snapshot.Snapshot) (float64, error) {
	if data.ExpectedSamples <= 0 {
		return 0, fmt.Errorf("expected sample count must be greater than zero")
	}

	if data.ObservedSamples < 0 {
		return 0, fmt.Errorf("observed sample count cannot be negative")
	}

	if data.ObservedSamples >= data.ExpectedSamples {
		return 1, nil
	}

	return float64(data.ObservedSamples) / float64(data.ExpectedSamples), nil
}
