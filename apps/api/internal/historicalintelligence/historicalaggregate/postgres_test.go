package historicalaggregate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
	"github.com/jackc/pgx/v5"
)

type fakeScanner struct {
	values []any
	err    error
}

func (scanner fakeScanner) Scan(
	destinations ...any,
) error {
	if scanner.err != nil {
		return scanner.err
	}
	if len(destinations) != len(scanner.values) {
		return fmt.Errorf(
			"destination count=%d values=%d",
			len(destinations),
			len(scanner.values),
		)
	}

	for index, destination := range destinations {
		if err := assignFakeValue(
			destination,
			scanner.values[index],
		); err != nil {
			return err
		}
	}

	return nil
}

type fakeRows struct {
	values [][]any
	index  int
	err    error
}

func (rows *fakeRows) Next() bool {
	return rows.index < len(rows.values)
}

func (rows *fakeRows) Scan(
	destinations ...any,
) error {
	if rows.index >= len(rows.values) {
		return pgx.ErrNoRows
	}
	scanner := fakeScanner{
		values: rows.values[rows.index],
	}
	rows.index++
	return scanner.Scan(destinations...)
}

func (rows *fakeRows) Err() error {
	return rows.err
}

func (rows *fakeRows) Close() {}

type fakeQueryCall struct {
	query string
	args  []any
}

type fakePostgresClient struct {
	queryRows []rowScanner
	rows      rowIterator
	queryErr  error

	rowCalls   []fakeQueryCall
	queryCalls []fakeQueryCall
}

func (client *fakePostgresClient) QueryRow(
	_ context.Context,
	query string,
	args ...any,
) rowScanner {
	client.rowCalls = append(
		client.rowCalls,
		fakeQueryCall{
			query: query,
			args: append(
				[]any(nil),
				args...,
			),
		},
	)
	if len(client.queryRows) == 0 {
		return fakeScanner{err: pgx.ErrNoRows}
	}

	result := client.queryRows[0]
	client.queryRows = client.queryRows[1:]
	return result
}

func (client *fakePostgresClient) Query(
	_ context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	client.queryCalls = append(
		client.queryCalls,
		fakeQueryCall{
			query: query,
			args: append(
				[]any(nil),
				args...,
			),
		},
	)
	if client.queryErr != nil {
		return nil, client.queryErr
	}
	return client.rows, nil
}

func TestPostgresStorePutReturnsInsertedRecord(
	t *testing.T,
) {
	result := aggregateFixture(
		t,
		"b",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	key := resultKey(result)
	encoded, err := encodeResultKey(key)
	if err != nil {
		t.Fatalf("encode key: %v", err)
	}
	recordID := makeRecordID(
		encoded,
		result.Provenance.InputFingerprint,
	)
	storedAt := aggregateTestTime().
		Add(time.Minute)

	client := &fakePostgresClient{
		queryRows: []rowScanner{
			fakeScanner{
				values: aggregateRowAt(t, result, storedAt),
			},
		},
	}
	store := newPostgresStore(
		client,
		func() time.Time {
			return storedAt
		},
	)

	record, err := store.Put(
		context.Background(),
		result,
	)
	if err != nil {
		t.Fatalf("put aggregate: %v", err)
	}

	if record.ID != recordID ||
		record.InputFingerprint !=
			result.Provenance.InputFingerprint ||
		!record.StoredAt.Equal(storedAt) {
		t.Fatalf(
			"unexpected stored record: %#v",
			record,
		)
	}
	if record.Result.Summary.Total !=
		result.Summary.Total {
		t.Fatalf(
			"stored result total=%f want=%f",
			record.Result.Summary.Total,
			result.Summary.Total,
		)
	}
}

func TestPostgresStoreRejectsConflictingReplay(
	t *testing.T,
) {
	incoming := aggregateFixture(
		t,
		"c",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	existing := incoming.Clone()
	existing.Provenance.InputFingerprint =
		"sha256:" + strings.Repeat("d", 64)
	client := &fakePostgresClient{
		queryRows: []rowScanner{
			fakeScanner{err: pgx.ErrNoRows},
			fakeScanner{
				values: aggregateRow(t, existing),
			},
		},
	}
	store := newPostgresStore(
		client,
		aggregateTestTime,
	)

	_, err := store.Put(
		context.Background(),
		incoming,
	)
	if !errors.Is(err, ErrResultConflict) {
		t.Fatalf(
			"expected conflict error, got %v",
			err,
		)
	}
}

func TestPostgresStoreListUsesSentinelPagination(
	t *testing.T,
) {
	first := aggregateFixture(
		t,
		"f",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	second := aggregateFixture(
		t,
		"1",
		aggregateTestTime().Add(-2*time.Hour),
		aggregateTestTime().Add(-time.Hour),
	)
	third := aggregateFixture(
		t,
		"2",
		aggregateTestTime().Add(-3*time.Hour),
		aggregateTestTime().Add(-2*time.Hour),
	)

	rows := &fakeRows{
		values: [][]any{
			aggregateRow(t, first),
			aggregateRow(t, second),
			aggregateRow(t, third),
		},
	}
	store := newPostgresStore(
		&fakePostgresClient{
			rows: rows,
		},
		aggregateTestTime,
	)

	page, err := store.List(
		context.Background(),
		ListQuery{
			SchemaVersion: historicalcontract.
				SchemaVersionV1,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			Granularity: historicalcontract.
				GranularityHour,
			Limit: 2,
		},
	)
	if err != nil {
		t.Fatalf("list aggregates: %v", err)
	}
	if len(page.Records) != 2 ||
		!page.HasMore {
		t.Fatalf(
			"unexpected page: %#v",
			page,
		)
	}
}

func aggregateFixture(
	t *testing.T,
	fingerprintCharacter string,
	startTime time.Time,
	endTime time.Time,
) historicalcontract.Result {
	t.Helper()

	window := historicalcontract.TimeWindow{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  aggregateTestTime(),
	}
	bucket := historicalwindow.Bucket{
		Key:       "aggregate-bucket",
		Sequence:  0,
		StartTime: startTime,
		EndTime:   endTime,
	}
	result, err := historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: historicalcontract.Metric{
				Name: historicalcontract.
					MetricNameFlightCount,
				Unit: "flights",
				Aggregation: historicalcontract.
					AggregationCount,
			},
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			Plan: historicalwindow.Plan{
				Version: historicalwindow.Version,
				Fingerprint: "aggregate-plan-" +
					startTime.Format(time.RFC3339Nano),
				RequestedStartTime: startTime,
				RequestedEndTime:   endTime,
				AsOfTime:           aggregateTestTime(),
				Granularity: historicalcontract.
					GranularityHour,
				EffectiveWindow: &window,
				Buckets: []historicalwindow.Bucket{
					bucket,
				},
				MaximumBucketCount: 100,
			},
			Values: []historicalseries.BucketValue{
				{
					Bucket:      bucket,
					Value:       3,
					SampleCount: 3,
					Coverage: historicalseries.CoverageEvidence{
						State: historicalseries.
							CoverageStateComplete,
						LoadedCount:  3,
						MatchedCount: 3,
						Ratio:        1,
					},
				},
			},
			BuilderVersion: Version,
			InputFingerprint: "sha256:" +
				strings.Repeat(
					fingerprintCharacter,
					64,
				),
			SourceNames:           []string{"test"},
			LatestSourceUpdatedAt: endTime,
			GeneratedAt:           aggregateTestTime(),
		},
	)
	if err != nil {
		t.Fatalf(
			"build aggregate fixture: %v",
			err,
		)
	}

	return result
}

func aggregatePayload(
	t *testing.T,
	result historicalcontract.Result,
) []byte {
	t.Helper()
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf(
			"marshal aggregate fixture: %v",
			err,
		)
	}
	return payload
}

func aggregateRow(
	t *testing.T,
	result historicalcontract.Result,
) []any {
	t.Helper()
	return aggregateRowAt(t, result, aggregateTestTime())
}

func aggregateRowAt(
	t *testing.T,
	result historicalcontract.Result,
	storedAt time.Time,
) []any {
	t.Helper()
	encoded, err := encodeResultKey(
		resultKey(result),
	)
	if err != nil {
		t.Fatalf(
			"encode aggregate key: %v",
			err,
		)
	}
	encodedScope, err := scopeKey(
		result.Scope,
	)
	if err != nil {
		t.Fatalf(
			"encode aggregate scope: %v",
			err,
		)
	}

	return []any{
		makeRecordID(
			encoded,
			result.Provenance.InputFingerprint,
		),
		string(result.SchemaVersion),
		string(result.Metric.Name),
		string(result.Scope.Type),
		encodedScope,
		result.Scope.RegionCode,
		result.Scope.AirportICAOCode,
		result.Scope.OriginICAOCode,
		result.Scope.DestinationICAOCode,
		string(result.Granularity),
		result.Provenance.InputFingerprint,
		string(result.Status),
		string(result.Confidence.Level),
		aggregatePayload(t, result),
		result.Window.StartTime,
		result.Window.StartTime.UnixNano(),
		result.Window.EndTime,
		result.Window.EndTime.UnixNano(),
		result.Window.AsOfTime,
		result.Window.AsOfTime.UnixNano(),
		storedAt,
		storedAt.UnixNano(),
	}
}

func assignFakeValue(
	destination any,
	value any,
) error {
	switch target := destination.(type) {
	case *string:
		typed, ok := value.(string)
		if !ok {
			return fmt.Errorf(
				"value %T is not string",
				value,
			)
		}
		*target = typed
		return nil

	case *time.Time:
		typed, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf(
				"value %T is not time.Time",
				value,
			)
		}
		*target = typed
		return nil

	case *[]byte:
		typed, ok := value.([]byte)
		if !ok {
			return fmt.Errorf(
				"value %T is not []byte",
				value,
			)
		}
		*target = append([]byte(nil), typed...)
		return nil

	case *int64:
		typed, ok := value.(int64)
		if !ok {
			return fmt.Errorf(
				"value %T is not int64",
				value,
			)
		}
		*target = typed
		return nil

	default:
		return fmt.Errorf(
			"unsupported destination %T",
			destination,
		)
	}
}

func TestPostgresStoreRejectsNilContext(
	t *testing.T,
) {
	store := newPostgresStore(
		&fakePostgresClient{},
		aggregateTestTime,
	)
	_, err := store.Get(
		nil,
		ResultKey{},
	)
	if !errors.Is(
		err,
		ErrContextRequired,
	) {
		t.Fatalf(
			"expected context error, got %v",
			err,
		)
	}
}

func TestPostgresStoreValidatesBeforeCanonicalization(
	t *testing.T,
) {
	result := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	result.Metric.Unit = " flights "

	client := &fakePostgresClient{}
	store := newPostgresStore(
		client,
		aggregateTestTime,
	)
	_, err := store.Put(
		context.Background(),
		result,
	)
	var validationError *ValidationError
	if !errors.As(
		err,
		&validationError,
	) {
		t.Fatalf(
			"expected raw validation error, got %v",
			err,
		)
	}
	if len(client.rowCalls) != 0 {
		t.Fatalf(
			"database was called for invalid raw result: %#v",
			client.rowCalls,
		)
	}
}

func TestPostgresStoreRejectsSameFingerprintDifferentPayload(
	t *testing.T,
) {
	incoming := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	existing := incoming.Clone()
	existing.Provenance.BuilderVersion =
		"different-valid-builder"

	client := &fakePostgresClient{
		queryRows: []rowScanner{
			fakeScanner{err: pgx.ErrNoRows},
			fakeScanner{
				values: aggregateRow(
					t,
					existing,
				),
			},
		},
	}
	store := newPostgresStore(
		client,
		aggregateTestTime,
	)

	_, err := store.Put(
		context.Background(),
		incoming,
	)
	if !errors.Is(
		err,
		ErrResultPayloadConflict,
	) {
		t.Fatalf(
			"expected payload conflict, got %v",
			err,
		)
	}
}

func TestScanRecordRejectsStoredMetadataMismatch(
	t *testing.T,
) {
	result := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	row := aggregateRow(t, result)
	row[2] = string(
		historicalcontract.
			MetricNameActiveAircraft,
	)

	_, err := scanRecord(
		fakeScanner{values: row},
	)
	if !errors.Is(
		err,
		ErrCorruptResult,
	) {
		t.Fatalf(
			"expected corrupt metadata error, got %v",
			err,
		)
	}
}

func TestScanRecordRejectsStoredIdentifierMismatch(
	t *testing.T,
) {
	result := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	row := aggregateRow(t, result)
	row[0] = recordIDPrefix +
		strings.Repeat("f", 64)

	_, err := scanRecord(
		fakeScanner{values: row},
	)
	if !errors.Is(
		err,
		ErrCorruptResult,
	) {
		t.Fatalf(
			"expected corrupt identifier error, got %v",
			err,
		)
	}
}

func TestPostgresStoreListUsesFullTupleCursorSQL(
	t *testing.T,
) {
	result := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	encoded, err := encodeResultKey(
		resultKey(result),
	)
	if err != nil {
		t.Fatalf(
			"encode cursor fixture key: %v",
			err,
		)
	}
	cursorID := makeRecordID(
		encoded,
		result.Provenance.InputFingerprint,
	)
	client := &fakePostgresClient{
		rows: &fakeRows{},
	}
	store := newPostgresStore(
		client,
		aggregateTestTime,
	)

	_, err = store.List(
		context.Background(),
		ListQuery{
			SchemaVersion: result.SchemaVersion,
			MetricName:    result.Metric.Name,
			Scope:         result.Scope,
			Granularity:   result.Granularity,
			Cursor: &ListCursor{
				WindowEnd: result.Window.EndTime,
				WindowStart: result.
					Window.StartTime,
				AsOfTime: result.Window.AsOfTime,
				ID:       cursorID,
			},
			Limit: 3,
		},
	)
	if err != nil {
		t.Fatalf(
			"list with full cursor: %v",
			err,
		)
	}
	if len(client.queryCalls) != 1 {
		t.Fatalf(
			"query calls=%d want=1",
			len(client.queryCalls),
		)
	}
	call := client.queryCalls[0]
	for _, fragment := range []string{
		"window_end_unix_nano = $5",
		"window_start_unix_nano = $6",
		"as_of_time_unix_nano = $7",
		"id > $8",
		"LIMIT $9",
	} {
		if !strings.Contains(
			call.query,
			fragment,
		) {
			t.Fatalf(
				"cursor SQL misses %q",
				fragment,
			)
		}
	}
	if len(call.args) != 9 ||
		call.args[7] != cursorID ||
		call.args[8] != 4 {
		t.Fatalf(
			"unexpected cursor arguments: %#v",
			call.args,
		)
	}
}

func TestPostgresStoreRejectsInvalidStoredAt(
	t *testing.T,
) {
	result := aggregateFixture(
		t,
		"a",
		aggregateTestTime().Add(-time.Hour),
		aggregateTestTime(),
	)
	store := newPostgresStore(
		&fakePostgresClient{},
		func() time.Time {
			return result.GeneratedAt.
				Add(-time.Nanosecond)
		},
	)

	_, err := store.Put(
		context.Background(),
		result,
	)
	if !errors.Is(
		err,
		ErrStoredAtInvalid,
	) {
		t.Fatalf(
			"expected stored-at error, got %v",
			err,
		)
	}
}
