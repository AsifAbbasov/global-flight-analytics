package sharedsnapshot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
)

var (
	ErrFanOutRunnerRequired = errors.New(
		"shared snapshot fan-out runner is required",
	)

	ErrEnvelopePublisherRequired = errors.New(
		"shared snapshot envelope publisher is required",
	)
)

type FanOutRunner interface {
	Run(
		ctx context.Context,
		tasks []providerfanout.Task[Payload],
	) ([]providerfanout.Result[Payload], error)
}

type EnvelopePublisher interface {
	PublishEnvelope(
		cycleStartedAt time.Time,
		envelope providerfanin.Envelope[Payload],
	) (Snapshot, error)
}

type CycleConfig struct {
	Runner    FanOutRunner
	Publisher EnvelopePublisher
	Now       func() time.Time
}

type Cycle struct {
	runner    FanOutRunner
	publisher EnvelopePublisher
	now       func() time.Time
}

func NewCycle(
	config CycleConfig,
) (*Cycle, error) {
	if config.Runner == nil {
		return nil,
			ErrFanOutRunnerRequired
	}

	if config.Publisher == nil {
		return nil,
			ErrEnvelopePublisherRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Cycle{
		runner:    config.Runner,
		publisher: config.Publisher,
		now:       now,
	}, nil
}

func (
	cycle *Cycle,
) Run(
	ctx context.Context,
	tasks []providerfanout.Task[Payload],
) (Snapshot, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cycleStartedAt := cycle.now()

	results, err := cycle.runner.Run(
		ctx,
		tasks,
	)
	if err != nil {
		return Snapshot{},
			fmt.Errorf(
				"run shared snapshot provider fan-out: %w",
				err,
			)
	}

	envelope := providerfanin.Aggregate(
		results,
	)

	snapshot, err := cycle.publisher.PublishEnvelope(
		cycleStartedAt,
		envelope,
	)
	if err != nil {
		return Snapshot{},
			fmt.Errorf(
				"publish shared snapshot cycle: %w",
				err,
			)
	}

	return snapshot, nil
}
