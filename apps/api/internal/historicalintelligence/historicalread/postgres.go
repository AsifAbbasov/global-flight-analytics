package historicalread

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

const (
	readFlightsSQL = `
		SELECT
			id::text,
			COALESCE(aircraft_id::text, ''),
			COALESCE(callsign, ''),
			status,
			first_seen_at,
			last_seen_at,
			updated_at
		FROM flights
		WHERE first_seen_at < $2
		  AND last_seen_at >= $1
		  AND updated_at <= $3
		ORDER BY first_seen_at ASC, id ASC
		LIMIT $4;
	`

	readTrajectoriesSQL = `
		SELECT
			id::text,
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			start_time,
			end_time,
			segment_count,
			point_count,
			coverage_gap_count,
			quality_score::double precision,
			source_name,
			updated_at
		FROM flight_trajectories
		WHERE start_time < $2
		  AND end_time >= $1
		  AND updated_at <= $3
		ORDER BY start_time ASC, id ASC
		LIMIT $4;
	`

	readObservationsSQL = `
		SELECT
			id::text,
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			latitude::double precision,
			longitude::double precision,
			on_ground,
			observed_at,
			source_name,
			created_at
		FROM flight_states
		WHERE observed_at >= $1
		  AND observed_at < $2
		  AND observed_at <= $3
		  AND created_at <= $3
		ORDER BY observed_at ASC, id ASC
		LIMIT $4;
	`

	readRoutesSQL = `
		SELECT
			id,
			trajectory_id::text,
			as_of_time,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at
		FROM flight_route_results
		WHERE as_of_time >= $1
		  AND as_of_time < $2
		  AND as_of_time <= $3
		  AND stored_at <= $3
		ORDER BY as_of_time ASC, id ASC
		LIMIT $4;
	`
)

type PostgresRepository struct {
	client postgresClient
}

func NewPostgres(config PostgresConfig) (*PostgresRepository, error) {
	if config.Pool == nil {
		return nil, ErrPostgresPoolRequired
	}

	return NewPostgresWithExecutor(
		config.Pool,
	)
}

func NewPostgresWithExecutor(
	executor Executor,
) (*PostgresRepository, error) {
	if executor == nil {
		return nil,
			ErrPostgresExecutorRequired
	}

	return &PostgresRepository{
		client: executorClient{
			executor: executor,
		},
	}, nil
}

func newPostgresRepository(client postgresClient) *PostgresRepository {
	return &PostgresRepository{client: client}
}

func (repository *PostgresRepository) Read(
	ctx context.Context,
	query Query,
) (Snapshot, error) {
	normalized, err := normalizeQuery(query)
	if err != nil {
		return Snapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	flightRows, flightLimitReached, err := repository.readFlights(ctx, normalized)
	if err != nil {
		return Snapshot{}, err
	}
	trajectoryRows, trajectoryLimitReached, err := repository.readTrajectories(ctx, normalized)
	if err != nil {
		return Snapshot{}, err
	}
	observationRows, observationLimitReached, err := repository.readObservations(ctx, normalized)
	if err != nil {
		return Snapshot{}, err
	}
	routeRows, routeLimitReached, err := repository.readRoutes(ctx, normalized)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Version: Version,
		Query:   normalized,

		Flights:      flightRows,
		Trajectories: trajectoryRows,
		Observations: observationRows,
		Routes:       routeRows,

		FlightLimitReached:      flightLimitReached,
		TrajectoryLimitReached:  trajectoryLimitReached,
		ObservationLimitReached: observationLimitReached,
		RouteLimitReached:       routeLimitReached,
	}.Clone(), nil
}

func normalizeQuery(query Query) (Query, error) {
	if query.Window.StartTime.IsZero() {
		return Query{}, ErrStartTimeRequired
	}
	if query.Window.EndTime.IsZero() {
		return Query{}, ErrEndTimeRequired
	}
	if query.Window.AsOfTime.IsZero() {
		return Query{}, ErrAsOfTimeRequired
	}

	startTime := query.Window.StartTime.UTC()
	endTime := query.Window.EndTime.UTC()
	asOfTime := query.Window.AsOfTime.UTC()

	if !startTime.Before(endTime) {
		return Query{}, ErrWindowNotPositive
	}
	if endTime.After(asOfTime) {
		return Query{}, ErrWindowExceedsAsOfTime
	}

	limit := query.Limit
	if limit == 0 {
		limit = DefaultDatasetLimit
	}
	if limit < 1 || limit > MaximumDatasetLimit {
		return Query{}, ErrInvalidDatasetLimit
	}

	return Query{
		Window: historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Limit: limit,
	}, nil
}

func (repository *PostgresRepository) readFlights(
	ctx context.Context,
	query Query,
) ([]FlightRecord, bool, error) {
	rows, err := repository.client.Query(
		ctx,
		readFlightsSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, false, databaseError("read flights", err)
	}
	defer rows.Close()

	items := make([]FlightRecord, 0, query.Limit)
	for rows.Next() {
		var item FlightRecord
		if err := rows.Scan(
			&item.ID,
			&item.AircraftID,
			&item.Callsign,
			&item.Status,
			&item.FirstSeenAt,
			&item.LastSeenAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, false, databaseError("scan flights", err)
		}
		items = append(items, normalizeFlight(item))
	}
	if err := rows.Err(); err != nil {
		return nil, false, databaseError("iterate flights", err)
	}

	return trimFlights(items, query.Limit)
}

func (repository *PostgresRepository) readTrajectories(
	ctx context.Context,
	query Query,
) ([]TrajectoryRecord, bool, error) {
	rows, err := repository.client.Query(
		ctx,
		readTrajectoriesSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, false, databaseError("read trajectories", err)
	}
	defer rows.Close()

	items := make([]TrajectoryRecord, 0, query.Limit)
	for rows.Next() {
		var item TrajectoryRecord
		if err := rows.Scan(
			&item.ID,
			&item.FlightID,
			&item.AircraftID,
			&item.ICAO24,
			&item.Callsign,
			&item.StartTime,
			&item.EndTime,
			&item.SegmentCount,
			&item.PointCount,
			&item.CoverageGapCount,
			&item.QualityScore,
			&item.SourceName,
			&item.UpdatedAt,
		); err != nil {
			return nil, false, databaseError("scan trajectories", err)
		}
		items = append(items, normalizeTrajectory(item))
	}
	if err := rows.Err(); err != nil {
		return nil, false, databaseError("iterate trajectories", err)
	}

	return trimTrajectories(items, query.Limit)
}

func (repository *PostgresRepository) readObservations(
	ctx context.Context,
	query Query,
) ([]ObservationRecord, bool, error) {
	rows, err := repository.client.Query(
		ctx,
		readObservationsSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, false, databaseError("read observations", err)
	}
	defer rows.Close()

	items := make([]ObservationRecord, 0, query.Limit)
	for rows.Next() {
		var item ObservationRecord
		if err := rows.Scan(
			&item.ID,
			&item.FlightID,
			&item.AircraftID,
			&item.ICAO24,
			&item.Callsign,
			&item.Latitude,
			&item.Longitude,
			&item.OnGround,
			&item.ObservedAt,
			&item.SourceName,
			&item.CreatedAt,
		); err != nil {
			return nil, false, databaseError("scan observations", err)
		}
		items = append(items, normalizeObservation(item))
	}
	if err := rows.Err(); err != nil {
		return nil, false, databaseError("iterate observations", err)
	}

	return trimObservations(items, query.Limit)
}

func (repository *PostgresRepository) readRoutes(
	ctx context.Context,
	query Query,
) ([]RouteRecord, bool, error) {
	rows, err := repository.client.Query(
		ctx,
		readRoutesSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, false, databaseError("read routes", err)
	}
	defer rows.Close()

	items := make([]RouteRecord, 0, query.Limit)
	for rows.Next() {
		var item RouteRecord
		if err := rows.Scan(
			&item.ID,
			&item.TrajectoryID,
			&item.AsOfTime,
			&item.InputFingerprint,
			&item.Status,
			&item.ConfidenceLevel,
			&item.ValidationWarningCount,
			&item.RouteJSON,
			&item.StoredAt,
		); err != nil {
			return nil, false, databaseError("scan routes", err)
		}
		items = append(items, normalizeRoute(item))
	}
	if err := rows.Err(); err != nil {
		return nil, false, databaseError("iterate routes", err)
	}

	return trimRoutes(items, query.Limit)
}

func databaseError(operation string, err error) error {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}

func normalizeFlight(item FlightRecord) FlightRecord {
	item.FirstSeenAt = item.FirstSeenAt.UTC()
	item.LastSeenAt = item.LastSeenAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item
}

func normalizeTrajectory(item TrajectoryRecord) TrajectoryRecord {
	item.StartTime = item.StartTime.UTC()
	item.EndTime = item.EndTime.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item
}

func normalizeObservation(item ObservationRecord) ObservationRecord {
	item.ObservedAt = item.ObservedAt.UTC()
	item.CreatedAt = item.CreatedAt.UTC()
	return item
}

func normalizeRoute(item RouteRecord) RouteRecord {
	item.AsOfTime = item.AsOfTime.UTC()
	item.StoredAt = item.StoredAt.UTC()
	item.RouteJSON = append([]byte(nil), item.RouteJSON...)
	return item
}

func trimFlights(items []FlightRecord, limit int) ([]FlightRecord, bool, error) {
	if len(items) <= limit {
		return items, false, nil
	}
	return items[:limit], true, nil
}

func trimTrajectories(items []TrajectoryRecord, limit int) ([]TrajectoryRecord, bool, error) {
	if len(items) <= limit {
		return items, false, nil
	}
	return items[:limit], true, nil
}

func trimObservations(items []ObservationRecord, limit int) ([]ObservationRecord, bool, error) {
	if len(items) <= limit {
		return items, false, nil
	}
	return items[:limit], true, nil
}

func trimRoutes(items []RouteRecord, limit int) ([]RouteRecord, bool, error) {
	if len(items) <= limit {
		return items, false, nil
	}
	return items[:limit], true, nil
}
