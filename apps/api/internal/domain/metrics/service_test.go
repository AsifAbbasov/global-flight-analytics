package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type activeAircraftRepositoryStub struct {
	summary ActiveAircraftObservationSummary
	err     error
	query   ActiveAircraftQuery
}

func (
	r *activeAircraftRepositoryStub,
) CountActiveAircraft(
	ctx context.Context,
	query ActiveAircraftQuery,
) (ActiveAircraftObservationSummary, error) {
	r.query = query

	return r.summary,
		r.err
}

func TestCalculateActiveAircraftBuildsExplainableHighConfidenceMetric(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		10,
		20,
		15,
		0,
		0,
		time.UTC,
	)

	repository := &activeAircraftRepositoryStub{
		summary: ActiveAircraftObservationSummary{
			Count:            2,
			FirstObservedAt:  now.Add(-5 * time.Minute),
			LatestObservedAt: now.Add(-1 * time.Minute),
			SourceNames: []string{
				"airplanes.live",
				"opensky",
			},
			HasObservations: true,
		},
	}

	service := mustNewServiceWithClock(
		repository,
		region.NewService(),
		func() time.Time {
			return now
		},
	)

	metric, err := service.CalculateActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			RegionCode:    "caucasus",
			WindowMinutes: 15,
		},
	)
	if err != nil {
		t.Fatalf(
			"calculate active aircraft metric: %v",
			err,
		)
	}

	if metric.Metric != ActiveAircraftMetricName {
		t.Fatalf(
			"expected metric %s, got %s",
			ActiveAircraftMetricName,
			metric.Metric,
		)
	}

	if metric.Value != 2 {
		t.Fatalf(
			"expected value 2, got %d",
			metric.Value,
		)
	}

	if metric.Scope.Type != MetricScopeRegion ||
		metric.Scope.Code != "caucasus" {
		t.Fatalf(
			"expected caucasus region scope, got %+v",
			metric.Scope,
		)
	}

	if !repository.query.Scope.IsBounded() {
		t.Fatal(
			"expected bounded repository query for region metric",
		)
	}

	if metric.Confidence.Level != ConfidenceLevelHigh {
		t.Fatalf(
			"expected high confidence, got %s",
			metric.Confidence.Level,
		)
	}

	if metric.Confidence.Score != 0.9 {
		t.Fatalf(
			"expected confidence score 0.9, got %f",
			metric.Confidence.Score,
		)
	}

	if len(metric.Sources) != 2 {
		t.Fatalf(
			"expected 2 sources, got %d",
			len(metric.Sources),
		)
	}

	if len(metric.Limitations) != 3 {
		t.Fatalf(
			"expected 3 limitations, got %d",
			len(metric.Limitations),
		)
	}
}

func TestCalculateActiveAircraftDefaultsToGlobalFifteenMinuteWindow(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		10,
		20,
		15,
		0,
		0,
		time.UTC,
	)

	repository := &activeAircraftRepositoryStub{
		summary: ActiveAircraftObservationSummary{},
	}

	service := mustNewServiceWithClock(
		repository,
		region.NewService(),
		func() time.Time {
			return now
		},
	)

	metric, err := service.CalculateActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{},
	)
	if err != nil {
		t.Fatalf(
			"calculate default active aircraft metric: %v",
			err,
		)
	}

	if metric.WindowMinutes != DefaultActiveAircraftWindowMinutes {
		t.Fatalf(
			"expected default window %d, got %d",
			DefaultActiveAircraftWindowMinutes,
			metric.WindowMinutes,
		)
	}

	if metric.Scope.Type != MetricScopeGlobal ||
		metric.Scope.Code != "world" {
		t.Fatalf(
			"expected global world scope, got %+v",
			metric.Scope,
		)
	}

	if repository.query.Scope.IsBounded() {
		t.Fatal(
			"expected unbounded repository query for global metric",
		)
	}

	expectedObservedFrom := now.Add(
		-DefaultActiveAircraftWindowMinutes * time.Minute,
	)
	if !repository.query.ObservedFrom.Equal(
		expectedObservedFrom,
	) {
		t.Fatalf(
			"expected observed_from %s, got %s",
			expectedObservedFrom,
			repository.query.ObservedFrom,
		)
	}

	if metric.Confidence.Level != ConfidenceLevelNone {
		t.Fatalf(
			"expected no confidence when there are no observations, got %s",
			metric.Confidence.Level,
		)
	}
}

func TestCalculateActiveAircraftRejectsInvalidWindow(
	t *testing.T,
) {
	service := mustNewServiceWithClock(
		&activeAircraftRepositoryStub{},
		region.NewService(),
		time.Now,
	)

	_, err := service.CalculateActiveAircraft(
		context.Background(),
		ActiveAircraftRequest{
			WindowMinutes: MaximumActiveAircraftWindowMinutes + 1,
		},
	)
	if !errors.Is(
		err,
		ErrInvalidWindowMinutes,
	) {
		t.Fatalf(
			"expected invalid window error, got %v",
			err,
		)
	}
}
