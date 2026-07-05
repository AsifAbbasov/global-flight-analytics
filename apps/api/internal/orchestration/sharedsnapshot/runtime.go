package sharedsnapshot

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
)

type RuntimeConfig struct {
	Executor providerfanout.Executor
	Now      func() time.Time
}

type Runtime struct {
	store *Store
	cycle *Cycle
}

func NewRuntime(
	config RuntimeConfig,
) (*Runtime, error) {
	runner, err := providerfanout.New(
		config.Executor,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create shared snapshot fan-out runner: %w",
			err,
		)
	}

	store := NewStore()

	publisher, err := NewPublisher(
		PublisherConfig{
			Store: store,
			Now:   config.Now,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create shared snapshot publisher: %w",
			err,
		)
	}

	cycle, err := NewCycle(
		CycleConfig{
			Runner:    runner,
			Publisher: publisher,
			Now:       config.Now,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create shared snapshot cycle: %w",
			err,
		)
	}

	return &Runtime{
		store: store,
		cycle: cycle,
	}, nil
}

func (runtime *Runtime) Run(
	ctx context.Context,
	tasks []providerfanout.Task,
) (Snapshot, error) {
	return runtime.cycle.Run(
		ctx,
		tasks,
	)
}

func (runtime *Runtime) Current() (
	Snapshot,
	bool,
) {
	return runtime.store.Current()
}
