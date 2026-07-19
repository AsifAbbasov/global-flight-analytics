package postgres

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestTelemetryDatabaseValuePreservesUnavailableAndZero(
	t *testing.T,
) {
	unavailable := telemetryFloatDatabaseValue(
		0,
		false,
	)
	if unavailable.Valid {
		t.Fatal(
			"unavailable telemetry must become PostgreSQL NULL",
		)
	}

	zero := telemetryFloatDatabaseValue(
		0,
		true,
	)
	if !zero.Valid || zero.Float64 != 0 {
		t.Fatalf(
			"available zero telemetry was not preserved: %#v",
			zero,
		)
	}
}

func TestApplyTelemetryDatabaseValuesRestoresAvailability(
	t *testing.T,
) {
	item := flightstate.FlightState{}

	applyTelemetryDatabaseValues(
		&item,
		pgtype.Float8{},
		pgtype.Float8{
			Float64: 0,
			Valid:   true,
		},
		pgtype.Float8{
			Float64: -2.5,
			Valid:   true,
		},
		pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	)

	if !item.TelemetryAvailabilityKnown {
		t.Fatal(
			"database read did not mark telemetry availability as known",
		)
	}
	if item.HasVelocity() {
		t.Fatal(
			"PostgreSQL NULL velocity was restored as available",
		)
	}
	if !item.HasHeading() ||
		item.HeadingDegrees != 0 {
		t.Fatal(
			"available zero heading was not restored",
		)
	}
	if !item.HasVerticalRate() ||
		item.VerticalRateMPS != -2.5 {
		t.Fatal(
			"available vertical rate was not restored",
		)
	}
	if !item.HasOnGroundState() ||
		item.OnGround {
		t.Fatal(
			"available on_ground=false was not restored",
		)
	}
}
