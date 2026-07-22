package metricquery

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type regionalRepositoryStub struct {
	repositoryStub
	bounds        Bounds
	regionalItems []trajectory.FlightTrajectory
	regionalCalls int
}

func (stub *regionalRepositoryStub) ListTrajectoriesWithinBounds(
	ctx context.Context,
	from time.Time,
	to time.Time,
	bounds Bounds,
	limit int,
) ([]trajectory.FlightTrajectory, error) {
	stub.from = from
	stub.to = to
	stub.bounds = bounds
	stub.limit = limit
	stub.regionalCalls++

	return stub.regionalItems, stub.err
}

func TestWorldBoundsAreaMatchesEarthSurfaceArea(t *testing.T) {
	bounds := Bounds{
		MinLatitude:  -90,
		MaxLatitude:  90,
		MinLongitude: -180,
		MaxLongitude: 180,
	}

	area, err := bounds.AreaSquareKilometers()
	if err != nil {
		t.Fatalf("expected world area, got %v", err)
	}

	expected := 4 * math.Pi *
		meanEarthRadiusKilometers *
		meanEarthRadiusKilometers
	if math.Abs(area-expected) > 0.001 {
		t.Fatalf("expected %.3f square kilometers, got %.3f", expected, area)
	}
}

func TestBoundsRejectInvalidLatitudeOrder(t *testing.T) {
	_, err := (Bounds{
		MinLatitude:  44,
		MaxLatitude:  38,
		MinLongitude: 38,
		MaxLongitude: 51,
	}).AreaSquareKilometers()

	if !errors.Is(err, ErrBoundsInvalid) {
		t.Fatalf("expected invalid bounds, got %v", err)
	}
}

func TestRecentWithinBoundsQueriesRegionalRepository(t *testing.T) {
	now := time.Date(2026, time.July, 14, 12, 0, 0, 0, time.UTC)
	bounds := Bounds{
		MinLatitude:  35,
		MaxLatitude:  43,
		MinLongitude: 25,
		MaxLongitude: 45,
	}
	stub := &regionalRepositoryStub{
		regionalItems: []trajectory.FlightTrajectory{
			{ID: "trajectory-one"},
		},
	}
	service, err := NewWithClock(stub, func() time.Time { return now })
	if err != nil {
		t.Fatalf("expected service, got %v", err)
	}

	items, err := service.RecentWithinBounds(
		nil,
		RecentRequest{
			WindowMinutes: 30,
			Limit:         200,
		},
		bounds,
	)
	if err != nil {
		t.Fatalf("expected regional trajectories, got %v", err)
	}

	if stub.regionalCalls != 1 ||
		stub.from != now.Add(-30*time.Minute) ||
		stub.to != now ||
		stub.limit != 200 ||
		stub.bounds != bounds ||
		len(items) != 1 {
		t.Fatalf(
			"unexpected regional query: calls=%d from=%s to=%s limit=%d bounds=%#v items=%#v",
			stub.regionalCalls,
			stub.from,
			stub.to,
			stub.limit,
			stub.bounds,
			items,
		)
	}

	items[0].ID = "mutated"
	if stub.regionalItems[0].ID != "trajectory-one" {
		t.Fatal("expected regional result slice to be copied")
	}
}

func TestRecentWithinBoundsRequiresRegionalRepository(t *testing.T) {
	service, err := New(&repositoryStub{})
	if err != nil {
		t.Fatalf("expected service, got %v", err)
	}

	_, err = service.RecentWithinBounds(
		context.Background(),
		RecentRequest{},
		Bounds{
			MinLatitude:  35,
			MaxLatitude:  43,
			MinLongitude: 25,
			MaxLongitude: 45,
		},
	)
	if !errors.Is(err, ErrRegionalRepositoryUnsupported) {
		t.Fatalf("expected regional repository requirement, got %v", err)
	}
}
