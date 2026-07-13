package snapshot

import (
	"sync"
	"time"
)

type Snapshot struct {
	Time time.Time

	ActiveAircraft       int
	AreaSquareKilometers float64
	ObservedSamples      int
	ExpectedSamples      int
}

type Store struct {
	mutex    sync.RWMutex
	snapshot Snapshot
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Save(snapshot Snapshot) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.snapshot = snapshot
}

func (s *Store) Load() Snapshot {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.snapshot
}
