package historicalaggregate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/jackc/pgx/v5"
)

const (
	insertResultSQL = `
		INSERT INTO historical_aggregate_results (
			id,
			schema_version,
			metric_name,
			scope_type,
			scope_key,
			region_code,
			airport_icao_code,
			origin_icao_code,
			destination_icao_code,
			granularity,
			window_start,
			window_start_unix_nano,
			window_end,
			window_end_unix_nano,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			series_status,
			confidence_level,
			result_json,
			stored_at,
			stored_at_unix_nano
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19,
			$20::jsonb,
			$21,
			$22
		)
		ON CONFLICT (
			schema_version,
			metric_name,
			scope_key,
			granularity,
			window_start_unix_nano,
			window_end_unix_nano,
			as_of_time_unix_nano
		)
		DO NOTHING
		RETURNING
			id,
			input_fingerprint,
			result_json,
			window_start,
			window_start_unix_nano,
			window_end,
			window_end_unix_nano,
			as_of_time,
			as_of_time_unix_nano,
			stored_at,
			stored_at_unix_nano;
	`

	getResultSQL = `
		SELECT
			id,
			input_fingerprint,
			result_json,
			window_start,
			window_start_unix_nano,
			window_end,
			window_end_unix_nano,
			as_of_time,
			as_of_time_unix_nano,
			stored_at,
			stored_at_unix_nano
		FROM historical_aggregate_results
		WHERE schema_version = $1
		  AND metric_name = $2
		  AND scope_key = $3
		  AND granularity = $4
		  AND window_start_unix_nano = $5
		  AND window_end_unix_nano = $6
		  AND as_of_time_unix_nano = $7;
	`

	getLatestResultSQL = `
		SELECT
			id,
			input_fingerprint,
			result_json,
			window_start,
			window_start_unix_nano,
			window_end,
			window_end_unix_nano,
			as_of_time,
			as_of_time_unix_nano,
			stored_at,
			stored_at_unix_nano
		FROM historical_aggregate_results
		WHERE schema_version = $1
		  AND metric_name = $2
		  AND scope_key = $3
		  AND granularity = $4
		ORDER BY
			window_end_unix_nano DESC,
			window_start_unix_nano DESC,
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT 1;
	`

	listResultsSQL = `
		SELECT
			id,
			input_fingerprint,
			result_json,
			window_start,
			window_start_unix_nano,
			window_end,
			window_end_unix_nano,
			as_of_time,
			as_of_time_unix_nano,
			stored_at,
			stored_at_unix_nano
		FROM historical_aggregate_results
		WHERE schema_version = $1
		  AND metric_name = $2
		  AND scope_key = $3
		  AND granularity = $4
		ORDER BY
			window_end_unix_nano DESC,
			window_start_unix_nano DESC,
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT $5;
	`

	listResultsAfterCursorSQL = `
		SELECT
			id,
			input_fingerprint,
			result_json,
			window_start,
			window_start_unix_nano,
			window_end,
			window_end_unix_nano,
			as_of_time,
			as_of_time_unix_nano,
			stored_at,
			stored_at_unix_nano
		FROM historical_aggregate_results
		WHERE schema_version = $1
		  AND metric_name = $2
		  AND scope_key = $3
		  AND granularity = $4
		  AND (
			window_end_unix_nano < $5
			OR (
				window_end_unix_nano = $5
				AND window_start_unix_nano < $6
			)
			OR (
				window_end_unix_nano = $5
				AND window_start_unix_nano = $6
				AND as_of_time_unix_nano < $7
			)
			OR (
				window_end_unix_nano = $5
				AND window_start_unix_nano = $6
				AND as_of_time_unix_nano = $7
				AND id > $8
			)
		  )
		ORDER BY
			window_end_unix_nano DESC,
			window_start_unix_nano DESC,
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT $9;
	`
)

type rowScanner interface {
	Scan(destinations ...any) error
}

type rowIterator interface {
	Next() bool
	Scan(destinations ...any) error
	Err() error
	Close()
}

type postgresClient interface {
	QueryRow(
		ctx context.Context,
		query string,
		args ...any,
	) rowScanner
	Query(
		ctx context.Context,
		query string,
		args ...any,
	) (rowIterator, error)
}

type executorClient struct {
	executor Executor
}

func (client executorClient) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) rowScanner {
	return client.executor.QueryRow(
		ctx,
		query,
		args...,
	)
}

func (client executorClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	rows, err := client.executor.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type PostgresStore struct {
	client postgresClient
	now    func() time.Time
}

func NewPostgres(
	config PostgresConfig,
) (*PostgresStore, error) {
	if config.Pool == nil {
		return nil, ErrPostgresPoolRequired
	}

	return NewPostgresWithExecutor(
		config.Pool,
		config.Now,
	)
}

func NewPostgresWithExecutor(
	executor Executor,
	now func() time.Time,
) (*PostgresStore, error) {
	if executor == nil {
		return nil,
			ErrPostgresExecutorRequired
	}
	if now == nil {
		now = time.Now
	}

	return &PostgresStore{
		client: executorClient{
			executor: executor,
		},
		now: now,
	}, nil
}

func newPostgresStore(
	client postgresClient,
	now func() time.Time,
) *PostgresStore {
	if now == nil {
		now = time.Now
	}
	return &PostgresStore{
		client: client,
		now:    now,
	}
}

func (store *PostgresStore) Put(
	ctx context.Context,
	result historicalcontract.Result,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalized := normalizeResult(result)
	if _, err := validateStorableResult(
		normalized,
	); err != nil {
		return Record{}, err
	}

	key := resultKey(normalized)
	normalizedKey, err := normalizeResultKey(key)
	if err != nil {
		return Record{}, err
	}
	compositeKey, err := encodeResultKey(
		normalizedKey,
	)
	if err != nil {
		return Record{}, err
	}
	encodedScope, err := scopeKey(
		normalizedKey.Scope,
	)
	if err != nil {
		return Record{}, err
	}

	payload, err := json.Marshal(normalized)
	if err != nil {
		return Record{},
			fmt.Errorf(
				"marshal historical aggregate result: %w",
				err,
			)
	}

	fingerprint := normalized.Provenance.
		InputFingerprint
	recordID := makeRecordID(
		compositeKey,
		fingerprint,
	)
	storedAt := store.now().UTC()

	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			insertResultSQL,
			recordID,
			string(normalizedKey.SchemaVersion),
			string(normalizedKey.MetricName),
			string(normalizedKey.Scope.Type),
			encodedScope,
			normalizedKey.Scope.RegionCode,
			normalizedKey.Scope.AirportICAOCode,
			normalizedKey.Scope.OriginICAOCode,
			normalizedKey.Scope.DestinationICAOCode,
			string(normalizedKey.Granularity),
			normalizedKey.Window.StartTime,
			normalizedKey.Window.StartTime.UnixNano(),
			normalizedKey.Window.EndTime,
			normalizedKey.Window.EndTime.UnixNano(),
			normalizedKey.Window.AsOfTime,
			normalizedKey.Window.AsOfTime.UnixNano(),
			fingerprint,
			string(normalized.Status),
			string(normalized.Confidence.Level),
			payload,
			storedAt,
			storedAt.UnixNano(),
		),
	)
	if err == nil {
		return record.Clone(), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Record{},
			databaseFailure(
				ctx,
				"insert result",
				err,
			)
	}

	existing, err := store.getNormalized(
		ctx,
		normalizedKey,
	)
	if err != nil {
		return Record{}, err
	}
	if existing.InputFingerprint !=
		fingerprint {
		return Record{}, ErrResultConflict
	}

	return existing.Clone(), nil
}

func (store *PostgresStore) Get(
	ctx context.Context,
	key ResultKey,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalized, err := normalizeResultKey(key)
	if err != nil {
		return Record{}, err
	}

	return store.getNormalized(
		ctx,
		normalized,
	)
}

func (store *PostgresStore) getNormalized(
	ctx context.Context,
	key ResultKey,
) (Record, error) {
	encodedScope, err := scopeKey(key.Scope)
	if err != nil {
		return Record{}, err
	}

	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			getResultSQL,
			string(key.SchemaVersion),
			string(key.MetricName),
			encodedScope,
			string(key.Granularity),
			key.Window.StartTime.UnixNano(),
			key.Window.EndTime.UnixNano(),
			key.Window.AsOfTime.UnixNano(),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrResultNotFound
	}
	if err != nil {
		return Record{},
			databaseFailure(
				ctx,
				"get result",
				err,
			)
	}

	return record.Clone(), nil
}

func (store *PostgresStore) GetLatest(
	ctx context.Context,
	query ListQuery,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalized, err := normalizeListQuery(
		query,
	)
	if err != nil {
		return Record{}, err
	}
	encodedScope, err := scopeKey(
		normalized.Scope,
	)
	if err != nil {
		return Record{}, err
	}

	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			getLatestResultSQL,
			string(normalized.SchemaVersion),
			string(normalized.MetricName),
			encodedScope,
			string(normalized.Granularity),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrResultNotFound
	}
	if err != nil {
		return Record{},
			databaseFailure(
				ctx,
				"get latest result",
				err,
			)
	}

	return record.Clone(), nil
}

func (store *PostgresStore) List(
	ctx context.Context,
	query ListQuery,
) (Page, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}

	normalized, err := normalizeListQuery(
		query,
	)
	if err != nil {
		return Page{}, err
	}
	encodedScope, err := scopeKey(
		normalized.Scope,
	)
	if err != nil {
		return Page{}, err
	}

	limitWithSentinel := normalized.Limit + 1
	var rows rowIterator

	if normalized.Cursor == nil {
		rows, err = store.client.Query(
			ctx,
			listResultsSQL,
			string(normalized.SchemaVersion),
			string(normalized.MetricName),
			encodedScope,
			string(normalized.Granularity),
			limitWithSentinel,
		)
	} else {
		cursor := normalized.Cursor
		rows, err = store.client.Query(
			ctx,
			listResultsAfterCursorSQL,
			string(normalized.SchemaVersion),
			string(normalized.MetricName),
			encodedScope,
			string(normalized.Granularity),
			cursor.WindowEnd.UnixNano(),
			cursor.WindowStart.UnixNano(),
			cursor.AsOfTime.UnixNano(),
			cursor.ID,
			limitWithSentinel,
		)
	}
	if err != nil {
		return Page{},
			databaseFailure(
				ctx,
				"list results",
				err,
			)
	}
	defer rows.Close()

	records := make(
		[]Record,
		0,
		limitWithSentinel,
	)
	for rows.Next() {
		record, scanErr := scanRecord(rows)
		if scanErr != nil {
			return Page{},
				databaseFailure(
					ctx,
					"scan listed result",
					scanErr,
				)
		}
		records = append(
			records,
			record.Clone(),
		)
	}
	if err := rows.Err(); err != nil {
		return Page{},
			databaseFailure(
				ctx,
				"iterate listed results",
				err,
			)
	}

	hasMore := len(records) >
		normalized.Limit
	var nextCursor *ListCursor
	if hasMore {
		records = records[:normalized.Limit]
		nextCursor = listCursorFromRecord(
			records[len(records)-1],
		)
	}

	return Page{
		Records:    records,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}.Clone(), nil
}

func scanRecord(
	scanner rowScanner,
) (Record, error) {
	var (
		id                  string
		inputFingerprint    string
		payload             []byte
		windowStartMirror   time.Time
		windowStartUnixNano int64
		windowEndMirror     time.Time
		windowEndUnixNano   int64
		asOfTimeMirror      time.Time
		asOfTimeUnixNano    int64
		storedAtMirror      time.Time
		storedAtUnixNano    int64
	)

	if err := scanner.Scan(
		&id,
		&inputFingerprint,
		&payload,
		&windowStartMirror,
		&windowStartUnixNano,
		&windowEndMirror,
		&windowEndUnixNano,
		&asOfTimeMirror,
		&asOfTimeUnixNano,
		&storedAtMirror,
		&storedAtUnixNano,
	); err != nil {
		return Record{}, err
	}

	exactWindowStart := time.Unix(0, windowStartUnixNano).UTC()
	exactWindowEnd := time.Unix(0, windowEndUnixNano).UTC()
	exactAsOfTime := time.Unix(0, asOfTimeUnixNano).UTC()
	exactStoredAt := time.Unix(0, storedAtUnixNano).UTC()

	for _, timestamp := range []struct {
		field  string
		mirror time.Time
		exact  time.Time
	}{
		{field: "window_start", mirror: windowStartMirror, exact: exactWindowStart},
		{field: "window_end", mirror: windowEndMirror, exact: exactWindowEnd},
		{field: "as_of_time", mirror: asOfTimeMirror, exact: exactAsOfTime},
		{field: "stored_at", mirror: storedAtMirror, exact: exactStoredAt},
	} {
		if err := validateTimestampMirror(
			timestamp.field,
			timestamp.mirror,
			timestamp.exact,
		); err != nil {
			return Record{}, err
		}
	}

	var result historicalcontract.Result
	if err := json.Unmarshal(
		payload,
		&result,
	); err != nil {
		return Record{},
			fmt.Errorf(
				"unmarshal historical aggregate result: %w",
				err,
			)
	}

	result = normalizeResult(result)
	for _, identity := range []struct {
		field string
		value time.Time
		exact time.Time
	}{
		{field: "window_start_unix_nano", value: result.Window.StartTime, exact: exactWindowStart},
		{field: "window_end_unix_nano", value: result.Window.EndTime, exact: exactWindowEnd},
		{field: "as_of_time_unix_nano", value: result.Window.AsOfTime, exact: exactAsOfTime},
	} {
		if !identity.value.UTC().Equal(identity.exact) {
			return Record{}, &CorruptResultError{Field: identity.field}
		}
	}
	if _, err := validateStorableResult(
		result,
	); err != nil {
		return Record{}, err
	}
	if result.Provenance.InputFingerprint !=
		inputFingerprint {
		return Record{},
			fmt.Errorf(
				"historical aggregate fingerprint mismatch",
			)
	}

	return Record{
		ID:               id,
		Key:              resultKey(result),
		InputFingerprint: inputFingerprint,
		Result:           result,
		StoredAt:         exactStoredAt,
	}, nil
}

func databaseFailure(
	ctx context.Context,
	operation string,
	err error,
) error {
	if ctx != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
	}

	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}
