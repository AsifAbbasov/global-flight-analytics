package historicalmaterialization

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
)

type fakePeriodRepository struct {
	snapshots historicalread.PeriodSnapshots
	queries   []historicalread.PeriodQueries
	err       error
	mutate    func(
		*historicalread.PeriodSnapshots,
		historicalread.PeriodQueries,
	)
}

func (repository *fakePeriodRepository) Read(
	_ context.Context,
	_ historicalread.Query,
) (historicalread.Snapshot, error) {
	return historicalread.Snapshot{},
		errors.New(
			"legacy combined read must not be used",
		)
}

func (repository *fakePeriodRepository) ReadPeriods(
	_ context.Context,
	queries historicalread.PeriodQueries,
) (historicalread.PeriodSnapshots, error) {
	repository.queries = append(
		repository.queries,
		queries,
	)
	if repository.err != nil {
		return historicalread.PeriodSnapshots{},
			repository.err
	}

	result := repository.snapshots.Clone()
	result.Previous.Version = historicalread.Version
	result.Previous.IsolationLevel =
		historicalread.SnapshotIsolationRepeatableRead
	result.Previous.Query = queries.Previous
	result.Current.Version = historicalread.Version
	result.Current.IsolationLevel =
		historicalread.SnapshotIsolationRepeatableRead
	result.Current.Query = queries.Current
	if repository.mutate != nil {
		repository.mutate(
			&result,
			queries,
		)
	}
	return result.Clone(), nil
}

type fakeAggregateWriter struct {
	results []historicalcontract.Result
	err     error
	mutate  func(*historicalaggregate.Record)
}

func (writer *fakeAggregateWriter) Put(
	_ context.Context,
	result historicalcontract.Result,
) (historicalaggregate.Record, error) {
	if writer.err != nil {
		return historicalaggregate.Record{},
			writer.err
	}
	writer.results = append(
		writer.results,
		result.Clone(),
	)

	record := historicalaggregate.Record{
		ID: "historical-aggregate-record-" +
			strings.Repeat("a", 64),
		Key: historicalaggregate.ResultKey{
			SchemaVersion: result.SchemaVersion,
			MetricName:    result.Metric.Name,
			Scope:         result.Scope,
			Granularity:   result.Granularity,
			Window:        result.Window,
		},
		InputFingerprint: result.
			Provenance.InputFingerprint,
		Result: result.Clone(),
		StoredAt: result.GeneratedAt.
			Add(time.Second),
	}
	if writer.mutate != nil {
		writer.mutate(&record)
	}
	return record.Clone(), nil
}

func TestMaterializeReadsAdjacentPeriodsIndependentlyAndPersistsComparison(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	repository := &fakePeriodRepository{
		snapshots: materializationPeriodSnapshots(
			asOfTime,
		),
	}
	writer := &fakeAggregateWriter{}
	materializer := newTestMaterializer(
		t,
		repository,
		writer,
		asOfTime,
	)

	outcome, err := materializer.Materialize(
		context.Background(),
		materializationRequest(asOfTime),
	)
	if err != nil {
		t.Fatalf(
			"materialize historical traffic: %v",
			err,
		)
	}

	if len(repository.queries) != 1 {
		t.Fatalf(
			"period read call count=%d want=1",
			len(repository.queries),
		)
	}
	queries := repository.queries[0]
	if !queries.Previous.Window.StartTime.Equal(
		asOfTime.Add(-4*time.Hour),
	) ||
		!queries.Previous.Window.EndTime.Equal(
			asOfTime.Add(-2*time.Hour),
		) ||
		!queries.Current.Window.StartTime.Equal(
			asOfTime.Add(-2*time.Hour),
		) ||
		!queries.Current.Window.EndTime.Equal(
			asOfTime,
		) {
		t.Fatalf(
			"unexpected period windows: %#v",
			queries,
		)
	}
	if queries.Previous.Limit != 100 ||
		queries.Current.Limit != 100 {
		t.Fatalf(
			"period limits are not independent: %#v",
			queries,
		)
	}

	comparison := outcome.CurrentResult.Comparison
	if comparison == nil ||
		comparison.CurrentValue != 2 ||
		comparison.PreviousValue != 1 ||
		comparison.AbsoluteChange != 1 ||
		comparison.Direction !=
			historicalcontract.TrendDirectionUp {
		t.Fatalf(
			"unexpected comparison: %#v",
			comparison,
		)
	}
	if comparison.PercentageChange == nil ||
		*comparison.PercentageChange != 100 {
		t.Fatalf(
			"unexpected percentage: %#v",
			comparison.PercentageChange,
		)
	}
	if outcome.ReadSummaries.Previous.FlightCount !=
		1 ||
		outcome.ReadSummaries.Current.FlightCount !=
			2 {
		t.Fatalf(
			"period read summaries are not separated: %#v",
			outcome.ReadSummaries,
		)
	}
	if outcome.ReadSummary.FlightCount != 3 {
		t.Fatalf(
			"legacy aggregate read count=%d want=3",
			outcome.ReadSummary.FlightCount,
		)
	}
	if windowsEqual(
		outcome.ReadSummaries.Previous.Window,
		outcome.ReadSummaries.Current.Window,
	) {
		t.Fatalf(
			"period summaries share one window: %#v",
			outcome.ReadSummaries,
		)
	}
	if !reflect.DeepEqual(
		outcome.CurrentResult,
		outcome.Record.Result,
	) {
		t.Fatal(
			"current result must be the canonical persisted result",
		)
	}
	if !strings.Contains(
		outcome.CurrentResult.Provenance.BuilderVersion,
		"historical-period-comparison-v2",
	) || strings.Contains(
		outcome.CurrentResult.Provenance.BuilderVersion,
		Version,
	) {
		t.Fatalf(
			"comparison provenance was repaired outside its owner: %q",
			outcome.CurrentResult.Provenance.BuilderVersion,
		)
	}
	if outcome.CurrentPeriodResult.Comparison != nil {
		t.Fatal(
			"raw current-period result unexpectedly contains a comparison",
		)
	}
}

func TestMaterializeRejectsSnapshotContractViolations(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	tests := []struct {
		name   string
		kind   error
		mutate func(
			*historicalread.PeriodSnapshots,
			historicalread.PeriodQueries,
		)
	}{
		{
			name: "version",
			kind: ErrSnapshotVersionMismatch,
			mutate: func(
				snapshots *historicalread.PeriodSnapshots,
				_ historicalread.PeriodQueries,
			) {
				snapshots.Current.Version =
					"stale-version"
			},
		},
		{
			name: "query",
			kind: ErrSnapshotQueryMismatch,
			mutate: func(
				snapshots *historicalread.PeriodSnapshots,
				_ historicalread.PeriodQueries,
			) {
				snapshots.Current.Query.Limit++
			},
		},
		{
			name: "isolation",
			kind: ErrSnapshotIsolationMismatch,
			mutate: func(
				snapshots *historicalread.PeriodSnapshots,
				_ historicalread.PeriodQueries,
			) {
				snapshots.Current.IsolationLevel =
					historicalread.
						SnapshotIsolationCallerTransaction
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				repository :=
					&fakePeriodRepository{
						snapshots: materializationPeriodSnapshots(
							asOfTime,
						),
						mutate: test.mutate,
					}
				materializer := newTestMaterializer(
					t,
					repository,
					&fakeAggregateWriter{},
					asOfTime,
				)

				_, err := materializer.Materialize(
					context.Background(),
					materializationRequest(
						asOfTime,
					),
				)
				if !errors.Is(
					err,
					test.kind,
				) {
					t.Fatalf(
						"error=%v want=%v",
						err,
						test.kind,
					)
				}
				var staged *StageError
				if !errors.As(
					err,
					&staged,
				) ||
					staged.Stage !=
						StageSnapshotContract {
					t.Fatalf(
						"unexpected stage: %#v",
						staged,
					)
				}
			},
		)
	}
}

func TestMaterializeRejectsNilContext(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	materializer := newTestMaterializer(
		t,
		&fakePeriodRepository{
			snapshots: materializationPeriodSnapshots(
				asOfTime,
			),
		},
		&fakeAggregateWriter{},
		asOfTime,
	)

	_, err := materializer.Materialize(
		nil,
		materializationRequest(asOfTime),
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

func TestMaterializeWrapsRepositoryAndPersistenceStages(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	readFailure := errors.New("read failed")
	storeFailure := errors.New("store failed")

	tests := []struct {
		name       string
		repository *fakePeriodRepository
		writer     *fakeAggregateWriter
		cause      error
		stage      Stage
	}{
		{
			name: "read",
			repository: &fakePeriodRepository{
				err: readFailure,
			},
			writer: &fakeAggregateWriter{},
			cause:  readFailure,
			stage:  StagePeriodRead,
		},
		{
			name: "persistence",
			repository: &fakePeriodRepository{
				snapshots: materializationPeriodSnapshots(
					asOfTime,
				),
			},
			writer: &fakeAggregateWriter{
				err: storeFailure,
			},
			cause: storeFailure,
			stage: StagePersistence,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				materializer :=
					newTestMaterializer(
						t,
						test.repository,
						test.writer,
						asOfTime,
					)
				_, err :=
					materializer.Materialize(
						context.Background(),
						materializationRequest(
							asOfTime,
						),
					)
				if !errors.Is(
					err,
					test.cause,
				) {
					t.Fatalf(
						"error=%v want cause=%v",
						err,
						test.cause,
					)
				}
				var staged *StageError
				if !errors.As(
					err,
					&staged,
				) ||
					staged.Stage != test.stage {
					t.Fatalf(
						"unexpected stage: %#v",
						staged,
					)
				}
			},
		)
	}
}

func TestMaterializeUsesPersistedResultAsCanonicalOutcome(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	const canonicalSuffix = " Canonicalized by aggregate writer."
	writer := &fakeAggregateWriter{
		mutate: func(
			record *historicalaggregate.Record,
		) {
			if len(record.Result.Limitations) == 0 {
				panic(
					"comparison result has no limitation to canonicalize",
				)
			}
			record.Result.Limitations[0].Message +=
				canonicalSuffix
		},
	}
	materializer := newTestMaterializer(
		t,
		&fakePeriodRepository{
			snapshots: materializationPeriodSnapshots(
				asOfTime,
			),
		},
		writer,
		asOfTime,
	)

	outcome, err := materializer.Materialize(
		context.Background(),
		materializationRequest(asOfTime),
	)
	if err != nil {
		t.Fatalf(
			"materialize: %v",
			err,
		)
	}
	if !reflect.DeepEqual(
		outcome.CurrentResult,
		outcome.Record.Result,
	) {
		t.Fatal(
			"outcome ignored the canonical persisted result",
		)
	}
	if !strings.HasSuffix(
		outcome.CurrentResult.Limitations[0].Message,
		canonicalSuffix,
	) {
		t.Fatal(
			"persisted canonical representation is absent from outcome",
		)
	}
}

func TestMaterializeRejectsPersistedRecordIdentityMismatch(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	writer := &fakeAggregateWriter{
		mutate: func(
			record *historicalaggregate.Record,
		) {
			record.InputFingerprint =
				"sha256:" +
					strings.Repeat("f", 64)
		},
	}
	materializer := newTestMaterializer(
		t,
		&fakePeriodRepository{
			snapshots: materializationPeriodSnapshots(
				asOfTime,
			),
		},
		writer,
		asOfTime,
	)

	_, err := materializer.Materialize(
		context.Background(),
		materializationRequest(asOfTime),
	)
	if !errors.Is(
		err,
		ErrPersistedRecordMismatch,
	) {
		t.Fatalf(
			"error=%v want=%v",
			err,
			ErrPersistedRecordMismatch,
		)
	}
}

func TestMaterializeBindsGeneratedAtIntoFingerprint(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	repository := &fakePeriodRepository{
		snapshots: materializationPeriodSnapshots(
			asOfTime,
		),
	}
	writer := &fakeAggregateWriter{}
	materializer := newTestMaterializer(
		t,
		repository,
		writer,
		asOfTime,
	)

	firstRequest := materializationRequest(
		asOfTime,
	)
	secondRequest := firstRequest
	secondRequest.GeneratedAt =
		firstRequest.GeneratedAt.Add(time.Second)

	_, err := materializer.Materialize(
		context.Background(),
		firstRequest,
	)
	if err != nil {
		t.Fatalf(
			"first materialization: %v",
			err,
		)
	}
	_, err = materializer.Materialize(
		context.Background(),
		secondRequest,
	)
	if err != nil {
		t.Fatalf(
			"second materialization: %v",
			err,
		)
	}
	if len(writer.results) != 2 {
		t.Fatalf(
			"stored result count=%d want=2",
			len(writer.results),
		)
	}
	if writer.results[0].Provenance.
		InputFingerprint ==
		writer.results[1].Provenance.
			InputFingerprint {
		t.Fatal(
			"generated time is absent from materialization identity",
		)
	}
}

func TestMaterializeUsesDefaultAndMaximumDatasetLimits(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	tests := []struct {
		name      string
		requested int
		expected  int
	}{
		{
			name:      "default",
			requested: 0,
			expected: historicalread.
				DefaultDatasetLimit,
		},
		{
			name: "maximum",
			requested: historicalread.
				MaximumDatasetLimit,
			expected: historicalread.
				MaximumDatasetLimit,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				repository :=
					&fakePeriodRepository{
						snapshots: materializationPeriodSnapshots(
							asOfTime,
						),
					}
				materializer :=
					newTestMaterializer(
						t,
						repository,
						&fakeAggregateWriter{},
						asOfTime,
					)
				request :=
					materializationRequest(
						asOfTime,
					)
				request.DatasetLimit =
					test.requested

				_, err :=
					materializer.Materialize(
						context.Background(),
						request,
					)
				if err != nil {
					t.Fatalf(
						"materialize: %v",
						err,
					)
				}
				queries :=
					repository.queries[0]
				if queries.Previous.Limit !=
					test.expected ||
					queries.Current.Limit !=
						test.expected {
					t.Fatalf(
						"unexpected limits: %#v",
						queries,
					)
				}
			},
		)
	}
}

func TestMaterializeRejectsMetricScopeMismatch(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	materializer := newTestMaterializer(
		t,
		&fakePeriodRepository{},
		&fakeAggregateWriter{},
		asOfTime,
	)
	request := materializationRequest(asOfTime)
	request.MetricName =
		historicalcontract.
			MetricNameAirportOperations
	request.Scope = historicalcontract.Scope{
		Type: historicalcontract.ScopeTypeGlobal,
	}

	_, err := materializer.Materialize(
		context.Background(),
		request,
	)
	if !errors.Is(
		err,
		ErrScopeUnsupported,
	) {
		t.Fatalf(
			"error=%v want=%v",
			err,
			ErrScopeUnsupported,
		)
	}
}

func TestMaterializeRejectsWindowWithoutCompleteBucket(
	t *testing.T,
) {
	asOfTime := materializationTestTime().
		Add(30 * time.Minute)
	materializer := newTestMaterializer(
		t,
		&fakePeriodRepository{},
		&fakeAggregateWriter{},
		asOfTime,
	)
	request := materializationRequest(asOfTime)
	request.StartTime =
		asOfTime.Add(-20 * time.Minute)

	_, err := materializer.Materialize(
		context.Background(),
		request,
	)
	if !errors.Is(
		err,
		ErrNoEffectiveWindow,
	) {
		t.Fatalf(
			"error=%v want=%v",
			err,
			ErrNoEffectiveWindow,
		)
	}
}

func TestOutcomeCloneIsolatesNestedResults(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	materializer := newTestMaterializer(
		t,
		&fakePeriodRepository{
			snapshots: materializationPeriodSnapshots(
				asOfTime,
			),
		},
		&fakeAggregateWriter{},
		asOfTime,
	)
	outcome, err := materializer.Materialize(
		context.Background(),
		materializationRequest(asOfTime),
	)
	if err != nil {
		t.Fatalf(
			"materialize: %v",
			err,
		)
	}

	cloned := outcome.Clone()
	cloned.CurrentResult.Limitations =
		append(
			cloned.CurrentResult.Limitations,
			historicalcontract.Limitation{
				Code:    "clone-only",
				Message: "clone-only",
				Scope:   "test",
			},
		)
	if reflect.DeepEqual(
		cloned.CurrentResult,
		outcome.CurrentResult,
	) {
		t.Fatal(
			"outcome clone shares nested result slices",
		)
	}
}

func newTestMaterializer(
	t *testing.T,
	repository *fakePeriodRepository,
	writer historicalaggregate.Writer,
	now time.Time,
) *Materializer {
	t.Helper()
	materializer, err := New(
		Config{
			Repository: repository,
			Store:      writer,
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create materializer: %v",
			err,
		)
	}
	return materializer
}

func materializationRequest(
	asOfTime time.Time,
) Request {
	return Request{
		StartTime: asOfTime.
			Add(-2 * time.Hour),
		EndTime:  asOfTime,
		AsOfTime: asOfTime,
		Granularity: historicalcontract.
			GranularityHour,
		MetricName: historicalcontract.
			MetricNameFlightCount,
		Scope: historicalcontract.Scope{
			Type: historicalcontract.
				ScopeTypeGlobal,
		},
		DatasetLimit: 100,
		GeneratedAt:  asOfTime,
	}
}

func materializationPeriodSnapshots(
	asOfTime time.Time,
) historicalread.PeriodSnapshots {
	return historicalread.PeriodSnapshots{
		Previous: historicalread.Snapshot{
			Flights: []historicalread.FlightRecord{
				{
					ID: "previous-flight",
					FirstSeenAt: asOfTime.
						Add(-3*time.Hour -
							30*time.Minute),
					UpdatedAt: asOfTime.
						Add(-3 * time.Hour),
				},
			},
			FlightMatchedCount: 1,
		},
		Current: historicalread.Snapshot{
			Flights: []historicalread.FlightRecord{
				{
					ID: "current-flight-one",
					FirstSeenAt: asOfTime.
						Add(-90 * time.Minute),
					UpdatedAt: asOfTime.
						Add(-80 * time.Minute),
				},
				{
					ID: "current-flight-two",
					FirstSeenAt: asOfTime.
						Add(-30 * time.Minute),
					UpdatedAt: asOfTime.
						Add(-20 * time.Minute),
				},
			},
			FlightMatchedCount: 2,
		},
	}
}

func materializationTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}
