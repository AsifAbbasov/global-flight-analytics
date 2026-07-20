package featurestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertSnapshotSQL = `
		INSERT INTO flight_feature_snapshots (
			id,
			trajectory_id,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			validation_status,
			features_json,
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
			$8::jsonb,
			$9,
			$10
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
			validation_status,
			features_json,
			stored_at,
			stored_at_unix_nano;
	`

	getSnapshotSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			validation_status,
			features_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_feature_snapshots
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		  AND as_of_time_unix_nano = $3;
	`

	getLatestSnapshotSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			validation_status,
			features_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_feature_snapshots
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		ORDER BY
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT 1;
	`

	listSnapshotsSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			validation_status,
			features_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_feature_snapshots
		WHERE trajectory_id = $1::uuid
		  AND schema_version = $2
		ORDER BY
			as_of_time_unix_nano DESC,
			id ASC
		LIMIT $3;
	`

	listSnapshotsBeforeSQL = `
		SELECT
			id,
			trajectory_id::text,
			schema_version,
			as_of_time,
			as_of_time_unix_nano,
			input_fingerprint,
			validation_status,
			features_json,
			stored_at,
			stored_at_unix_nano
		FROM flight_feature_snapshots
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
	features flightfeatures.FlightFeatures,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	if err := validateStorableFeatures(features); err != nil {
		return Record{}, err
	}

	normalized := normalizeFeatures(features)
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
			"marshal feature snapshot payload: %w",
			err,
		)
	}

	key := snapshotKey(normalized)
	compositeKey := encodeSnapshotKey(key)
	fingerprint := normalized.Provenance.InputFingerprint
	recordID := makeRecordID(
		compositeKey,
		fingerprint,
	)
	storedAt := store.now().UTC()

	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			insertSnapshotSQL,
			recordID,
			key.TrajectoryID,
			string(key.SchemaVersion),
			key.AsOfTime,
			key.AsOfTime.UnixNano(),
			fingerprint,
			string(normalized.Quality.Status),
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
			"insert snapshot",
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
		return Record{}, ErrSnapshotConflict
	}

	return existing.Clone(), nil
}

func (store *PostgresStore) Get(
	ctx context.Context,
	key SnapshotKey,
) (Record, error) {
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}

	normalizedKey, err := normalizeSnapshotKey(key)
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
	key SnapshotKey,
) (Record, error) {
	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			getSnapshotSQL,
			key.TrajectoryID,
			string(key.SchemaVersion),
			key.AsOfTime.UnixNano(),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrSnapshotNotFound
	}
	if err != nil {
		return Record{}, databaseFailure(
			ctx,
			"get snapshot",
			err,
		)
	}

	return record.Clone(), nil
}

func (store *PostgresStore) GetLatest(
	ctx context.Context,
	trajectoryID string,
	schemaVersion flightfeatures.SchemaVersion,
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
	if schemaVersion != flightfeatures.SchemaVersionV1 {
		return Record{}, ErrUnsupportedSchemaVersion
	}

	record, err := scanRecord(
		store.client.QueryRow(
			ctx,
			getLatestSnapshotSQL,
			normalizedTrajectoryID,
			string(schemaVersion),
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrSnapshotNotFound
	}
	if err != nil {
		return Record{}, databaseFailure(
			ctx,
			"get latest snapshot",
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

	limitWithSentinel := normalizedQuery.Limit + 1
	var rows rowIterator

	if normalizedQuery.BeforeAsOfTime.IsZero() {
		rows, err = store.client.Query(
			ctx,
			listSnapshotsSQL,
			normalizedQuery.TrajectoryID,
			string(normalizedQuery.SchemaVersion),
			limitWithSentinel,
		)
	} else {
		rows, err = store.client.Query(
			ctx,
			listSnapshotsBeforeSQL,
			normalizedQuery.TrajectoryID,
			string(normalizedQuery.SchemaVersion),
			normalizedQuery.BeforeAsOfTime.UnixNano(),
			limitWithSentinel,
		)
	}
	if err != nil {
		return Page{}, databaseFailure(
			ctx,
			"list snapshots",
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
				"scan listed snapshot",
				scanErr,
			)
		}

		records = append(records, record.Clone())
	}
	if err := rows.Err(); err != nil {
		return Page{}, databaseFailure(
			ctx,
			"iterate listed snapshots",
			err,
		)
	}

	hasMore := len(records) > normalizedQuery.Limit
	if hasMore {
		records = records[:normalizedQuery.Limit]
	}

	return Page{
		Records: records,
		HasMore: hasMore,
	}.Clone(), nil
}

func scanRecord(scanner rowScanner) (Record, error) {
	var (
		id               string
		trajectoryID     string
		schemaVersion    string
		asOfTime         time.Time
		asOfTimeUnixNano int64
		inputFingerprint string
		validationStatus string
		payload          []byte
		storedAt         time.Time
		storedAtUnixNano int64
	)

	if err := scanner.Scan(
		&id,
		&trajectoryID,
		&schemaVersion,
		&asOfTime,
		&asOfTimeUnixNano,
		&inputFingerprint,
		&validationStatus,
		&payload,
		&storedAt,
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
		asOfTime,
		keyAsOfTime,
	); err != nil {
		return Record{}, err
	}
	if err := validateTimestampMirror(
		"stored_at",
		storedAt,
		exactStoredAt,
	); err != nil {
		return Record{}, err
	}

	var features flightfeatures.FlightFeatures
	if err := json.Unmarshal(payload, &features); err != nil {
		return Record{}, &DatabaseError{
			Operation: "decode snapshot payload",
			Err:       err,
		}
	}

	key := SnapshotKey{
		TrajectoryID:  trajectoryID,
		SchemaVersion: flightfeatures.SchemaVersion(schemaVersion),
		AsOfTime:      keyAsOfTime,
	}
	record := Record{
		ID:               id,
		Key:              key,
		InputFingerprint: inputFingerprint,
		Features:         features,
		StoredAt:         exactStoredAt,
	}

	if err := validateDecodedRecord(
		record,
		flightfeatures.ValidationStatus(
			validationStatus,
		),
	); err != nil {
		return Record{}, err
	}

	return record.Clone(), nil
}

func validateDecodedRecord(
	record Record,
	validationStatus flightfeatures.ValidationStatus,
) error {
	expectedID := makeRecordID(
		encodeSnapshotKey(record.Key),
		record.InputFingerprint,
	)
	if record.ID != expectedID {
		return &CorruptSnapshotError{
			Field: "id",
		}
	}
	if record.Features.TrajectoryID !=
		record.Key.TrajectoryID {
		return &CorruptSnapshotError{
			Field: "trajectory_id",
		}
	}
	if record.Features.SchemaVersion !=
		record.Key.SchemaVersion {
		return &CorruptSnapshotError{
			Field: "schema_version",
		}
	}
	if !record.Features.Window.AsOfTime.UTC().Equal(
		record.Key.AsOfTime,
	) {
		return &CorruptSnapshotError{
			Field: "as_of_time_unix_nano",
		}
	}
	if record.Features.Provenance.InputFingerprint !=
		record.InputFingerprint {
		return &CorruptSnapshotError{
			Field: "input_fingerprint",
		}
	}
	if record.Features.Quality.Status !=
		validationStatus {
		return &CorruptSnapshotError{
			Field: "validation_status",
		}
	}
	if err := validateStorableFeatures(
		record.Features,
	); err != nil {
		return &CorruptSnapshotError{
			Field: "features_json",
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

func nonNilContext(
	ctx context.Context,
) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
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
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, ErrCorruptSnapshot) {
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
