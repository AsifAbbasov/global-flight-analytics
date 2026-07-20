package postgres

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestFlightStateRepositoryPersistsExplicitAltitudeIntegerPolicy(
	t *testing.T,
) {
	fixture := newFlightStateAltitudeIntegrationFixture(
		t,
		false,
	)

	flightID := "77777777-7777-7777-7777-777777777777"
	observedAt := time.Date(
		2026,
		time.July,
		20,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	state := makeAltitudePersistenceState(
		flightID,
		"RND001",
		observedAt,
		9753.5,
		flightstate.AltitudeStatusObserved,
		-12.5,
		flightstate.AltitudeStatusObserved,
		false,
	)

	if err := fixture.repository.SaveFlightStates(
		context.Background(),
		[]flightstate.FlightState{state},
	); err != nil {
		t.Fatalf("save explicitly rounded altitude state: %v", err)
	}

	var barometric int32
	var geometric int32
	if err := fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT
				barometric_altitude_m,
				geometric_altitude_m
			FROM flight_states
			WHERE icao24 = $1
			  AND observed_at = $2
		`,
		"RND001",
		observedAt,
	).Scan(
		&barometric,
		&geometric,
	); err != nil {
		t.Fatalf("read persisted altitude integers: %v", err)
	}

	if barometric != 9754 || geometric != -13 {
		t.Fatalf(
			"persisted altitude integers = (%d, %d), want (9754, -13)",
			barometric,
			geometric,
		)
	}
}

func TestFlightStateRepositoryRejectsNonFiniteObservedAltitude(
	t *testing.T,
) {
	fixture := newFlightStateAltitudeIntegrationFixture(
		t,
		false,
	)

	flightID := "88888888-8888-8888-8888-888888888888"
	state := makeAltitudePersistenceState(
		flightID,
		"RND002",
		time.Date(
			2026,
			time.July,
			20,
			10,
			1,
			0,
			0,
			time.UTC,
		),
		math.NaN(),
		flightstate.AltitudeStatusObserved,
		1000,
		flightstate.AltitudeStatusObserved,
		false,
	)

	err := fixture.repository.SaveFlightStates(
		context.Background(),
		[]flightstate.FlightState{state},
	)
	if !errors.Is(err, ErrAltitudeMetersNotFinite) {
		t.Fatalf(
			"expected ErrAltitudeMetersNotFinite, got %v",
			err,
		)
	}

	var count int
	if err := fixture.pool.QueryRow(
		context.Background(),
		"SELECT count(*) FROM flight_states WHERE icao24 = $1",
		"RND002",
	).Scan(&count); err != nil {
		t.Fatalf("count rejected altitude rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rejected batch rollback, found %d rows", count)
	}
}
