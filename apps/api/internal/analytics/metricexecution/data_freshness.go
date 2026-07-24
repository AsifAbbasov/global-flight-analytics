package metricexecution

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
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
		MetricIDDataFreshness,
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

			if request.Snapshot.Time.IsZero() {
				return metricCalculation[float64]{
					Value: 0,
					Factors: serverOwnedSnapshotConfidenceFactors(
						false,
						"Data freshness uses the latest server-owned retained trajectory observation.",
					),
					Limitations: []analyticalresult.Notice{
						{
							Code:    NoticeCodeNoTrajectoryObservations,
							Message: "No usable retained trajectory observation was available for freshness calculation.",
						},
					},
				}, nil
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
				Factors: serverOwnedSnapshotConfidenceFactors(
					true,
					"Data freshness uses the latest server-owned retained trajectory observation.",
				),
			}, nil
		},
	)
}
