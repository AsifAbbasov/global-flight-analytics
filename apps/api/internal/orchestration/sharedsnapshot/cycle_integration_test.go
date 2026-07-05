package sharedsnapshot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
)

type integrationFanOutRunner struct {
	results []providerfanout.Result
}

func (runner *integrationFanOutRunner) Run(
	ctx context.Context,
	tasks []providerfanout.Task,
) ([]providerfanout.Result, error) {
	return runner.results, nil
}

func TestCyclePublishesAggregatedEnvelopeIntoStore(
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

	assembledAt := time.Date(
		2026,
		time.July,
		5,
		12,
		1,
		0,
		0,
		time.UTC,
	)

	weatherError := errors.New(
		"weather provider unavailable",
	)

	runner := &integrationFanOutRunner{
		results: []providerfanout.Result{
			{
				TaskID:     "traffic",
				RequestKey: "regional-traffic",
				Value:      "traffic-value",
				Shared:     false,
			},
			{
				TaskID:     "weather",
				RequestKey: "current-weather",
				Err:        weatherError,
			},
		},
	}

	store := NewStore()

	publisher, err := NewPublisher(
		PublisherConfig{
			Store: store,
			Now: func() time.Time {
				return assembledAt
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot publisher: %v",
			err,
		)
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

	publishedSnapshot, err := cycle.Run(
		context.Background(),
		nil,
	)
	if err != nil {
		t.Fatalf(
			"run shared snapshot cycle: %v",
			err,
		)
	}

	if publishedSnapshot.Status != providerfanin.BatchStatusPartial {
		t.Fatalf(
			"unexpected published snapshot status: %q",
			publishedSnapshot.Status,
		)
	}

	if publishedSnapshot.TotalCount != 2 {
		t.Fatalf(
			"unexpected published total count: %d",
			publishedSnapshot.TotalCount,
		)
	}

	if publishedSnapshot.SuccessCount != 1 {
		t.Fatalf(
			"unexpected published success count: %d",
			publishedSnapshot.SuccessCount,
		)
	}

	if publishedSnapshot.FailureCount != 1 {
		t.Fatalf(
			"unexpected published failure count: %d",
			publishedSnapshot.FailureCount,
		)
	}

	if !publishedSnapshot.CycleStartedAt.Equal(
		cycleStartedAt,
	) {
		t.Fatalf(
			"unexpected cycle start time: got %s, want %s",
			publishedSnapshot.CycleStartedAt,
			cycleStartedAt,
		)
	}

	if !publishedSnapshot.AssembledAt.Equal(
		assembledAt,
	) {
		t.Fatalf(
			"unexpected assembled time: got %s, want %s",
			publishedSnapshot.AssembledAt,
			assembledAt,
		)
	}

	currentSnapshot, exists := store.Current()
	if !exists {
		t.Fatal(
			"expected current shared snapshot in store",
		)
	}

	if currentSnapshot.Status != providerfanin.BatchStatusPartial {
		t.Fatalf(
			"unexpected stored snapshot status: %q",
			currentSnapshot.Status,
		)
	}

	if len(currentSnapshot.Successes) != 1 {
		t.Fatalf(
			"unexpected stored success count: %d",
			len(currentSnapshot.Successes),
		)
	}

	if len(currentSnapshot.Failures) != 1 {
		t.Fatalf(
			"unexpected stored failure count: %d",
			len(currentSnapshot.Failures),
		)
	}

	if currentSnapshot.Successes[0].TaskID != "traffic" {
		t.Fatalf(
			"unexpected stored success task identifier: %q",
			currentSnapshot.Successes[0].TaskID,
		)
	}

	if currentSnapshot.Failures[0].TaskID != "weather" {
		t.Fatalf(
			"unexpected stored failure task identifier: %q",
			currentSnapshot.Failures[0].TaskID,
		)
	}

	if !errors.Is(
		currentSnapshot.Failures[0].Err,
		weatherError,
	) {
		t.Fatalf(
			"unexpected stored provider error: %v",
			currentSnapshot.Failures[0].Err,
		)
	}
}
