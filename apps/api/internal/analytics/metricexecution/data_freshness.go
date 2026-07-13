package metricexecution

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func (
	service *Service,
) DataFreshness(
	ctx context.Context,
	request DataFreshnessRequest,
) (Execution[float64], error) {
	return executeSnapshotMetric(
		ctx,
		service,
		metrics.DataFreshnessMetricID,
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
				(metrics.DataFreshnessMetric{
					MaxAge: request.MaxAge,
				}).Calculate(
					request.Snapshot,
					evaluatedAt,
				)
			if err != nil {
				return metricCalculation[float64]{},
					err
			}

			return metricCalculation[float64]{
				Value: value,
				Factors: methodConfidenceFactors(
					"Data freshness uses a deterministic timestamp-age calculation.",
				),
			}, nil
		},
	)
}
