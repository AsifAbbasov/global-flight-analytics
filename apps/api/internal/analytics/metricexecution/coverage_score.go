package metricexecution

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func (
	service *Service,
) CoverageScore(
	ctx context.Context,
	request CoverageScoreRequest,
) (Execution[float64], error) {
	return executeSnapshotMetric(
		ctx,
		service,
		MetricIDCoverageScore,
		trajectoryeligibility.
			CapabilityTrafficMetrics,
		request.PublicationMetadata,
		func(
			ctx context.Context,
			evaluatedAt time.Time,
		) (metricCalculation[float64], error) {
			if err := ctx.Err(); err != nil {
				return metricCalculation[float64]{},
					err
			}

			value, err :=
				(metrics.CoverageScoreMetric{}).
					Calculate(
						request.Snapshot,
					)
			if err != nil {
				return metricCalculation[float64]{},
					err
			}

			return metricCalculation[float64]{
				Value: value,
				Factors: serverOwnedSnapshotConfidenceFactors(
					request.Snapshot.ObservedSamples > 0,
					"Coverage score uses server-owned covered-to-expected observation intervals.",
				),
			}, nil
		},
	)
}
