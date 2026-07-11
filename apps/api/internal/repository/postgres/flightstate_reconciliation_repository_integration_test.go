package postgres

import (
	"context"
	"testing"
	"time"
)

func TestFlightStateRepositoryListsReconciliationScopeUsingCanonicalICAO24(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	const ingestionRunID = "11111111-1111-1111-1111-111111111111"

	observedAt := time.Date(
		2026,
		time.July,
		11,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY,
				flight_id uuid,
				aircraft_id uuid,
				ingestion_run_id uuid,
				icao24 text NOT NULL,
				callsign text,
				latitude numeric,
				longitude numeric,
				barometric_altitude_m integer,
				barometric_altitude_status text NOT NULL,
				geometric_altitude_m integer,
				geometric_altitude_status text NOT NULL,
				velocity_mps numeric,
				heading_degrees numeric,
				vertical_rate_mps numeric,
				on_ground boolean,
				origin_country text,
				observed_at timestamptz NOT NULL,
				source_name text NOT NULL
			);
		`,
	)
	if err != nil {
		t.Fatalf(
			"create reconciliation flight state table: %v",
			err,
		)
	}

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO flight_states (
				id,
				ingestion_run_id,
				icao24,
				callsign,
				latitude,
				longitude,
				barometric_altitude_m,
				barometric_altitude_status,
				geometric_altitude_m,
				geometric_altitude_status,
				velocity_mps,
				heading_degrees,
				vertical_rate_mps,
				on_ground,
				origin_country,
				observed_at,
				source_name
			)
			VALUES (
				'22222222-2222-2222-2222-222222222222',
				$1,
				'ABC123',
				'TEST123',
				40.4675,
				50.0467,
				1000,
				'observed',
				1100,
				'observed',
				200,
				90,
				0,
				false,
				'Azerbaijan',
				$2,
				'test'
			);
		`,
		ingestionRunID,
		observedAt,
	)
	if err != nil {
		t.Fatalf(
			"insert reconciliation flight state: %v",
			err,
		)
	}

	repository := NewFlightStateRepository(
		fixture.pool,
	)

	states, err := repository.ListByReconciliationScope(
		context.Background(),
		"abc123",
		ingestionRunID,
		observedAt,
		observedAt,
	)
	if err != nil {
		t.Fatalf(
			"list reconciliation flight states: %v",
			err,
		)
	}

	if len(states) != 1 {
		t.Fatalf(
			"expected one reconciliation flight state, got %d",
			len(states),
		)
	}

	if states[0].ICAO24 != "ABC123" {
		t.Fatalf(
			"expected persisted canonical icao24 ABC123, got %q",
			states[0].ICAO24,
		)
	}

	if states[0].IngestionRunID != ingestionRunID {
		t.Fatalf(
			"expected ingestion run id %s, got %s",
			ingestionRunID,
			states[0].IngestionRunID,
		)
	}
}
