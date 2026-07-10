package sharedsnapshot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
)

type integrationFanOutRunner struct {
	results []providerfanout.Result[Payload]
}

func (
	runner *integrationFanOutRunner,
) Run(
	_ context.Context,
	_ []providerfanout.Task[Payload],
) ([]providerfanout.Result[Payload], error) {
	return runner.results,
		nil
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

	assembledAt := cycleStartedAt.Add(
		time.Minute,
	)

	weatherError := errors.New(
		"weather provider unavailable",
	)

	runner := &integrationFanOutRunner{
		results: []providerfanout.Result[Payload]{
			{
				TaskID:     TaskIDRegionalTraffic,
				RequestKey: "regional-traffic",
				Value: NewRegionalTrafficPayload(
					[]flightstate.FlightState{
						{},
					},
				),
			},
			{
				TaskID:     TaskIDCurrentWeather,
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

	currentSnapshot, exists := store.Current()
	if !exists {
		t.Fatal(
			"expected current shared snapshot in store",
		)
	}

	if len(currentSnapshot.Successes) != 1 {
		t.Fatalf(
			"unexpected stored success count: %d",
			len(currentSnapshot.Successes),
		)
	}

	trafficPayload, ok := currentSnapshot.Successes[0].Payload.RegionalTraffic()
	if !ok {
		t.Fatalf(
			"expected stored regional traffic payload, got kind %q",
			currentSnapshot.Successes[0].Payload.Kind(),
		)
	}

	if len(trafficPayload.States) != 1 {
		t.Fatalf(
			"unexpected stored traffic state count: got %d, want 1",
			len(trafficPayload.States),
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
