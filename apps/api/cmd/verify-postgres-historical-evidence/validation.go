package main

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/jackc/pgx/v5"
)

type metricExpectation struct {
	Name  historicalcontract.MetricName
	Scope historicalcontract.Scope

	CurrentPoints  []float64
	PreviousPoints []float64

	CurrentTotal   float64
	PreviousTotal  float64
	AbsoluteChange float64
	Percentage     float64
}

func evidenceMetricExpectations() []metricExpectation {
	return []metricExpectation{
		{
			Name: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			CurrentPoints: []float64{
				2,
				3,
			},
			PreviousPoints: []float64{
				1,
				1,
			},
			CurrentTotal:   5,
			PreviousTotal:  2,
			AbsoluteChange: 3,
			Percentage:     150,
		},
		{
			Name: historicalcontract.
				MetricNameTrajectoryCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			CurrentPoints: []float64{
				2,
				3,
			},
			PreviousPoints: []float64{
				1,
				1,
			},
			CurrentTotal:   5,
			PreviousTotal:  2,
			AbsoluteChange: 3,
			Percentage:     150,
		},
		{
			Name: historicalcontract.
				MetricNameObservationCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			CurrentPoints: []float64{
				4,
				6,
			},
			PreviousPoints: []float64{
				2,
				3,
			},
			CurrentTotal:   10,
			PreviousTotal:  5,
			AbsoluteChange: 5,
			Percentage:     100,
		},
		{
			Name: historicalcontract.
				MetricNameAirportDepartures,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeAirport,
				AirportICAOCode: fixtureOriginICAO,
			},
			CurrentPoints: []float64{
				2,
				3,
			},
			PreviousPoints: []float64{
				1,
				1,
			},
			CurrentTotal:   5,
			PreviousTotal:  2,
			AbsoluteChange: 3,
			Percentage:     150,
		},
		{
			Name: historicalcontract.
				MetricNameRouteObservations,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeRoute,
				OriginICAOCode:      fixtureOriginICAO,
				DestinationICAOCode: fixtureDestinationICAO,
			},
			CurrentPoints: []float64{
				2,
				3,
			},
			PreviousPoints: []float64{
				1,
				1,
			},
			CurrentTotal:   5,
			PreviousTotal:  2,
			AbsoluteChange: 3,
			Percentage:     150,
		},
	}
}

func validateMetricOutcome(
	outcome historicalmaterialization.Outcome,
	expectation metricExpectation,
	schedule evidenceSchedule,
) error {
	if outcome.Version !=
		historicalmaterialization.Version {
		return fmt.Errorf(
			"materialization version = %q",
			outcome.Version,
		)
	}
	if outcome.CurrentResult.Metric.Name !=
		expectation.Name {
		return fmt.Errorf(
			"current metric = %s, want %s",
			outcome.CurrentResult.Metric.Name,
			expectation.Name,
		)
	}
	if !reflect.DeepEqual(
		outcome.CurrentResult.Scope,
		expectation.Scope,
	) {
		return fmt.Errorf(
			"current scope = %#v, want %#v",
			outcome.CurrentResult.Scope,
			expectation.Scope,
		)
	}
	if !outcome.CurrentResult.Window.StartTime.Equal(
		schedule.CurrentStart,
	) ||
		!outcome.CurrentResult.Window.EndTime.Equal(
			schedule.CurrentEnd,
		) ||
		!outcome.PreviousResult.Window.StartTime.Equal(
			schedule.PreviousStart,
		) ||
		!outcome.PreviousResult.Window.EndTime.Equal(
			schedule.PreviousEnd,
		) {
		return fmt.Errorf(
			"unexpected current or previous analytical window",
		)
	}

	if err := validatePoints(
		outcome.CurrentResult,
		expectation.CurrentPoints,
	); err != nil {
		return fmt.Errorf(
			"current points: %w",
			err,
		)
	}
	if err := validatePoints(
		outcome.PreviousResult,
		expectation.PreviousPoints,
	); err != nil {
		return fmt.Errorf(
			"previous points: %w",
			err,
		)
	}

	if !almostEqual(
		outcome.CurrentResult.Summary.Total,
		expectation.CurrentTotal,
	) {
		return fmt.Errorf(
			"current total = %v, want %v",
			outcome.CurrentResult.Summary.Total,
			expectation.CurrentTotal,
		)
	}
	if !almostEqual(
		outcome.PreviousResult.Summary.Total,
		expectation.PreviousTotal,
	) {
		return fmt.Errorf(
			"previous total = %v, want %v",
			outcome.PreviousResult.Summary.Total,
			expectation.PreviousTotal,
		)
	}

	comparison := outcome.CurrentResult.Comparison
	if comparison == nil {
		return fmt.Errorf(
			"period comparison is absent",
		)
	}
	if !almostEqual(
		comparison.CurrentValue,
		expectation.CurrentTotal,
	) ||
		!almostEqual(
			comparison.PreviousValue,
			expectation.PreviousTotal,
		) ||
		!almostEqual(
			comparison.AbsoluteChange,
			expectation.AbsoluteChange,
		) {
		return fmt.Errorf(
			"unexpected comparison values: %#v",
			comparison,
		)
	}
	if comparison.PercentageChange == nil ||
		!almostEqual(
			*comparison.PercentageChange,
			expectation.Percentage,
		) {
		return fmt.Errorf(
			"percentage change = %#v, want %v",
			comparison.PercentageChange,
			expectation.Percentage,
		)
	}
	if comparison.Direction !=
		historicalcontract.TrendDirectionUp {
		return fmt.Errorf(
			"comparison direction = %s, want up",
			comparison.Direction,
		)
	}

	if outcome.ReadSummary.FlightCount != 7 ||
		outcome.ReadSummary.TrajectoryCount != 7 ||
		outcome.ReadSummary.ObservationCount != 15 ||
		outcome.ReadSummary.RouteCount != 7 {
		return fmt.Errorf(
			"unexpected read summary: %#v",
			outcome.ReadSummary,
		)
	}
	if outcome.ReadSummary.FlightLimitReached ||
		outcome.ReadSummary.TrajectoryLimitReached ||
		outcome.ReadSummary.ObservationLimitReached ||
		outcome.ReadSummary.RouteLimitReached {
		return fmt.Errorf(
			"fixture unexpectedly reached a dataset limit",
		)
	}

	if outcome.Record.ID == "" ||
		!reflect.DeepEqual(
			outcome.Record.Result,
			outcome.CurrentResult,
		) {
		return fmt.Errorf(
			"persisted aggregate record is incomplete",
		)
	}

	report := historicalcontract.Validate(
		outcome.CurrentResult,
	)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		return fmt.Errorf(
			"current contract invalid: errors=%d warnings=%d",
			report.ErrorCount,
			report.WarningCount,
		)
	}

	return nil
}

func validatePoints(
	result historicalcontract.Result,
	expected []float64,
) error {
	if result.Status !=
		historicalcontract.SeriesStatusComplete {
		return fmt.Errorf(
			"series status = %s, want complete",
			result.Status,
		)
	}
	if len(result.Points) != len(expected) {
		return fmt.Errorf(
			"point count = %d, want %d",
			len(result.Points),
			len(expected),
		)
	}

	for index, point := range result.Points {
		if point.Status !=
			historicalcontract.BucketStatusComplete {
			return fmt.Errorf(
				"point[%d] status = %s, want complete",
				index,
				point.Status,
			)
		}
		if !almostEqual(
			point.Value,
			expected[index],
		) {
			return fmt.Errorf(
				"point[%d] value = %v, want %v",
				index,
				point.Value,
				expected[index],
			)
		}
	}

	return nil
}

func validateReplayEvidence(
	result historicalreplay.Result,
	schedule evidenceSchedule,
) error {
	if result.Version != historicalreplay.Version {
		return fmt.Errorf(
			"replay version = %q",
			result.Version,
		)
	}
	if len(result.Windows) != 2 {
		return fmt.Errorf(
			"replay window count = %d, want 2",
			len(result.Windows),
		)
	}

	expectedCurrent := []float64{
		2,
		3,
	}
	expectedPrevious := []float64{
		1,
		2,
	}
	expectedStart := schedule.CurrentStart

	for index, window := range result.Windows {
		expectedEnd := expectedStart.Add(
			time.Hour,
		)
		if !window.Bucket.StartTime.Equal(
			expectedStart,
		) ||
			!window.Bucket.EndTime.Equal(
				expectedEnd,
			) {
			return fmt.Errorf(
				"replay window[%d] = %#v",
				index,
				window.Bucket,
			)
		}
		if window.Record.Result.Metric.Name !=
			historicalcontract.MetricNameFlightCount {
			return fmt.Errorf(
				"replay metric[%d] = %s",
				index,
				window.Record.Result.Metric.Name,
			)
		}
		if !almostEqual(
			window.Record.Result.Summary.Total,
			expectedCurrent[index],
		) {
			return fmt.Errorf(
				"replay total[%d] = %v, want %v",
				index,
				window.Record.Result.Summary.Total,
				expectedCurrent[index],
			)
		}
		comparison :=
			window.Record.Result.Comparison
		if comparison == nil ||
			!almostEqual(
				comparison.PreviousValue,
				expectedPrevious[index],
			) {
			return fmt.Errorf(
				"replay comparison[%d] = %#v",
				index,
				comparison,
			)
		}

		report := historicalcontract.Validate(
			window.Record.Result,
		)
		if report.Status !=
			historicalcontract.ValidationStatusValid {
			return fmt.Errorf(
				"replay contract[%d] invalid: errors=%d warnings=%d",
				index,
				report.ErrorCount,
				report.WarningCount,
			)
		}

		expectedStart = expectedEnd
	}

	if !expectedStart.Equal(schedule.CurrentEnd) {
		return fmt.Errorf(
			"replay ended at %s, want %s",
			expectedStart,
			schedule.CurrentEnd,
		)
	}

	return nil
}

type evidenceCounts struct {
	Flights      int
	Trajectories int
	Observations int
	Routes       int
	Aggregates   int
}

type rowQuerier interface {
	QueryRow(
		context.Context,
		string,
		...any,
	) pgx.Row
}

func countEvidence(
	ctx context.Context,
	querier rowQuerier,
	fixture evidenceFixture,
	asOfTime time.Time,
) (evidenceCounts, error) {
	var counts evidenceCounts
	err := querier.QueryRow(
		ctx,
		`
			SELECT
				(
					SELECT count(*)::integer
					FROM flights
					WHERE id::text =
						ANY($1::text[])
				),
				(
					SELECT count(*)::integer
					FROM flight_trajectories
					WHERE id::text =
						ANY($2::text[])
				),
				(
					SELECT count(*)::integer
					FROM flight_states
					WHERE id::text =
						ANY($3::text[])
				),
				(
					SELECT count(*)::integer
					FROM flight_route_results
					WHERE id =
						ANY($4::text[])
				),
				(
					SELECT count(*)::integer
					FROM historical_aggregate_results
					WHERE as_of_time_unix_nano = $5
				);
		`,
		fixture.FlightIDs,
		fixture.TrajectoryIDs,
		fixture.ObservationIDs,
		fixture.RouteRecordIDs,
		asOfTime.UnixNano(),
	).Scan(
		&counts.Flights,
		&counts.Trajectories,
		&counts.Observations,
		&counts.Routes,
		&counts.Aggregates,
	)
	if err != nil {
		return evidenceCounts{},
			fmt.Errorf(
				"count verification evidence: %w",
				err,
			)
	}

	return counts, nil
}

func validateTransactionalCounts(
	counts evidenceCounts,
	expectedAggregateCount int,
) error {
	expected := evidenceCounts{
		Flights:      7,
		Trajectories: 7,
		Observations: 15,
		Routes:       7,
		Aggregates:   expectedAggregateCount,
	}
	if counts != expected {
		return fmt.Errorf(
			"transactional counts = %#v, want %#v",
			counts,
			expected,
		)
	}

	return nil
}

func validateRollbackCounts(
	counts evidenceCounts,
) error {
	if counts != (evidenceCounts{}) {
		return fmt.Errorf(
			"verification evidence remained after rollback: %#v",
			counts,
		)
	}

	return nil
}

func reloadAggregate(
	ctx context.Context,
	store historicalaggregate.Store,
	outcome historicalmaterialization.Outcome,
) error {
	loaded, err := store.Get(
		ctx,
		outcome.Record.Key,
	)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(
		loaded,
		outcome.Record,
	) {
		return fmt.Errorf(
			"reloaded aggregate differs from materialized record",
		)
	}

	return nil
}

func almostEqual(
	left float64,
	right float64,
) bool {
	return math.Abs(left-right) <= 1e-9
}
