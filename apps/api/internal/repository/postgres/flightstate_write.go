package postgres

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5"
)

const insertFlightStateQuery = `
	INSERT INTO flight_states (
		flight_id,
		aircraft_id,
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
		squawk_code,
		special_purpose_indicator,
		position_source,
		aircraft_category,
		aircraft_category_available,
		observed_at,
		source_name,
		ingestion_run_id
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11,
		$12,
		$13,
		$14,
		$15,
		$16,
		$17,
		$18,
		$19,
		$20,
		$21,
		$22,
		$23
	);
`

func saveFlightStateBatch(
	ctx context.Context,
	tx pgx.Tx,
	items []flightstate.FlightState,
) error {
	for index, item := range items {
		arguments, err := prepareFlightStateInsertArguments(
			index,
			item,
		)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			insertFlightStateQuery,
			arguments...,
		); err != nil {
			return fmt.Errorf(
				"insert flight state at index %d for icao24 %s: %w",
				index,
				item.ICAO24,
				err,
			)
		}
	}

	return nil
}

func prepareFlightStateInsertArguments(
	index int,
	item flightstate.FlightState,
) ([]any, error) {
	barometricAltitude, barometricStatus, err := altitudeDatabaseValue(
		item.BarometricAltitudeM,
		item.BarometricAltitudeStatus,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"prepare barometric altitude at index %d for icao24 %s: %w",
			index,
			item.ICAO24,
			err,
		)
	}

	geometricAltitude, geometricStatus, err := altitudeDatabaseValue(
		item.GeometricAltitudeM,
		item.GeometricAltitudeStatus,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"prepare geometric altitude at index %d for icao24 %s: %w",
			index,
			item.ICAO24,
			err,
		)
	}

	squawkCode, err := flightstate.NormalizeSquawkCode(item.SquawkCode)
	if err != nil {
		return nil, fmt.Errorf(
			"prepare squawk code at index %d for icao24 %s: %w",
			index,
			item.ICAO24,
			err,
		)
	}

	positionSource, err := flightstate.NormalizePositionSource(
		item.PositionSource,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"prepare position source at index %d for icao24 %s: %w",
			index,
			item.ICAO24,
			err,
		)
	}

	if err := flightstate.ValidateAircraftCategory(
		item.AircraftCategory,
		item.AircraftCategoryAvailable,
	); err != nil {
		return nil, fmt.Errorf(
			"prepare aircraft category at index %d for icao24 %s: %w",
			index,
			item.ICAO24,
			err,
		)
	}

	var aircraftCategory any
	if item.AircraftCategoryAvailable {
		aircraftCategory = item.AircraftCategory
	}

	return []any{
		nullableUUID(item.FlightID),
		nullableUUID(item.AircraftID),
		item.ICAO24,
		nullableText(item.Callsign),
		item.Latitude,
		item.Longitude,
		barometricAltitude,
		barometricStatus,
		geometricAltitude,
		geometricStatus,
		telemetryFloatDatabaseValue(
			item.VelocityMPS,
			item.HasVelocity(),
		),
		telemetryFloatDatabaseValue(
			item.HeadingDegrees,
			item.HasHeading(),
		),
		telemetryFloatDatabaseValue(
			item.VerticalRateMPS,
			item.HasVerticalRate(),
		),
		telemetryBoolDatabaseValue(
			item.OnGround,
			item.HasOnGroundState(),
		),
		nullableText(item.OriginCountry),
		squawkCode,
		item.SpecialPurposeIndicator,
		string(positionSource),
		aircraftCategory,
		item.AircraftCategoryAvailable,
		item.ObservedAt,
		requiredSourceNameValue(item.SourceName),
		nullableUUID(item.IngestionRunID),
	}, nil
}
