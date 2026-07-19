package projectionread

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestLoadCurrentTrajectoryHydratesPointsAtPersistedEnd(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	persistedEnd := asOfTime.Add(
		-2 * time.Minute,
	)
	item := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		persistedEnd,
	)
	item.Points = nil
	item.PointCount = 3

	client := &scriptedClient{
		rowsQueue: []*scriptedRows{
			{
				values: [][]any{
					projectionReadPointRow(
						"state-a",
						item.StartTime,
						40.40,
						49.80,
					),
					projectionReadPointRow(
						"state-b",
						persistedEnd,
						40.50,
						50.00,
					),
				},
			},
		},
	}
	repository := &trajectoryRepositoryStub{
		items: map[string]trajectory.FlightTrajectory{
			item.ID: item,
		},
		errs: map[string]error{},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		repository,
	)

	result, err :=
		source.LoadCurrentTrajectory(
			context.Background(),
			item.ID,
			asOfTime,
		)
	if err != nil {
		t.Fatalf(
			"LoadCurrentTrajectory() error = %v",
			err,
		)
	}

	if len(result.Points) != 2 ||
		result.PointCount != 2 ||
		!result.EndTime.Equal(
			persistedEnd,
		) ||
		result.Points[0].FlightStateID !=
			"state-a" {
		t.Fatalf(
			"unexpected hydrated trajectory: %#v",
			result,
		)
	}
	if len(client.queryCalls) != 1 {
		t.Fatalf(
			"point query calls = %d, want 1",
			len(client.queryCalls),
		)
	}
	if len(client.queryCalls[0].args) != 4 {
		t.Fatalf(
			"point query args = %#v",
			client.queryCalls[0].args,
		)
	}
	cutoff, ok := client.queryCalls[0].args[2].(time.Time)
	if !ok ||
		!cutoff.Equal(persistedEnd) {
		t.Fatalf(
			"point cutoff = %#v, want %s",
			client.queryCalls[0].args[2],
			persistedEnd,
		)
	}
}

func TestLoadRouteUsesLatestResultAtOrBeforeAsOf(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	payload, err := json.Marshal(route)
	if err != nil {
		t.Fatalf(
			"marshal route fixture: %v",
			err,
		)
	}

	client := &scriptedClient{
		rowQueue: []scriptedRow{
			{
				values: []any{
					payload,
				},
			},
		},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		&trajectoryRepositoryStub{
			items: map[string]trajectory.FlightTrajectory{},
			errs:  map[string]error{},
		},
	)

	result, err := source.LoadRoute(
		context.Background(),
		current.ID,
		asOfTime,
	)
	if err != nil {
		t.Fatalf(
			"LoadRoute() error = %v",
			err,
		)
	}
	if result.TrajectoryID != current.ID ||
		result.Status !=
			routecontract.RouteStatusComplete {
		t.Fatalf(
			"unexpected route result: %#v",
			result,
		)
	}
	if len(client.queryRowCalls) != 1 ||
		!strings.Contains(
			client.queryRowCalls[0].query,
			"as_of_time <= $3",
		) {
		t.Fatalf(
			"route query was not bounded by as-of time: %#v",
			client.queryRowCalls,
		)
	}
}

func TestLoadHistoricalCandidatesUsesRouteScopedBoundedSelection(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	candidate := projectionReadTrajectory(
		"83aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime.Add(
			-24*time.Hour,
		),
	)
	candidate.Points = nil
	candidate.PointCount = 3

	client := &scriptedClient{
		rowsQueue: []*scriptedRows{
			{
				values: [][]any{
					{
						candidate.ID,
					},
				},
			},
			{
				values: [][]any{
					projectionReadPointRow(
						"candidate-state-a",
						candidate.StartTime,
						40.10,
						49.20,
					),
					projectionReadPointRow(
						"candidate-state-b",
						candidate.EndTime,
						40.20,
						49.30,
					),
				},
			},
		},
	}
	repository := &trajectoryRepositoryStub{
		items: map[string]trajectory.FlightTrajectory{
			candidate.ID: candidate,
		},
		errs: map[string]error{},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		repository,
	)

	result, err :=
		source.LoadHistoricalCandidates(
			context.Background(),
			current,
			route,
			asOfTime,
		)
	if err != nil {
		t.Fatalf(
			"LoadHistoricalCandidates() error = %v",
			err,
		)
	}
	if len(result) != 1 ||
		result[0].ID != candidate.ID ||
		len(result[0].Points) != 2 {
		t.Fatalf(
			"unexpected historical candidates: %#v",
			result,
		)
	}
	if len(client.queryCalls) != 2 {
		t.Fatalf(
			"query calls = %d, want 2",
			len(client.queryCalls),
		)
	}
	selectionCall := client.queryCalls[0]
	for _, fragment := range []string{
		"{Origin,Airport,ICAOCode}",
		"{Destination,Airport,ICAOCode}",
		"trajectory.end_time < $7",
		"LIMIT $8",
	} {
		if !strings.Contains(
			selectionCall.query,
			fragment,
		) {
			t.Fatalf(
				"candidate selection query is missing %q",
				fragment,
			)
		}
	}
	if selectionCall.args[4] != "UBBB" ||
		selectionCall.args[5] != "LTBA" {
		t.Fatalf(
			"candidate route scope args = %#v",
			selectionCall.args,
		)
	}
}

func TestLoadRouteHistoryBuildsDeterministicValidatedSummary(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	route := projectionReadCompleteRoute(
		current,
		asOfTime,
	)
	lastObservedAt := asOfTime.Add(
		-24 * time.Hour,
	)

	client := &scriptedClient{
		rowQueue: []scriptedRow{
			{
				values: []any{
					int64(10),
					int64(8),
					int64(7),
					int64(4),
					pgtype.Timestamptz{
						Time:  lastObservedAt,
						Valid: true,
					},
				},
			},
			{
				values: []any{
					int64(10),
					int64(8),
					int64(7),
					int64(4),
					pgtype.Timestamptz{
						Time:  lastObservedAt,
						Valid: true,
					},
				},
			},
		},
	}
	source := newProjectionReadTestSource(
		t,
		client,
		&trajectoryRepositoryStub{
			items: map[string]trajectory.FlightTrajectory{},
			errs:  map[string]error{},
		},
	)

	first, err := source.LoadRouteHistory(
		context.Background(),
		route,
		asOfTime,
	)
	if err != nil {
		t.Fatalf(
			"first LoadRouteHistory() error = %v",
			err,
		)
	}
	second, err := source.LoadRouteHistory(
		context.Background(),
		route,
		asOfTime,
	)
	if err != nil {
		t.Fatalf(
			"second LoadRouteHistory() error = %v",
			err,
		)
	}

	if first.RouteKey != "UBBB>LTBA" ||
		first.ObservationCount != 10 ||
		first.DistinctFlightCount != 8 ||
		first.DistinctDayCount != 7 ||
		first.RecentObservationCount != 4 ||
		!first.LastObservedAt.Equal(
			lastObservedAt,
		) ||
		first.InputFingerprint !=
			second.InputFingerprint {
		t.Fatalf(
			"unexpected route-history summary: first=%#v second=%#v",
			first,
			second,
		)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf(
			"route-history validation error = %v",
			err,
		)
	}
	if len(client.queryRowCalls) != 2 ||
		!strings.Contains(
			client.queryRowCalls[0].query,
			"as_of_time <= $3",
		) ||
		strings.Contains(
			strings.ToUpper(
				client.queryRowCalls[0].query,
			),
			"NOW()",
		) {
		t.Fatalf(
			"route-history query is not reproducibly bounded: %#v",
			client.queryRowCalls,
		)
	}
}

func TestPostgresQueriesNeverUseDatabaseCurrentTime(
	t *testing.T,
) {
	for name, query := range map[string]string{
		"route":              routeAtOrBeforeSQL,
		"candidates":         historicalCandidateIDsSQL,
		"history":            routeHistorySummarySQL,
		"points by flight":   trajectoryPointsByFlightSQL,
		"points by aircraft": trajectoryPointsByAircraftSQL,
	} {
		upper := strings.ToUpper(query)
		for _, forbidden := range []string{
			"NOW()",
			"CURRENT_TIMESTAMP",
			"LOCALTIMESTAMP",
		} {
			if strings.Contains(
				upper,
				forbidden,
			) {
				t.Fatalf(
					"%s query contains forbidden current-time expression %s",
					name,
					forbidden,
				)
			}
		}
	}
}

func newProjectionReadTestSource(
	t *testing.T,
	client postgresClient,
	repository trajectoryRepository,
) *PostgresDataSource {
	t.Helper()

	source, err := newPostgresDataSource(
		client,
		repository,
		DefaultPolicy().DataSource,
	)
	if err != nil {
		t.Fatalf(
			"newPostgresDataSource() error = %v",
			err,
		)
	}

	return source
}

func projectionReadPointRow(
	id string,
	observedAt time.Time,
	latitude float64,
	longitude float64,
) []any {
	return []any{
		id,
		"6b57d421-9f75-4f1b-931d-d4e658515d92",
		"a20eef16-c12c-41fd-870e-cd5a814ef3ad",
		"4A1234",
		"AHY123",
		pgtype.Float8{
			Float64: latitude,
			Valid:   true,
		},
		pgtype.Float8{
			Float64: longitude,
			Valid:   true,
		},
		pgtype.Float8{
			Float64: 10000,
			Valid:   true,
		},
		string(
			flightstate.
				AltitudeStatusObserved,
		),
		pgtype.Float8{
			Float64: 10100,
			Valid:   true,
		},
		string(
			flightstate.
				AltitudeStatusObserved,
		),
		pgtype.Float8{
			Float64: 220,
			Valid:   true,
		},
		pgtype.Float8{
			Float64: 270,
			Valid:   true,
		},
		pgtype.Float8{
			Float64: 0,
			Valid:   true,
		},
		pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
		"Azerbaijan",
		observedAt.UTC(),
		"projection-read-test",
	}
}
