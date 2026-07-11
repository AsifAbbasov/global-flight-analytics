package ingestdaemon

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCycleRunnerRequired = errors.New(
		"ingest daemon cycle runner is required",
	)
	ErrIntervalInvalid = errors.New(
		"ingest daemon interval must be greater than zero",
	)
)

type Clock func() time.Time

type CycleRunner func(
	ctx context.Context,
) error

type WaitFunction func(
	ctx context.Context,
	duration time.Duration,
) error

type Observer func(
	result CycleResult,
)

type Config struct {
	RunCycle CycleRunner
	Interval time.Duration
	Now      Clock
	Wait     WaitFunction
	Observe  Observer
}

type Daemon struct {
	runCycle CycleRunner
	interval time.Duration
	now      Clock
	wait     WaitFunction
	observe  Observer
}

type CycleResult struct {
	Number     int
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error
}

func New(
	config Config,
) (*Daemon, error) {
	if config.RunCycle == nil {
		return nil, ErrCycleRunnerRequired
	}

	if config.Interval <= 0 {
		return nil, ErrIntervalInvalid
	}

	now := config.Now
	if now == nil {
		now = func() time.Time {
			return time.Now().UTC()
		}
	}

	wait := config.Wait
	if wait == nil {
		wait = waitForDuration
	}

	return &Daemon{
		runCycle: config.RunCycle,
		interval: config.Interval,
		now:      now,
		wait:     wait,
		observe:  config.Observe,
	}, nil
}

func (
	daemon *Daemon,
) Run(
	ctx context.Context,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cycleNumber := 0

	for {
		if ctx.Err() != nil {
			return nil
		}

		cycleNumber++

		startedAt := daemon.now().
			UTC()

		cycleErr := daemon.runCycle(
			ctx,
		)

		finishedAt := daemon.now().
			UTC()

		if daemon.observe != nil {
			daemon.observe(
				CycleResult{
					Number:     cycleNumber,
					StartedAt:  startedAt,
					FinishedAt: finishedAt,
					Err:        cycleErr,
				},
			)
		}

		if ctx.Err() != nil {
			return nil
		}

		if err := daemon.wait(
			ctx,
			daemon.interval,
		); err != nil {
			if ctx.Err() != nil ||
				errors.Is(
					err,
					context.Canceled,
				) ||
				errors.Is(
					err,
					context.DeadlineExceeded,
				) {
				return nil
			}

			return fmt.Errorf(
				"wait before next ingest cycle: %w",
				err,
			)
		}
	}
}

func waitForDuration(
	ctx context.Context,
	duration time.Duration,
) error {
	timer := time.NewTimer(
		duration,
	)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-timer.C:
		return nil
	}
}
