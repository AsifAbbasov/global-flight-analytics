package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/ingestdaemon"
)

func TestRunSingleIngestionCycleRunsExactlyOnce(
	t *testing.T,
) {
	t.Parallel()

	startedAt := time.Date(
		2026,
		time.August,
		3,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	finishedAt := startedAt.Add(2 * time.Second)
	clockValues := []time.Time{
		startedAt,
		finishedAt,
	}
	clockIndex := 0
	cycleCount := 0
	observed := ingestdaemon.CycleResult{}

	err := runSingleIngestionCycle(
		context.Background(),
		func(context.Context) error {
			cycleCount++
			return nil
		},
		func() time.Time {
			value := clockValues[clockIndex]
			clockIndex++
			return value
		},
		func(result ingestdaemon.CycleResult) {
			observed = result
		},
	)
	if err != nil {
		t.Fatalf("run one ingestion cycle: %v", err)
	}
	if cycleCount != 1 {
		t.Fatalf("expected one cycle, got %d", cycleCount)
	}
	if observed.Number != 1 {
		t.Fatalf("expected cycle number 1, got %d", observed.Number)
	}
	if !observed.StartedAt.Equal(startedAt) {
		t.Fatalf("unexpected start time: %s", observed.StartedAt)
	}
	if !observed.FinishedAt.Equal(finishedAt) {
		t.Fatalf("unexpected finish time: %s", observed.FinishedAt)
	}
	if observed.NextDelay != 0 {
		t.Fatalf("one-shot execution must not schedule a next delay: %s", observed.NextDelay)
	}
}

func TestRunSingleIngestionCycleReturnsWrappedCycleError(
	t *testing.T,
) {
	t.Parallel()

	expectedErr := errors.New("provider unavailable")
	observedErr := error(nil)

	err := runSingleIngestionCycle(
		context.Background(),
		func(context.Context) error {
			return expectedErr
		},
		func() time.Time {
			return time.Date(
				2026,
				time.August,
				3,
				8,
				0,
				0,
				0,
				time.UTC,
			)
		},
		func(result ingestdaemon.CycleResult) {
			observedErr = result.Err
		},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped cycle error, got %v", err)
	}
	if !errors.Is(observedErr, expectedErr) {
		t.Fatalf("observer did not receive cycle error: %v", observedErr)
	}
}

func TestRunSingleIngestionCycleRequiresContext(
	t *testing.T,
) {
	t.Parallel()

	err := runSingleIngestionCycle(
		nil,
		func(context.Context) error { return nil },
		nil,
		nil,
	)
	if !errors.Is(err, ingestdaemon.ErrContextRequired) {
		t.Fatalf("expected context error, got %v", err)
	}
}

func TestRunSingleIngestionCycleRequiresRunner(
	t *testing.T,
) {
	t.Parallel()

	err := runSingleIngestionCycle(
		context.Background(),
		nil,
		nil,
		nil,
	)
	if !errors.Is(err, ingestdaemon.ErrCycleRunnerRequired) {
		t.Fatalf("expected runner error, got %v", err)
	}
}
