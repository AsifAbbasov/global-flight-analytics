package readsbcompat

import (
	"testing"
	"time"
)

func TestMapAircraftSeparatesPositionAndMessageObservationTimes(t *testing.T) {
	snapshotAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	item := AircraftItem{Hex: "abc123", Latitude: 40.4093, Longitude: 49.8671, Seen: OptionalFloat64{Value: 0.5, Available: true}, SeenPos: OptionalFloat64{Value: 18, Available: true}}
	mapped := MapAircraft("test-readsb", item, snapshotAt)
	expectedPosition := snapshotAt.Add(-18 * time.Second)
	if !mapped.ObservedAt.Equal(expectedPosition) {
		t.Fatalf("position observed_at = %s, want %s", mapped.ObservedAt, expectedPosition)
	}
	if mapped.MessageObservedAt == nil {
		t.Fatal("expected message observation time")
	}
	expectedMessage := snapshotAt.Add(-500 * time.Millisecond)
	if !mapped.MessageObservedAt.Equal(expectedMessage) {
		t.Fatalf("message observed_at = %s, want %s", mapped.MessageObservedAt, expectedMessage)
	}
}
func TestMapAircraftFallsBackToMessageAgeWhenSeenPosUnavailable(t *testing.T) {
	snapshotAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	item := AircraftItem{Hex: "abc123", Latitude: 40.4093, Longitude: 49.8671, Seen: OptionalFloat64{Value: 2, Available: true}}
	mapped := MapAircraft("test-readsb", item, snapshotAt)
	expected := snapshotAt.Add(-2 * time.Second)
	if !mapped.ObservedAt.Equal(expected) {
		t.Fatalf("fallback observed_at = %s, want %s", mapped.ObservedAt, expected)
	}
	if mapped.MessageObservedAt == nil || !mapped.MessageObservedAt.Equal(expected) {
		t.Fatalf("message observation fallback = %#v, want %s", mapped.MessageObservedAt, expected)
	}
}
func TestMapAircraftKeepsMessageTimeUnavailableWhenSeenMissing(t *testing.T) {
	snapshotAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	item := AircraftItem{Hex: "abc123", Latitude: 40.4093, Longitude: 49.8671, SeenPos: OptionalFloat64{Value: 3, Available: true}}
	mapped := MapAircraft("test-readsb", item, snapshotAt)
	if mapped.MessageObservedAt != nil {
		t.Fatalf("unexpected message observation time: %v", mapped.MessageObservedAt)
	}
	expectedPosition := snapshotAt.Add(-3 * time.Second)
	if !mapped.ObservedAt.Equal(expectedPosition) {
		t.Fatalf("position observed_at = %s, want %s", mapped.ObservedAt, expectedPosition)
	}
}
