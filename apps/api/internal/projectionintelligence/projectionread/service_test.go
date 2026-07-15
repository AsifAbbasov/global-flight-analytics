package projectionread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type dataSourceStub struct {
	current    trajectory.FlightTrajectory
	currentErr error

	route    routecontract.Result
	routeErr error

	candidates    []trajectory.FlightTrajectory
	candidatesErr error

	history    projectionroutefrequency.HistorySummary
	historyErr error

	currentCalls   int
	routeCalls     int
	candidateCalls int
	historyCalls   int
}

func (
	stub *dataSourceStub,
) LoadCurrentTrajectory(
	context.Context,
	string,
	time.Time,
) (trajectory.FlightTrajectory, error) {
	stub.currentCalls++
	return stub.current,
		stub.currentErr
}

func (
	stub *dataSourceStub,
) LoadRoute(
	context.Context,
	string,
	time.Time,
) (routecontract.Result, error) {
	stub.routeCalls++
	return stub.route,
		stub.routeErr
}

func (
	stub *dataSourceStub,
) LoadHistoricalCandidates(
	context.Context,
	trajectory.FlightTrajectory,
	routecontract.Result,
	time.Time,
) ([]trajectory.FlightTrajectory, error) {
	stub.candidateCalls++
	return append(
			[]trajectory.FlightTrajectory(nil),
			stub.candidates...,
		),
		stub.candidatesErr
}

func (
	stub *dataSourceStub,
) LoadRouteHistory(
	context.Context,
	routecontract.Result,
	time.Time,
) (
	projectionroutefrequency.HistorySummary,
	error,
) {
	stub.historyCalls++
	return stub.history.Clone(),
		stub.historyErr
}

type composerStub struct {
	result  projectionproduction.Result
	err     error
	request projectionproduction.Request
	calls   int
}

func (
	stub *composerStub,
) Compose(
	request projectionproduction.Request,
) (projectionproduction.Result, error) {
	stub.calls++
	stub.request = request
	return stub.result.Clone(),
		stub.err
}

func TestServiceLoadsCompleteProductionInputs(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	history := projectionReadHistory(
		asOfTime,
	)
	source := &dataSourceStub{
		current: current,
		route:   route,
		candidates: []trajectory.FlightTrajectory{
			projectionReadTrajectory(
				"83aa02ab-7061-4e9e-a238-d32710371ee3",
				asOfTime.Add(
					-24*time.Hour,
				),
			),
		},
		history: history,
	}
	composer := &composerStub{
		result: projectionproduction.Result{
			Version: projectionproduction.Version,
		},
	}
	service, err := NewService(
		ServiceConfig{
			DataSource: source,
			Composer:   composer,
			Policy:     DefaultPolicy(),
			Now: func() time.Time {
				return asOfTime.Add(
					time.Second,
				)
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      current.ID,
			AsOfTime:          asOfTime,
			RequestedDuration: 5 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf(
			"Get() error = %v",
			err,
		)
	}

	if source.currentCalls != 1 ||
		source.routeCalls != 1 ||
		source.candidateCalls != 1 ||
		source.historyCalls != 1 ||
		composer.calls != 1 {
		t.Fatalf(
			"unexpected call counts: source=%#v composer=%d",
			source,
			composer.calls,
		)
	}
	if len(
		composer.request.
			HistoricalCandidates,
	) != 1 ||
		composer.request.RouteHistory == nil ||
		composer.request.Route.Status !=
			routecontract.RouteStatusComplete ||
		!composer.request.GeneratedAt.Equal(
			asOfTime.Add(time.Second),
		) {
		t.Fatalf(
			"unexpected composition request: %#v",
			composer.request,
		)
	}
}

func TestServiceUsesAuditableUnavailableRouteWithoutMaterializedRoute(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	source := &dataSourceStub{
		current:  current,
		routeErr: ErrRouteNotFound,
	}
	composer := &composerStub{
		result: projectionproduction.Result{
			Version: projectionproduction.Version,
		},
	}
	service, err := NewService(
		ServiceConfig{
			DataSource: source,
			Composer:   composer,
			Policy:     DefaultPolicy(),
			Now: func() time.Time {
				return asOfTime.Add(
					time.Second,
				)
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      current.ID,
			AsOfTime:          asOfTime,
			RequestedDuration: 5 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf(
			"Get() error = %v",
			err,
		)
	}

	report := routecontract.Validate(
		composer.request.Route,
	)
	if report.Status !=
		routecontract.ValidationStatusValid ||
		composer.request.Route.Status !=
			routecontract.RouteStatusUnavailable ||
		composer.request.RouteHistory != nil ||
		len(
			composer.request.
				HistoricalCandidates,
		) != 0 ||
		source.candidateCalls != 0 ||
		source.historyCalls != 0 {
		t.Fatalf(
			"unexpected unavailable-route composition: report=%#v request=%#v source=%#v",
			report,
			composer.request,
			source,
		)
	}
}

func TestServiceRejectsFutureAsOfBeforeLoadingData(
	t *testing.T,
) {
	now := projectionReadTestAsOfTime()
	source := &dataSourceStub{}
	composer := &composerStub{}
	service, err := NewService(
		ServiceConfig{
			DataSource: source,
			Composer:   composer,
			Policy:     DefaultPolicy(),
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
			AsOfTime:          now.Add(time.Minute),
			RequestedDuration: 5 * time.Minute,
		},
	)
	if !errors.Is(
		err,
		ErrInvalidRequest,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidRequest",
			err,
		)
	}
	if source.currentCalls != 0 ||
		composer.calls != 0 {
		t.Fatalf(
			"future request reached dependencies: source=%d composer=%d",
			source.currentCalls,
			composer.calls,
		)
	}
}

func TestServiceMapsCurrentTrajectoryNotFound(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	source := &dataSourceStub{
		currentErr: ErrTrajectoryNotFound,
	}
	service, err := NewService(
		ServiceConfig{
			DataSource: source,
			Composer:   &composerStub{},
			Policy:     DefaultPolicy(),
			Now: func() time.Time {
				return asOfTime
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"NewService() error = %v",
			err,
		)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(
		err,
		ErrTrajectoryNotFound,
	) {
		t.Fatalf(
			"error = %v, want ErrTrajectoryNotFound",
			err,
		)
	}
}
