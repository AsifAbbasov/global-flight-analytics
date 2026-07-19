package postgres

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrafficRepository struct {
	db *pgxpool.Pool
}

func NewTrafficRepository(
	db *pgxpool.Pool,
) *TrafficRepository {
	return &TrafficRepository{
		db: db,
	}
}

func (
	r *TrafficRepository,
) GetCurrent(
	ctx context.Context,
) ([]traffic.CurrentTrafficItem, error) {
	const query = `
		WITH latest_successful_run AS (
			SELECT ir.id
			FROM ingestion_runs ir
			WHERE ir.status = $1
			AND ir.finished_at IS NOT NULL
			AND EXISTS (
				SELECT 1
				FROM flight_states candidate
				WHERE candidate.ingestion_run_id = ir.id
			)
			ORDER BY
				ir.finished_at DESC,
				ir.created_at DESC
			LIMIT 1
		)
		SELECT DISTINCT ON (fs.icao24)
			fs.icao24,
			COALESCE(fs.callsign, ''),
			fs.latitude,
			fs.longitude,
			COALESCE(
				NULLIF(fs.geometric_altitude_m, 0),
				fs.barometric_altitude_m,
				0
			),
			fs.velocity_mps,
			fs.heading_degrees,
			fs.on_ground,
			fs.observed_at,
			COALESCE(am.model, ''),
			COALESCE(al.name, ''),
			COALESCE(fs.origin_country, '')
		FROM flight_states fs
		JOIN latest_successful_run latest_run
			ON latest_run.id = fs.ingestion_run_id
		LEFT JOIN aircraft a ON a.id = fs.aircraft_id
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		WHERE fs.latitude IS NOT NULL
		  AND fs.longitude IS NOT NULL
		  AND fs.velocity_mps IS NOT NULL
		  AND fs.heading_degrees IS NOT NULL
		  AND fs.on_ground IS NOT NULL
		ORDER BY fs.icao24, fs.observed_at DESC;
	`

	rows, err := r.db.Query(
		ctx,
		query,
		string(
			ingestionrun.StatusSuccess,
		),
	)
	if err != nil {
		return nil,
			err
	}

	return scanCurrentTrafficRows(
		rows,
	)
}

func (
	r *TrafficRepository,
) GetCurrentByBounds(
	ctx context.Context,
	bounds traffic.Bounds,
) ([]traffic.CurrentTrafficItem, error) {
	const query = `
		WITH latest_successful_run AS (
			SELECT ir.id
			FROM ingestion_runs ir
			WHERE ir.status = $1
			AND ir.finished_at IS NOT NULL
			AND EXISTS (
				SELECT 1
				FROM flight_states candidate
				WHERE candidate.ingestion_run_id = ir.id
			)
			ORDER BY
				ir.finished_at DESC,
				ir.created_at DESC
			LIMIT 1
		)
		SELECT DISTINCT ON (fs.icao24)
			fs.icao24,
			COALESCE(fs.callsign, ''),
			fs.latitude,
			fs.longitude,
			COALESCE(
				NULLIF(fs.geometric_altitude_m, 0),
				fs.barometric_altitude_m,
				0
			),
			fs.velocity_mps,
			fs.heading_degrees,
			fs.on_ground,
			fs.observed_at,
			COALESCE(am.model, ''),
			COALESCE(al.name, ''),
			COALESCE(fs.origin_country, '')
		FROM flight_states fs
		JOIN latest_successful_run latest_run
			ON latest_run.id = fs.ingestion_run_id
		LEFT JOIN aircraft a ON a.id = fs.aircraft_id
		LEFT JOIN aircraft_models am ON am.id = a.model_id
		LEFT JOIN airlines al ON al.id = a.airline_id
		WHERE fs.latitude BETWEEN $2 AND $3
		AND fs.longitude BETWEEN $4 AND $5
		AND fs.velocity_mps IS NOT NULL
		AND fs.heading_degrees IS NOT NULL
		AND fs.on_ground IS NOT NULL
		ORDER BY fs.icao24, fs.observed_at DESC;
	`

	rows, err := r.db.Query(
		ctx,
		query,
		string(
			ingestionrun.StatusSuccess,
		),
		bounds.MinLatitude,
		bounds.MaxLatitude,
		bounds.MinLongitude,
		bounds.MaxLongitude,
	)
	if err != nil {
		return nil,
			err
	}

	return scanCurrentTrafficRows(
		rows,
	)
}

func scanCurrentTrafficRows(
	rows pgx.Rows,
) ([]traffic.CurrentTrafficItem, error) {
	defer rows.Close()

	items := make(
		[]traffic.CurrentTrafficItem,
		0,
	)

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
			return nil,
				err
		}

		items = append(
			items,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			err
	}

	return items,
		nil
}
