package ingestdaemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunStartsFirstCycleImmediately(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	runCount := 0
	waitCount := 0

	daemon := mustNewDaemon(
		t,
		Config{
			RunCycle: func(
				context.Context,
			) error {
				runCount++

				return nil
			},
			Interval: time.Minute,
			Wait: func(
				context.Context,
				time.Duration,
			) error {
				waitCount++
				cancel()

				return context.Canceled
			},
		},
	)

	if err := daemon.Run(
		ctx,
	); err != nil {
		t.Fatalf(
			"run ingest daemon: %v",
			err,
		)
	}

	if runCount != 1 {
		t.Fatalf(
			"expected one immediate cycle, got %d",
			runCount,
		)
	}

	if waitCount != 1 {
		t.Fatalf(
			"expected one wait after the cycle, got %d",
			waitCount,
		)
	}
}

func TestRunContinuesAfterCycleFailure(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	temporaryErr := errors.New(
		"temporary provider failure",
	)

	runCount := 0
	waitCount := 0
	results := make(
		[]CycleResult,
		0,
		2,
	)

	daemon := mustNewDaemon(
		t,
		Config{
			RunCycle: func(
				context.Context,
			) error {
				runCount++

				if runCount == 1 {
					return temporaryErr
				}

				cancel()

				return nil
			},
			Interval: time.Minute,
			Wait: func(
				context.Context,
				time.Duration,
			) error {
				waitCount++

				return nil
			},
			Observe: func(
				result CycleResult,
			) {
				results = append(
					results,
					result,
				)
			},
		},
	)

	if err := daemon.Run(
		ctx,
	); err != nil {
		t.Fatalf(
			"run ingest daemon: %v",
			err,
		)
	}

	if runCount != 2 {
		t.Fatalf(
			"expected two cycles, got %d",
			runCount,
		)
	}

	if waitCount != 1 {
		t.Fatalf(
			"expected one wait between cycles, got %d",
			waitCount,
		)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected two observed results, got %d",
			len(results),
		)
	}

	if !errors.Is(
		results[0].Err,
		temporaryErr,
	) {
		t.Fatalf(
			"expected first cycle error %v, got %v",
			temporaryErr,
			results[0].Err,
		)
	}

	if results[1].Err != nil {
		t.Fatalf(
			"expected successful second cycle, got %v",
			results[1].Err,
		)
	}
}

func TestRunStopsGracefullyWhenContextIsCancelledDuringCycle(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	waitCalled := false

	daemon := mustNewDaemon(
		t,
		Config{
			RunCycle: func(
				context.Context,
			) error {
				cancel()

				return context.Canceled
			},
			Interval: time.Minute,
			Wait: func(
				context.Context,
				time.Duration,
			) error {
				waitCalled = true

				return nil
			},
		},
	)

	if err := daemon.Run(
		ctx,
	); err != nil {
		t.Fatalf(
			"run ingest daemon: %v",
			err,
		)
	}

	if waitCalled {
		t.Fatal(
			"expected cancellation to stop before waiting",
		)
	}
}

func TestRunDoesNotOverlapCycles(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	var active int32
	var maximumActive int32
	runCount := 0

	daemon := mustNewDaemon(
		t,
		Config{
			RunCycle: func(
				context.Context,
			) error {
				currentActive := atomic.AddInt32(
					&active,
					1,
				)

				for {
					previousMaximum := atomic.LoadInt32(
						&maximumActive,
					)
					if currentActive <= previousMaximum ||
						atomic.CompareAndSwapInt32(
							&maximumActive,
							previousMaximum,
							currentActive,
						) {
						break
					}
				}

				runCount++

				atomic.AddInt32(
					&active,
					-1,
				)

				if runCount == 2 {
					cancel()
				}

				return nil
			},
			Interval: time.Minute,
			Wait: func(
				context.Context,
				time.Duration,
			) error {
				return nil
			},
		},
	)

	if err := daemon.Run(
		ctx,
	); err != nil {
		t.Fatalf(
			"run ingest daemon: %v",
			err,
		)
	}

	if maximumActive != 1 {
		t.Fatalf(
			"expected at most one active cycle, got %d",
			maximumActive,
		)
	}
}

func TestRunSkipsCycleWhenContextIsAlreadyCancelled(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	runCount := 0

	daemon := mustNewDaemon(
		t,
		Config{
			RunCycle: func(
				context.Context,
			) error {
				runCount++

				return nil
			},
			Interval: time.Minute,
		},
	)

	if err := daemon.Run(
		ctx,
	); err != nil {
		t.Fatalf(
			"run ingest daemon: %v",
			err,
		)
	}

	if runCount != 0 {
		t.Fatalf(
			"expected no cycle, got %d",
			runCount,
		)
	}
}

func TestNewValidatesRequiredConfiguration(
	t *testing.T,
) {
	_, err := New(
		Config{
			Interval: time.Minute,
		},
	)
	if !errors.Is(
		err,
		ErrCycleRunnerRequired,
	) {
		t.Fatalf(
			"expected cycle runner error, got %v",
			err,
		)
	}

	_, err = New(
		Config{
			RunCycle: func(
				context.Context,
			) error {
				return nil
			},
		},
	)
	if !errors.Is(
		err,
		ErrIntervalInvalid,
	) {
		t.Fatalf(
			"expected interval error, got %v",
			err,
		)
	}
}

func mustNewDaemon(
	t *testing.T,
	config Config,
) *Daemon {
	t.Helper()

	daemon, err := New(
		config,
	)
	if err != nil {
		t.Fatalf(
			"create ingest daemon: %v",
			err,
		)
	}

	return daemon
}
