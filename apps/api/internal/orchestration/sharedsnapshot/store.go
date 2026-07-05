package sharedsnapshot

import (
	"errors"
	"sync"
)

var ErrSnapshotOlderThanCurrent = errors.New(
	"shared snapshot collection cycle is older than current snapshot",
)

type Store struct {
	mutex sync.RWMutex

	current *Snapshot
}

func NewStore() *Store {
	return &Store{}
}

func (store *Store) Publish(
	snapshot Snapshot,
) error {
	if snapshot.AssembledAt.IsZero() {
		return ErrAssembledAtRequired
	}

	if !snapshot.CycleStartedAt.IsZero() &&
		snapshot.AssembledAt.Before(
			snapshot.CycleStartedAt,
		) {
		return ErrAssembledBeforeCycleStart
	}

	candidate := snapshot.Clone()
	candidateOrderTime := snapshotOrderTime(
		candidate,
	)

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.current != nil {
		currentOrderTime := snapshotOrderTime(
			*store.current,
		)

		if candidateOrderTime.Before(
			currentOrderTime,
		) {
			return ErrSnapshotOlderThanCurrent
		}
	}

	store.current = &candidate

	return nil
}

func (store *Store) Current() (
	Snapshot,
	bool,
) {
	if store == nil {
		return Snapshot{}, false
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if store.current == nil {
		return Snapshot{}, false
	}

	return store.current.Clone(), true
}
