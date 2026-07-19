package projectionread

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type ServiceConfig struct {
	DataSource DataSource
	Composer   Composer
	Policy     Policy
	Now        func() time.Time
}

type Service struct {
	dataSource DataSource
	composer   Composer
	policy     Policy
	now        func() time.Time
}

func NewService(
	config ServiceConfig,
) (*Service, error) {
	if config.DataSource == nil {
		return nil, ErrDataSourceRequired
	}
	if config.Composer == nil {
		return nil, ErrComposerRequired
	}
	if err := config.Policy.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate Projection Intelligence read policy: %w",
			err,
		)
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		dataSource: config.DataSource,
		composer:   config.Composer,
		policy:     config.Policy,
		now:        now,
	}, nil
}

func (
	service *Service,
) Get(
	ctx context.Context,
	request Request,
) (projectionproduction.Result, error) {
	if service == nil ||
		service.dataSource == nil ||
		service.composer == nil {
		return projectionproduction.Result{},
			ErrServiceUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return projectionproduction.Result{}, err
	}

	trajectoryID := strings.TrimSpace(
		request.TrajectoryID,
	)
	asOfTime := request.AsOfTime.UTC()
	if trajectoryID == "" ||
		asOfTime.IsZero() ||
		request.RequestedDuration <= 0 {
		return projectionproduction.Result{},
			ErrInvalidRequest
	}

	generatedAt := service.now().UTC()
	if generatedAt.IsZero() ||
		asOfTime.After(generatedAt) {
		return projectionproduction.Result{},
			fmt.Errorf(
				"%w: as-of time must not exceed generation time",
				ErrInvalidRequest,
			)
	}

	snapshot, err := service.dataSource.LoadSnapshot(
		ctx,
		SnapshotRequest{
			TrajectoryID: trajectoryID,
			AsOfTime:     asOfTime,
		},
	)
	if err != nil {
		return projectionproduction.Result{},
			classifySnapshotError(err)
	}

	route := unavailableRoute(
		snapshot.CurrentTrajectory,
		asOfTime,
		generatedAt,
	)
	if snapshot.Route != nil {
		route = snapshot.Route.Clone()
	}

	historicalCandidates :=
		[]trajectory.FlightTrajectory{}
	var routeHistoryPointer *projectionroutefrequency.HistorySummary
	if route.Status == routecontract.RouteStatusComplete {
		historicalCandidates = append(
			[]trajectory.FlightTrajectory(nil),
			snapshot.HistoricalCandidates...,
		)
		if snapshot.RouteHistory != nil {
			historyCopy := snapshot.RouteHistory.Clone()
			routeHistoryPointer = &historyCopy
		}
	}

	result, err := service.composer.Compose(
		projectionproduction.Request{
			CurrentTrajectory: snapshot.
				CurrentTrajectory,
			HistoricalCandidates: historicalCandidates,
			Route:                route,
			RouteHistory:         routeHistoryPointer,
			AsOfTime:             asOfTime,
			RequestedDuration: request.
				RequestedDuration,
			GeneratedAt: generatedAt,
		},
	)
	if err != nil {
		return projectionproduction.Result{},
			fmt.Errorf(
				"compose Projection Intelligence result: %w",
				err,
			)
	}

	return result.Clone(), nil
}

func classifySnapshotError(
	err error,
) error {
	if errors.Is(err, ErrTrajectoryNotFound) {
		return ErrTrajectoryNotFound
	}

	return fmt.Errorf(
		"load Projection Intelligence snapshot: %w",
		err,
	)
}
