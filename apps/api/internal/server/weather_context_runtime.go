package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontext"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type weatherContextApplicationReader interface {
	Get(
		context.Context,
		weathercontext.Request,
	) (weathercontext.Result, error)
}

type weatherContextReaderAdapter struct {
	reader weatherContextApplicationReader
}

func (
	adapter weatherContextReaderAdapter,
) GetWeatherContext(
	ctx context.Context,
	request handlers.WeatherContextReadRequest,
) (handlers.WeatherContextReadResult, error) {
	if adapter.reader == nil {
		return handlers.WeatherContextReadResult{},
			handlers.ErrWeatherContextServiceUnavailable
	}

	result, err := adapter.reader.Get(
		ctx,
		weathercontext.Request{
			TrajectoryID:      request.TrajectoryID,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		return handlers.WeatherContextReadResult{},
			mapWeatherContextReadError(err)
	}

	return handlers.WeatherContextReadResult{
		Version:          handlers.WeatherContextReadResultVersion,
		Weather:          result.Weather.Clone(),
		Trust:            result.Trust.Clone(),
		Alignment:        result.Alignment.Clone(),
		Encounter:        result.Encounter.Clone(),
		Uncertainty:      result.Uncertainty.Clone(),
		InputFingerprint: result.InputFingerprint,
		GeneratedAt:      result.GeneratedAt,
	}, nil
}

type weatherContextTrajectoryApplicationReader interface {
	GetTrajectoryByID(
		context.Context,
		string,
	) (trajectory.FlightTrajectory, error)
}

type weatherContextTrajectoryReaderAdapter struct {
	reader weatherContextTrajectoryApplicationReader
}

func (
	adapter weatherContextTrajectoryReaderAdapter,
) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	if adapter.reader == nil {
		return trajectory.FlightTrajectory{},
			weathercontext.ErrServiceUnavailable
	}

	result, err := adapter.reader.GetTrajectoryByID(
		ctx,
		trajectoryID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return trajectory.FlightTrajectory{},
			weathercontext.ErrTrajectoryNotFound
	}
	if errors.Is(
		err,
		trafficquery.ErrTrajectoryRepositoryRequired,
	) {
		return trajectory.FlightTrajectory{},
			weathercontext.ErrServiceUnavailable
	}
	if errors.Is(
		err,
		trafficquery.ErrInvalidTrajectoryID,
	) {
		return trajectory.FlightTrajectory{},
			weathercontext.ErrInvalidRequest
	}
	if err != nil {
		return trajectory.FlightTrajectory{}, err
	}

	return result, nil
}

type weatherContextProjectionReaderAdapter struct {
	reader handlers.ProjectionIntelligenceReader
}

func (
	adapter weatherContextProjectionReaderAdapter,
) GetProjection(
	ctx context.Context,
	request weathercontext.ProjectionRequest,
) (projectionproduction.Result, error) {
	if adapter.reader == nil {
		return projectionproduction.Result{},
			weathercontext.ErrServiceUnavailable
	}

	result, err := adapter.reader.GetProjectionIntelligence(
		ctx,
		handlers.ProjectionIntelligenceReadRequest{
			TrajectoryID:      request.TrajectoryID,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	switch {
	case errors.Is(
		err,
		handlers.ErrProjectionIntelligenceNotFound,
	):
		return projectionproduction.Result{},
			weathercontext.ErrProjectionNotFound
	case errors.Is(
		err,
		handlers.ErrProjectionIntelligenceServiceUnavailable,
	):
		return projectionproduction.Result{},
			weathercontext.ErrServiceUnavailable
	case errors.Is(
		err,
		handlers.ErrProjectionIntelligenceInvalidRequest,
	):
		return projectionproduction.Result{},
			weathercontext.ErrInvalidRequest
	case err != nil:
		return projectionproduction.Result{}, err
	default:
		return result.Clone(), nil
	}
}

func newWeatherContextPostgresReader(
	pool *pgxpool.Pool,
	projectionReader handlers.ProjectionIntelligenceReader,
) (handlers.WeatherContextReader, error) {
	trajectoryRepository := postgres.NewTrajectoryRepository(
		pool,
	)
	trajectoryService := trafficquery.New(
		trafficquery.Config{
			TrajectoryRepository: trajectoryRepository,
		},
	)

	weatherSnapshotReader, err :=
		weathercontext.NewPostgresSnapshotReader(
			pool,
			weathercontext.DefaultPostgresSnapshotPolicy(),
		)
	if err != nil {
		return nil, fmt.Errorf(
			"compose PostgreSQL Weather Context snapshot reader: %w",
			err,
		)
	}

	service, err := weathercontext.NewService(
		weathercontext.Config{
			TrajectoryReader: weatherContextTrajectoryReaderAdapter{
				reader: trajectoryService,
			},
			WeatherSnapshotReader: weatherSnapshotReader,
			ProjectionReader: weatherContextProjectionReaderAdapter{
				reader: projectionReader,
			},
			TrustPolicy:       weathertrust.DefaultPolicy(),
			AlignmentPolicy:   weatheralignment.DefaultPolicy(),
			EncounterPolicy:   weatherencounter.DefaultPolicy(),
			UncertaintyPolicy: weatheruncertainty.DefaultPolicy(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"compose production Weather Context service: %w",
			err,
		)
	}

	return weatherContextReaderAdapter{
		reader: service,
	}, nil
}

func mapWeatherContextReadError(
	err error,
) error {
	switch {
	case errors.Is(err, weathercontext.ErrTrajectoryNotFound),
		errors.Is(err, weathercontext.ErrWeatherNotFound),
		errors.Is(err, weathercontext.ErrProjectionNotFound):
		return handlers.ErrWeatherContextNotFound

	case errors.Is(err, weathercontext.ErrServiceUnavailable):
		return handlers.ErrWeatherContextServiceUnavailable

	case errors.Is(err, weathercontext.ErrInvalidRequest):
		return handlers.ErrWeatherContextInvalidRequest

	default:
		return err
	}
}
