package routestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertResultSQL = `
		INSERT INTO flight_route_results (
			id,
			trajectory_id,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at,
			stored_at_unix_nano
		)
		VALUES (
			$1,
			$2::uuid,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10::jsonb,
			$11,
			$12
		)
		ON CONFLICT (
			trajectory_id,
			schema_version,
			as_of_time_unix_nano
		)
		DO NOTHING
		RETURNING
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at,
			stored_at_unix_nano;
	`

	getResultSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_route_results
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		  AND as_of_time_unix_nano = $3;
	`

	getLatestResultSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_route_results
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		ORDER BY
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT 1;
	`

	listResultsSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_route_results
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		ORDER BY
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT $3;
	`

	listResultsBeforeSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			route_status,
			confidence_level,
			validation_warning_count,
			route_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_route_results
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		  AND as_of_time_unix_nano < $3
		ORDER BY
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT $4;
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

type pgxPoolClient struct {
	pool *pgxpool.Pool
}

func (client pgxPoolClient) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) rowScanner {
	return client.pool.QueryRow(ctx, query, args...)
}

func (client pgxPoolClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	rows, err := client.pool.Query(ctx, query, args...)
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

	return newPostgresStore(
		pgxPoolClient{pool: config.Pool},
		config.Now,
	), nil
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
	result routecontract.Result,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalized := normalizeResult(result)
	report, err := validateStorableResult(normalized)
	if err != nil {
		return Record{}, err
	}

	trajectoryID, err := normalizePostgresTrajectoryID(
		normalized.TrajectoryID,
	)
	if err != nil {
		return Record{}, err
	}
	normalized.TrajectoryID = trajectoryID

	payload, err := json.Marshal(normalized)
	if err != nil {
		return Record{}, fmt.Errorf(
			"marshal route result payload: %w",
			err,
		)
	}

	key := resultKey(normalized)
	compositeKey := encodeResultKey(key)
	fingerprint :=
		normalized.Provenance.InputFingerprint
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
			key.TrajectoryID,
			string(key.SchemaVersion),
			key.AsOfTime,
			key.AsOfTime.UnixNano(),
			fingerprint,
			string(normalized.Status),
			string(normalized.Confidence.Level),
			report.WarningCount,
			payload,
			storedAt,
			storedAt.UnixNano(),
		),
	)
	if err == nil {
		return record.Clone(), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Record{}, databaseFailure(
			ctx,
			"insert route result",
			err,
		)
	}

	existing, err := store.getNormalized(
		ctx,
		key,
	)
	if err != nil {
		return Record{}, err
	}
	if existing.InputFingerprint != fingerprint {
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

	normalizedKey, err := normalizeResultKey(key)
	if err != nil {
		return Record{}, err
	}
	trajectoryID, err := normalizePostgresTrajectoryID(
		normalizedKey.TrajectoryID,
	)
	if err != nil {
		return Record{}, err
	}
	normalizedKey.TrajectoryID = trajectoryID

	return store.getNormalized(ctx, normalizedKey)
}

func (store *PostgresStore) getNormalized(
	ctx context.Context,
	key ResultKey,
) (Record, error) {
	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			getResultSQL,
			key.TrajectoryID,
			string(key.SchemaVersion),
			key.AsOfTime.UnixNano(),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrResultNotFound
	}
	if err != nil {
		return Record{}, databaseFailure(
			ctx,
			"get route result",
			err,
		)
	}

	return record.Clone(), nil
}

func (store *PostgresStore) GetLatest(
	ctx context.Context,
	trajectoryID string,
	schemaVersion routecontract.SchemaVersion,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalizedTrajectoryID, err :=
		normalizePostgresTrajectoryID(trajectoryID)
	if err != nil {
		return Record{}, err
	}
	if schemaVersion !=
		routecontract.SchemaVersionV1 {
		return Record{},
			ErrUnsupportedSchemaVersion
	}

	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			getLatestResultSQL,
			normalizedTrajectoryID,
			string(schemaVersion),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrResultNotFound
	}
	if err != nil {
		return Record{}, databaseFailure(
			ctx,
			"get latest route result",
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

	normalizedQuery, err := normalizeListQuery(query)
	if err != nil {
		return Page{}, err
	}
	trajectoryID, err := normalizePostgresTrajectoryID(
		normalizedQuery.TrajectoryID,
	)
	if err != nil {
		return Page{}, err
	}
	normalizedQuery.TrajectoryID = trajectoryID

	limitWithSentinel :=
		normalizedQuery.Limit + 1
	var rows rowIterator

	if normalizedQuery.BeforeAsOfTime.IsZero() {
		rows, err = store.client.Query(
			ctx,
			listResultsSQL,
			normalizedQuery.TrajectoryID,
			string(normalizedQuery.SchemaVersion),
			limitWithSentinel,
		)
	} else {
		rows, err = store.client.Query(
			ctx,
			listResultsBeforeSQL,
			normalizedQuery.TrajectoryID,
			string(normalizedQuery.SchemaVersion),
			normalizedQuery.BeforeAsOfTime.
				UnixNano(),
			limitWithSentinel,
		)
	}
	if err != nil {
		return Page{}, databaseFailure(
			ctx,
			"list route results",
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
			return Page{}, databaseFailure(
				ctx,
				"scan listed route result",
				scanErr,
			)
		}
		records = append(
			records,
			record.Clone(),
		)
	}
	if err := rows.Err(); err != nil {
		return Page{}, databaseFailure(
			ctx,
			"iterate listed route results",
			err,
		)
	}

	hasMore :=
		len(records) > normalizedQuery.Limit
	if hasMore {
		records =
			records[:normalizedQuery.Limit]
	}

	return Page{
		Records: records,
		HasMore: hasMore,
	}.Clone(), nil
}

func scanRecord(
	scanner rowScanner,
) (Record, error) {
	var (
		id                     string
		trajectoryID           string
		schemaVersion          string
		asOfTimeMirror         time.Time
		asOfTimeUnixNano       int64
		inputFingerprint       string
		routeStatus            string
		confidenceLevel        string
		validationWarningCount int
		payload                []byte
		storedAtMirror         time.Time
		storedAtUnixNano       int64
	)

	if err := scanner.Scan(
		&id,
		&trajectoryID,
		&schemaVersion,
		&asOfTimeMirror,
		&asOfTimeUnixNano,
		&inputFingerprint,
		&routeStatus,
		&confidenceLevel,
		&validationWarningCount,
		&payload,
		&storedAtMirror,
		&storedAtUnixNano,
	); err != nil {
		return Record{}, err
	}

	keyAsOfTime := time.Unix(
		0,
		asOfTimeUnixNano,
	).UTC()
	exactStoredAt := time.Unix(
		0,
		storedAtUnixNano,
	).UTC()
	if err := validateTimestampMirror(
		"as_of_time",
		asOfTimeMirror,
		keyAsOfTime,
	); err != nil {
		return Record{}, err
	}
	if err := validateTimestampMirror(
		"stored_at",
		storedAtMirror,
		exactStoredAt,
	); err != nil {
		return Record{}, err
	}

	var result routecontract.Result
	if err := json.Unmarshal(
		payload,
		&result,
	); err != nil {
		return Record{}, &DatabaseError{
			Operation: "decode route result payload",
			Err:       err,
		}
	}

	record := Record{
		ID: id,
		Key: ResultKey{
			TrajectoryID: trajectoryID,
			SchemaVersion: routecontract.SchemaVersion(
				schemaVersion,
			),
			AsOfTime: keyAsOfTime,
		},
		InputFingerprint: inputFingerprint,
		Result:           result,
		StoredAt:         exactStoredAt,
	}

	if err := validateDecodedRecord(
		record,
		routecontract.RouteStatus(routeStatus),
		routecontract.ConfidenceLevel(
			confidenceLevel,
		),
		validationWarningCount,
	); err != nil {
		return Record{}, err
	}

	return record.Clone(), nil
}

func validateDecodedRecord(
	record Record,
	routeStatus routecontract.RouteStatus,
	confidenceLevel routecontract.ConfidenceLevel,
	validationWarningCount int,
) error {
	expectedID := makeRecordID(
		encodeResultKey(record.Key),
		record.InputFingerprint,
	)
	if record.ID != expectedID {
		return &CorruptResultError{
			Field: "id",
		}
	}
	if record.Result.TrajectoryID !=
		record.Key.TrajectoryID {
		return &CorruptResultError{
			Field: "trajectory_id",
		}
	}
	if record.Result.SchemaVersion !=
		record.Key.SchemaVersion {
		return &CorruptResultError{
			Field: "schema_version",
		}
	}
	if !record.Result.Window.AsOfTime.UTC().
		Equal(record.Key.AsOfTime) {
		return &CorruptResultError{
			Field: "as_of_time_unix_nano",
		}
	}
	if record.Result.Provenance.
		InputFingerprint !=
		record.InputFingerprint {
		return &CorruptResultError{
			Field: "input_fingerprint",
		}
	}
	if record.Result.Status != routeStatus {
		return &CorruptResultError{
			Field: "route_status",
		}
	}
	if record.Result.Confidence.Level !=
		confidenceLevel {
		return &CorruptResultError{
			Field: "confidence_level",
		}
	}
	if validationWarningCount < 0 {
		return &CorruptResultError{
			Field: "validation_warning_count",
		}
	}

	report := routecontract.Validate(
		record.Result,
	)
	if report.Status !=
		routecontract.ValidationStatusValid ||
		report.WarningCount !=
			validationWarningCount {
		return &CorruptResultError{
			Field: "route_json",
		}
	}

	return nil
}

func normalizePostgresTrajectoryID(
	trajectoryID string,
) (string, error) {
	normalized, err := normalizeTrajectoryID(
		trajectoryID,
	)
	if err != nil {
		return "", err
	}

	parsed, err := uuid.Parse(normalized)
	if err != nil {
		return "", ErrInvalidTrajectoryID
	}

	return strings.ToLower(parsed.String()), nil
}

func databaseFailure(
	ctx context.Context,
	operation string,
	err error,
) error {
	if err == nil {
		return nil
	}
	if ctx != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
	}
	if errors.Is(err, context.Canceled) ||
		errors.Is(
			err,
			context.DeadlineExceeded,
		) ||
		errors.Is(err, ErrCorruptResult) {
		return err
	}

	var databaseErr *DatabaseError
	if errors.As(err, &databaseErr) {
		return err
	}

	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}
