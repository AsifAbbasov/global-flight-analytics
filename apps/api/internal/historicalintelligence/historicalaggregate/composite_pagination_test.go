package historicalaggregate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

type paginationCapturingClient struct {
	rows  rowIterator
	query string
	args  []any
}

func (
	client *paginationCapturingClient,
) QueryRow(
	context.Context,
	string,
	...any,
) rowScanner {
	return fakeScanner{}
}

func (
	client *paginationCapturingClient,
) Query(
	_ context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	client.query = query
	client.args = append(
		[]any(nil),
		args...,
	)
	return client.rows, nil
}

func TestPostgresStoreBuildsCompositeNextCursor(
	t *testing.T,
) {
	sharedEnd := aggregateTestTime()
	first := aggregateFixture(
		t,
		"3",
		sharedEnd.Add(-time.Hour),
		sharedEnd,
	)
	second := aggregateFixture(
		t,
		"4",
		sharedEnd.Add(-time.Hour),
		sharedEnd,
	)
	third := aggregateFixture(
		t,
		"5",
		sharedEnd.Add(-time.Hour),
		sharedEnd,
	)

	client := &paginationCapturingClient{
		rows: &fakeRows{
			values: [][]any{
				aggregateRow(t, first),
				aggregateRow(t, second),
				aggregateRow(t, third),
			},
		},
	}
	store := newPostgresStore(
		client,
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
		t.Fatalf(
			"List() error = %v",
			err,
		)
	}
	if !page.HasMore ||
		len(page.Records) != 2 ||
		page.NextCursor == nil {
		t.Fatalf(
			"unexpected first page: %#v",
			page,
		)
	}

	last := page.Records[1]
	cursor := page.NextCursor
	if !cursor.WindowEnd.Equal(
		last.Key.Window.EndTime,
	) ||
		!cursor.WindowStart.Equal(
			last.Key.Window.StartTime,
		) ||
		!cursor.AsOfTime.Equal(
			last.Key.Window.AsOfTime,
		) ||
		cursor.ID != last.ID {
		t.Fatalf(
			"next cursor does not match the last returned record: cursor=%#v record=%#v",
			cursor,
			last,
		)
	}
}

func TestPostgresStoreUsesFullCompositeCursorPredicate(
	t *testing.T,
) {
	cursor := &ListCursor{
		WindowEnd: aggregateTestTime(),
		WindowStart: aggregateTestTime().
			Add(-time.Hour),
		AsOfTime: aggregateTestTime().
			Add(time.Minute),
		ID: "historical-aggregate-record-" +
			strings.Repeat("a", 64),
	}
	client := &paginationCapturingClient{
		rows: &fakeRows{},
	}
	store := newPostgresStore(
		client,
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
			Cursor: cursor,
			Limit:  2,
		},
	)
	if err != nil {
		t.Fatalf(
			"List() error = %v",
			err,
		)
	}
	if page.HasMore ||
		page.NextCursor != nil ||
		len(page.Records) != 0 {
		t.Fatalf(
			"unexpected empty page: %#v",
			page,
		)
	}

	for _, fragment := range []string{
		"window_end_unix_nano < $5",
		"window_end_unix_nano = $5",
		"window_start_unix_nano < $6",
		"window_start_unix_nano = $6",
		"as_of_time_unix_nano < $7",
		"as_of_time_unix_nano = $7",
		"id > $8",
		"LIMIT $9",
	} {
		if !strings.Contains(
			client.query,
			fragment,
		) {
			t.Fatalf(
				"composite cursor query is missing %q",
				fragment,
			)
		}
	}

	expectedArgs := []any{
		string(
			historicalcontract.SchemaVersionV1,
		),
		string(
			historicalcontract.
				MetricNameFlightCount,
		),
		"global",
		string(
			historicalcontract.
				GranularityHour,
		),
		cursor.WindowEnd.UnixNano(),
		cursor.WindowStart.UnixNano(),
		cursor.AsOfTime.UnixNano(),
		cursor.ID,
		3,
	}
	if len(client.args) != len(expectedArgs) {
		t.Fatalf(
			"cursor argument count = %d, want %d: %#v",
			len(client.args),
			len(expectedArgs),
			client.args,
		)
	}
	for index := range expectedArgs {
		if client.args[index] !=
			expectedArgs[index] {
			t.Fatalf(
				"cursor argument %d = %#v, want %#v",
				index,
				client.args[index],
				expectedArgs[index],
			)
		}
	}
}

func TestNormalizeListQueryRejectsPartialCursor(
	t *testing.T,
) {
	_, err := normalizeListQuery(
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
			Cursor: &ListCursor{
				WindowEnd: aggregateTestTime(),
			},
			Limit: 2,
		},
	)
	if !errors.Is(
		err,
		ErrInvalidListCursor,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidListCursor",
			err,
		)
	}
}
