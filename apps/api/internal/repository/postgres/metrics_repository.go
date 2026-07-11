package postgres

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsRepository struct {
	db *pgxpool.Pool
}

func NewMetricsRepository(
	db *pgxpool.Pool,
) *MetricsRepository {
	return &MetricsRepository{
		db: db,
	}
}

func (
	r *MetricsRepository,
) CountActiveAircraft(
	ctx context.Context,
	query metrics.ActiveAircraftQuery,
) (metrics.ActiveAircraftObservationSummary, error) {
	const statement = `
		WITH candidate_states AS (
			SELECT
				fs.icao24,
				fs.source_name,
				fs.observed_at
			FROM flight_states fs
			WHERE fs.observed_at >= $1
			AND fs.observed_at <= $2
			AND btrim(fs.icao24) <> ''
			AND (
				$3::boolean = false
				OR (
					fs.latitude BETWEEN $4 AND $5
					AND fs.longitude BETWEEN $6 AND $7
				)
			)
		)
		SELECT
			COUNT(DISTINCT icao24)::integer,
			MIN(observed_at),
			MAX(observed_at),
			COALESCE(
				array_agg(DISTINCT source_name) FILTER (
					WHERE btrim(source_name) <> ''
				),
				ARRAY[]::text[]
			)
		FROM candidate_states;
	`

	var (
		count            int
		firstObservedAt  pgtype.Timestamptz
		latestObservedAt pgtype.Timestamptz
		sourceNames      []string
	)

	err := r.db.QueryRow(
		ctx,
		statement,
		query.ObservedFrom,
		query.ObservedTo,
		query.UseBounds,
		query.Bounds.MinLatitude,
		query.Bounds.MaxLatitude,
		query.Bounds.MinLongitude,
		query.Bounds.MaxLongitude,
	).Scan(
		&count,
		&firstObservedAt,
		&latestObservedAt,
		&sourceNames,
	)
	if err != nil {
		return metrics.ActiveAircraftObservationSummary{},
			err
	}

	summary := metrics.ActiveAircraftObservationSummary{
		Count:       count,
		SourceNames: sourceNames,
	}

	if count > 0 &&
		firstObservedAt.Valid &&
		latestObservedAt.Valid {
		summary.FirstObservedAt = firstObservedAt.Time
		summary.LatestObservedAt = latestObservedAt.Time
		summary.HasObservations = true
	}

	return summary,
		nil
}
