package metricquery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type contextContractRepository struct {
	recentCalls   int
	byIDCalls     int
	regionalCalls int
}

func (repository *contextContractRepository) ListTrajectoriesByEndTime(
	context.Context,
	time.Time,
	time.Time,
	int,
) ([]trajectory.FlightTrajectory, error) {
	repository.recentCalls++
	return nil, nil
}

func (repository *contextContractRepository) ListTrajectoriesByIDs(
	context.Context,
	[]string,
) ([]trajectory.FlightTrajectory, error) {
	repository.byIDCalls++
	return nil, nil
}

func (repository *contextContractRepository) ListTrajectoriesWithinBounds(
	context.Context,
	time.Time,
	time.Time,
	Bounds,
	int,
) ([]trajectory.FlightTrajectory, error) {
	repository.regionalCalls++
	return nil, nil
}

func TestRecentRejectsNilContextBeforeRepositoryRead(t *testing.T) {
	repository := &contextContractRepository{}
	service := &Service{
		repository: repository,
		now: func() time.Time {
			return time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
		},
	}

	_, err := service.Recent(nil, RecentRequest{})
	if !errors.Is(err, ErrQueryContextRequired) {
		t.Fatalf("error = %v, want context required", err)
	}
	if repository.recentCalls != 0 {
		t.Fatalf("recent repository calls = %d, want 0", repository.recentCalls)
	}
}

func TestByIDsRejectsNilContextBeforeRepositoryRead(t *testing.T) {
	repository := &contextContractRepository{}
	service := &Service{
		repository: repository,
		now:        time.Now,
	}

	_, err := service.ByIDs(nil, nil)
	if !errors.Is(err, ErrQueryContextRequired) {
		t.Fatalf("error = %v, want context required", err)
	}
	if repository.byIDCalls != 0 {
		t.Fatalf("by-id repository calls = %d, want 0", repository.byIDCalls)
	}
}

func TestRecentWithinBoundsRejectsNilContextBeforeRepositoryRead(t *testing.T) {
	repository := &contextContractRepository{}
	service := &Service{
		repository: repository,
		now:        time.Now,
	}

	_, err := service.RecentWithinBounds(nil, RecentRequest{}, Bounds{})
	if !errors.Is(err, ErrQueryContextRequired) {
		t.Fatalf("error = %v, want context required", err)
	}
	if repository.regionalCalls != 0 {
		t.Fatalf("regional repository calls = %d, want 0", repository.regionalCalls)
	}
}
