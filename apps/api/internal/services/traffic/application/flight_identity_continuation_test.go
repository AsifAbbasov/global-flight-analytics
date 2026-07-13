package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/flightcontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

func TestApplyFlightIdentityContinuationsContinuesSingleTrajectory(
	t *testing.T,
) {
	previous, current := continuationTrajectoryPair()
	repository := &fakeContinuationRepository{
		items: map[string]trajectory.FlightTrajectory{
			"ABC123": previous,
		},
	}
	service := &Service{
		trajectoryContinuationRepository: repository,
		identityContinuationConfig: flightcontinuation.Config{
			MaxGap: 5 * time.Minute,
		},
	}
	result := processor.ProcessingResult{
		Trajectories: map[string]trajectory.FlightTrajectory{
			"ABC123": current,
		},
	}

	count, err := service.applyFlightIdentityContinuations(
		context.Background(),
		&result,
	)
	if err != nil {
		t.Fatalf("expected no continuation error, got %v", err)
	}

	if count != 1 {
		t.Fatalf(
			"expected 1 continued trajectory, got %d",
			count,
		)
	}

	item, exists := result.Trajectories["ABC123"]
	if !exists {
		t.Fatal("expected the single-aircraft collection key to remain ABC123")
	}

	if item.IdentityKey != previous.IdentityKey {
		t.Fatalf(
			"expected previous identity key, got %s",
			item.IdentityKey,
		)
	}

	if item.SplitReason !=
		trajectory.FlightSplitReasonContinuedFromPreviousBatch {
		t.Fatalf(
			"expected continuation split reason, got %s",
			item.SplitReason,
		)
	}
}

func TestApplyFlightIdentityContinuationsOnlyUsesEarliestCurrentGroup(
	t *testing.T,
) {
	previous, earliest := continuationTrajectoryPair()
	later := earliest
	later.IdentityKey = testContinuationIdentityKey("c")
	later.Callsign = "AHY102"
	later.StartTime = earliest.StartTime.Add(time.Minute)
	later.EndTime = earliest.EndTime.Add(time.Minute)
	later.SplitReason =
		trajectory.FlightSplitReasonCallsignChanged

	service := &Service{
		trajectoryContinuationRepository: &fakeContinuationRepository{
			items: map[string]trajectory.FlightTrajectory{
				"ABC123": previous,
			},
		},
		identityContinuationConfig: flightcontinuation.Config{
			MaxGap: 5 * time.Minute,
		},
	}
	result := processor.ProcessingResult{
		Trajectories: map[string]trajectory.FlightTrajectory{
			earliest.IdentityKey: earliest,
			later.IdentityKey:    later,
		},
	}

	count, err := service.applyFlightIdentityContinuations(
		context.Background(),
		&result,
	)
	if err != nil {
		t.Fatalf("expected no continuation error, got %v", err)
	}

	if count != 1 {
		t.Fatalf(
			"expected 1 continued trajectory, got %d",
			count,
		)
	}

	if _, exists := result.Trajectories[previous.IdentityKey]; !exists {
		t.Fatal("expected earliest group to be re-keyed to previous identity")
	}

	if _, exists := result.Trajectories[later.IdentityKey]; !exists {
		t.Fatal("expected later split group to keep its new identity")
	}
}

func TestApplyFlightIdentityContinuationsIgnoresNotFound(
	t *testing.T,
) {
	_, current := continuationTrajectoryPair()
	service := &Service{
		trajectoryContinuationRepository: &fakeContinuationRepository{},
		identityContinuationConfig: flightcontinuation.Config{
			MaxGap: 5 * time.Minute,
		},
	}
	result := processor.ProcessingResult{
		Trajectories: map[string]trajectory.FlightTrajectory{
			"ABC123": current,
		},
	}

	count, err := service.applyFlightIdentityContinuations(
		context.Background(),
		&result,
	)
	if err != nil {
		t.Fatalf("expected not found to be ignored, got %v", err)
	}

	if count != 0 {
		t.Fatalf(
			"expected zero continuations, got %d",
			count,
		)
	}
}

func TestApplyFlightIdentityContinuationsReturnsRepositoryErrorWithoutMutation(
	t *testing.T,
) {
	_, current := continuationTrajectoryPair()
	expectedError := errors.New("database unavailable")
	service := &Service{
		trajectoryContinuationRepository: &fakeContinuationRepository{
			err: expectedError,
		},
		identityContinuationConfig: flightcontinuation.Config{
			MaxGap: 5 * time.Minute,
		},
	}
	result := processor.ProcessingResult{
		Trajectories: map[string]trajectory.FlightTrajectory{
			"ABC123": current,
		},
	}

	count, err := service.applyFlightIdentityContinuations(
		context.Background(),
		&result,
	)

	if count != 0 {
		t.Fatalf(
			"expected zero applied continuations, got %d",
			count,
		)
	}

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if result.Trajectories["ABC123"].IdentityKey !=
		current.IdentityKey {
		t.Fatal("expected result to remain unchanged after repository error")
	}
}

func TestNewRejectsNegativeIdentityContinuationMaximumGap(
	t *testing.T,
) {
	service, err := New(
		Config{
			IdentityContinuationMaxGap: -time.Second,
		},
	)

	if err == nil {
		t.Fatal("expected negative continuation maximum gap error")
	}

	if service != nil {
		t.Fatal("expected nil service for invalid continuation configuration")
	}
}

type fakeContinuationRepository struct {
	items map[string]trajectory.FlightTrajectory
	err   error
}

func (
	repository *fakeContinuationRepository,
) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	if repository.err != nil {
		return trajectory.FlightTrajectory{},
			repository.err
	}

	item, exists := repository.items[icao24]
	if !exists {
		return trajectory.FlightTrajectory{},
			trajectory.ErrNotFound
	}

	return item, nil
}

func continuationTrajectoryPair() (
	trajectory.FlightTrajectory,
	trajectory.FlightTrajectory,
) {
	now := time.Date(
		2026,
		time.July,
		13,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	previous := trajectory.FlightTrajectory{
		IdentityKey: testContinuationIdentityKey("a"),
		IdentityBasis: trajectory.
			FlightIdentityBasisCallsignAndStartTime,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		ICAO24:   "ABC123",
		Callsign: "AHY101",
		StartTime: now.Add(
			-10 * time.Minute,
		),
		EndTime: now,
	}

	current := trajectory.FlightTrajectory{
		IdentityKey: testContinuationIdentityKey("b"),
		IdentityBasis: trajectory.
			FlightIdentityBasisCallsignAndStartTime,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		ICAO24:   "ABC123",
		Callsign: "AHY101",
		StartTime: now.Add(
			time.Minute,
		),
		EndTime: now.Add(
			2 * time.Minute,
		),
	}

	return previous, current
}

func testContinuationIdentityKey(
	character string,
) string {
	return "flight-identity-" +
		strings.Repeat(
			character,
			64,
		)
}
