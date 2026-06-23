package postgres

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAirportNotFound = errors.New("airport not found")

type AirportRepository struct {
	pool *pgxpool.Pool
}

func NewAirportRepository(pool *pgxpool.Pool) *AirportRepository {
	return &AirportRepository{
		pool: pool,
	}
}

func (r *AirportRepository) List(ctx context.Context) ([]airport.Airport, error) {
	const query = `
		SELECT
			COALESCE(a.icao_code, ''),
			COALESCE(a.iata_code, ''),
			a.name,
			COALESCE(a.city, ''),
			COALESCE(c.name, ''),
			a.latitude,
			a.longitude,
			COALESCE(a.elevation_ft, 0),
			COALESCE(a.timezone, ''),
			COALESCE(ap.description, '')
		FROM airports a
		LEFT JOIN countries c ON c.id = a.country_id
		LEFT JOIN airport_profiles ap ON ap.airport_id = a.id
		ORDER BY a.name ASC
		LIMIT 100;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	airports := make([]airport.Airport, 0)

	for rows.Next() {
		var item airport.Airport

		if err := rows.Scan(
			&item.ICAOCode,
			&item.IATACode,
			&item.Name,
			&item.City,
			&item.Country,
			&item.Latitude,
			&item.Longitude,
			&item.ElevationFt,
			&item.Timezone,
			&item.Description,
		); err != nil {
			return nil, err
		}

		airports = append(airports, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return airports, nil
}

func (r *AirportRepository) GetByICAO(ctx context.Context, icao string) (airport.Airport, error) {
	const query = `
		SELECT
			COALESCE(a.icao_code, ''),
			COALESCE(a.iata_code, ''),
			a.name,
			COALESCE(a.city, ''),
			COALESCE(c.name, ''),
			a.latitude,
			a.longitude,
			COALESCE(a.elevation_ft, 0),
			COALESCE(a.timezone, ''),
			COALESCE(ap.description, '')
		FROM airports a
		LEFT JOIN countries c ON c.id = a.country_id
		LEFT JOIN airport_profiles ap ON ap.airport_id = a.id
		WHERE a.icao_code = $1
		LIMIT 1;
	`

	var item airport.Airport

	err := r.pool.QueryRow(ctx, query, icao).Scan(
		&item.ICAOCode,
		&item.IATACode,
		&item.Name,
		&item.City,
		&item.Country,
		&item.Latitude,
		&item.Longitude,
		&item.ElevationFt,
		&item.Timezone,
		&item.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return airport.Airport{}, ErrAirportNotFound
		}

		return airport.Airport{}, err
	}

	return item, nil
}
