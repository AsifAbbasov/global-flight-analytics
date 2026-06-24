package postgres

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flight"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrFlightNotFound = errors.New("flight not found")

type FlightRepository struct {
	db *pgxpool.Pool
}

func NewFlightRepository(db *pgxpool.Pool) *FlightRepository {
	return &FlightRepository{
		db: db,
	}
}

func (r *FlightRepository) List(ctx context.Context) ([]flight.Flight, error) {
	const query = `
		SELECT
			f.id::text,
			COALESCE(f.aircraft_id::text, ''),
			COALESCE(a.icao24, ''),
			COALESCE(f.callsign, ''),
			f.status,
			f.first_seen_at,
			f.last_seen_at,
			COALESCE(am.model, ''),
			COALESCE(al.name, ''),
			COALESCE(c.name, '')
		FROM flights f
		LEFT JOIN aircraft a ON a.id = f.aircraft_id
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		LEFT JOIN countries c ON c.id = a.country_id
		ORDER BY f.last_seen_at DESC
		LIMIT 100;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]flight.Flight, 0)

	for rows.Next() {
		var item flight.Flight

		if err := rows.Scan(
			&item.ID,
			&item.AircraftID,
			&item.ICAO24,
			&item.Callsign,
			&item.Status,
			&item.FirstSeenAt,
			&item.LastSeenAt,
			&item.AircraftModel,
			&item.Airline,
			&item.Country,
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

func (r *FlightRepository) GetByID(ctx context.Context, id string) (flight.Flight, error) {
	const query = `
		SELECT
			f.id::text,
			COALESCE(f.aircraft_id::text, ''),
			COALESCE(a.icao24, ''),
			COALESCE(f.callsign, ''),
			f.status,
			f.first_seen_at,
			f.last_seen_at,
			COALESCE(am.model, ''),
			COALESCE(al.name, ''),
			COALESCE(c.name, '')
		FROM flights f
		LEFT JOIN aircraft a ON a.id = f.aircraft_id
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		LEFT JOIN countries c ON c.id = a.country_id
		WHERE f.id = $1
		LIMIT 1;
	`

	var item flight.Flight

	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.AircraftID,
		&item.ICAO24,
		&item.Callsign,
		&item.Status,
		&item.FirstSeenAt,
		&item.LastSeenAt,
		&item.AircraftModel,
		&item.Airline,
		&item.Country,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return flight.Flight{}, ErrFlightNotFound
		}

		return flight.Flight{}, err
	}

	return item, nil
}
