package routestore

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5"
)

func TestNewPostgresWithExecutorRequiresExecutor(t *testing.T) {
	_, err := NewPostgresWithExecutor(nil, time.Now)
	if !errors.Is(err, ErrPostgresExecutorRequired) {
		t.Fatalf("error = %v", err)
	}
}

func TestPostgresStorePutAndRead(t *testing.T) {
	result := validRouteResult()
	storedAt := result.GeneratedAt.Add(time.Second)
	row := rowForResult(t, result, storedAt)
	client := &fakeClient{rows: []rowScanner{row}}
	store := newPostgresStore(client, func() time.Time { return storedAt })

	record, err := store.Put(context.Background(), result)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if record.Result.Status != routecontract.RouteStatusComplete ||
		record.Key.AsOfTime.UnixNano() != result.Window.AsOfTime.UnixNano() ||
		record.StoredAt.UnixNano() != storedAt.UnixNano() {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestPostgresStorePutIsIdempotent(t *testing.T) {
	result := validRouteResult()
	storedAt := result.GeneratedAt.Add(time.Second)
	client := &fakeClient{
		rows: []rowScanner{
			fakeRow{err: pgx.ErrNoRows},
			rowForResult(t, result, storedAt),
		},
	}
	store := newPostgresStore(client, func() time.Time { return storedAt })

	record, err := store.Put(context.Background(), result)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if record.InputFingerprint != result.Provenance.InputFingerprint {
		t.Fatalf("unexpected record: %#v", record)
	}
}

func TestPostgresStorePutDetectsConflict(t *testing.T) {
	result := validRouteResult()
	existing := result.Clone()
	existing.Provenance.InputFingerprint =
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	storedAt := result.GeneratedAt.Add(time.Second)
	client := &fakeClient{
		rows: []rowScanner{
			fakeRow{err: pgx.ErrNoRows},
			rowForResult(t, existing, storedAt),
		},
	}
	store := newPostgresStore(client, func() time.Time { return storedAt })

	_, err := store.Put(context.Background(), result)
	if !errors.Is(err, ErrResultConflict) {
		t.Fatalf("error = %v, want conflict", err)
	}
}

func TestPostgresStoreGetLatestAndList(t *testing.T) {
	first := validRouteResult()
	second := first.Clone()
	second.Window.AsOfTime = first.Window.AsOfTime.Add(-time.Minute)
	second.Provenance.InputFingerprint =
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	storedAt := first.GeneratedAt.Add(time.Second)

	client := &fakeClient{
		rows: []rowScanner{
			rowForResult(t, first, storedAt),
		},
		queryRows: &fakeRows{
			rows: []fakeRow{
				rowForResult(t, first, storedAt),
				rowForResult(t, second, storedAt),
			},
		},
	}
	store := newPostgresStore(client, func() time.Time { return storedAt })

	latest, err := store.GetLatest(
		context.Background(),
		first.TrajectoryID,
		routecontract.SchemaVersionV1,
	)
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if latest.Key.AsOfTime.UnixNano() != first.Window.AsOfTime.UnixNano() {
		t.Fatalf("unexpected latest: %#v", latest)
	}

	page, err := store.List(context.Background(), ListQuery{
		TrajectoryID:  first.TrajectoryID,
		SchemaVersion: routecontract.SchemaVersionV1,
		Limit:         1,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 1 || !page.HasMore {
		t.Fatalf("unexpected page: %#v", page)
	}
}

func TestPostgresStoreRejectsInvalidTrajectoryUUID(t *testing.T) {
	store := newPostgresStore(&fakeClient{}, time.Now)
	result := validRouteResult()
	result.TrajectoryID = "not-a-uuid"

	_, err := store.Put(context.Background(), result)
	if !errors.Is(err, ErrInvalidTrajectoryID) {
		t.Fatalf("error = %v", err)
	}
}

func TestScanRecordRejectsCorruptStatus(t *testing.T) {
	result := validRouteResult()
	storedAt := result.GeneratedAt.Add(time.Second)
	row := rowForResult(t, result, storedAt)
	row.values[6] = string(routecontract.RouteStatusPartial)

	_, err := scanRecord(row)
	if !errors.Is(err, ErrCorruptResult) {
		t.Fatalf("error = %v", err)
	}
}

func TestPostgresStorePreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := newPostgresStore(&fakeClient{}, time.Now)

	_, err := store.Put(ctx, validRouteResult())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}

func rowForResult(
	t *testing.T,
	result routecontract.Result,
	storedAt time.Time,
) fakeRow {
	t.Helper()
	normalized := normalizeResult(result)
	report, err := validateStorableResult(normalized)
	if err != nil {
		t.Fatalf("fixture validation error = %v", err)
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("marshal error = %v", err)
	}
	key := resultKey(normalized)
	return fakeRow{values: []any{
		makeRecordID(
			encodeResultKey(key),
			normalized.Provenance.InputFingerprint,
		),
		normalized.TrajectoryID,
		string(normalized.SchemaVersion),
		normalized.Window.AsOfTime,
		normalized.Window.AsOfTime.UnixNano(),
		normalized.Provenance.InputFingerprint,
		string(normalized.Status),
		string(normalized.Confidence.Level),
		report.WarningCount,
		payload,
		storedAt,
		storedAt.UnixNano(),
	}}
}

type fakeClient struct {
	rows      []rowScanner
	queryRows rowIterator
	queryErr  error
}

func (client *fakeClient) QueryRow(
	context.Context,
	string,
	...any,
) rowScanner {
	if len(client.rows) == 0 {
		return fakeRow{err: pgx.ErrNoRows}
	}
	row := client.rows[0]
	client.rows = client.rows[1:]
	return row
}

func (client *fakeClient) Query(
	context.Context,
	string,
	...any,
) (rowIterator, error) {
	if client.queryErr != nil {
		return nil, client.queryErr
	}
	return client.queryRows, nil
}

type fakeRow struct {
	values []any
	err    error
}

func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("unexpected destination count")
	}
	for index, destination := range destinations {
		if err := assign(destination, row.values[index]); err != nil {
			return err
		}
	}
	return nil
}

type fakeRows struct {
	rows  []fakeRow
	index int
	err   error
}

func (rows *fakeRows) Next() bool {
	return rows != nil && rows.index < len(rows.rows)
}

func (rows *fakeRows) Scan(destinations ...any) error {
	row := rows.rows[rows.index]
	rows.index++
	return row.Scan(destinations...)
}

func (rows *fakeRows) Err() error {
	if rows == nil {
		return nil
	}
	return rows.err
}

func (rows *fakeRows) Close() {}

func assign(destination any, value any) error {
	switch target := destination.(type) {
	case *string:
		typed, ok := value.(string)
		if !ok {
			return errors.New("expected string")
		}
		*target = typed
	case *int:
		typed, ok := value.(int)
		if !ok {
			return errors.New("expected int")
		}
		*target = typed
	case *int64:
		typed, ok := value.(int64)
		if !ok {
			return errors.New("expected int64")
		}
		*target = typed
	case *time.Time:
		typed, ok := value.(time.Time)
		if !ok {
			return errors.New("expected time.Time")
		}
		*target = typed
	case *[]byte:
		typed, ok := value.([]byte)
		if !ok {
			return errors.New("expected bytes")
		}
		*target = append([]byte(nil), typed...)
	default:
		return errors.New("unsupported destination")
	}
	return nil
}

func TestRecordFixtureRoundTrip(t *testing.T) {
	result := validRouteResult()
	record, err := scanRecord(
		rowForResult(t, result, result.GeneratedAt.Add(time.Second)),
	)
	if err != nil {
		t.Fatalf("scanRecord() error = %v", err)
	}
	if !reflect.DeepEqual(record.Result, normalizeResult(result)) {
		t.Fatalf("round trip mismatch")
	}
}
