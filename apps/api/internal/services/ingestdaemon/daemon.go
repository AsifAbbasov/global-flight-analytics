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
	ErrMaxFailureBackoffInvalid = errors.New(
		"ingest daemon maximum failure backoff must be at least the interval",
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
	RunCycle          CycleRunner
	Interval          time.Duration
	MaxFailureBackoff time.Duration
	Now               Clock
	Wait              WaitFunction
	Observe           Observer
}

type Daemon struct {
	runCycle          CycleRunner
	interval          time.Duration
	maxFailureBackoff time.Duration
	now               Clock
	wait              WaitFunction
	observe           Observer
}

type CycleResult struct {
	Number              int
	StartedAt           time.Time
	FinishedAt          time.Time
	Err                 error
	ConsecutiveFailures int
	RetryAt             time.Time
	NextDelay           time.Duration
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

	maxFailureBackoff := config.MaxFailureBackoff
	if maxFailureBackoff <= 0 {
		maxFailureBackoff = config.Interval
	}
	if maxFailureBackoff < config.Interval {
		return nil, ErrMaxFailureBackoffInvalid
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
		runCycle:          config.RunCycle,
		interval:          config.Interval,
		maxFailureBackoff: maxFailureBackoff,
		now:               now,
		wait:              wait,
		observe:           config.Observe,
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
	consecutiveFailures := 0

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

		if cycleErr == nil {
			consecutiveFailures = 0
		} else {
			consecutiveFailures++
		}

		retryAt := retryAtFromError(
			cycleErr,
		)
		nextDelay := daemon.nextDelay(
			finishedAt,
			cycleErr,
			consecutiveFailures,
			retryAt,
		)

		if daemon.observe != nil {
			daemon.observe(
				CycleResult{
					Number:              cycleNumber,
					StartedAt:           startedAt,
					FinishedAt:          finishedAt,
					Err:                 cycleErr,
					ConsecutiveFailures: consecutiveFailures,
					RetryAt:             retryAt,
					NextDelay:           nextDelay,
				},
			)
		}

		if ctx.Err() != nil {
			return nil
		}

		if err := daemon.wait(
			ctx,
			nextDelay,
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

func (
	daemon *Daemon,
) nextDelay(
	finishedAt time.Time,
	cycleErr error,
	consecutiveFailures int,
	retryAt time.Time,
) time.Duration {
	delay := daemon.interval
	if cycleErr != nil {
		delay = exponentialBackoff(
			daemon.interval,
			daemon.maxFailureBackoff,
			consecutiveFailures,
		)
	}

	if !retryAt.IsZero() &&
		retryAt.After(finishedAt) {
		providerDelay := retryAt.Sub(
			finishedAt,
		)
		if providerDelay > delay {
			delay = providerDelay
		}
	}

	return delay
}

func exponentialBackoff(
	base time.Duration,
	maximum time.Duration,
	consecutiveFailures int,
) time.Duration {
	if consecutiveFailures <= 1 {
		return base
	}

	delay := base
	for step := 1; step < consecutiveFailures; step++ {
		if delay >= maximum {
			return maximum
		}
		if delay > maximum/2 {
			return maximum
		}
		delay *= 2
	}

	if delay > maximum {
		return maximum
	}

	return delay
}

func retryAtFromError(
	err error,
) time.Time {
	if err == nil {
		return time.Time{}
	}

	var evidence interface {
		RetryAtTime() time.Time
	}
	if errors.As(
		err,
		&evidence,
	) {
		return evidence.RetryAtTime().
			UTC()
	}

	return time.Time{}
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
