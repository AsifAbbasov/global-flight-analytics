package projectionread

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestRouteScopedHistoricalCandidateEvidenceValidatesCandidates(
	t *testing.T,
) {
	route := routecontract.Result{
		Status: routecontract.RouteStatusComplete,
		Origin: &routecontract.EndpointInference{
			Airport: routecontract.AirportReference{
				ICAOCode: "UBBB",
			},
		},
		Destination: &routecontract.EndpointInference{
			Airport: routecontract.AirportReference{
				ICAOCode: "LTBA",
			},
		},
	}
	candidates := []trajectory.FlightTrajectory{
		{ID: "historical-a"},
		{ID: "historical-b"},
	}

	scope, err := routeScopedHistoricalCandidateEvidence(
		route,
		candidates,
		DefaultSourceName,
	)
	if err != nil {
		t.Fatalf(
			"routeScopedHistoricalCandidateEvidence() error = %v",
			err,
		)
	}
	if err := scope.ValidateForCandidates(candidates); err != nil {
		t.Fatalf("route scope validation error = %v", err)
	}
	if scope.Route.OriginICAO != "UBBB" ||
		scope.Route.DestinationICAO != "LTBA" {
		t.Fatalf("unexpected route scope: %#v", scope)
	}
}

func TestSnapshotCloneDoesNotShareRouteScope(
	t *testing.T,
) {
	scope := routeScopedTestScope(t)
	snapshot := Snapshot{
		HistoricalCandidateRouteScope: routeScopePointer(scope),
	}

	cloned := snapshot.Clone()
	cloned.HistoricalCandidateRouteScope.SourceName = "changed"
	if snapshot.HistoricalCandidateRouteScope.SourceName == "changed" {
		t.Fatal("Snapshot.Clone() shared route-scope state")
	}
}

func TestServiceForwardsSourceOwnedRouteScope(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(current, asOfTime)
	candidate := projectionReadTrajectory(
		"83aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime.Add(-24*time.Hour),
	)
	candidates := []trajectory.FlightTrajectory{candidate}
	scope, err := routeScopedHistoricalCandidateEvidence(
		route,
		candidates,
		DefaultSourceName,
	)
	if err != nil {
		t.Fatalf("build route scope error = %v", err)
	}

	source := &dataSourceStub{
		snapshot: Snapshot{
			CurrentTrajectory:             current,
			Route:                         routePointer(route),
			HistoricalCandidates:          candidates,
			HistoricalCandidateRouteScope: routeScopePointer(scope),
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
			RequestedDuration: time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if composer.request.HistoricalCandidateRouteScope.InputFingerprint !=
		scope.InputFingerprint {
		t.Fatalf(
			"forwarded route scope = %#v, want %#v",
			composer.request.HistoricalCandidateRouteScope,
			scope,
		)
	}
}

func routeScopedTestScope(t *testing.T) projectionneighbors.RouteScope {
	t.Helper()
	scope := projectionneighbors.UniformRouteScope(
		"UBBB",
		"LTBA",
		DefaultSourceName+":"+routeScopedCandidateEvidenceVersion,
	)
	if err := scope.ValidateForCandidates(nil); err != nil {
		t.Fatalf("route scope validation error = %v", err)
	}
	return scope
}
