package server

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/etareliability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionevaluation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/jackc/pgx/v5/pgxpool"
)

type etaReliabilityApplicationReader interface {
	Get(context.Context, projectionread.Request) (etareliability.Result, error)
}

type etaReliabilityReaderAdapter struct {
	reader etaReliabilityApplicationReader
}

func (adapter etaReliabilityReaderAdapter) GetETAReliability(
	ctx context.Context,
	request handlers.ProjectionIntelligenceReadRequest,
) (etareliability.Result, error) {
	if adapter.reader == nil {
		return etareliability.Result{}, etareliability.ErrServiceUnavailable
	}
	return adapter.reader.Get(ctx, projectionread.Request{
		TrajectoryID:      request.TrajectoryID,
		AsOfTime:          request.AsOfTime,
		RequestedDuration: request.RequestedDuration,
	})
}

func newETAReliabilityPostgresReader(pool *pgxpool.Pool) (handlers.ETAReliabilityReader, error) {
	composition, err := projectionread.NewPostgres(projectionread.PostgresConfig{
		Pool:   pool,
		Policy: projectionread.DefaultPolicy(),
	})
	if err != nil {
		return nil, fmt.Errorf("compose PostgreSQL Projection Intelligence dependency for ETA Reliability: %w", err)
	}
	evaluator, err := projectionevaluation.New(projectionevaluation.Config{
		MaximumInterpolationGap:     3 * time.Minute,
		MaximumTruthGroundSpeedMPS:  400,
		MaximumTruthVerticalRateMPS: 100,
		MinimumEvaluatedPointCount:  1,
		MaximumHorizontalErrorM:     10000,
		MaximumAltitudeErrorM:       1000,
		LeadTimeBucketSize:          time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("compose projection evaluator for ETA Reliability: %w", err)
	}
	service, err := etareliability.New(etareliability.ServiceConfig{
		ProjectionReader: composition.Service,
		SnapshotReader:   composition.DataSource,
		Evaluator:        evaluator,
		Policy:           etareliability.DefaultPolicy(),
	})
	if err != nil {
		return nil, fmt.Errorf("compose ETA Reliability service: %w", err)
	}
	return etaReliabilityReaderAdapter{reader: service}, nil
}
