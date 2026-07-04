package postgres

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrafficRepository struct {
	db *pgxpool.Pool
}

func NewTrafficRepository(db *pgxpool.Pool) *TrafficRepository {
	return &TrafficRepository{
		db: db,
	}
}

func (r *TrafficRepository) GetCurrent(
	ctx context.Context,
) ([]traffic.CurrentTrafficItem, error) {
	const query = `
		SELECT DISTINCT ON (fs.icao24)
			fs.icao24,
			COALESCE(fs.callsign, ''),
			COALESCE(fs.latitude, 0),
			COALESCE(fs.longitude, 0),
			COALESCE(
				NULLIF(fs.geometric_altitude_m, 0),
				fs.barometric_altitude_m,
				0
			),
			COALESCE(fs.velocity_mps, 0),
			COALESCE(fs.heading_degrees, 0),
			COALESCE(fs.on_ground, false),
			fs.observed_at,
			COALESCE(am.model, ''),
			COALESCE(al.name, ''),
			COALESCE(fs.origin_country, '')
		FROM flight_states fs
		LEFT JOIN aircraft a ON a.id = fs.aircraft_id
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		ORDER BY fs.icao24, fs.observed_at DESC
		LIMIT 500;
	`

	return r.queryCurrentTraffic(ctx, query)
}

func (r *TrafficRepository) GetCurrentByBounds(
	ctx context.Context,
	bounds traffic.Bounds,
) ([]traffic.CurrentTrafficItem, error) {
	const query = `
		SELECT DISTINCT ON (fs.icao24)
			fs.icao24,
			COALESCE(fs.callsign, ''),
			COALESCE(fs.latitude, 0),
			COALESCE(fs.longitude, 0),
			COALESCE(
				NULLIF(fs.geometric_altitude_m, 0),
				fs.barometric_altitude_m,
				0
			),
			COALESCE(fs.velocity_mps, 0),
			COALESCE(fs.heading_degrees, 0),
			COALESCE(fs.on_ground, false),
			fs.observed_at,
			COALESCE(am.model, ''),
			COALESCE(al.name, ''),
			COALESCE(fs.origin_country, '')
		FROM flight_states fs
		LEFT JOIN aircraft a ON a.id = fs.aircraft_id
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		WHERE fs.latitude BETWEEN $1 AND $2
		AND fs.longitude BETWEEN $3 AND $4
		ORDER BY fs.icao24, fs.observed_at DESC
		LIMIT 500;
	`

	return r.queryCurrentTraffic(
		ctx,
		query,
		bounds.MinLatitude,
		bounds.MaxLatitude,
		bounds.MinLongitude,
		bounds.MaxLongitude,
	)
}

func (r *TrafficRepository) queryCurrentTraffic(
	ctx context.Context,
	query string,
	args ...any,
) ([]traffic.CurrentTrafficItem, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]traffic.CurrentTrafficItem, 0)

	for rows.Next() {
		var item traffic.CurrentTrafficItem

		if err := rows.Scan(
			&item.ICAO24,
			&item.Callsign,
			&item.Latitude,
			&item.Longitude,
			&item.AltitudeM,
			&item.VelocityMPS,
			&item.HeadingDegrees,
			&item.OnGround,
			&item.ObservedAt,
			&item.AircraftModel,
			&item.Airline,
			&item.OriginCountry,
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
