package sharedsnapshot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
)

type recordingFanOutRunner struct {
	receivedTasks []providerfanout.Task[Payload]
	results       []providerfanout.Result[Payload]
	err           error
}

func (
	runner *recordingFanOutRunner,
) Run(
	_ context.Context,
	tasks []providerfanout.Task[Payload],
) ([]providerfanout.Result[Payload], error) {
	runner.receivedTasks = append(
		[]providerfanout.Task[Payload](nil),
		tasks...,
	)

	if runner.err != nil {
		return nil,
			runner.err
	}

	return runner.results,
		nil
}

type recordingEnvelopePublisher struct {
	cycleStartedAt time.Time
	envelope       providerfanin.Envelope[Payload]
	snapshot       Snapshot
	err            error
}

func (
	publisher *recordingEnvelopePublisher,
) PublishEnvelope(
	cycleStartedAt time.Time,
	envelope providerfanin.Envelope[Payload],
) (Snapshot, error) {
	publisher.cycleStartedAt = cycleStartedAt
	publisher.envelope = envelope

	if publisher.err != nil {
		return Snapshot{},
			publisher.err
	}

	return publisher.snapshot,
		nil
}

func TestNewCycleRequiresRunner(
	t *testing.T,
) {
	_, err := NewCycle(
		CycleConfig{
			Publisher: &recordingEnvelopePublisher{},
		},
	)

	if !errors.Is(
		err,
		ErrFanOutRunnerRequired,
	) {
		t.Fatalf(
			"expected ErrFanOutRunnerRequired, got %v",
			err,
		)
	}
}

func TestNewCycleRequiresPublisher(
	t *testing.T,
) {
	_, err := NewCycle(
		CycleConfig{
			Runner: &recordingFanOutRunner{},
		},
	)

	if !errors.Is(
		err,
		ErrEnvelopePublisherRequired,
	) {
		t.Fatalf(
			"expected ErrEnvelopePublisherRequired, got %v",
			err,
		)
	}
}

func TestCycleRunsFanOutAggregatesAndPublishes(
	t *testing.T,
) {
	cycleStartedAt := time.Date(
		2026,
		time.July,
		5,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	tasks := []providerfanout.Task[Payload]{
		{
			ID:         "traffic",
			RequestKey: "regional-traffic",
		},
		{
			ID:         "weather",
			RequestKey: "current-weather",
		},
	}

	runner := &recordingFanOutRunner{
		results: []providerfanout.Result[Payload]{
			{
				TaskID:     "traffic",
				RequestKey: "regional-traffic",
				Value: NewRegionalTrafficPayload(
					nil,
				),
			},
			{
				TaskID:     "weather",
				RequestKey: "current-weather",
				Err: errors.New(
					"weather unavailable",
				),
			},
		},
	}

	expectedSnapshot := Snapshot{
		CycleStartedAt: cycleStartedAt,
		AssembledAt:    cycleStartedAt,
	}

	publisher := &recordingEnvelopePublisher{
		snapshot: expectedSnapshot,
	}

	cycle, err := NewCycle(
		CycleConfig{
			Runner:    runner,
			Publisher: publisher,
			Now: func() time.Time {
				return cycleStartedAt
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot cycle: %v",
			err,
		)
	}

	snapshot, err := cycle.Run(
		context.Background(),
		tasks,
	)
	if err != nil {
		t.Fatalf(
			"run shared snapshot cycle: %v",
			err,
		)
	}

	if len(runner.receivedTasks) != len(tasks) {
		t.Fatalf(
			"unexpected task count: got %d, want %d",
			len(runner.receivedTasks),
			len(tasks),
		)
	}

	if publisher.envelope.Status != providerfanin.BatchStatusPartial {
		t.Fatalf(
			"unexpected envelope status: %q",
			publisher.envelope.Status,
		)
	}

	if publisher.envelope.TotalCount != 2 {
		t.Fatalf(
			"unexpected total count: %d",
			publisher.envelope.TotalCount,
		)
	}

	if !snapshot.CycleStartedAt.Equal(
		expectedSnapshot.CycleStartedAt,
	) {
		t.Fatal(
			"expected published snapshot to be returned",
		)
	}
}

func TestCyclePropagatesFanOutError(
	t *testing.T,
) {
	expectedError := errors.New(
		"fan-out unavailable",
	)

	cycle, err := NewCycle(
		CycleConfig{
			Runner: &recordingFanOutRunner{
				err: expectedError,
			},
			Publisher: &recordingEnvelopePublisher{},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot cycle: %v",
			err,
		)
	}

	_, err = cycle.Run(
		context.Background(),
		nil,
	)

	if !errors.Is(
		err,
		expectedError,
	) {
		t.Fatalf(
			"expected wrapped fan-out error, got %v",
			err,
		)
	}
}

func TestCyclePropagatesPublisherError(
	t *testing.T,
) {
	expectedError := errors.New(
		"snapshot publisher unavailable",
	)

	cycleStartedAt := time.Date(
		2026,
		time.July,
		5,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	cycle, err := NewCycle(
		CycleConfig{
			Runner: &recordingFanOutRunner{
				results: []providerfanout.Result[Payload]{
					{
						TaskID: "traffic",
						Value: NewRegionalTrafficPayload(
							nil,
						),
					},
				},
			},
			Publisher: &recordingEnvelopePublisher{
				err: expectedError,
			},
			Now: func() time.Time {
				return cycleStartedAt
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot cycle: %v",
			err,
		)
	}

	_, err = cycle.Run(
		context.Background(),
		nil,
	)

	if !errors.Is(
		err,
		expectedError,
	) {
		t.Fatalf(
			"expected wrapped publisher error, got %v",
			err,
		)
	}
}
