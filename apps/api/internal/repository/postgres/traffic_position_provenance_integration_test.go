package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestTrafficRepositoryPreservesPositionAndFeedProvenance(t *testing.T) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	runID := "22222222-2222-2222-2222-222222222222"
	observedAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)

	mustExecTrafficAltitudeSQL(
		t,
		fixture.pool,
		`
			INSERT INTO ingestion_runs (id, finished_at, status, created_at)
			VALUES ($1, $2, 'success', $2);
		`,
		runID,
		observedAt,
	)

	mustExecTrafficAltitudeSQL(
		t,
		fixture.pool,
		`
			INSERT INTO flight_states (
				ingestion_run_id,
				icao24,
				callsign,
				latitude,
				longitude,
				geometric_altitude_status,
				barometric_altitude_status,
				velocity_mps,
				heading_degrees,
				on_ground,
				observed_at,
				position_source,
				source_name,
				origin_country
			)
			VALUES (
				$1,
				'4B1801',
				'AZAL101',
				40.4,
				49.8,
				'unavailable',
				'observed',
				230,
				285,
				false,
				$2,
				'mlat',
				'adsb.lol',
				'Azerbaijan'
			)
		`,
		runID,
		observedAt,
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("current traffic count = %d, want 1", len(items))
	}
	if items[0].PositionSource != flightstate.PositionSourceMLAT {
		t.Fatalf("position source = %q, want %q", items[0].PositionSource, flightstate.PositionSourceMLAT)
	}
	if items[0].SourceName != "adsb.lol" {
		t.Fatalf("source name = %q, want adsb.lol", items[0].SourceName)
	}
}
