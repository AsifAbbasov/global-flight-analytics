package postgres

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// The international foot is defined exactly as 0.3048 metres.
	internationalFootInMeters = 0.3048
)

type AirportRepository struct {
	pool *pgxpool.Pool
}

func NewAirportRepository(
	pool *pgxpool.Pool,
) *AirportRepository {
	return &AirportRepository{
		pool: pool,
	}
}

func (repository *AirportRepository) List(
	ctx context.Context,
) ([]airport.Airport, error) {
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
		ORDER BY a.name ASC;
	`

	rows, err := repository.pool.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	airports := make(
		[]airport.Airport,
		0,
	)

	for rows.Next() {
		var item airport.Airport
		var elevationFeet float64

		if err := rows.Scan(
			&item.ICAOCode,
			&item.IATACode,
			&item.Name,
			&item.City,
			&item.Country,
			&item.Latitude,
			&item.Longitude,
			&elevationFeet,
			&item.Timezone,
			&item.Description,
		); err != nil {
			return nil, err
		}

		item.ElevationM = feetToMeters(
			elevationFeet,
		)

		airports = append(
			airports,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return airports, nil
}

func (repository *AirportRepository) GetByICAO(
	ctx context.Context,
	icao string,
) (airport.Airport, error) {
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
	var elevationFeet float64

	err := repository.pool.QueryRow(
		ctx,
		query,
		icao,
	).Scan(
		&item.ICAOCode,
		&item.IATACode,
		&item.Name,
		&item.City,
		&item.Country,
		&item.Latitude,
		&item.Longitude,
		&elevationFeet,
		&item.Timezone,
		&item.Description,
	)
	if err != nil {
		if errors.Is(
			err,
			pgx.ErrNoRows,
		) {
			return airport.Airport{}, airport.ErrNotFound
		}

		return airport.Airport{}, err
	}

	item.ElevationM = feetToMeters(
		elevationFeet,
	)

	return item, nil
}

func feetToMeters(
	feet float64,
) float64 {
	return feet * internationalFootInMeters
}
