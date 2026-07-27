package historicalread

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	readHistoryCoverageSQL = `
		SELECT coverage_started_at
		FROM historical_read_history_state
		WHERE singleton = true;
	`

	readFlightsSQL = `
		WITH version_at_cutoff AS (
			SELECT DISTINCT ON (flight_id)
				flight_id,
				aircraft_id,
				callsign,
				status,
				first_seen_at,
				last_seen_at,
				source_updated_at,
				version_id
			FROM historical_read_flight_versions
			WHERE recorded_from <= $3
			  AND (recorded_to IS NULL OR recorded_to > $3)
			ORDER BY flight_id, recorded_from DESC, version_id DESC
		)
		SELECT
			flight_id::text,
			aircraft_id::text,
			callsign,
			status,
			first_seen_at,
			last_seen_at,
			source_updated_at,
			COUNT(*) OVER ()::bigint
		FROM version_at_cutoff
		WHERE first_seen_at < $2
		  AND last_seen_at > $1
		ORDER BY first_seen_at ASC, flight_id ASC
		LIMIT $4;
	`

	readTrajectoriesSQL = `
		WITH version_at_cutoff AS (
			SELECT DISTINCT ON (trajectory_id)
				trajectory_id,
				flight_id,
				aircraft_id,
				icao24,
				callsign,
				start_time,
				end_time,
				segment_count,
				point_count,
				coverage_gap_count,
				round(quality_score, 12)::double precision AS quality_score,
				source_name,
				source_updated_at,
				version_id
			FROM historical_read_trajectory_versions
			WHERE recorded_from <= $3
			  AND (recorded_to IS NULL OR recorded_to > $3)
			ORDER BY trajectory_id, recorded_from DESC, version_id DESC
		)
		SELECT
			trajectory_id::text,
			flight_id::text,
			aircraft_id::text,
			icao24,
			callsign,
			start_time,
			end_time,
			segment_count,
			point_count,
			coverage_gap_count,
			quality_score,
			source_name,
			source_updated_at,
			COUNT(*) OVER ()::bigint
		FROM version_at_cutoff
		WHERE start_time < $2
		  AND end_time > $1
		ORDER BY start_time ASC, trajectory_id ASC
		LIMIT $4;
	`

	readObservationsSQL = `
		SELECT
			id::text,
			flight_id::text,
			aircraft_id::text,
			icao24,
			callsign,
			round(latitude, 8)::double precision,
			round(longitude, 8)::double precision,
			on_ground,
			observed_at,
			source_name,
			created_at,
			COUNT(*) OVER ()::bigint
		FROM flight_states
		WHERE observed_at >= $1
		  AND observed_at < $2
		  AND observed_at <= $3
		  AND created_at <= $3
		ORDER BY observed_at ASC, id ASC
		LIMIT $4;
	`

	readRoutesSQL = `
		WITH trajectory_at_cutoff AS (
			SELECT DISTINCT ON (trajectory_id)
				trajectory_id,
				start_time,
				end_time,
				version_id
			FROM historical_read_trajectory_versions
			WHERE recorded_from <= $3
			  AND (recorded_to IS NULL OR recorded_to > $3)
			ORDER BY trajectory_id, recorded_from DESC, version_id DESC
		),
		latest_route AS (
			SELECT DISTINCT ON (result.trajectory_id)
				result.id,
				result.trajectory_id,
				result.as_of_time,
				result.input_fingerprint,
				result.route_status,
				result.confidence_level,
				result.validation_warning_count,
				result.route_json,
				result.stored_at,
				trajectory.start_time AS event_time,
				trajectory.end_time AS event_end_time,
				octet_length(result.route_json::text)::bigint AS payload_bytes
			FROM flight_route_results AS result
			JOIN trajectory_at_cutoff AS trajectory
			  ON trajectory.trajectory_id = result.trajectory_id
			WHERE trajectory.start_time < $2
			  AND trajectory.end_time > $1
			  AND result.as_of_time <= $3
			  AND result.stored_at <= $3
			ORDER BY
				result.trajectory_id,
				result.as_of_time DESC,
				result.stored_at DESC,
				result.id ASC
		),
		ordered_route AS (
			SELECT
				latest_route.*,
				COUNT(*) OVER ()::bigint AS total_count,
				COALESCE(
					SUM(payload_bytes) OVER (),
					0
				)::bigint AS total_payload_bytes,
				SUM(payload_bytes) OVER (
					ORDER BY event_time ASC, id ASC
					ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
				)::bigint AS cumulative_payload_bytes
			FROM latest_route
		)
		SELECT
			id,
			trajectory_id::text,
			event_time,
			event_end_time,
			as_of_time,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			CASE
				WHEN cumulative_payload_bytes <= $5 THEN route_json
				ELSE NULL
			END,
			stored_at,
			payload_bytes,
			total_count,
			total_payload_bytes,
			cumulative_payload_bytes > $5
		FROM ordered_route
		WHERE cumulative_payload_bytes - payload_bytes < $5
		ORDER BY event_time ASC, id ASC
		LIMIT $4;
	`
)

type PostgresRepository struct {
	beginner       snapshotBeginner
	client         postgresClient
	isolationLevel string
}

func NewPostgres(config PostgresConfig) (*PostgresRepository, error) {
	if config.Pool == nil {
		return nil, ErrPostgresPoolRequired
	}

	return &PostgresRepository{
		beginner:       poolSnapshotBeginner{pool: config.Pool},
		isolationLevel: SnapshotIsolationRepeatableRead,
	}, nil
}

func NewPostgresInTransaction(
	transaction pgx.Tx,
) (*PostgresRepository, error) {
	if transaction == nil {
		return nil, ErrPostgresTransactionRequired
	}

	return &PostgresRepository{
		client:         transactionClient{transaction: transaction},
		isolationLevel: SnapshotIsolationCallerTransaction,
	}, nil
}

func newPostgresRepository(client postgresClient) *PostgresRepository {
	return &PostgresRepository{
		client:         client,
		isolationLevel: SnapshotIsolationCallerTransaction,
	}
}

func newManagedPostgresRepository(
	beginner snapshotBeginner,
) *PostgresRepository {
	return &PostgresRepository{
		beginner:       beginner,
		isolationLevel: SnapshotIsolationRepeatableRead,
	}
}

func (repository *PostgresRepository) Read(
	ctx context.Context,
	query Query,
) (Snapshot, error) {
	if ctx == nil {
		return Snapshot{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	normalized, err := normalizeQuery(query)
	if err != nil {
		return Snapshot{}, err
	}

	if repository == nil ||
		(repository.beginner == nil && repository.client == nil) {
		return Snapshot{}, ErrPostgresClientRequired
	}

	if repository.beginner == nil {
		return repository.readSnapshot(
			ctx,
			repository.client,
			normalized,
			repository.isolationLevel,
		)
	}

	transaction, err := repository.beginner.BeginSnapshot(ctx)
	if err != nil {
		return Snapshot{}, databaseError("begin repeatable-read snapshot", err)
	}

	snapshot, readErr := repository.readSnapshot(
		ctx,
		transaction,
		normalized,
		SnapshotIsolationRepeatableRead,
	)
	if readErr != nil {
		rollbackErr := transaction.Rollback(context.Background())
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return Snapshot{}, errors.Join(
				readErr,
				databaseError("rollback repeatable-read snapshot", rollbackErr),
			)
		}
		return Snapshot{}, readErr
	}

	if err := transaction.Commit(ctx); err != nil {
		return Snapshot{}, databaseError("commit repeatable-read snapshot", err)
	}

	return snapshot.Clone(), nil
}

func (repository *PostgresRepository) readSnapshot(
	ctx context.Context,
	client postgresClient,
	query Query,
	isolationLevel string,
) (Snapshot, error) {
	coverageStartedAt, err := readHistoryCoverage(ctx, client)
	if err != nil {
		return Snapshot{}, err
	}
	if query.Window.AsOfTime.Before(coverageStartedAt) {
		return Snapshot{}, ErrTemporalHistoryUnavailable
	}

	flightRows, flightMatchedCount, flightLimitReached, err :=
		readFlights(ctx, client, query)
	if err != nil {
		return Snapshot{}, err
	}
	trajectoryRows, trajectoryMatchedCount, trajectoryLimitReached, err :=
		readTrajectories(ctx, client, query)
	if err != nil {
		return Snapshot{}, err
	}
	observationRows, observationMatchedCount, observationLimitReached, err :=
		readObservations(ctx, client, query)
	if err != nil {
		return Snapshot{}, err
	}
	routeRows, routeMatchedCount, routePayloadBytes, routeTotalPayloadBytes,
		routeLimitReached, routeByteLimitReached, err :=
		readRoutes(ctx, client, query)
	if err != nil {
		return Snapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Version:        Version,
		IsolationLevel: isolationLevel,
		Query:          query,

		Flights:      flightRows,
		Trajectories: trajectoryRows,
		Observations: observationRows,
		Routes:       routeRows,

		FlightMatchedCount:      flightMatchedCount,
		TrajectoryMatchedCount:  trajectoryMatchedCount,
		ObservationMatchedCount: observationMatchedCount,
		RouteMatchedCount:       routeMatchedCount,

		RoutePayloadBytes:      routePayloadBytes,
		RouteTotalPayloadBytes: routeTotalPayloadBytes,

		FlightLimitReached:      flightLimitReached,
		TrajectoryLimitReached:  trajectoryLimitReached,
		ObservationLimitReached: observationLimitReached,
		RouteLimitReached:       routeLimitReached,
		RouteByteLimitReached:   routeByteLimitReached,
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

	routePayloadByteLimit := query.RoutePayloadByteLimit
	if routePayloadByteLimit == 0 {
		routePayloadByteLimit = DefaultRoutePayloadByteLimit
	}
	if routePayloadByteLimit < 1 ||
		routePayloadByteLimit > MaximumRoutePayloadByteLimit {
		return Query{}, ErrInvalidRoutePayloadByteLimit
	}

	return Query{
		Window: historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Limit:                 limit,
		RoutePayloadByteLimit: routePayloadByteLimit,
	}, nil
}

func readHistoryCoverage(
	ctx context.Context,
	client postgresClient,
) (time.Time, error) {
	rows, err := client.Query(ctx, readHistoryCoverageSQL)
	if err != nil {
		return time.Time{}, databaseError("read temporal history coverage", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return time.Time{}, databaseError("iterate temporal history coverage", err)
		}
		return time.Time{}, ErrTemporalHistoryUnavailable
	}

	var coverageStartedAt time.Time
	if err := rows.Scan(&coverageStartedAt); err != nil {
		return time.Time{}, databaseError("scan temporal history coverage", err)
	}
	if rows.Next() {
		return time.Time{}, ErrSnapshotMetadataInvalid
	}
	if err := rows.Err(); err != nil {
		return time.Time{}, databaseError("iterate temporal history coverage", err)
	}
	if coverageStartedAt.IsZero() {
		return time.Time{}, ErrSnapshotMetadataInvalid
	}
	return coverageStartedAt.UTC(), nil
}

func readFlights(
	ctx context.Context,
	client postgresClient,
	query Query,
) ([]FlightRecord, int64, bool, error) {
	rows, err := client.Query(
		ctx,
		readFlightsSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, 0, false, databaseError("read flights", err)
	}
	defer rows.Close()

	items := make([]FlightRecord, 0, initialCapacity(query.Limit))
	var matchedCount int64
	for rows.Next() {
		var item FlightRecord
		var aircraftID pgtype.Text
		var callsign pgtype.Text
		var rowMatchedCount int64
		if err := rows.Scan(
			&item.ID,
			&aircraftID,
			&callsign,
			&item.Status,
			&item.FirstSeenAt,
			&item.LastSeenAt,
			&item.UpdatedAt,
			&rowMatchedCount,
		); err != nil {
			return nil, 0, false, databaseError("scan flights", err)
		}
		if err := reconcileMatchedCount(&matchedCount, rowMatchedCount); err != nil {
			return nil, 0, false, err
		}
		applyText(&item.AircraftID, &item.AircraftIDAvailable, aircraftID)
		applyText(&item.Callsign, &item.CallsignAvailable, callsign)
		item = normalizeFlight(item)
		if err := validateFlightRecord(item, query, len(items)); err != nil {
			return nil, 0, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, databaseError("iterate flights", err)
	}

	items, limited := trimToLimit(items, query.Limit)
	if err := validateMatchedCount(DatasetFlights, matchedCount, len(items)); err != nil {
		return nil, 0, false, err
	}
	return items, matchedCount, limited || matchedCount > int64(len(items)), nil
}

func readTrajectories(
	ctx context.Context,
	client postgresClient,
	query Query,
) ([]TrajectoryRecord, int64, bool, error) {
	rows, err := client.Query(
		ctx,
		readTrajectoriesSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, 0, false, databaseError("read trajectories", err)
	}
	defer rows.Close()

	items := make([]TrajectoryRecord, 0, initialCapacity(query.Limit))
	var matchedCount int64
	for rows.Next() {
		var item TrajectoryRecord
		var flightID pgtype.Text
		var aircraftID pgtype.Text
		var callsign pgtype.Text
		var rowMatchedCount int64
		if err := rows.Scan(
			&item.ID,
			&flightID,
			&aircraftID,
			&item.ICAO24,
			&callsign,
			&item.StartTime,
			&item.EndTime,
			&item.SegmentCount,
			&item.PointCount,
			&item.CoverageGapCount,
			&item.QualityScore,
			&item.SourceName,
			&item.UpdatedAt,
			&rowMatchedCount,
		); err != nil {
			return nil, 0, false, databaseError("scan trajectories", err)
		}
		if err := reconcileMatchedCount(&matchedCount, rowMatchedCount); err != nil {
			return nil, 0, false, err
		}
		applyText(&item.FlightID, &item.FlightIDAvailable, flightID)
		applyText(&item.AircraftID, &item.AircraftIDAvailable, aircraftID)
		applyText(&item.Callsign, &item.CallsignAvailable, callsign)
		item = normalizeTrajectory(item)
		if err := validateTrajectoryRecord(item, query, len(items)); err != nil {
			return nil, 0, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, databaseError("iterate trajectories", err)
	}

	items, limited := trimToLimit(items, query.Limit)
	if err := validateMatchedCount(DatasetTrajectories, matchedCount, len(items)); err != nil {
		return nil, 0, false, err
	}
	return items, matchedCount, limited || matchedCount > int64(len(items)), nil
}

func readObservations(
	ctx context.Context,
	client postgresClient,
	query Query,
) ([]ObservationRecord, int64, bool, error) {
	rows, err := client.Query(
		ctx,
		readObservationsSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
	)
	if err != nil {
		return nil, 0, false, databaseError("read observations", err)
	}
	defer rows.Close()

	items := make([]ObservationRecord, 0, initialCapacity(query.Limit))
	var matchedCount int64
	for rows.Next() {
		var item ObservationRecord
		var flightID pgtype.Text
		var aircraftID pgtype.Text
		var callsign pgtype.Text
		var rowMatchedCount int64
		if err := rows.Scan(
			&item.ID,
			&flightID,
			&aircraftID,
			&item.ICAO24,
			&callsign,
			&item.Latitude,
			&item.Longitude,
			&item.OnGround,
			&item.ObservedAt,
			&item.SourceName,
			&item.CreatedAt,
			&rowMatchedCount,
		); err != nil {
			return nil, 0, false, databaseError("scan observations", err)
		}
		if err := reconcileMatchedCount(&matchedCount, rowMatchedCount); err != nil {
			return nil, 0, false, err
		}
		applyText(&item.FlightID, &item.FlightIDAvailable, flightID)
		applyText(&item.AircraftID, &item.AircraftIDAvailable, aircraftID)
		applyText(&item.Callsign, &item.CallsignAvailable, callsign)
		item = normalizeObservation(item)
		if err := validateObservationRecord(item, query, len(items)); err != nil {
			return nil, 0, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, databaseError("iterate observations", err)
	}

	items, limited := trimToLimit(items, query.Limit)
	if err := validateMatchedCount(DatasetObservations, matchedCount, len(items)); err != nil {
		return nil, 0, false, err
	}
	return items, matchedCount, limited || matchedCount > int64(len(items)), nil
}

func readRoutes(
	ctx context.Context,
	client postgresClient,
	query Query,
) (
	[]RouteRecord,
	int64,
	int64,
	int64,
	bool,
	bool,
	error,
) {
	rows, err := client.Query(
		ctx,
		readRoutesSQL,
		query.Window.StartTime,
		query.Window.EndTime,
		query.Window.AsOfTime,
		query.Limit+1,
		query.RoutePayloadByteLimit,
	)
	if err != nil {
		return nil, 0, 0, 0, false, false, databaseError("read routes", err)
	}
	defer rows.Close()

	items := make([]RouteRecord, 0, initialCapacity(query.Limit))
	var matchedCount int64
	var totalPayloadBytes int64
	var payloadBytes int64
	byteLimitReached := false

	for rows.Next() {
		var item RouteRecord
		var payload []byte
		var rowMatchedCount int64
		var rowTotalPayloadBytes int64
		var overByteLimit bool
		if err := rows.Scan(
			&item.ID,
			&item.TrajectoryID,
			&item.EventStartTime,
			&item.EventEndTime,
			&item.AsOfTime,
			&item.InputFingerprint,
			&item.Status,
			&item.ConfidenceLevel,
			&item.ValidationWarningCount,
			&payload,
			&item.StoredAt,
			&item.PayloadBytes,
			&rowMatchedCount,
			&rowTotalPayloadBytes,
			&overByteLimit,
		); err != nil {
			return nil, 0, 0, 0, false, false, databaseError("scan routes", err)
		}
		if err := reconcileMatchedCount(&matchedCount, rowMatchedCount); err != nil {
			return nil, 0, 0, 0, false, false, err
		}
		if totalPayloadBytes != 0 && totalPayloadBytes != rowTotalPayloadBytes {
			return nil, 0, 0, 0, false, false, ErrSnapshotMetadataInvalid
		}
		totalPayloadBytes = rowTotalPayloadBytes
		if overByteLimit {
			byteLimitReached = true
			continue
		}

		item.RouteJSON = payload
		payloadHash := sha256.Sum256(payload)
		item.PayloadFingerprint = "sha256:" + hex.EncodeToString(payloadHash[:])
		item = normalizeRoute(item)
		if result, valid := item.ResultAt(query.Window.AsOfTime); valid {
			item.Result = result
			item.ResultAvailable = true
			item.RouteJSON = nil
		}
		if err := validateRouteRecord(item, query, len(items)); err != nil {
			return nil, 0, 0, 0, false, false, err
		}
		payloadBytes += item.PayloadBytes
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, 0, false, false, databaseError("iterate routes", err)
	}

	items, limited := trimToLimit(items, query.Limit)
	if err := validateMatchedCount(DatasetRoutes, matchedCount, len(items)); err != nil {
		return nil, 0, 0, 0, false, false, err
	}
	rowLimitReached := limited || matchedCount > int64(query.Limit)
	byteLimitReached = byteLimitReached || totalPayloadBytes > query.RoutePayloadByteLimit

	return items,
		matchedCount,
		payloadBytes,
		totalPayloadBytes,
		rowLimitReached,
		byteLimitReached,
		nil
}

func databaseError(operation string, err error) error {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}

func normalizeFlight(item FlightRecord) FlightRecord {
	item.FirstSeenAt = normalizeTime(item.FirstSeenAt)
	item.LastSeenAt = normalizeTime(item.LastSeenAt)
	item.UpdatedAt = normalizeTime(item.UpdatedAt)
	return item
}

func normalizeTrajectory(item TrajectoryRecord) TrajectoryRecord {
	item.StartTime = normalizeTime(item.StartTime)
	item.EndTime = normalizeTime(item.EndTime)
	item.UpdatedAt = normalizeTime(item.UpdatedAt)
	item.ICAO24 = strings.ToLower(strings.TrimSpace(item.ICAO24))
	item.SourceName = strings.TrimSpace(item.SourceName)
	return item
}

func normalizeObservation(item ObservationRecord) ObservationRecord {
	item.ObservedAt = normalizeTime(item.ObservedAt)
	item.CreatedAt = normalizeTime(item.CreatedAt)
	item.ICAO24 = strings.ToLower(strings.TrimSpace(item.ICAO24))
	item.SourceName = strings.TrimSpace(item.SourceName)
	return item
}

func normalizeRoute(item RouteRecord) RouteRecord {
	item.EventStartTime = normalizeTime(item.EventStartTime)
	item.EventEndTime = normalizeTime(item.EventEndTime)
	item.AsOfTime = normalizeTime(item.AsOfTime)
	item.StoredAt = normalizeTime(item.StoredAt)
	item.ID = strings.TrimSpace(item.ID)
	item.TrajectoryID = strings.TrimSpace(item.TrajectoryID)
	item.InputFingerprint = strings.TrimSpace(item.InputFingerprint)
	item.Status = strings.TrimSpace(item.Status)
	item.ConfidenceLevel = strings.TrimSpace(item.ConfidenceLevel)
	return item
}

func applyText(
	destination *string,
	available *bool,
	value pgtype.Text,
) {
	if !value.Valid {
		*destination = ""
		*available = false
		return
	}
	*destination = value.String
	*available = true
}

func reconcileMatchedCount(current *int64, incoming int64) error {
	if incoming < 0 {
		return ErrSnapshotMetadataInvalid
	}
	if *current == 0 {
		*current = incoming
		return nil
	}
	if *current != incoming {
		return ErrSnapshotMetadataInvalid
	}
	return nil
}

func initialCapacity(limit int) int {
	const maximumInitialCapacity = 1_024
	if limit < maximumInitialCapacity {
		return limit
	}
	return maximumInitialCapacity
}

func trimToLimit[T any](items []T, limit int) ([]T, bool) {
	if len(items) <= limit {
		return items, false
	}
	return items[:limit], true
}
