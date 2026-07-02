package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestGetLatestTrajectoryByICAO24(t *testing.T) {
	repository := &fakeTrajectoryReadRepository{
		latestByICAO24: map[string]trajectory.FlightTrajectory{
			"ABC123": makeQueryTrajectory("trajectory-1", "ABC123"),
		},
	}

	service := New(Config{
		TrajectoryRepository: repository,
	})

	result, err := service.GetLatestTrajectoryByICAO24(context.Background(), "  abc123  ")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != "trajectory-1" {
		t.Fatalf("expected trajectory-1, got %s", result.ID)
	}

	if repository.lastICAO24 != "ABC123" {
		t.Fatalf("expected normalized ICAO24 ABC123, got %s", repository.lastICAO24)
	}
}

func TestGetLatestTrajectoryByICAO24RejectsEmptyICAO24(t *testing.T) {
	repository := &fakeTrajectoryReadRepository{}

	service := New(Config{
		TrajectoryRepository: repository,
	})

	_, err := service.GetLatestTrajectoryByICAO24(context.Background(), "   ")

	if !errors.Is(err, ErrInvalidICAO24) {
		t.Fatalf("expected ErrInvalidICAO24, got %v", err)
	}

	if repository.latestCallCount != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repository.latestCallCount)
	}
}

func TestGetLatestTrajectoryByICAO24RequiresRepository(t *testing.T) {
	service := New(Config{})

	_, err := service.GetLatestTrajectoryByICAO24(context.Background(), "ABC123")

	if !errors.Is(err, ErrTrajectoryRepositoryRequired) {
		t.Fatalf("expected ErrTrajectoryRepositoryRequired, got %v", err)
	}
}

func TestGetLatestTrajectoryByICAO24RejectsMalformedICAO24(t *testing.T) {
	repository := &fakeTrajectoryReadRepository{}

	service := New(Config{
		TrajectoryRepository: repository,
	})

	_, err := service.GetLatestTrajectoryByICAO24(context.Background(), "BAD")

	if !errors.Is(err, ErrInvalidICAO24) {
		t.Fatalf("expected ErrInvalidICAO24, got %v", err)
	}

	if repository.latestCallCount != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repository.latestCallCount)
	}
}

func TestGetLatestTrajectoryByICAO24WrapsRepositoryError(t *testing.T) {
	expectedError := errors.New("repository failed")

	repository := &fakeTrajectoryReadRepository{
		err: expectedError,
	}

	service := New(Config{
		TrajectoryRepository: repository,
	})

	_, err := service.GetLatestTrajectoryByICAO24(context.Background(), "ABC123")

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestGetTrajectoryByID(t *testing.T) {
	repository := &fakeTrajectoryReadRepository{
		byID: map[string]trajectory.FlightTrajectory{
			"trajectory-1": makeQueryTrajectory("trajectory-1", "ABC123"),
		},
	}

	service := New(Config{
		TrajectoryRepository: repository,
	})

	result, err := service.GetTrajectoryByID(context.Background(), "  trajectory-1  ")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID != "trajectory-1" {
		t.Fatalf("expected trajectory-1, got %s", result.ID)
	}

	if repository.lastTrajectoryID != "trajectory-1" {
		t.Fatalf("expected trimmed trajectory id, got %s", repository.lastTrajectoryID)
	}
}

func TestGetTrajectoryByIDRejectsEmptyID(t *testing.T) {
	repository := &fakeTrajectoryReadRepository{}

	service := New(Config{
		TrajectoryRepository: repository,
	})

	_, err := service.GetTrajectoryByID(context.Background(), "   ")

	if !errors.Is(err, ErrInvalidTrajectoryID) {
		t.Fatalf("expected ErrInvalidTrajectoryID, got %v", err)
	}

	if repository.byIDCallCount != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repository.byIDCallCount)
	}
}

func TestGetTrajectoryByIDRequiresRepository(t *testing.T) {
	service := New(Config{})

	_, err := service.GetTrajectoryByID(context.Background(), "trajectory-1")

	if !errors.Is(err, ErrTrajectoryRepositoryRequired) {
		t.Fatalf("expected ErrTrajectoryRepositoryRequired, got %v", err)
	}
}

func TestNormalizeICAO24(t *testing.T) {
	result := normalizeICAO24("  abc123  ")

	if result != "ABC123" {
		t.Fatalf("expected ABC123, got %s", result)
	}
}

type fakeTrajectoryReadRepository struct {
	latestByICAO24 map[string]trajectory.FlightTrajectory
	byID           map[string]trajectory.FlightTrajectory

	lastICAO24       string
	lastTrajectoryID string
	latestCallCount  int
	byIDCallCount    int

	err error
}

func (repository *fakeTrajectoryReadRepository) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	repository.latestCallCount++
	repository.lastICAO24 = icao24

	if repository.err != nil {
		return trajectory.FlightTrajectory{}, repository.err
	}

	item, ok := repository.latestByICAO24[icao24]
	if !ok {
		return trajectory.FlightTrajectory{}, errors.New("not found")
	}

	return item, nil
}

func (repository *fakeTrajectoryReadRepository) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	repository.byIDCallCount++
	repository.lastTrajectoryID = trajectoryID

	if repository.err != nil {
		return trajectory.FlightTrajectory{}, repository.err
	}

	item, ok := repository.byID[trajectoryID]
	if !ok {
		return trajectory.FlightTrajectory{}, errors.New("not found")
	}

	return item, nil
}

func makeQueryTrajectory(id string, icao24 string) trajectory.FlightTrajectory {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	return trajectory.FlightTrajectory{
		ID:               id,
		ICAO24:           icao24,
		Callsign:         "AHY101",
		StartTime:        now.Add(-5 * time.Minute),
		EndTime:          now,
		DurationSeconds:  300,
		SegmentCount:     1,
		PointCount:       5,
		CoverageGapCount: 0,
		QualityScore:     0.95,
		SourceName:       "test",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}
