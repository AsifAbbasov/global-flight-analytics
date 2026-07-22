package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMetricsRepositoryPoolRequired = errors.New(
	"metrics repository pool is required",
)

const activeAircraftSummaryProjection = `
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

const activeAircraftGlobalStatement = `
		WITH candidate_states AS (
			SELECT
				fs.icao24,
				fs.source_name,
				fs.observed_at
			FROM flight_states fs
			WHERE fs.observed_at >= $1
			  AND fs.observed_at <= $2
			  AND btrim(fs.icao24) <> ''
		)
` + activeAircraftSummaryProjection

const activeAircraftBoundedStatement = `
		WITH candidate_states AS (
			SELECT
				fs.icao24,
				fs.source_name,
				fs.observed_at
			FROM flight_states fs
			WHERE fs.observed_at >= $1
			  AND fs.observed_at <= $2
			  AND btrim(fs.icao24) <> ''
			  AND fs.latitude BETWEEN $3 AND $4
			  AND fs.longitude BETWEEN $5 AND $6
		)
` + activeAircraftSummaryProjection

type MetricsRepository struct {
	db *pgxpool.Pool
}

func NewMetricsRepository(
	db *pgxpool.Pool,
) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (
	r *MetricsRepository,
) CountActiveAircraft(
	ctx context.Context,
	query metrics.ActiveAircraftQuery,
) (metrics.ActiveAircraftObservationSummary, error) {
	if r == nil || r.db == nil {
		return metrics.ActiveAircraftObservationSummary{},
			ErrMetricsRepositoryPoolRequired
	}
	if err := requireRepositoryContext(ctx); err != nil {
		return metrics.ActiveAircraftObservationSummary{}, err
	}
	if err := query.Validate(); err != nil {
		return metrics.ActiveAircraftObservationSummary{}, err
	}

	statement := activeAircraftGlobalStatement
	arguments := []any{
		query.ObservedFrom,
		query.ObservedTo,
	}
	if query.Scope.IsBounded() {
		statement = activeAircraftBoundedStatement
		arguments = append(
			arguments,
			query.Scope.Bounds.MinLatitude,
			query.Scope.Bounds.MaxLatitude,
			query.Scope.Bounds.MinLongitude,
			query.Scope.Bounds.MaxLongitude,
		)
	}

	var (
		count            int
		firstObservedAt  pgtype.Timestamptz
		latestObservedAt pgtype.Timestamptz
		sourceNames      []string
	)

	err := r.db.QueryRow(
		ctx,
		statement,
		arguments...,
	).Scan(
		&count,
		&firstObservedAt,
		&latestObservedAt,
		&sourceNames,
	)
	if err != nil {
		return metrics.ActiveAircraftObservationSummary{},
			fmt.Errorf("count active aircraft: %w", err)
	}

	summary := metrics.ActiveAircraftObservationSummary{
		Count:       count,
		SourceNames: sourceNames,
	}
	if count > 0 && firstObservedAt.Valid && latestObservedAt.Valid {
		summary.FirstObservedAt = firstObservedAt.Time
		summary.LatestObservedAt = latestObservedAt.Time
		summary.HasObservations = true
	}
	return summary, nil
}
