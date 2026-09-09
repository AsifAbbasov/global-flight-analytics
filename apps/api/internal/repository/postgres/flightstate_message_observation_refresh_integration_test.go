package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestFlightStateDuplicatePositionRefreshesOnlyNewerMessageObservation(t *testing.T) {
	fixture := newReconciliationFixture(t)
	defer fixture.close(t)
	ctx := context.Background()

	_, err := fixture.pool.Exec(ctx, `
		CREATE TABLE flight_states (
			id bigserial PRIMARY KEY,
			flight_id uuid,
			aircraft_id uuid,
			icao24 text NOT NULL,
			callsign text,
			latitude numeric NOT NULL,
			longitude numeric NOT NULL,
			barometric_altitude_m integer,
			barometric_altitude_status text NOT NULL,
			geometric_altitude_m integer,
			geometric_altitude_status text NOT NULL,
			velocity_mps numeric,
			heading_degrees numeric,
			vertical_rate_mps numeric,
			on_ground boolean,
			origin_country text,
			squawk_code text,
			special_purpose_indicator boolean NOT NULL,
			position_source text NOT NULL,
			aircraft_category smallint,
			aircraft_category_available boolean NOT NULL,
			observed_at timestamptz NOT NULL,
			message_observed_at timestamptz,
			source_name text NOT NULL,
			ingestion_run_id uuid,
			UNIQUE (source_name, icao24, observed_at)
		);
	`)
	if err != nil {
		t.Fatalf("create flight_states fixture: %v", err)
	}

	positionAt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	firstMessageAt := positionAt.Add(2 * time.Second)
	newerMessageAt := positionAt.Add(11 * time.Second)
	olderMessageAt := positionAt.Add(7 * time.Second)

	state := flightstate.FlightState{
		ICAO24:                     "ABC123",
		Latitude:                   40.4093,
		Longitude:                  49.8671,
		BarometricAltitudeStatus:   flightstate.AltitudeStatusUnavailable,
		GeometricAltitudeStatus:    flightstate.AltitudeStatusUnavailable,
		PositionSource:             flightstate.PositionSourceADSB,
		ObservedAt:                 positionAt,
		MessageObservedAt:          &firstMessageAt,
		SourceName:                 "test-readsb",
		TelemetryAvailabilityKnown: true,
	}

	if inserted := saveFlightStateBatchForTest(t, ctx, fixture, state); inserted != 1 {
		t.Fatalf("first inserted count = %d, want 1", inserted)
	}
	assertPersistedMessageObservation(t, ctx, fixture, firstMessageAt, 1)

	state.MessageObservedAt = &newerMessageAt
	if inserted := saveFlightStateBatchForTest(t, ctx, fixture, state); inserted != 0 {
		t.Fatalf("newer duplicate inserted count = %d, want 0", inserted)
	}
	assertPersistedMessageObservation(t, ctx, fixture, newerMessageAt, 1)

	state.MessageObservedAt = &olderMessageAt
	if inserted := saveFlightStateBatchForTest(t, ctx, fixture, state); inserted != 0 {
		t.Fatalf("older duplicate inserted count = %d, want 0", inserted)
	}
	assertPersistedMessageObservation(t, ctx, fixture, newerMessageAt, 1)

	state.MessageObservedAt = nil
	if inserted := saveFlightStateBatchForTest(t, ctx, fixture, state); inserted != 0 {
		t.Fatalf("nil duplicate inserted count = %d, want 0", inserted)
	}
	assertPersistedMessageObservation(t, ctx, fixture, newerMessageAt, 1)
}

func saveFlightStateBatchForTest(
	t *testing.T,
	ctx context.Context,
	fixture *reconciliationFixture,
	state flightstate.FlightState,
) int {
	t.Helper()
	tx, err := fixture.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin flight state transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	inserted, err := saveFlightStateBatch(ctx, tx, []flightstate.FlightState{state})
	if err != nil {
		t.Fatalf("save flight state batch: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit flight state transaction: %v", err)
	}
	return inserted
}

func assertPersistedMessageObservation(
	t *testing.T,
	ctx context.Context,
	fixture *reconciliationFixture,
	want time.Time,
	wantRows int,
) {
	t.Helper()
	var got time.Time
	var rows int
	err := fixture.pool.QueryRow(ctx, `
		SELECT message_observed_at, COUNT(*) OVER ()
		FROM flight_states
		WHERE source_name = 'test-readsb'
			AND icao24 = 'ABC123';
	`).Scan(&got, &rows)
	if err != nil {
		t.Fatalf("load persisted message observation: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("message_observed_at = %s, want %s", got, want)
	}
	if rows != wantRows {
		t.Fatalf("flight state row count = %d, want %d", rows, wantRows)
	}
}
