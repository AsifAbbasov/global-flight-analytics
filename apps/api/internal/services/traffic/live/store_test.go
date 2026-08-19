package live

import (
	"testing"
	"time"
)

func TestStoreUsesNewestObservationAndPreservesSelectedOutsideBounds(t *testing.T) {
	store, err := NewStore(DefaultConfig())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	oldAltitude := 1000.0
	newAltitude := 2000.0

	result := store.UpsertBatch([]Aircraft{
		{
			ICAO24:     "ABC123",
			Latitude:   40.4,
			Longitude:  49.8,
			AltitudeM:  &oldAltitude,
			ObservedAt: now.Add(-20 * time.Second),
			ReceivedAt: now.Add(-19 * time.Second),
			Source:     "provider-a",
		},
		{
			ICAO24:     "abc123",
			Latitude:   40.5,
			Longitude:  49.9,
			AltitudeM:  &newAltitude,
			ObservedAt: now.Add(-10 * time.Second),
			ReceivedAt: now.Add(-9 * time.Second),
			Source:     "provider-a",
		},
		{
			ICAO24:     "def456",
			Latitude:   42.0,
			Longitude:  51.0,
			ObservedAt: now.Add(-5 * time.Second),
			ReceivedAt: now.Add(-4 * time.Second),
			Source:     "provider-b",
		},
	})
	if result.Accepted != 3 || result.Rejected != 0 {
		t.Fatalf("unexpected upsert result: %+v", result)
	}

	bounds := Bounds{
		MinLatitude:  40,
		MinLongitude: 49,
		MaxLatitude:  41,
		MaxLongitude: 50,
	}
	snapshot, err := store.Snapshot(now, SnapshotQuery{
		Bounds:         &bounds,
		SelectedICAO24: []string{"DEF456"},
		Limit:          10,
	})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.TotalActive != 2 || snapshot.Matching != 2 || len(snapshot.Aircraft) != 2 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	if snapshot.Aircraft[0].ICAO24 != "def456" {
		t.Fatalf("selected aircraft must sort first, got %q", snapshot.Aircraft[0].ICAO24)
	}
	if snapshot.Aircraft[1].AltitudeM == nil || *snapshot.Aircraft[1].AltitudeM != newAltitude {
		t.Fatalf("newest altitude was not retained: %+v", snapshot.Aircraft[1])
	}
}

func TestStoreTTLAndCapacityAreBounded(t *testing.T) {
	config := DefaultConfig()
	config.TTL = 30 * time.Second
	config.Capacity = 2
	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)

	result := store.UpsertBatch([]Aircraft{
		liveTestAircraft("000001", now.Add(-20*time.Second), now),
		liveTestAircraft("000002", now.Add(-10*time.Second), now),
		liveTestAircraft("000003", now.Add(-5*time.Second), now),
	})
	if result.Evicted != 1 {
		t.Fatalf("evicted = %d, want 1", result.Evicted)
	}

	snapshot, err := store.Snapshot(now.Add(25*time.Second), SnapshotQuery{})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snapshot.TotalActive != 1 || len(snapshot.Aircraft) != 1 {
		t.Fatalf("TTL did not prune stale state: %+v", snapshot)
	}
	if snapshot.Aircraft[0].ICAO24 != "000003" {
		t.Fatalf("unexpected surviving aircraft: %+v", snapshot.Aircraft)
	}
}

func TestStoreDeterministicallyPrefersConfiguredSourceAtEqualObservationTime(t *testing.T) {
	config := DefaultConfig()
	config.SourcePriority = map[string]int{"preferred": 20, "secondary": 10}
	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	preferred := liveTestAircraft("abc123", now, now)
	preferred.Source = "preferred"
	preferred.Latitude = 41
	secondary := liveTestAircraft("abc123", now, now.Add(time.Second))
	secondary.Source = "secondary"
	secondary.Latitude = 42

	store.UpsertBatch([]Aircraft{secondary, preferred})
	snapshot, err := store.Snapshot(now.Add(time.Second), SnapshotQuery{})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Aircraft) != 1 || snapshot.Aircraft[0].Source != "preferred" || snapshot.Aircraft[0].Latitude != 41 {
		t.Fatalf("source priority not applied: %+v", snapshot.Aircraft)
	}
}

func TestStoreRejectsInvalidQuery(t *testing.T) {
	store, err := NewStore(DefaultConfig())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	_, err = store.Snapshot(now, SnapshotQuery{
		SelectedICAO24: []string{"not-icao"},
	})
	if err == nil {
		t.Fatal("expected invalid selected ICAO24 error")
	}
}

func liveTestAircraft(icao24 string, observedAt, receivedAt time.Time) Aircraft {
	return Aircraft{
		ICAO24:     icao24,
		Latitude:   40.4,
		Longitude:  49.8,
		ObservedAt: observedAt,
		ReceivedAt: receivedAt,
		Source:     "test",
	}
}
