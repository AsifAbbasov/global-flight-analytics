package featurestore

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/jackc/pgx/v5"
)

const testTrajectoryID = "8A3D6E20-2C68-4B35-A512-7D91E6A90C31"

func TestNewPostgresRequiresPool(t *testing.T) {
	_, err := NewPostgres(PostgresConfig{})
	if !errors.Is(err, ErrPostgresPoolRequired) {
		t.Fatalf(
			"NewPostgres() error = %v, want %v",
			err,
			ErrPostgresPoolRequired,
		)
	}
}

func TestPostgresStoreImplementsStore(t *testing.T) {
	var store Store = newPostgresStore(
		&fakePostgresClient{},
		time.Now,
	)
	if store == nil {
		t.Fatal("PostgresStore does not implement Store")
	}
}

func TestPostgresStorePutInsertsCanonicalSnapshot(
	t *testing.T,
) {
	storedAt := time.Date(
		2026,
		time.July,
		14,
		18,
		30,
		0,
		987654321,
		time.UTC,
	)
	client := &fakePostgresClient{}
	client.queryRow = func(
		ctx context.Context,
		query string,
		args ...any,
	) rowScanner {
		if !strings.Contains(query, "INSERT INTO flight_feature_snapshots") ||
			!strings.Contains(query, "ON CONFLICT") {
			t.Fatalf("unexpected insert query:\n%s", query)
		}
		if len(args) != 10 {
			t.Fatalf("insert args = %d, want 10", len(args))
		}

		return rowFromInsertArguments(t, args)
	}

	store := newPostgresStore(
		client,
		func() time.Time {
			return storedAt
		},
	)
	features := validPostgresFeatures(
		testTrajectoryID,
		time.Date(
			2026,
			time.July,
			14,
			17,
			0,
			0,
			123456789,
			time.UTC,
		),
		"a",
	)

	record, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	canonicalTrajectoryID :=
		"8a3d6e20-2c68-4b35-a512-7d91e6a90c31"
	if record.Key.TrajectoryID != canonicalTrajectoryID ||
		record.Features.TrajectoryID != canonicalTrajectoryID {
		t.Fatalf(
			"trajectory id was not canonicalized: %#v",
			record,
		)
	}
	if !record.Key.AsOfTime.Equal(
		features.Window.AsOfTime,
	) {
		t.Fatalf(
			"AsOfTime = %v, want %v",
			record.Key.AsOfTime,
			features.Window.AsOfTime,
		)
	}
	if !record.StoredAt.Equal(storedAt) {
		t.Fatalf(
			"StoredAt = %v, want %v",
			record.StoredAt,
			storedAt,
		)
	}
	if !strings.HasPrefix(record.ID, recordIDPrefix) ||
		len(record.ID) != len(recordIDPrefix)+64 {
		t.Fatalf("record id = %q", record.ID)
	}
}

func TestPostgresStorePutIsIdempotentForSameFingerprint(
	t *testing.T,
) {
	features := validPostgresFeatures(
		testTrajectoryID,
		time.Date(
			2026,
			time.July,
			14,
			17,
			0,
			0,
			0,
			time.UTC,
		),
		"a",
	)
	existing := expectedRecord(
		features,
		time.Date(
			2026,
			time.July,
			14,
			18,
			0,
			0,
			0,
			time.UTC,
		),
	)
	call := 0
	client := &fakePostgresClient{
		queryRow: func(
			ctx context.Context,
			query string,
			args ...any,
		) rowScanner {
			call++
			if call == 1 {
				return errorRow{err: pgx.ErrNoRows}
			}

			return rowFromRecord(t, existing)
		},
	}

	store := newPostgresStore(client, time.Now)
	record, err := store.Put(
		context.Background(),
		features,
	)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if !reflect.DeepEqual(record, existing) {
		t.Fatalf(
			"record = %#v, want %#v",
			record,
			existing,
		)
	}
	if call != 2 {
		t.Fatalf("QueryRow calls = %d, want 2", call)
	}
}

func TestPostgresStorePutRejectsConflictingFingerprint(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		14,
		17,
		0,
		0,
		0,
		time.UTC,
	)
	incoming := validPostgresFeatures(
		testTrajectoryID,
		asOfTime,
		"b",
	)
	existingFeatures := validPostgresFeatures(
		testTrajectoryID,
		asOfTime,
		"a",
	)
	existing := expectedRecord(
		existingFeatures,
		time.Now().UTC(),
	)
	call := 0
	client := &fakePostgresClient{
		queryRow: func(
			ctx context.Context,
			query string,
			args ...any,
		) rowScanner {
			call++
			if call == 1 {
				return errorRow{err: pgx.ErrNoRows}
			}

			return rowFromRecord(t, existing)
		},
	}

	store := newPostgresStore(client, time.Now)
	_, err := store.Put(
		context.Background(),
		incoming,
	)
	if !errors.Is(err, ErrSnapshotConflict) {
		t.Fatalf(
			"Put() error = %v, want %v",
			err,
			ErrSnapshotConflict,
		)
	}
}

func TestPostgresStoreGetMapsNoRowsToNotFound(
	t *testing.T,
) {
	client := &fakePostgresClient{
		queryRow: func(
			context.Context,
			string,
			...any,
		) rowScanner {
			return errorRow{err: pgx.ErrNoRows}
		},
	}
	store := newPostgresStore(client, time.Now)

	_, err := store.Get(
		context.Background(),
		SnapshotKey{
			TrajectoryID:  testTrajectoryID,
			SchemaVersion: flightfeatures.SchemaVersionV1,
			AsOfTime:      time.Now(),
		},
	)
	if !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf(
			"Get() error = %v, want %v",
			err,
			ErrSnapshotNotFound,
		)
	}
}

func TestPostgresStoreGetLatestUsesDescendingOrder(
	t *testing.T,
) {
	features := validPostgresFeatures(
		testTrajectoryID,
		time.Now().UTC(),
		"a",
	)
	record := expectedRecord(
		features,
		time.Now().UTC(),
	)
	client := &fakePostgresClient{
		queryRow: func(
			ctx context.Context,
			query string,
			args ...any,
		) rowScanner {
			if !strings.Contains(
				query,
				"as_of_time_unix_nano DESC",
			) {
				t.Fatalf(
					"latest query is not descending:\n%s",
					query,
				)
			}

			return rowFromRecord(t, record)
		},
	}
	store := newPostgresStore(client, time.Now)

	loaded, err := store.GetLatest(
		context.Background(),
		testTrajectoryID,
		flightfeatures.SchemaVersionV1,
	)
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if !reflect.DeepEqual(loaded, record) {
		t.Fatalf(
			"loaded = %#v, want %#v",
			loaded,
			record,
		)
	}
}

func TestPostgresStoreListUsesSentinelAndCursor(
	t *testing.T,
) {
	base := time.Date(
		2026,
		time.July,
		14,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	records := []Record{
		expectedRecord(
			validPostgresFeatures(
				testTrajectoryID,
				base.Add(3*time.Hour),
				"a",
			),
			base.Add(4*time.Hour),
		),
		expectedRecord(
			validPostgresFeatures(
				testTrajectoryID,
				base.Add(2*time.Hour),
				"b",
			),
			base.Add(4*time.Hour),
		),
		expectedRecord(
			validPostgresFeatures(
				testTrajectoryID,
				base.Add(time.Hour),
				"c",
			),
			base.Add(4*time.Hour),
		),
	}
	client := &fakePostgresClient{
		query: func(
			ctx context.Context,
			query string,
			args ...any,
		) (rowIterator, error) {
			if !strings.Contains(
				query,
				"as_of_time_unix_nano < $3",
			) {
				t.Fatalf(
					"cursor query missing boundary:\n%s",
					query,
				)
			}
			if len(args) != 4 ||
				args[2] != base.Add(4*time.Hour).UnixNano() ||
				args[3] != 3 {
				t.Fatalf(
					"list args = %#v",
					args,
				)
			}

			return rowsFromRecords(t, records), nil
		},
	}
	store := newPostgresStore(client, time.Now)

	page, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:   testTrajectoryID,
			SchemaVersion:  flightfeatures.SchemaVersionV1,
			BeforeAsOfTime: base.Add(4 * time.Hour),
			Limit:          2,
		},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 2 || !page.HasMore {
		t.Fatalf("page = %#v", page)
	}
	if !page.Records[0].Key.AsOfTime.Equal(
		base.Add(3*time.Hour),
	) || !page.Records[1].Key.AsOfTime.Equal(
		base.Add(2*time.Hour),
	) {
		t.Fatalf(
			"unexpected order: %#v",
			page.Records,
		)
	}
}

func TestPostgresStoreRejectsInvalidUUID(t *testing.T) {
	store := newPostgresStore(
		&fakePostgresClient{},
		time.Now,
	)
	features := validPostgresFeatures(
		"not-a-uuid",
		time.Now().UTC(),
		"a",
	)

	_, err := store.Put(
		context.Background(),
		features,
	)
	if !errors.Is(err, ErrInvalidTrajectoryID) {
		t.Fatalf(
			"Put() error = %v, want %v",
			err,
			ErrInvalidTrajectoryID,
		)
	}
}

func TestPostgresStorePreservesContextCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	store := newPostgresStore(
		&fakePostgresClient{},
		time.Now,
	)
	_, err := store.GetLatest(
		ctx,
		testTrajectoryID,
		flightfeatures.SchemaVersionV1,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"GetLatest() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestPostgresStoreWrapsDatabaseFailure(t *testing.T) {
	sentinel := errors.New("database unavailable")
	client := &fakePostgresClient{
		queryRow: func(
			context.Context,
			string,
			...any,
		) rowScanner {
			return errorRow{err: sentinel}
		},
	}
	store := newPostgresStore(client, time.Now)

	_, err := store.GetLatest(
		context.Background(),
		testTrajectoryID,
		flightfeatures.SchemaVersionV1,
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf(
			"GetLatest() error = %v, want sentinel",
			err,
		)
	}
	var databaseErr *DatabaseError
	if !errors.As(err, &databaseErr) ||
		databaseErr.Operation !=
			"get latest snapshot" {
		t.Fatalf(
			"database error = %#v",
			databaseErr,
		)
	}
}

func TestPostgresStoreDetectsCorruptSnapshot(
	t *testing.T,
) {
	features := validPostgresFeatures(
		testTrajectoryID,
		time.Now().UTC(),
		"a",
	)
	record := expectedRecord(
		features,
		time.Now().UTC(),
	)
	record.ID = "feature-record-" +
		strings.Repeat("0", 64)

	client := &fakePostgresClient{
		queryRow: func(
			context.Context,
			string,
			...any,
		) rowScanner {
			return rowFromRecord(t, record)
		},
	}
	store := newPostgresStore(client, time.Now)

	_, err := store.GetLatest(
		context.Background(),
		testTrajectoryID,
		flightfeatures.SchemaVersionV1,
	)
	if !errors.Is(err, ErrCorruptSnapshot) {
		t.Fatalf(
			"GetLatest() error = %v, want %v",
			err,
			ErrCorruptSnapshot,
		)
	}
}

func TestPostgresStoreVersionRemainsStable(t *testing.T) {
	if PostgresVersion !=
		"flight-feature-postgres-store-v1" {
		t.Fatalf(
			"PostgresVersion = %q",
			PostgresVersion,
		)
	}
}

type fakePostgresClient struct {
	queryRow func(
		context.Context,
		string,
		...any,
	) rowScanner
	query func(
		context.Context,
		string,
		...any,
	) (rowIterator, error)
}

func (client *fakePostgresClient) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) rowScanner {
	if client.queryRow == nil {
		return errorRow{
			err: errors.New(
				"unexpected QueryRow call",
			),
		}
	}

	return client.queryRow(ctx, query, args...)
}

func (client *fakePostgresClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	if client.query == nil {
		return nil, errors.New(
			"unexpected Query call",
		)
	}

	return client.query(ctx, query, args...)
}

type errorRow struct {
	err error
}

func (row errorRow) Scan(...any) error {
	return row.err
}

type valueRow struct {
	scan func(...any) error
}

func (row valueRow) Scan(
	destinations ...any,
) error {
	return row.scan(destinations...)
}

type fakeRows struct {
	rows  []rowScanner
	index int
	err   error
}

func (rows *fakeRows) Next() bool {
	return rows.index < len(rows.rows)
}

func (rows *fakeRows) Scan(
	destinations ...any,
) error {
	if rows.index >= len(rows.rows) {
		return errors.New("scan without current row")
	}

	current := rows.rows[rows.index]
	rows.index++

	return current.Scan(destinations...)
}

func (rows *fakeRows) Err() error {
	return rows.err
}

func (rows *fakeRows) Close() {}

func rowFromInsertArguments(
	t *testing.T,
	args []any,
) rowScanner {
	t.Helper()

	return valueRow{
		scan: func(destinations ...any) error {
			schemaVersion, ok := args[2].(string)
			if !ok {
				t.Fatalf(
					"schema version argument type = %T",
					args[2],
				)
			}
			validationStatus, ok := args[6].(string)
			if !ok {
				t.Fatalf(
					"validation status argument type = %T",
					args[6],
				)
			}

			assignDatabaseRow(
				t,
				destinations,
				args[0].(string),
				args[1].(string),
				schemaVersion,
				args[3].(time.Time),
				args[4].(int64),
				args[5].(string),
				validationStatus,
				args[7].([]byte),
				args[8].(time.Time),
				args[9].(int64),
			)

			return nil
		},
	}
}

func rowFromRecord(
	t *testing.T,
	record Record,
) rowScanner {
	t.Helper()

	payload, err := json.Marshal(record.Features)
	if err != nil {
		t.Fatalf(
			"json.Marshal() error = %v",
			err,
		)
	}

	return valueRow{
		scan: func(destinations ...any) error {
			assignDatabaseRow(
				t,
				destinations,
				record.ID,
				record.Key.TrajectoryID,
				string(record.Key.SchemaVersion),
				record.Key.AsOfTime,
				record.Key.AsOfTime.UnixNano(),
				record.InputFingerprint,
				string(record.Features.Quality.Status),
				payload,
				record.StoredAt,
				record.StoredAt.UnixNano(),
			)

			return nil
		},
	}
}

func rowsFromRecords(
	t *testing.T,
	records []Record,
) rowIterator {
	t.Helper()

	rows := make(
		[]rowScanner,
		0,
		len(records),
	)
	for _, record := range records {
		rows = append(
			rows,
			rowFromRecord(t, record),
		)
	}

	return &fakeRows{rows: rows}
}

func assignDatabaseRow(
	t *testing.T,
	destinations []any,
	id string,
	trajectoryID string,
	schemaVersion string,
	asOfTime time.Time,
	asOfTimeUnixNano int64,
	inputFingerprint string,
	validationStatus string,
	payload []byte,
	storedAt time.Time,
	storedAtUnixNano int64,
) {
	t.Helper()

	if len(destinations) != 10 {
		t.Fatalf(
			"scan destinations = %d, want 10",
			len(destinations),
		)
	}

	*destinations[0].(*string) = id
	*destinations[1].(*string) = trajectoryID
	*destinations[2].(*string) = schemaVersion
	*destinations[3].(*time.Time) = asOfTime
	*destinations[4].(*int64) = asOfTimeUnixNano
	*destinations[5].(*string) = inputFingerprint
	*destinations[6].(*string) = validationStatus
	*destinations[7].(*[]byte) = append(
		[]byte(nil),
		payload...,
	)
	*destinations[8].(*time.Time) = storedAt
	*destinations[9].(*int64) = storedAtUnixNano
}

func validPostgresFeatures(
	trajectoryID string,
	asOfTime time.Time,
	suffix string,
) flightfeatures.FlightFeatures {
	return flightfeatures.FlightFeatures{
		SchemaVersion: flightfeatures.SchemaVersionV1,
		TrajectoryID:  trajectoryID,
		Window: flightfeatures.FeatureWindow{
			StartTime: asOfTime.Add(-time.Hour),
			EndTime:   asOfTime.Add(-time.Minute),
			AsOfTime:  asOfTime,
		},
		ExtractedAt: asOfTime,
		Quality: flightfeatures.FeatureQuality{
			Status: flightfeatures.ValidationStatusValid,
		},
		Provenance: flightfeatures.FeatureProvenance{
			ExtractorVersion: "flight-feature-extractor-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat(
					suffix,
					64,
				),
			TrajectoryUpdatedAt: asOfTime.Add(-time.Minute),
			SourceNames: []string{
				"test",
			},
		},
	}
}

func expectedRecord(
	features flightfeatures.FlightFeatures,
	storedAt time.Time,
) Record {
	normalized := normalizeFeatures(features)
	trajectoryID, err :=
		normalizePostgresTrajectoryID(
			normalized.TrajectoryID,
		)
	if err != nil {
		panic(err)
	}
	normalized.TrajectoryID = trajectoryID
	key := snapshotKey(normalized)
	fingerprint :=
		normalized.Provenance.InputFingerprint

	return Record{
		ID: makeRecordID(
			encodeSnapshotKey(key),
			fingerprint,
		),
		Key:              key,
		InputFingerprint: fingerprint,
		Features:         normalized,
		StoredAt:         storedAt.UTC(),
	}
}
