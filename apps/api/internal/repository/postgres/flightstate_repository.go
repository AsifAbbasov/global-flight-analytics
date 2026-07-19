package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrFlightStateRepositoryPoolRequired = errors.New(
	"flight state repository pool is required",
)

type FlightStateRepository struct {
	db *pgxpool.Pool
}

func NewFlightStateRepository(
	db *pgxpool.Pool,
) *FlightStateRepository {
	return &FlightStateRepository{
		db: db,
	}
}

func (r *FlightStateRepository) SaveFlightStates(
	ctx context.Context,
	items []flightstate.FlightState,
) error {
	if len(items) == 0 {
		return nil
	}

	if r == nil || r.db == nil {
		return ErrFlightStateRepositoryPoolRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := r.db.BeginTx(
		ctx,
		pgx.TxOptions{},
	)
	if err != nil {
		return fmt.Errorf(
			"begin flight states transaction: %w",
			err,
		)
	}

	committed := false

	defer func() {
		if !committed {
			_ = tx.Rollback(
				ctx,
			)
		}
	}()

	const query = `
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
			CAST($7::double precision AS integer),
			$8,
			CAST($9::double precision AS integer),
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

	for index, item := range items {
		barometricAltitude, barometricStatus, err :=
			altitudeDatabaseValue(
				item.BarometricAltitudeM,
				item.BarometricAltitudeStatus,
			)
		if err != nil {
			return fmt.Errorf(
				"prepare barometric altitude at index %d for icao24 %s: %w",
				index,
				item.ICAO24,
				err,
			)
		}

		geometricAltitude, geometricStatus, err :=
			altitudeDatabaseValue(
				item.GeometricAltitudeM,
				item.GeometricAltitudeStatus,
			)
		if err != nil {
			return fmt.Errorf(
				"prepare geometric altitude at index %d for icao24 %s: %w",
				index,
				item.ICAO24,
				err,
			)
		}

		squawkCode, err := flightstate.NormalizeSquawkCode(
			item.SquawkCode,
		)
		if err != nil {
			return fmt.Errorf(
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
			return fmt.Errorf(
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
			return fmt.Errorf(
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

		_, err = tx.Exec(
			ctx,
			query,
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
			sourceNameOrUnknown(item.SourceName),
			nullableUUID(item.IngestionRunID),
		)
		if err != nil {
			return fmt.Errorf(
				"insert flight state at index %d for icao24 %s: %w",
				index,
				item.ICAO24,
				err,
			)
		}
	}

	if err := tx.Commit(
		ctx,
	); err != nil {
		return fmt.Errorf(
			"commit flight states transaction: %w",
			err,
		)
	}

	committed = true

	return nil
}

func (r *FlightStateRepository) ListByFlightID(
	ctx context.Context,
	flightID string,
) ([]flightstate.FlightState, error) {
	const query = `
		SELECT
			id::text,
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			COALESCE(ingestion_run_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			latitude::double precision,
			longitude::double precision,
			barometric_altitude_m::double precision,
			barometric_altitude_status,
			geometric_altitude_m::double precision,
			geometric_altitude_status,
			velocity_mps::double precision,
			heading_degrees::double precision,
			vertical_rate_mps::double precision,
			on_ground,
			COALESCE(origin_country, ''),
			COALESCE(squawk_code, ''),
			COALESCE(special_purpose_indicator, false),
			COALESCE(position_source, ''),
			aircraft_category,
			COALESCE(aircraft_category_available, false),
			observed_at,
			source_name
		FROM flight_states
		WHERE flight_id = $1
		  AND latitude IS NOT NULL
		  AND longitude IS NOT NULL
		ORDER BY observed_at ASC;
	`

	rows, err := r.db.Query(
		ctx,
		query,
		flightID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(
		[]flightstate.FlightState,
		0,
	)

	for rows.Next() {
		var item flightstate.FlightState
		var barometricAltitude pgtype.Float8
		var geometricAltitude pgtype.Float8
		var velocity pgtype.Float8
		var heading pgtype.Float8
		var verticalRate pgtype.Float8
		var onGround pgtype.Bool
		var barometricStatus string
		var geometricStatus string
		var positionSource string
		var aircraftCategory pgtype.Int2

		if err := rows.Scan(
			&item.ID,
			&item.FlightID,
			&item.AircraftID,
			&item.IngestionRunID,
			&item.ICAO24,
			&item.Callsign,
			&item.Latitude,
			&item.Longitude,
			&barometricAltitude,
			&barometricStatus,
			&geometricAltitude,
			&geometricStatus,
			&velocity,
			&heading,
			&verticalRate,
			&onGround,
			&item.OriginCountry,
			&item.SquawkCode,
			&item.SpecialPurposeIndicator,
			&positionSource,
			&aircraftCategory,
			&item.AircraftCategoryAvailable,
			&item.ObservedAt,
			&item.SourceName,
		); err != nil {
			return nil, err
		}

		item.PositionSource = flightstate.PositionSource(
			positionSource,
		)
		item.AircraftCategory = 0
		if aircraftCategory.Valid {
			item.AircraftCategory = int(
				aircraftCategory.Int16,
			)
		}

		applyAltitudeDatabaseValues(
			&item,
			barometricAltitude,
			barometricStatus,
			geometricAltitude,
			geometricStatus,
		)
		applyTelemetryDatabaseValues(
			&item,
			velocity,
			heading,
			verticalRate,
			onGround,
		)

		items = append(
			items,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *FlightStateRepository) GetLatestByICAO24(
	ctx context.Context,
	icao24 string,
) (flightstate.FlightState, error) {
	const query = `
		SELECT
			id::text,
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			COALESCE(ingestion_run_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			latitude::double precision,
			longitude::double precision,
			barometric_altitude_m::double precision,
			barometric_altitude_status,
			geometric_altitude_m::double precision,
			geometric_altitude_status,
			velocity_mps::double precision,
			heading_degrees::double precision,
			vertical_rate_mps::double precision,
			on_ground,
			COALESCE(origin_country, ''),
			COALESCE(squawk_code, ''),
			COALESCE(special_purpose_indicator, false),
			COALESCE(position_source, ''),
			aircraft_category,
			COALESCE(aircraft_category_available, false),
			observed_at,
			source_name
		FROM flight_states
		WHERE icao24 = $1
		  AND latitude IS NOT NULL
		  AND longitude IS NOT NULL
		ORDER BY observed_at DESC
		LIMIT 1;
	`

	var item flightstate.FlightState
	var barometricAltitude pgtype.Float8
	var geometricAltitude pgtype.Float8
	var velocity pgtype.Float8
	var heading pgtype.Float8
	var verticalRate pgtype.Float8
	var onGround pgtype.Bool
	var barometricStatus string
	var geometricStatus string
	var positionSource string
	var aircraftCategory pgtype.Int2

	err := r.db.QueryRow(
		ctx,
		query,
		icao24,
	).Scan(
		&item.ID,
		&item.FlightID,
		&item.AircraftID,
		&item.IngestionRunID,
		&item.ICAO24,
		&item.Callsign,
		&item.Latitude,
		&item.Longitude,
		&barometricAltitude,
		&barometricStatus,
		&geometricAltitude,
		&geometricStatus,
		&velocity,
		&heading,
		&verticalRate,
		&onGround,
		&item.OriginCountry,
		&item.SquawkCode,
		&item.SpecialPurposeIndicator,
		&positionSource,
		&aircraftCategory,
		&item.AircraftCategoryAvailable,
		&item.ObservedAt,
		&item.SourceName,
	)
	if err != nil {
		if errors.Is(
			err,
			pgx.ErrNoRows,
		) {
			return flightstate.FlightState{},
				flightstate.ErrNotFound
		}

		return flightstate.FlightState{}, err
	}

	item.PositionSource = flightstate.PositionSource(
		positionSource,
	)
	item.AircraftCategory = 0
	if aircraftCategory.Valid {
		item.AircraftCategory = int(
			aircraftCategory.Int16,
		)
	}

	applyAltitudeDatabaseValues(
		&item,
		barometricAltitude,
		barometricStatus,
		geometricAltitude,
		geometricStatus,
	)
	applyTelemetryDatabaseValues(
		&item,
		velocity,
		heading,
		verticalRate,
		onGround,
	)

	return item, nil
}

func altitudeDatabaseValue(
	value float64,
	status flightstate.AltitudeStatus,
) (
	pgtype.Float8,
	string,
	error,
) {
	effectiveStatus := flightstate.ResolveAltitudeStatus(
		value,
		status,
	)

	if !flightstate.IsKnownAltitudeStatus(
		effectiveStatus,
	) {
		return pgtype.Float8{},
			"",
			fmt.Errorf(
				"unsupported altitude status %q",
				effectiveStatus,
			)
	}

	switch effectiveStatus {
	case flightstate.AltitudeStatusObserved:
		return pgtype.Float8{
				Float64: value,
				Valid:   true,
			},
			string(effectiveStatus),
			nil

	case flightstate.AltitudeStatusGround:
		return pgtype.Float8{
				Float64: 0,
				Valid:   true,
			},
			string(effectiveStatus),
			nil

	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable,
		flightstate.AltitudeStatusInvalid:
		return pgtype.Float8{
				Valid: false,
			},
			string(effectiveStatus),
			nil

	default:
		return pgtype.Float8{},
			"",
			fmt.Errorf(
				"unsupported altitude status %q",
				effectiveStatus,
			)
	}
}

func applyAltitudeDatabaseValues(
	item *flightstate.FlightState,
	barometricAltitude pgtype.Float8,
	barometricStatus string,
	geometricAltitude pgtype.Float8,
	geometricStatus string,
) {
	item.BarometricAltitudeM = 0
	if barometricAltitude.Valid {
		item.BarometricAltitudeM = barometricAltitude.Float64
	}

	item.BarometricAltitudeStatus = flightstate.AltitudeStatus(
		barometricStatus,
	)

	item.GeometricAltitudeM = 0
	if geometricAltitude.Valid {
		item.GeometricAltitudeM = geometricAltitude.Float64
	}

	item.GeometricAltitudeStatus = flightstate.AltitudeStatus(
		geometricStatus,
	)
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2
