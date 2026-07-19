package postgres

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5/pgtype"
)

func telemetryFloatDatabaseValue(
	value float64,
	available bool,
) pgtype.Float8 {
	if !available {
		return pgtype.Float8{}
	}
	return pgtype.Float8{
		Float64: value,
		Valid:   true,
	}
}

func telemetryBoolDatabaseValue(
	value bool,
	available bool,
) pgtype.Bool {
	if !available {
		return pgtype.Bool{}
	}
	return pgtype.Bool{
		Bool:  value,
		Valid: true,
	}
}

func applyTelemetryDatabaseValues(
	item *flightstate.FlightState,
	velocity pgtype.Float8,
	heading pgtype.Float8,
	verticalRate pgtype.Float8,
	onGround pgtype.Bool,
) {
	item.TelemetryAvailabilityKnown = true

	item.VelocityMPS = 0
	item.VelocityAvailable = velocity.Valid
	if velocity.Valid {
		item.VelocityMPS = velocity.Float64
	}

	item.HeadingDegrees = 0
	item.HeadingAvailable = heading.Valid
	if heading.Valid {
		item.HeadingDegrees = heading.Float64
	}

	item.VerticalRateMPS = 0
	item.VerticalRateAvailable = verticalRate.Valid
	if verticalRate.Valid {
		item.VerticalRateMPS = verticalRate.Float64
	}

	item.OnGround = false
	item.OnGroundAvailable = onGround.Valid
	if onGround.Valid {
		item.OnGround = onGround.Bool
	}
}
