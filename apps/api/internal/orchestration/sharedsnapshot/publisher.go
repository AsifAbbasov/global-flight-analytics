package sharedsnapshot

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

var ErrSnapshotStoreRequired = errors.New(
	"shared snapshot store is required",
)

type SnapshotStore interface {
	Publish(
		snapshot Snapshot,
	) error
}

type PublisherConfig struct {
	Store SnapshotStore
	Now   func() time.Time
}

type Publisher struct {
	store SnapshotStore
	now   func() time.Time
}

func NewPublisher(
	config PublisherConfig,
) (*Publisher, error) {
	if config.Store == nil {
		return nil,
			ErrSnapshotStoreRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Publisher{
		store: config.Store,
		now:   now,
	}, nil
}

func (
	publisher *Publisher,
) PublishEnvelope(
	cycleStartedAt time.Time,
	envelope providerfanin.Envelope[Payload],
) (Snapshot, error) {
	assembledAt := publisher.now()

	snapshot, err := FromEnvelopeForCycle(
		cycleStartedAt,
		assembledAt,
		envelope,
	)
	if err != nil {
		return Snapshot{},
			fmt.Errorf(
				"assemble shared snapshot: %w",
				err,
			)
	}

	if err := publisher.store.Publish(
		snapshot,
	); err != nil {
		return Snapshot{},
			fmt.Errorf(
				"publish shared snapshot: %w",
				err,
			)
	}

	return snapshot.Clone(),
		nil
}
