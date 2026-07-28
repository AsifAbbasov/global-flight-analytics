package historicalread

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestPostgresRepositoryReadPeriodsUsesOneManagedSnapshotAndIndependentLimits(
	t *testing.T,
) {
	current := validQuery()
	current.Limit = 3
	previous := current
	previous.Window = historicalcontract.TimeWindow{
		StartTime: current.Window.StartTime.
			Add(-time.Hour),
		EndTime:  current.Window.StartTime,
		AsOfTime: current.Window.AsOfTime,
	}
	previous.Limit = 1

	client := emptyPeriodFakeClient(
		previous.Window.StartTime.
			Add(-time.Hour),
	)
	managed := &fakeManagedSnapshot{
		fakeClient: client,
	}
	beginner := &fakeSnapshotBeginner{
		snapshot: managed,
	}
	repository :=
		newManagedPostgresRepository(beginner)

	snapshots, err := repository.ReadPeriods(
		context.Background(),
		PeriodQueries{
			Previous: previous,
			Current:  current,
		},
	)
	if err != nil {
		t.Fatalf(
			"ReadPeriods() error = %v",
			err,
		)
	}

	if beginner.beginCount != 1 ||
		!managed.committed ||
		managed.rolledBack {
		t.Fatalf(
			"unexpected transaction lifecycle: %#v %#v",
			beginner,
			managed,
		)
	}
	if len(client.calls) != 10 {
		t.Fatalf(
			"query call count=%d want=10",
			len(client.calls),
		)
	}
	if got := client.calls[1].args[3]; got != 2 {
		t.Fatalf(
			"previous flight sentinel limit=%v want=2",
			got,
		)
	}
	if got := client.calls[6].args[3]; got != 4 {
		t.Fatalf(
			"current flight sentinel limit=%v want=4",
			got,
		)
	}
	if !snapshots.Previous.Query.Equal(
		previous,
	) || !snapshots.Current.Query.Equal(
		current,
	) {
		t.Fatalf(
			"period query metadata mismatch: %#v",
			snapshots,
		)
	}
	if snapshots.Previous.IsolationLevel !=
		SnapshotIsolationRepeatableRead ||
		snapshots.Current.IsolationLevel !=
			SnapshotIsolationRepeatableRead {
		t.Fatalf(
			"unexpected period isolation: %#v",
			snapshots,
		)
	}
}

func TestPostgresRepositoryReadPeriodsRejectsNonAdjacentWindows(
	t *testing.T,
) {
	current := validQuery()
	previous := current
	previous.Window.StartTime =
		current.Window.StartTime.Add(-2 * time.Hour)
	previous.Window.EndTime =
		current.Window.StartTime.Add(-time.Minute)

	repository := newPostgresRepository(
		&fakeClient{},
	)
	_, err := repository.ReadPeriods(
		context.Background(),
		PeriodQueries{
			Previous: previous,
			Current:  current,
		},
	)
	if !errors.Is(
		err,
		ErrPeriodWindowsNotAdjacent,
	) {
		t.Fatalf(
			"error=%v want=%v",
			err,
			ErrPeriodWindowsNotAdjacent,
		)
	}
}

func TestPostgresRepositoryReadPeriodsRejectsAsOfTimeMismatch(
	t *testing.T,
) {
	current := validQuery()
	previous := current
	previous.Window = historicalcontract.TimeWindow{
		StartTime: current.Window.StartTime.
			Add(-time.Hour),
		EndTime: current.Window.StartTime,
		AsOfTime: current.Window.AsOfTime.
			Add(-time.Second),
	}

	repository := newPostgresRepository(
		&fakeClient{},
	)
	_, err := repository.ReadPeriods(
		context.Background(),
		PeriodQueries{
			Previous: previous,
			Current:  current,
		},
	)
	if !errors.Is(
		err,
		ErrPeriodAsOfTimeMismatch,
	) {
		t.Fatalf(
			"error=%v want=%v",
			err,
			ErrPeriodAsOfTimeMismatch,
		)
	}
}

func TestPostgresRepositoryReadPeriodsRollsBackCurrentPeriodFailure(
	t *testing.T,
) {
	current := validQuery()
	previous := current
	previous.Window = historicalcontract.TimeWindow{
		StartTime: current.Window.StartTime.
			Add(-time.Hour),
		EndTime:  current.Window.StartTime,
		AsOfTime: current.Window.AsOfTime,
	}

	client := emptyPeriodFakeClient(
		previous.Window.StartTime.
			Add(-time.Hour),
	)
	readFailure := errors.New(
		"current period read failed",
	)
	client.errs = make([]error, 10)
	client.errs[6] = readFailure

	managed := &fakeManagedSnapshot{
		fakeClient: client,
	}
	repository := newManagedPostgresRepository(
		&fakeSnapshotBeginner{
			snapshot: managed,
		},
	)

	_, err := repository.ReadPeriods(
		context.Background(),
		PeriodQueries{
			Previous: previous,
			Current:  current,
		},
	)
	if err == nil ||
		!errors.Is(err, readFailure) ||
		!managed.rolledBack ||
		managed.committed {
		t.Fatalf(
			"unexpected failure lifecycle: err=%v snapshot=%#v",
			err,
			managed,
		)
	}
}

func TestPostgresRepositoryReadPeriodsRejectsNilContext(
	t *testing.T,
) {
	repository := newPostgresRepository(
		&fakeClient{},
	)
	_, err := repository.ReadPeriods(
		nil,
		PeriodQueries{},
	)
	if !errors.Is(
		err,
		ErrContextRequired,
	) {
		t.Fatalf(
			"error=%v want=%v",
			err,
			ErrContextRequired,
		)
	}
}

func emptyPeriodFakeClient(
	coverageStartedAt time.Time,
) *fakeClient {
	results := make(
		[]*fakeRows,
		0,
		10,
	)
	for period := 0; period < 2; period++ {
		results = append(
			results,
			&fakeRows{
				rows: []fakeRow{
					{
						values: []any{
							coverageStartedAt,
						},
					},
				},
			},
			&fakeRows{},
			&fakeRows{},
			&fakeRows{},
			&fakeRows{},
		)
	}
	return &fakeClient{
		results: results,
	}
}
