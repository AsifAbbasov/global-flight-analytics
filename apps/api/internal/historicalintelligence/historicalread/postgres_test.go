package historicalread

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

type fakeRow struct {
	values []any
	err    error
}

type fakeRows struct {
	rows      []fakeRow
	index     int
	err       error
	closed    bool
	scanError error
}

func (rows *fakeRows) Next() bool {
	return rows.index < len(rows.rows)
}

func (rows *fakeRows) Scan(destinations ...any) error {
	if rows.scanError != nil {
		return rows.scanError
	}
	if rows.index >= len(rows.rows) {
		return errors.New("scan called without row")
	}

	row := rows.rows[rows.index]
	rows.index++
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("destination count mismatch")
	}

	for index, value := range row.values {
		switch destination := destinations[index].(type) {
		case *string:
			*destination = value.(string)
		case *int:
			*destination = value.(int)
		case *float64:
			*destination = value.(float64)
		case *time.Time:
			*destination = value.(time.Time)
		case *[]byte:
			*destination = append([]byte(nil), value.([]byte)...)
		case **float64:
			if value == nil {
				*destination = nil
			} else {
				v := value.(float64)
				*destination = &v
			}
		case **bool:
			if value == nil {
				*destination = nil
			} else {
				v := value.(bool)
				*destination = &v
			}
		default:
			return errors.New("unsupported destination")
		}
	}

	return nil
}

func (rows *fakeRows) Err() error {
	return rows.err
}

func (rows *fakeRows) Close() {
	rows.closed = true
}

type fakeQueryCall struct {
	query string
	args  []any
}

type fakeClient struct {
	results []*fakeRows
	errs    []error
	calls   []fakeQueryCall
}

func (client *fakeClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	client.calls = append(
		client.calls,
		fakeQueryCall{
			query: query,
			args:  append([]any(nil), args...),
		},
	)

	index := len(client.calls) - 1
	if index < len(client.errs) &&
		client.errs[index] != nil {
		return nil, client.errs[index]
	}
	if index >= len(client.results) {
		return &fakeRows{}, nil
	}

	return client.results[index], nil
}

func TestNormalizeQuery(t *testing.T) {
	location := time.FixedZone("Asia/Baku", 4*60*60)
	startTime := time.Date(2026, time.July, 1, 12, 0, 0, 0, location)
	endTime := startTime.Add(time.Hour)
	asOfTime := endTime.Add(time.Hour)

	query, err := normalizeQuery(Query{
		Window: historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
	})
	if err != nil {
		t.Fatalf("normalizeQuery() error = %v", err)
	}

	if query.Limit != DefaultDatasetLimit {
		t.Fatalf("limit = %d", query.Limit)
	}
	if query.Window.StartTime.Location() != time.UTC ||
		query.Window.EndTime.Location() != time.UTC ||
		query.Window.AsOfTime.Location() != time.UTC {
		t.Fatal("query times are not normalized to UTC")
	}
	if !query.Window.StartTime.Equal(startTime.UTC()) {
		t.Fatal("query start instant changed")
	}
}

func TestNormalizeQueryRejectsInvalidInput(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	asOf := end.Add(time.Hour)

	tests := []struct {
		name  string
		query Query
		want  error
	}{
		{
			name: "start",
			query: Query{
				Window: historicalcontract.TimeWindow{
					EndTime:  end,
					AsOfTime: asOf,
				},
			},
			want: ErrStartTimeRequired,
		},
		{
			name: "end",
			query: Query{
				Window: historicalcontract.TimeWindow{
					StartTime: start,
					AsOfTime:  asOf,
				},
			},
			want: ErrEndTimeRequired,
		},
		{
			name: "as of",
			query: Query{
				Window: historicalcontract.TimeWindow{
					StartTime: start,
					EndTime:   end,
				},
			},
			want: ErrAsOfTimeRequired,
		},
		{
			name: "non-positive",
			query: Query{
				Window: historicalcontract.TimeWindow{
					StartTime: end,
					EndTime:   start,
					AsOfTime:  asOf,
				},
			},
			want: ErrWindowNotPositive,
		},
		{
			name: "future",
			query: Query{
				Window: historicalcontract.TimeWindow{
					StartTime: start,
					EndTime:   asOf,
					AsOfTime:  end,
				},
			},
			want: ErrWindowExceedsAsOfTime,
		},
		{
			name: "limit low",
			query: Query{
				Window: historicalcontract.TimeWindow{
					StartTime: start,
					EndTime:   end,
					AsOfTime:  asOf,
				},
				Limit: -1,
			},
			want: ErrInvalidDatasetLimit,
		},
		{
			name: "limit high",
			query: Query{
				Window: historicalcontract.TimeWindow{
					StartTime: start,
					EndTime:   end,
					AsOfTime:  asOf,
				},
				Limit: MaximumDatasetLimit + 1,
			},
			want: ErrInvalidDatasetLimit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizeQuery(test.query)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestPostgresRepositoryRead(t *testing.T) {
	base := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		123,
		time.FixedZone("source", 2*60*60),
	)
	client := &fakeClient{
		results: []*fakeRows{
			{
				rows: []fakeRow{
					{
						values: []any{
							"flight-1",
							"aircraft-1",
							"J2001",
							"completed",
							base,
							base.Add(time.Hour),
							base.Add(2 * time.Hour),
						},
					},
					{
						values: []any{
							"flight-2",
							"",
							"",
							"unknown",
							base.Add(time.Minute),
							base.Add(30 * time.Minute),
							base.Add(31 * time.Minute),
						},
					},
				},
			},
			{
				rows: []fakeRow{
					{
						values: []any{
							"trajectory-1",
							"flight-1",
							"aircraft-1",
							"abc123",
							"J2001",
							base,
							base.Add(time.Hour),
							2,
							10,
							1,
							0.9,
							"airplaneslive",
							base.Add(2 * time.Hour),
						},
					},
				},
			},
			{
				rows: []fakeRow{
					{
						values: []any{
							"state-1",
							"flight-1",
							"aircraft-1",
							"abc123",
							"J2001",
							40.1,
							49.9,
							false,
							base.Add(10 * time.Minute),
							"airplaneslive",
							base.Add(11 * time.Minute),
						},
					},
				},
			},
			{
				rows: []fakeRow{
					{
						values: []any{
							"route-record-1",
							"trajectory-1",
							base.Add(time.Hour),
							"sha256:" + strings.Repeat("a", 64),
							"complete",
							"high",
							0,
							[]byte(`{"status":"complete"}`),
							base.Add(2 * time.Hour),
						},
					},
				},
			},
		},
	}
	repository := newPostgresRepository(client)

	start := base.UTC().Add(-time.Hour)
	end := base.UTC().Add(3 * time.Hour)
	asOf := end.Add(time.Hour)

	snapshot, err := repository.Read(
		context.Background(),
		Query{
			Window: historicalcontract.TimeWindow{
				StartTime: start,
				EndTime:   end,
				AsOfTime:  asOf,
			},
			Limit: 1,
		},
	)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if snapshot.Version != Version ||
		len(snapshot.Flights) != 1 ||
		len(snapshot.Trajectories) != 1 ||
		len(snapshot.Observations) != 1 ||
		len(snapshot.Routes) != 1 ||
		!snapshot.FlightLimitReached ||
		snapshot.TrajectoryLimitReached ||
		snapshot.ObservationLimitReached ||
		snapshot.RouteLimitReached {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}

	if len(client.calls) != 4 {
		t.Fatalf("query call count = %d", len(client.calls))
	}
	for _, call := range client.calls {
		if len(call.args) != 4 {
			t.Fatalf("query args = %#v", call.args)
		}
		if call.args[3] != 2 {
			t.Fatalf("database limit = %v, want 2", call.args[3])
		}
	}

	if snapshot.Flights[0].FirstSeenAt.Location() != time.UTC ||
		snapshot.Trajectories[0].StartTime.Location() != time.UTC ||
		snapshot.Observations[0].ObservedAt.Location() != time.UTC ||
		snapshot.Routes[0].AsOfTime.Location() != time.UTC {
		t.Fatal("snapshot timestamps are not normalized to UTC")
	}

	snapshot.Routes[0].RouteJSON[0] = '['
	if string(client.results[3].rows[0].values[7].([]byte)) !=
		`{"status":"complete"}` {
		t.Fatal("Read() returned shared route JSON")
	}
}

func TestPostgresRepositoryUsesStableSQLPredicates(t *testing.T) {
	for name, query := range map[string]string{
		"flights":      readFlightsSQL,
		"trajectories": readTrajectoriesSQL,
		"observations": readObservationsSQL,
		"routes":       readRoutesSQL,
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(query, "ORDER BY") ||
				!strings.Contains(query, "LIMIT $4") {
				t.Fatalf("query is not deterministically bounded: %s", query)
			}
			if !strings.Contains(query, "$3") {
				t.Fatalf("query does not enforce as-of time: %s", query)
			}
		})
	}
}

func TestPostgresRepositoryPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := newPostgresRepository(&fakeClient{})
	_, err := repository.Read(
		ctx,
		validQuery(),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestPostgresRepositoryWrapsDatabaseErrors(t *testing.T) {
	sentinel := errors.New("database unavailable")
	repository := newPostgresRepository(
		&fakeClient{
			errs: []error{sentinel},
		},
	)

	_, err := repository.Read(
		context.Background(),
		validQuery(),
	)

	var databaseErr *DatabaseError
	if !errors.As(err, &databaseErr) ||
		!errors.Is(err, sentinel) ||
		databaseErr.Operation != "read flights" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestPostgresRepositoryWrapsScanAndIteratorErrors(t *testing.T) {
	sentinel := errors.New("scan failed")
	repository := newPostgresRepository(
		&fakeClient{
			results: []*fakeRows{
				{scanError: sentinel, rows: []fakeRow{{}}},
			},
		},
	)

	_, err := repository.Read(
		context.Background(),
		validQuery(),
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("scan error = %v", err)
	}

	sentinel = errors.New("iteration failed")
	repository = newPostgresRepository(
		&fakeClient{
			results: []*fakeRows{
				{err: sentinel},
			},
		},
	)

	_, err = repository.Read(
		context.Background(),
		validQuery(),
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("iteration error = %v", err)
	}
}

func TestTrimFunctions(t *testing.T) {
	flights, limited, err := trimFlights(
		[]FlightRecord{{ID: "1"}, {ID: "2"}},
		1,
	)
	if err != nil ||
		!limited ||
		!reflect.DeepEqual(
			flights,
			[]FlightRecord{{ID: "1"}},
		) {
		t.Fatalf("unexpected trim result: %#v %v %v", flights, limited, err)
	}
}

func validQuery() Query {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	return Query{
		Window: historicalcontract.TimeWindow{
			StartTime: start,
			EndTime:   end,
			AsOfTime:  end.Add(time.Hour),
		},
		Limit: 10,
	}
}
