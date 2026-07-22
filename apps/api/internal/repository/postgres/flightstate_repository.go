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

	if err := requireRepositoryContext(ctx); err != nil {
		return err
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
			rollbackRepositoryTransaction(tx)
		}
	}()

	if err := saveFlightStateBatch(ctx, tx, items); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
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
	pgtype.Int4,
	string,
	error,
) {
	altitude, err := flightstate.NewAltitude(
		value,
		status,
	)
	if err != nil {
		switch {
		case status == flightstate.AltitudeStatusObserved:
			_, conversionErr := altitudeMetersToPostgresInteger(value)
			if conversionErr != nil {
				return pgtype.Int4{}, "", conversionErr
			}
			return pgtype.Int4{}, "", err

		case errors.Is(err, flightstate.ErrAltitudeStatusInvalid):
			return pgtype.Int4{},
				"",
				fmt.Errorf(
					"unsupported altitude status %q: %w",
					status,
					err,
				)

		case status == flightstate.AltitudeStatusInvalid,
			status == "":
			return pgtype.Int4{
					Valid: false,
				},
				string(flightstate.AltitudeStatusInvalid),
				nil

		default:
			return pgtype.Int4{}, "", err
		}
	}

	effectiveStatus := altitude.Status()

	switch effectiveStatus {
	case flightstate.AltitudeStatusObserved:
		integerValue, err := altitudeMetersToPostgresInteger(
			altitude.Meters(),
		)
		if err != nil {
			return pgtype.Int4{}, "", err
		}

		return pgtype.Int4{
				Int32: integerValue,
				Valid: true,
			},
			string(effectiveStatus),
			nil

	case flightstate.AltitudeStatusGround:
		return pgtype.Int4{
				Int32: 0,
				Valid: true,
			},
			string(effectiveStatus),
			nil

	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable,
		flightstate.AltitudeStatusInvalid:
		return pgtype.Int4{
				Valid: false,
			},
			string(effectiveStatus),
			nil

	default:
		return pgtype.Int4{},
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
