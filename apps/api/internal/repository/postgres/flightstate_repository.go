package postgres

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrFlightStateNotFound = errors.New("flight state not found")

type FlightStateRepository struct {
	db *pgxpool.Pool
}

func NewFlightStateRepository(db *pgxpool.Pool) *FlightStateRepository {
	return &FlightStateRepository{
		db: db,
	}
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
			icao24,
			COALESCE(callsign, ''),
			COALESCE(latitude, 0),
			COALESCE(longitude, 0),
			COALESCE(barometric_altitude_m, 0),
			COALESCE(geometric_altitude_m, 0),
			COALESCE(velocity_mps, 0),
			COALESCE(heading_degrees, 0),
			COALESCE(vertical_rate_mps, 0),
			COALESCE(on_ground, false),
			COALESCE(origin_country, ''),
			observed_at,
			source_name
		FROM flight_states
		WHERE flight_id = $1
		ORDER BY observed_at ASC;
	`

	rows, err := r.db.Query(ctx, query, flightID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]flightstate.FlightState, 0)

	for rows.Next() {
		var item flightstate.FlightState

		if err := rows.Scan(
			&item.ID,
			&item.FlightID,
			&item.AircraftID,
			&item.ICAO24,
			&item.Callsign,
			&item.Latitude,
			&item.Longitude,
			&item.BarometricAltitudeM,
			&item.GeometricAltitudeM,
			&item.VelocityMPS,
			&item.HeadingDegrees,
			&item.VerticalRateMPS,
			&item.OnGround,
			&item.OriginCountry,
			&item.ObservedAt,
			&item.SourceName,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
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
			icao24,
			COALESCE(callsign, ''),
			COALESCE(latitude, 0),
			COALESCE(longitude, 0),
			COALESCE(barometric_altitude_m, 0),
			COALESCE(geometric_altitude_m, 0),
			COALESCE(velocity_mps, 0),
			COALESCE(heading_degrees, 0),
			COALESCE(vertical_rate_mps, 0),
			COALESCE(on_ground, false),
			COALESCE(origin_country, ''),
			observed_at,
			source_name
		FROM flight_states
		WHERE icao24 = $1
		ORDER BY observed_at DESC
		LIMIT 1;
	`

	var item flightstate.FlightState

	err := r.db.QueryRow(ctx, query, icao24).Scan(
		&item.ID,
		&item.FlightID,
		&item.AircraftID,
		&item.ICAO24,
		&item.Callsign,
		&item.Latitude,
		&item.Longitude,
		&item.BarometricAltitudeM,
		&item.GeometricAltitudeM,
		&item.VelocityMPS,
		&item.HeadingDegrees,
		&item.VerticalRateMPS,
		&item.OnGround,
		&item.OriginCountry,
		&item.ObservedAt,
		&item.SourceName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return flightstate.FlightState{}, ErrFlightStateNotFound
		}

		return flightstate.FlightState{}, err
	}

	return item, nil
}
