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
		return nil,
			ErrDataSourceRequired
	}
	if config.Composer == nil {
		return nil,
			ErrComposerRequired
	}
	if err := config.Policy.Validate(); err != nil {
		return nil,
			fmt.Errorf(
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
		return projectionproduction.Result{},
			err
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
		asOfTime.After(
			generatedAt,
		) {
		return projectionproduction.Result{},
			fmt.Errorf(
				"%w: as-of time must not exceed generation time",
				ErrInvalidRequest,
			)
	}

	current, err :=
		service.dataSource.LoadCurrentTrajectory(
			ctx,
			trajectoryID,
			asOfTime,
		)
	if err != nil {
		return projectionproduction.Result{},
			classifyCurrentTrajectoryError(
				err,
			)
	}

	route, err :=
		service.dataSource.LoadRoute(
			ctx,
			trajectoryID,
			asOfTime,
		)
	if errors.Is(err, ErrRouteNotFound) {
		route = unavailableRoute(
			current,
			asOfTime,
			generatedAt,
		)
	} else if err != nil {
		return projectionproduction.Result{},
			fmt.Errorf(
				"load Projection Intelligence route: %w",
				err,
			)
	}

	historicalCandidates := []trajectory.FlightTrajectory{}
	var routeHistoryPointer *projectionroutefrequency.HistorySummary

	if route.Status ==
		routecontract.RouteStatusComplete {
		historicalCandidates, err =
			service.dataSource.
				LoadHistoricalCandidates(
					ctx,
					current,
					route,
					asOfTime,
				)
		if err != nil {
			return projectionproduction.Result{},
				fmt.Errorf(
					"load Projection Intelligence historical candidates: %w",
					err,
				)
		}

		routeHistory, historyErr :=
			service.dataSource.LoadRouteHistory(
				ctx,
				route,
				asOfTime,
			)
		switch {
		case historyErr == nil:
			routeHistoryCopy :=
				routeHistory.Clone()
			routeHistoryPointer =
				&routeHistoryCopy
		case errors.Is(
			historyErr,
			ErrRouteHistoryNotFound,
		):
			routeHistoryPointer = nil
		default:
			return projectionproduction.Result{},
				fmt.Errorf(
					"load Projection Intelligence route history: %w",
					historyErr,
				)
		}
	}

	result, err := service.composer.Compose(
		projectionproduction.Request{
			CurrentTrajectory:    current,
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

func classifyCurrentTrajectoryError(
	err error,
) error {
	if errors.Is(
		err,
		ErrTrajectoryNotFound,
	) {
		return ErrTrajectoryNotFound
	}

	return fmt.Errorf(
		"load Projection Intelligence current trajectory: %w",
		err,
	)
}
