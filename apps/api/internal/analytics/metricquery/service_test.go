package metricquery

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type repositoryStub struct {
	recentItems []trajectory.FlightTrajectory
	idItems     []trajectory.FlightTrajectory
	err         error

	from  time.Time
	to    time.Time
	limit int
	ids   []string
}

func (stub *repositoryStub) ListTrajectoriesByEndTime(
	ctx context.Context,
	from time.Time,
	to time.Time,
	limit int,
) ([]trajectory.FlightTrajectory, error) {
	stub.from = from
	stub.to = to
	stub.limit = limit
	return stub.recentItems, stub.err
}

func (stub *repositoryStub) ListTrajectoriesByIDs(
	ctx context.Context,
	ids []string,
) ([]trajectory.FlightTrajectory, error) {
	stub.ids = append([]string(nil), ids...)
	return stub.idItems, stub.err
}

func TestNewRequiresRepository(t *testing.T) {
	service, err := New(nil)
	if service != nil || !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository requirement, got %#v %v", service, err)
	}
}

func TestRecentQueriesNormalizedWindow(t *testing.T) {
	now := time.Date(2026, time.July, 14, 11, 0, 0, 0, time.UTC)
	stub := &repositoryStub{
		recentItems: []trajectory.FlightTrajectory{{ID: "trajectory-1"}},
	}
	service, err := NewWithClock(stub, func() time.Time { return now })
	if err != nil {
		t.Fatalf("expected service, got %v", err)
	}

	items, err := service.Recent(nil, RecentRequest{WindowMinutes: 30, Limit: 50})
	if err != nil {
		t.Fatalf("expected recent trajectories, got %v", err)
	}

	if stub.from != now.Add(-30*time.Minute) ||
		stub.to != now ||
		stub.limit != 50 ||
		len(items) != 1 {
		t.Fatalf("unexpected recent query: from=%s to=%s limit=%d items=%#v", stub.from, stub.to, stub.limit, items)
	}

	items[0].ID = "mutated"
	if stub.recentItems[0].ID != "trajectory-1" {
		t.Fatal("expected result slice to be copied")
	}
}

func TestByIDsNormalizesAndDeduplicates(t *testing.T) {
	stub := &repositoryStub{}
	service, err := New(stub)
	if err != nil {
		t.Fatalf("expected service, got %v", err)
	}

	_, err = service.ByIDs(context.Background(), []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222", "11111111-1111-4111-8111-111111111111"})
	if err != nil {
		t.Fatalf("expected id query, got %v", err)
	}

	expected := []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"}
	if !reflect.DeepEqual(stub.ids, expected) {
		t.Fatalf("expected %#v, got %#v", expected, stub.ids)
	}
}
