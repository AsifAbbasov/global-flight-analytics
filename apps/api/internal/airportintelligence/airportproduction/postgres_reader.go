package airportproduction

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const listDailyObservationsSQL = `
    WITH base AS (
        SELECT
            a.id AS airport_id,
            UPPER(a.icao_code) AS icao_code,
            s.observation_date,
            SUM(s.arrivals)::bigint AS arrivals,
            SUM(s.departures)::bigint AS departures
        FROM airport_statistics s
        JOIN airports a ON a.id = s.airport_id
        WHERE s.observation_date >= $1::date
          AND s.observation_date < $2::date
          AND a.icao_code IS NOT NULL
          AND ($3 = '' OR UPPER(a.icao_code) = UPPER($3))
        GROUP BY a.id, UPPER(a.icao_code), s.observation_date
    ),
    route_edges AS (
        SELECT origin_airport_id AS airport_id, destination_airport_id AS counterpart_airport_id, observation_date
        FROM route_statistics
        WHERE observation_date >= $1::date AND observation_date < $2::date
        UNION ALL
        SELECT destination_airport_id AS airport_id, origin_airport_id AS counterpart_airport_id, observation_date
        FROM route_statistics
        WHERE observation_date >= $1::date AND observation_date < $2::date
    ),
    daily_routes AS (
        SELECT airport_id, observation_date, COUNT(DISTINCT counterpart_airport_id)::bigint AS active_routes
        FROM route_edges
        WHERE airport_id IS NOT NULL AND counterpart_airport_id IS NOT NULL
        GROUP BY airport_id, observation_date
    ),
    aircraft_edges AS (
        SELECT origin_airport_id AS airport_id, aircraft_id, flight_id, (calculated_at AT TIME ZONE 'UTC')::date AS observation_date
        FROM route_predictions
        WHERE calculated_at >= $1::date AND calculated_at < $2::date
        UNION ALL
        SELECT destination_airport_id AS airport_id, aircraft_id, flight_id, (calculated_at AT TIME ZONE 'UTC')::date AS observation_date
        FROM route_predictions
        WHERE calculated_at >= $1::date AND calculated_at < $2::date
    ),
    daily_aircraft AS (
        SELECT airport_id, observation_date,
            COUNT(DISTINCT COALESCE(aircraft_id::text, flight_id::text))::bigint AS active_aircraft
        FROM aircraft_edges
        WHERE airport_id IS NOT NULL AND (aircraft_id IS NOT NULL OR flight_id IS NOT NULL)
        GROUP BY airport_id, observation_date
    )
    SELECT
        base.icao_code,
        base.observation_date::timestamp AT TIME ZONE 'UTC' AS window_start,
        (base.observation_date + 1)::timestamp AT TIME ZONE 'UTC' AS window_end,
        base.arrivals,
        base.departures,
        COALESCE(daily_aircraft.active_aircraft, 0)::bigint,
        COALESCE(daily_routes.active_routes, 0)::bigint,
        ((base.observation_date + 1)::timestamp AT TIME ZONE 'UTC') - interval '1 microsecond' AS observed_at
    FROM base
    LEFT JOIN daily_routes
      ON daily_routes.airport_id = base.airport_id
     AND daily_routes.observation_date = base.observation_date
    LEFT JOIN daily_aircraft
      ON daily_aircraft.airport_id = base.airport_id
     AND daily_aircraft.observation_date = base.observation_date
    ORDER BY base.icao_code ASC, base.observation_date ASC;
`

type PostgresObservationReader struct{ pool *pgxpool.Pool }

func NewPostgresObservationReader(pool *pgxpool.Pool) (*PostgresObservationReader, error) {
	if pool == nil {
		return nil, ErrPostgresPoolRequired
	}
	return &PostgresObservationReader{pool: pool}, nil
}

func (reader *PostgresObservationReader) ListDaily(ctx context.Context, query DailyQuery) ([]DailyObservation, error) {
	if reader == nil || reader.pool == nil {
		return nil, ErrPostgresPoolRequired
	}
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if query.WindowStart.IsZero() || query.WindowEnd.IsZero() || !query.WindowEnd.After(query.WindowStart) {
		return nil, fmt.Errorf("%w: valid observation window is required", ErrInvalidRequest)
	}
	icaoCode := strings.ToUpper(strings.TrimSpace(query.ICAOCode))
	if icaoCode != "" {
		normalized, err := normalizeICAO(icaoCode)
		if err != nil {
			return nil, err
		}
		icaoCode = normalized
	}
	rows, err := reader.pool.Query(ctx, listDailyObservationsSQL, query.WindowStart.UTC(), query.WindowEnd.UTC(), icaoCode)
	if err != nil {
		return nil, fmt.Errorf("query Airport Intelligence daily observations: %w", err)
	}
	defer rows.Close()
	result := make([]DailyObservation, 0)
	for rows.Next() {
		var observation DailyObservation
		var arrivals, departures, activeAircraft, activeRoutes int64
		if err := rows.Scan(&observation.ICAOCode, &observation.WindowStart, &observation.WindowEnd, &arrivals, &departures, &activeAircraft, &activeRoutes, &observation.ObservedAt); err != nil {
			return nil, fmt.Errorf("scan Airport Intelligence daily observation: %w", err)
		}
		observation.Arrivals, err = safeInt(arrivals)
		if err != nil {
			return nil, err
		}
		observation.Departures, err = safeInt(departures)
		if err != nil {
			return nil, err
		}
		observation.ActiveAircraft, err = safeInt(activeAircraft)
		if err != nil {
			return nil, err
		}
		observation.ActiveRoutes, err = safeInt(activeRoutes)
		if err != nil {
			return nil, err
		}
		observation.WindowStart = observation.WindowStart.UTC()
		observation.WindowEnd = observation.WindowEnd.UTC()
		observation.ObservedAt = observation.ObservedAt.UTC()
		result = append(result, observation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Airport Intelligence daily observations: %w", err)
	}
	return result, nil
}

func safeInt(value int64) (int, error) {
	maximum := int64(^uint(0) >> 1)
	if value < 0 || value > maximum {
		return 0, fmt.Errorf("%w: PostgreSQL counter is outside the supported integer range", ErrInvalidRequest)
	}
	return int(value), nil
}

var _ ObservationReader = (*PostgresObservationReader)(nil)
