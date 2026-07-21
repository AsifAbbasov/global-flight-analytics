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

type fakePostgresClient struct {
	queryRows []rowScanner
	rows      rowIterator
	queryErr  error
}

func (client *fakePostgresClient) QueryRow(
	_ context.Context,
	_ string,
	_ ...any,
) rowScanner {
	if len(client.queryRows) == 0 {
		return fakeScanner{err: pgx.ErrNoRows}
	}

	result := client.queryRows[0]
	client.queryRows = client.queryRows[1:]
	return result
}

func (client *fakePostgresClient) Query(
	_ context.Context,
	_ string,
	_ ...any,
) (rowIterator, error) {
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
				},
			},
			DataCoverageRatio: 1,
			BuilderVersion:    Version,
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
		t.Fatalf("encode aggregate key: %v", err)
	}

	return []any{
		makeRecordID(
			encoded,
			result.Provenance.InputFingerprint,
		),
		result.Provenance.InputFingerprint,
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
