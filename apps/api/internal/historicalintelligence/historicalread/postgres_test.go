package historicalread

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5/pgtype"
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
		if err := assignFakeValue(destinations[index], value); err != nil {
			return err
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
	_ context.Context,
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
	if index < len(client.errs) && client.errs[index] != nil {
		return nil, client.errs[index]
	}
	if index >= len(client.results) {
		return &fakeRows{}, nil
	}

	return client.results[index], nil
}

type fakeManagedSnapshot struct {
	*fakeClient
	committed  bool
	rolledBack bool
}

func (snapshot *fakeManagedSnapshot) Commit(context.Context) error {
	snapshot.committed = true
	return nil
}

func (snapshot *fakeManagedSnapshot) Rollback(context.Context) error {
	snapshot.rolledBack = true
	return nil
}

type fakeSnapshotBeginner struct {
	snapshot   *fakeManagedSnapshot
	beginCount int
	err        error
}

func (beginner *fakeSnapshotBeginner) BeginSnapshot(
	context.Context,
) (managedSnapshot, error) {
	beginner.beginCount++
	if beginner.err != nil {
		return nil, beginner.err
	}
	return beginner.snapshot, nil
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

	if query.Limit != DefaultDatasetLimit ||
		query.RoutePayloadByteLimit != DefaultRoutePayloadByteLimit {
		t.Fatalf("unexpected defaults: %#v", query)
	}
	if query.Window.StartTime.Location() != time.UTC ||
		query.Window.EndTime.Location() != time.UTC ||
		query.Window.AsOfTime.Location() != time.UTC {
		t.Fatal("query times are not normalized to UTC")
	}
}

func TestNormalizeQueryRejectsInvalidInput(t *testing.T) {
	start := historicalReadTestTime()
	end := start.Add(time.Hour)
	asOf := end.Add(time.Hour)

	tests := []struct {
		name  string
		query Query
		want  error
	}{
		{
			name: "start",
			query: Query{Window: historicalcontract.TimeWindow{
				EndTime: end, AsOfTime: asOf,
			}},
			want: ErrStartTimeRequired,
		},
		{
			name: "end",
			query: Query{Window: historicalcontract.TimeWindow{
				StartTime: start, AsOfTime: asOf,
			}},
			want: ErrEndTimeRequired,
		},
		{
			name: "as of",
			query: Query{Window: historicalcontract.TimeWindow{
				StartTime: start, EndTime: end,
			}},
			want: ErrAsOfTimeRequired,
		},
		{
			name: "window",
			query: Query{Window: historicalcontract.TimeWindow{
				StartTime: end, EndTime: start, AsOfTime: asOf,
			}},
			want: ErrWindowNotPositive,
		},
		{
			name: "future",
			query: Query{Window: historicalcontract.TimeWindow{
				StartTime: start, EndTime: asOf, AsOfTime: end,
			}},
			want: ErrWindowExceedsAsOfTime,
		},
		{
			name: "limit",
			query: Query{Window: historicalcontract.TimeWindow{
				StartTime: start, EndTime: end, AsOfTime: asOf,
			}, Limit: MaximumDatasetLimit + 1},
			want: ErrInvalidDatasetLimit,
		},
		{
			name: "byte limit",
			query: Query{Window: historicalcontract.TimeWindow{
				StartTime: start, EndTime: end, AsOfTime: asOf,
			}, RoutePayloadByteLimit: MaximumRoutePayloadByteLimit + 1},
			want: ErrInvalidRoutePayloadByteLimit,
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

func TestPostgresRepositoryReadUsesOneManagedSnapshot(t *testing.T) {
	client := validFakeClient(t, 2)
	managed := &fakeManagedSnapshot{fakeClient: client}
	beginner := &fakeSnapshotBeginner{snapshot: managed}
	repository := newManagedPostgresRepository(beginner)

	snapshot, err := repository.Read(context.Background(), validQuery())
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if beginner.beginCount != 1 || !managed.committed || managed.rolledBack {
		t.Fatalf("unexpected transaction lifecycle: %#v %#v", beginner, managed)
	}
	if len(client.calls) != 5 {
		t.Fatalf("query call count = %d, want 5 in one snapshot", len(client.calls))
	}
	if snapshot.IsolationLevel != SnapshotIsolationRepeatableRead ||
		snapshot.FlightMatchedCount != 2 ||
		!snapshot.FlightLimitReached ||
		snapshot.TrajectoryMatchedCount != 1 ||
		snapshot.ObservationMatchedCount != 1 ||
		snapshot.RouteMatchedCount != 1 {
		t.Fatalf("unexpected snapshot metadata: %#v", snapshot)
	}
	if len(snapshot.Flights) != 1 || len(snapshot.Routes) != 1 ||
		!snapshot.Routes[0].ResultAvailable {
		t.Fatalf("unexpected snapshot rows: %#v", snapshot)
	}
	if snapshot.Flights[0].AircraftIDAvailable ||
		!snapshot.Flights[0].CallsignAvailable {
		t.Fatalf("nullable provenance was not preserved: %#v", snapshot.Flights[0])
	}
}

func TestPostgresRepositoryRejectsNilContext(t *testing.T) {
	repository := newPostgresRepository(&fakeClient{})
	_, err := repository.Read(nil, validQuery())
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("error = %v, want %v", err, ErrContextRequired)
	}
}

func TestPostgresRepositoryRejectsUnavailableTemporalHistory(t *testing.T) {
	coverage := validQuery().Window.AsOfTime.Add(time.Hour)
	repository := newPostgresRepository(&fakeClient{
		results: []*fakeRows{{rows: []fakeRow{{values: []any{coverage}}}}},
	})
	_, err := repository.Read(context.Background(), validQuery())
	if !errors.Is(err, ErrTemporalHistoryUnavailable) {
		t.Fatalf("error = %v, want %v", err, ErrTemporalHistoryUnavailable)
	}
}

func TestPostgresRepositoryRollsBackFailedSnapshot(t *testing.T) {
	client := &fakeClient{
		results: []*fakeRows{{rows: []fakeRow{{values: []any{historicalReadTestTime()}}}}},
		errs:    []error{nil, errors.New("read failed")},
	}
	managed := &fakeManagedSnapshot{fakeClient: client}
	repository := newManagedPostgresRepository(
		&fakeSnapshotBeginner{snapshot: managed},
	)

	_, err := repository.Read(context.Background(), validQuery())
	if err == nil || !managed.rolledBack || managed.committed {
		t.Fatalf("unexpected failure lifecycle: err=%v snapshot=%#v", err, managed)
	}
}

func TestPostgresSQLUsesCorrectTemporalSemantics(t *testing.T) {
	checks := map[string][]string{
		"flights": {
			"last_seen_at > $1",
			"historical_read_flight_versions",
			"COUNT(*) OVER ()",
		},
		"trajectories": {
			"end_time > $1",
			"historical_read_trajectory_versions",
			"round(quality_score, 12)",
		},
		"observations": {
			"observed_at >= $1",
			"observed_at < $2",
			"round(latitude, 8)",
		},
		"routes": {
			"DISTINCT ON (result.trajectory_id)",
			"trajectory.start_time < $2",
			"trajectory.end_time > $1",
			"result.as_of_time <= $3",
			"cumulative_payload_bytes <= $5",
		},
	}
	queries := map[string]string{
		"flights":      readFlightsSQL,
		"trajectories": readTrajectoriesSQL,
		"observations": readObservationsSQL,
		"routes":       readRoutesSQL,
	}

	for name, fragments := range checks {
		for _, fragment := range fragments {
			if !strings.Contains(queries[name], fragment) {
				t.Fatalf("%s SQL misses %q: %s", name, fragment, queries[name])
			}
		}
	}
	if strings.Contains(readRoutesSQL, "as_of_time >= $1") {
		t.Fatal("route SQL still filters event membership by calculation time")
	}
}

func TestRecordValidationRejectsInvalidAlternativeExecutorValues(t *testing.T) {
	query := validQuery()
	trajectory := TrajectoryRecord{
		ID: "trajectory-1", ICAO24: "abc123", SourceName: "test",
		StartTime:    query.Window.StartTime,
		EndTime:      query.Window.EndTime,
		UpdatedAt:    query.Window.AsOfTime,
		QualityScore: 1.1,
	}
	if !errors.Is(validateTrajectoryRecord(trajectory, query, 0), ErrRecordInvalid) {
		t.Fatal("invalid quality score was accepted")
	}

	latitude := 91.0
	observation := ObservationRecord{
		ID: "state-1", ICAO24: "abc123", SourceName: "test",
		ObservedAt: query.Window.StartTime,
		CreatedAt:  query.Window.StartTime,
		Latitude:   &latitude,
	}
	if !errors.Is(validateObservationRecord(observation, query, 0), ErrRecordInvalid) {
		t.Fatal("invalid latitude was accepted")
	}
}

func TestRepresentedCoverageUsesExactDenominator(t *testing.T) {
	if got := RepresentedCoverage(10_000, 1_000_000); got != 0.01 {
		t.Fatalf("coverage = %f, want 0.01", got)
	}
	if got := RepresentedCoverage(0, 0); got != 1 {
		t.Fatalf("empty complete coverage = %f, want 1", got)
	}
}

func TestSnapshotTotalForSourcePreservesLegacyFixtureSemantics(
	t *testing.T,
) {
	snapshot := Snapshot{
		Flights: []FlightRecord{{ID: "flight-1"}},
	}
	if got := snapshot.TotalForSource(DatasetFlights); got != 1 {
		t.Fatalf("unlimited inferred total = %d, want 1", got)
	}

	snapshot.FlightLimitReached = true
	if got := snapshot.TotalForSource(DatasetFlights); got != 2 {
		t.Fatalf("limited inferred total = %d, want 2", got)
	}

	snapshot.FlightMatchedCount = 1_000_000
	if got := snapshot.TotalForSource(DatasetFlights); got != 1_000_000 {
		t.Fatalf("explicit total = %d, want 1000000", got)
	}
}

func TestTrimToLimit(t *testing.T) {
	items, limited := trimToLimit([]int{1, 2}, 1)
	if !limited || !reflect.DeepEqual(items, []int{1}) {
		t.Fatalf("unexpected trim result: %#v %t", items, limited)
	}
}

func validFakeClient(t *testing.T, flightTotal int64) *fakeClient {
	t.Helper()
	query := validQuery()
	base := query.Window.StartTime
	routeResult := routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  "trajectory-1",
		Window: routecontract.RouteWindow{
			StartTime: base,
			EndTime:   base.Add(30 * time.Minute),
			AsOfTime:  query.Window.AsOfTime,
		},
		Confidence: routecontract.Confidence{
			Score: 1,
			Level: routecontract.ConfidenceLevelHigh,
		},
	}
	payload, err := json.Marshal(routeResult)
	if err != nil {
		t.Fatalf("marshal route fixture: %v", err)
	}

	return &fakeClient{results: []*fakeRows{
		{rows: []fakeRow{{values: []any{base.Add(-time.Hour)}}}},
		{rows: []fakeRow{
			{values: []any{
				"flight-1", nil, "J2001", "completed",
				base, base.Add(30 * time.Minute), query.Window.AsOfTime, flightTotal,
			}},
			{values: []any{
				"flight-2", "aircraft-2", nil, "completed",
				base.Add(time.Minute), base.Add(40 * time.Minute), query.Window.AsOfTime, flightTotal,
			}},
		}},
		{rows: []fakeRow{{values: []any{
			"trajectory-1", "flight-1", nil, "abc123", "J2001",
			base, base.Add(30 * time.Minute), 1, 2, 0, 1.0,
			"test", query.Window.AsOfTime, int64(1),
		}}}},
		{rows: []fakeRow{{values: []any{
			"state-1", "flight-1", nil, "abc123", nil,
			40.1, 49.9, false, base.Add(time.Minute), "test", base.Add(time.Minute), int64(1),
		}}}},
		{rows: []fakeRow{{values: []any{
			"route-record-1", "trajectory-1", base, base.Add(30 * time.Minute), query.Window.AsOfTime,
			"sha256:" + strings.Repeat("a", 64), "complete", "high", 0,
			payload, query.Window.AsOfTime, int64(len(payload)), int64(1), int64(len(payload)), false,
		}}}},
	}}
}

func validQuery() Query {
	start := historicalReadTestTime()
	end := start.Add(time.Hour)
	return Query{
		Window: historicalcontract.TimeWindow{
			StartTime: start,
			EndTime:   end,
			AsOfTime:  end.Add(time.Hour),
		},
		Limit:                 1,
		RoutePayloadByteLimit: DefaultRoutePayloadByteLimit,
	}
}

func historicalReadTestTime() time.Time {
	return time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)
}

func assignFakeValue(destination any, value any) error {
	switch target := destination.(type) {
	case *string:
		*target = value.(string)
	case *int:
		*target = value.(int)
	case *int64:
		*target = value.(int64)
	case *float64:
		*target = value.(float64)
	case *bool:
		*target = value.(bool)
	case *time.Time:
		*target = value.(time.Time)
	case *[]byte:
		if value == nil {
			*target = nil
		} else {
			*target = append([]byte(nil), value.([]byte)...)
		}
	case **float64:
		if value == nil {
			*target = nil
		} else {
			copied := value.(float64)
			*target = &copied
		}
	case **bool:
		if value == nil {
			*target = nil
		} else {
			copied := value.(bool)
			*target = &copied
		}
	case *pgtype.Text:
		if value == nil {
			*target = pgtype.Text{}
		} else {
			*target = pgtype.Text{String: value.(string), Valid: true}
		}
	default:
		return errors.New("unsupported fake destination")
	}
	return nil
}
