package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontext"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
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
	LoadCurrentTrajectory(
		context.Context,
		string,
		time.Time,
	) (trajectory.FlightTrajectory, error)
}

type weatherContextTrajectoryReaderAdapter struct {
	reader weatherContextTrajectoryApplicationReader
}

func (
	adapter weatherContextTrajectoryReaderAdapter,
) GetTrajectory(
	ctx context.Context,
	request weathercontext.TrajectoryRequest,
) (trajectory.FlightTrajectory, error) {
	if adapter.reader == nil {
		return trajectory.FlightTrajectory{},
			weathercontext.ErrServiceUnavailable
	}

	result, err := adapter.reader.LoadCurrentTrajectory(
		ctx,
		request.TrajectoryID,
		request.AsOfTime,
	)
	switch {
	case errors.Is(
		err,
		projectionread.ErrTrajectoryNotFound,
	):
		return trajectory.FlightTrajectory{},
			weathercontext.ErrTrajectoryNotFound
	case errors.Is(
		err,
		projectionread.ErrServiceUnavailable,
	):
		return trajectory.FlightTrajectory{},
			weathercontext.ErrServiceUnavailable
	case errors.Is(
		err,
		projectionread.ErrInvalidRequest,
	):
		return trajectory.FlightTrajectory{},
			weathercontext.ErrInvalidRequest
	case err != nil:
		return trajectory.FlightTrajectory{}, err
	default:
		return result, nil
	}
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

// NewWeatherContextPostgresReader exposes the production Weather Context
// composition for bounded runtime verification and server integration checks.
func NewWeatherContextPostgresReader(
	pool *pgxpool.Pool,
	projectionReader handlers.ProjectionIntelligenceReader,
) (handlers.WeatherContextReader, error) {
	return newWeatherContextPostgresReader(
		pool,
		projectionReader,
	)
}

func newWeatherContextPostgresReader(
	pool *pgxpool.Pool,
	projectionReader handlers.ProjectionIntelligenceReader,
) (handlers.WeatherContextReader, error) {
	trajectoryDataSource, err :=
		projectionread.NewPostgresDataSource(
			projectionread.PostgresDataSourceConfig{
				Pool: pool,
				Policy: projectionread.
					DefaultPolicy().DataSource,
			},
		)
	if err != nil {
		return nil, fmt.Errorf(
			"compose PostgreSQL Weather Context trajectory reader: %w",
			err,
		)
	}

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
				reader: trajectoryDataSource,
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
