package live

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestUpsertFlightStatesPreservesAvailabilityAndValidZeroTelemetry(t *testing.T) {
	store, err := NewStore(DefaultConfig())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	result := store.UpsertFlightStates([]flightstate.FlightState{
		{
			ICAO24:                   "abc123",
			Latitude:                 40.4,
			Longitude:                49.8,
			BarometricAltitudeM:      0,
			BarometricAltitudeStatus: flightstate.AltitudeStatusGround,
			VelocityMPS:              0,
			VelocityAvailable:        true,
			HeadingDegrees:           0,
			HeadingAvailable:         true,
			VerticalRateMPS:          0,
			VerticalRateAvailable:    true,
			OnGround:                 false,
			OnGroundAvailable:        false,
			ObservedAt:               now.Add(-time.Second),
			SourceName:               "test-provider",
		},
	}, now)
	if result.Accepted != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	snapshot, err := store.Snapshot(now, SnapshotQuery{})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	item := snapshot.Aircraft[0]
	if item.AltitudeM == nil || *item.AltitudeM != 0 {
		t.Fatalf("ground altitude zero must remain observed: %+v", item)
	}
	if item.VelocityMPS == nil || *item.VelocityMPS != 0 ||
		item.HeadingDegrees == nil || *item.HeadingDegrees != 0 ||
		item.VerticalRateMPS == nil || *item.VerticalRateMPS != 0 {
		t.Fatalf("valid zero telemetry was lost: %+v", item)
	}
	if item.OnGround != nil {
		t.Fatalf("unavailable on-ground telemetry must remain nil: %+v", item)
	}
}
