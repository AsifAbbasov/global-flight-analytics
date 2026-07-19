package projectionread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type dataSourceStub struct {
	snapshot Snapshot
	err      error
	request  SnapshotRequest
	calls    int
}

func (
	stub *dataSourceStub,
) LoadSnapshot(
	_ context.Context,
	request SnapshotRequest,
) (Snapshot, error) {
	stub.calls++
	stub.request = request
	return stub.snapshot.Clone(), stub.err
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
	return stub.result.Clone(), stub.err
}

func TestServiceLoadsOneConsistentProductionSnapshot(
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
	history := projectionReadHistory(asOfTime)
	candidate := projectionReadTrajectory(
		"83aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime.Add(-24*time.Hour),
	)
	source := &dataSourceStub{
		snapshot: Snapshot{
			CurrentTrajectory: current,
			Route:             routePointer(route),
			HistoricalCandidates: []trajectory.FlightTrajectory{
				candidate,
			},
			RouteHistory: historyPointer(history),
		},
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
				return asOfTime.Add(time.Second)
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
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
		t.Fatalf("Get() error = %v", err)
	}

	if source.calls != 1 || composer.calls != 1 {
		t.Fatalf(
			"unexpected call counts: snapshot=%d composer=%d",
			source.calls,
			composer.calls,
		)
	}
	if source.request.TrajectoryID != current.ID ||
		!source.request.AsOfTime.Equal(asOfTime) {
		t.Fatalf(
			"unexpected snapshot request: %#v",
			source.request,
		)
	}
	if len(composer.request.HistoricalCandidates) != 1 ||
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

func TestServiceUsesAuditableUnavailableRouteWithoutSnapshotRoute(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	source := &dataSourceStub{
		snapshot: Snapshot{
			CurrentTrajectory: current,
		},
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
				return asOfTime.Add(time.Second)
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
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
		t.Fatalf("Get() error = %v", err)
	}

	report := routecontract.Validate(
		composer.request.Route,
	)
	if report.Status != routecontract.ValidationStatusValid ||
		composer.request.Route.Status !=
			routecontract.RouteStatusUnavailable ||
		composer.request.RouteHistory != nil ||
		len(composer.request.HistoricalCandidates) != 0 ||
		source.calls != 1 {
		t.Fatalf(
			"unexpected unavailable-route composition: report=%#v request=%#v calls=%d",
			report,
			composer.request,
			source.calls,
		)
	}
}

func TestServiceRejectsFutureAsOfBeforeLoadingSnapshot(
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
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
			AsOfTime:          now.Add(time.Minute),
			RequestedDuration: 5 * time.Minute,
		},
	)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
	if source.calls != 0 || composer.calls != 0 {
		t.Fatalf(
			"future request reached dependencies: source=%d composer=%d",
			source.calls,
			composer.calls,
		)
	}
}

func TestServiceMapsSnapshotTrajectoryNotFound(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	source := &dataSourceStub{
		err: ErrTrajectoryNotFound,
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
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      "73aa02ab-7061-4e9e-a238-d32710371ee3",
			AsOfTime:          asOfTime,
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(err, ErrTrajectoryNotFound) {
		t.Fatalf(
			"error = %v, want ErrTrajectoryNotFound",
			err,
		)
	}
}
