package postgres

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AircraftRepository struct {
	db *pgxpool.Pool
}

func NewAircraftRepository(db *pgxpool.Pool) *AircraftRepository {
	return &AircraftRepository{
		db: db,
	}
}

func (r *AircraftRepository) List(ctx context.Context) ([]aircraft.Aircraft, error) {
	const query = `
		SELECT
			COALESCE(a.icao24, ''),
			COALESCE(a.registration, ''),
			COALESCE(am.model, ''),
			COALESCE(am.manufacturer, ''),
			COALESCE(am.aircraft_type, ''),
			COALESCE(al.name, ''),
			COALESCE(c.name, '')
		FROM aircraft a
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		LEFT JOIN countries c ON c.id = a.country_id
		ORDER BY a.last_seen_at DESC NULLS LAST, a.created_at DESC
		LIMIT 100;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]aircraft.Aircraft, 0)

	for rows.Next() {
		var item aircraft.Aircraft

		if err := rows.Scan(
			&item.ICAO24,
			&item.Registration,
			&item.Model,
			&item.Manufacturer,
			&item.AircraftType,
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

func (r *AircraftRepository) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error) {
	const query = `
		SELECT
			COALESCE(a.icao24, ''),
			COALESCE(a.registration, ''),
			COALESCE(am.model, ''),
			COALESCE(am.manufacturer, ''),
			COALESCE(am.aircraft_type, ''),
			COALESCE(al.name, ''),
			COALESCE(c.name, '')
		FROM aircraft a
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		LEFT JOIN countries c ON c.id = a.country_id
		WHERE a.icao24 = $1
		LIMIT 1;
	`

	var item aircraft.Aircraft

	err := r.db.QueryRow(ctx, query, icao24).Scan(
		&item.ICAO24,
		&item.Registration,
		&item.Model,
		&item.Manufacturer,
		&item.AircraftType,
		&item.Airline,
		&item.Country,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aircraft.Aircraft{}, aircraft.ErrNotFound
		}

		return aircraft.Aircraft{}, err
	}

	return item, nil
}
