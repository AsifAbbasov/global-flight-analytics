package sharedsnapshot

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

type recordingSnapshotStore struct {
	publishedSnapshots []Snapshot
	publishErr         error
}

func (
	store *recordingSnapshotStore,
) Publish(
	snapshot Snapshot,
) error {
	if store.publishErr != nil {
		return store.publishErr
	}

	store.publishedSnapshots = append(
		store.publishedSnapshots,
		snapshot.Clone(),
	)

	return nil
}

func TestNewPublisherRequiresStore(
	t *testing.T,
) {
	_, err := NewPublisher(
		PublisherConfig{},
	)

	if !errors.Is(
		err,
		ErrSnapshotStoreRequired,
	) {
		t.Fatalf(
			"expected ErrSnapshotStoreRequired, got %v",
			err,
		)
	}
}

func TestPublisherPublishesEnvelopeWithInjectedAssemblyTime(
	t *testing.T,
) {
	store := &recordingSnapshotStore{}

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

	envelope := providerfanin.Envelope[Payload]{
		Status: providerfanin.BatchStatusPartial,

		TotalCount:   2,
		SuccessCount: 1,
		FailureCount: 1,

		Successes: []providerfanin.Success[Payload]{
			{
				TaskID:     TaskIDRegionalTraffic,
				RequestKey: "regional-traffic",
				Value: NewRegionalTrafficPayload(
					[]flightstate.FlightState{
						{},
					},
				),
				Shared: true,
			},
		},

		Failures: []providerfanin.Failure{
			{
				TaskID:     TaskIDCurrentWeather,
				RequestKey: "current-weather",
				Err: errors.New(
					"weather provider unavailable",
				),
			},
		},
	}

	snapshot, err := publisher.PublishEnvelope(
		cycleStartedAt,
		envelope,
	)
	if err != nil {
		t.Fatalf(
			"publish shared snapshot envelope: %v",
			err,
		)
	}

	if len(store.publishedSnapshots) != 1 {
		t.Fatalf(
			"expected one published snapshot, got %d",
			len(store.publishedSnapshots),
		)
	}

	if !snapshot.CycleStartedAt.Equal(
		cycleStartedAt,
	) {
		t.Fatalf(
			"unexpected cycle start time: got %s, want %s",
			snapshot.CycleStartedAt,
			cycleStartedAt,
		)
	}

	if !snapshot.AssembledAt.Equal(
		assembledAt,
	) {
		t.Fatalf(
			"unexpected assembled time: got %s, want %s",
			snapshot.AssembledAt,
			assembledAt,
		)
	}
}

func TestPublisherPropagatesStoreError(
	t *testing.T,
) {
	expectedError := errors.New(
		"snapshot store unavailable",
	)

	store := &recordingSnapshotStore{
		publishErr: expectedError,
	}

	cycleStartedAt := time.Now().UTC()

	publisher, err := NewPublisher(
		PublisherConfig{
			Store: store,
			Now: func() time.Time {
				return cycleStartedAt.Add(
					time.Minute,
				)
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot publisher: %v",
			err,
		)
	}

	_, err = publisher.PublishEnvelope(
		cycleStartedAt,
		providerfanin.Envelope[Payload]{},
	)

	if !errors.Is(
		err,
		expectedError,
	) {
		t.Fatalf(
			"expected wrapped snapshot store error, got %v",
			err,
		)
	}
}

func TestPublisherDoesNotWriteInvalidCycleTiming(
	t *testing.T,
) {
	store := &recordingSnapshotStore{}

	cycleStartedAt := time.Now().UTC()

	publisher, err := NewPublisher(
		PublisherConfig{
			Store: store,
			Now: func() time.Time {
				return cycleStartedAt.Add(
					-time.Minute,
				)
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot publisher: %v",
			err,
		)
	}

	_, err = publisher.PublishEnvelope(
		cycleStartedAt,
		providerfanin.Envelope[Payload]{},
	)

	if !errors.Is(
		err,
		ErrAssembledBeforeCycleStart,
	) {
		t.Fatalf(
			"expected ErrAssembledBeforeCycleStart, got %v",
			err,
		)
	}

	if len(store.publishedSnapshots) != 0 {
		t.Fatalf(
			"expected zero published snapshots, got %d",
			len(store.publishedSnapshots),
		)
	}
}
