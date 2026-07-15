package historicalmaterialization

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
)

type fakeReadRepository struct {
	snapshot historicalread.Snapshot
	queries  []historicalread.Query
	err      error
}

func (repository *fakeReadRepository) Read(
	_ context.Context,
	query historicalread.Query,
) (historicalread.Snapshot, error) {
	repository.queries = append(
		repository.queries,
		query,
	)
	if repository.err != nil {
		return historicalread.Snapshot{},
			repository.err
	}

	result := repository.snapshot.Clone()
	result.Query = query
	return result, nil
}

type fakeAggregateStore struct {
	results []historicalcontract.Result
	err     error
}

func (store *fakeAggregateStore) Put(
	_ context.Context,
	result historicalcontract.Result,
) (historicalaggregate.Record, error) {
	if store.err != nil {
		return historicalaggregate.Record{},
			store.err
	}
	store.results = append(
		store.results,
		result.Clone(),
	)

	return historicalaggregate.Record{
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
		Result:   result.Clone(),
		StoredAt: materializationTestTime(),
	}, nil
}

func (store *fakeAggregateStore) Get(
	context.Context,
	historicalaggregate.ResultKey,
) (historicalaggregate.Record, error) {
	return historicalaggregate.Record{},
		historicalaggregate.ErrResultNotFound
}

func (store *fakeAggregateStore) GetLatest(
	context.Context,
	historicalaggregate.ListQuery,
) (historicalaggregate.Record, error) {
	return historicalaggregate.Record{},
		historicalaggregate.ErrResultNotFound
}

func (store *fakeAggregateStore) List(
	context.Context,
	historicalaggregate.ListQuery,
) (historicalaggregate.Page, error) {
	return historicalaggregate.Page{}, nil
}

func TestMaterializeBuildsComparisonAndPersistsAggregate(
	t *testing.T,
) {
	asOfTime := materializationTestTime()
	repository := &fakeReadRepository{
		snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Flights: []historicalread.FlightRecord{
				{
					ID: "previous-flight",
					FirstSeenAt: asOfTime.
						Add(-3*time.Hour - 30*time.Minute),
					UpdatedAt: asOfTime.
						Add(-3 * time.Hour),
				},
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
		},
	}
	store := &fakeAggregateStore{}
	materializer, err := New(
		Config{
			Repository: repository,
			Store:      store,
			Now: func() time.Time {
				return asOfTime
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create materializer: %v",
			err,
		)
	}

	outcome, err := materializer.Materialize(
		context.Background(),
		Request{
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
		},
	)
	if err != nil {
		t.Fatalf(
			"materialize historical traffic: %v",
			err,
		)
	}

	if len(repository.queries) != 1 {
		t.Fatalf(
			"read call count=%d want=1",
			len(repository.queries),
		)
	}
	readWindow := repository.queries[0].Window
	if !readWindow.StartTime.Equal(
		asOfTime.Add(-4*time.Hour),
	) ||
		!readWindow.EndTime.Equal(asOfTime) {
		t.Fatalf(
			"unexpected combined read window: %#v",
			readWindow,
		)
	}
	if repository.queries[0].Limit != 100 {
		t.Fatalf(
			"read limit=%d want=100",
			repository.queries[0].Limit,
		)
	}

	if outcome.CurrentResult.Comparison == nil {
		t.Fatal(
			"expected adjacent-period comparison",
		)
	}
	comparison := outcome.CurrentResult.Comparison
	if comparison.CurrentValue != 2 ||
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
			"expected one-hundred-percent increase, got %#v",
			comparison.PercentageChange,
		)
	}

	if len(store.results) != 1 ||
		store.results[0].Comparison == nil {
		t.Fatalf(
			"expected compared aggregate to be persisted: %#v",
			store.results,
		)
	}
	if store.results[0].Provenance.InputFingerprint ==
		outcome.PreviousResult.Provenance.InputFingerprint {
		t.Fatal(
			"materialized comparison must use a combined current-and-previous fingerprint",
		)
	}
	if !strings.Contains(
		store.results[0].Provenance.BuilderVersion,
		Version,
	) || !strings.Contains(
		store.results[0].Provenance.BuilderVersion,
		"historical-period-comparison-v1",
	) {
		t.Fatalf(
			"unexpected materialized builder version: %q",
			store.results[0].Provenance.BuilderVersion,
		)
	}
	if outcome.Record.Result.Comparison == nil ||
		outcome.Record.Result.Summary.Total != 2 {
		t.Fatalf(
			"unexpected stored aggregate: %#v",
			outcome.Record,
		)
	}
	if outcome.ReadSummary.FlightCount != 3 {
		t.Fatalf(
			"read summary flight count=%d want=3",
			outcome.ReadSummary.FlightCount,
		)
	}
}

func TestMaterializeRejectsMetricScopeMismatch(
	t *testing.T,
) {
	materializer, err := New(
		Config{
			Repository: &fakeReadRepository{},
			Store:      &fakeAggregateStore{},
			Now:        materializationTestTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"create materializer: %v",
			err,
		)
	}

	_, err = materializer.Materialize(
		context.Background(),
		Request{
			StartTime: materializationTestTime().
				Add(-time.Hour),
			EndTime:  materializationTestTime(),
			AsOfTime: materializationTestTime(),
			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameAirportOperations,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			GeneratedAt: materializationTestTime(),
		},
	)
	if !errors.Is(
		err,
		ErrScopeUnsupported,
	) {
		t.Fatalf(
			"expected scope mismatch error, got %v",
			err,
		)
	}
}

func TestMaterializeRejectsWindowWithoutCompleteBucket(
	t *testing.T,
) {
	asOfTime := materializationTestTime().
		Add(30 * time.Minute)
	materializer, err := New(
		Config{
			Repository: &fakeReadRepository{},
			Store:      &fakeAggregateStore{},
			Now: func() time.Time {
				return asOfTime
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create materializer: %v",
			err,
		)
	}

	_, err = materializer.Materialize(
		context.Background(),
		Request{
			StartTime: asOfTime.
				Add(-20 * time.Minute),
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
			GeneratedAt: asOfTime,
		},
	)
	if !errors.Is(
		err,
		ErrNoEffectiveWindow,
	) {
		t.Fatalf(
			"expected no effective window error, got %v",
			err,
		)
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

func TestMaterializationFingerprintIncludesPreviousEvidence(
	t *testing.T,
) {
	current := historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Metric: historicalcontract.Metric{
			Name: historicalcontract.MetricNameFlightCount,
		},
		Scope: historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		},
		Granularity: historicalcontract.GranularityHour,
		Window: historicalcontract.TimeWindow{
			StartTime: materializationTestTime().Add(-time.Hour),
			EndTime:   materializationTestTime(),
			AsOfTime:  materializationTestTime(),
		},
		Provenance: historicalcontract.Provenance{
			BuilderVersion:   "current-builder",
			InputFingerprint: "sha256:" + strings.Repeat("a", 64),
		},
	}
	previous := current.Clone()
	previous.Window = historicalcontract.TimeWindow{
		StartTime: materializationTestTime().Add(-2 * time.Hour),
		EndTime:   materializationTestTime().Add(-time.Hour),
		AsOfTime:  materializationTestTime(),
	}
	previous.Provenance.BuilderVersion = "previous-builder"
	previous.Provenance.InputFingerprint =
		"sha256:" + strings.Repeat("b", 64)

	first := materializationFingerprint(current, previous)
	second := materializationFingerprint(current, previous)
	if first != second {
		t.Fatal("materialization fingerprint must be deterministic")
	}

	changedPrevious := previous.Clone()
	changedPrevious.Provenance.InputFingerprint =
		"sha256:" + strings.Repeat("c", 64)
	if first == materializationFingerprint(current, changedPrevious) {
		t.Fatal(
			"materialization fingerprint must change when previous-period evidence changes",
		)
	}
}
