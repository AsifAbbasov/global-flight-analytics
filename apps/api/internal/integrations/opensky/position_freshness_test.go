package opensky

import (
	"testing"
	"time"
)

func TestMapStateVectorPreservesPositionAndLastContactTimesSeparately(t *testing.T) {
	positionAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	contactAt := positionAt.Add(7 * time.Second)
	snapshotAt := contactAt.Add(time.Second)
	latitude, longitude := 40.4093, 49.8671
	mapped, usable, err := MapStateVector(StateVector{ICAO24: "abc123", Latitude: &latitude, Longitude: &longitude, TimePosition: &positionAt, LastContact: contactAt, SnapshotTime: snapshotAt})
	if err != nil {
		t.Fatalf("map state vector: %v", err)
	}
	if !usable {
		t.Fatal("expected state vector to be usable")
	}
	if !mapped.ObservedAt.Equal(positionAt) {
		t.Fatalf("position observed_at = %s, want %s", mapped.ObservedAt, positionAt)
	}
	if mapped.MessageObservedAt == nil || !mapped.MessageObservedAt.Equal(contactAt) {
		t.Fatalf("message observed_at = %#v, want %s", mapped.MessageObservedAt, contactAt)
	}
}
