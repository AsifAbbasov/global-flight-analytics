package ranking

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

func TestRankerSeparatesActivityFromDataConfidence(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(4 * time.Hour)

	result, err := NewRanker().Rank(Input{
		Statistics: []statistics.Statistics{
			airportStatistics(" ubbb ", windowStart, windowEnd, 80, 8, 80, 20, 0.90, 0.80, 12),
			airportStatistics("UBBA", windowStart, windowEnd, 40, 4, 40, 10, 1.00, 1.00, 8),
			airportStatistics("UGTB", windowStart, windowEnd, 60, 6, 60, 15, 0.50, 0.50, 10),
		},
	})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	wantOrder := []string{"UBBB", "UGTB", "UBBA"}
	for index, wantICAO := range wantOrder {
		if result.Airports[index].ICAOCode != wantICAO {
			t.Fatalf("position %d ICAO = %q, want %q", index+1, result.Airports[index].ICAOCode, wantICAO)
		}
		if result.Airports[index].Position != index+1 {
			t.Fatalf("position field = %d, want %d", result.Airports[index].Position, index+1)
		}
	}

	leader := result.Airports[0]
	assertFloatEqual(t, leader.ActivityScore, 100)
	assertFloatEqual(t, leader.DataConfidence, 85)
	assertFloatEqual(t, leader.MovementsComponent, 100)
	assertFloatEqual(t, leader.RoutesComponent, 100)
	assertFloatEqual(t, leader.ObservationsComponent, 100)
	assertFloatEqual(t, leader.IntensityComponent, 100)

	mostTrusted := result.Airports[2]
	assertFloatEqual(t, mostTrusted.ActivityScore, 50)
	assertFloatEqual(t, mostTrusted.DataConfidence, 100)

	if leader.ActivityScore <= mostTrusted.ActivityScore {
		t.Fatal("higher data confidence must not replace operational activity in the activity ranking")
	}
}

func TestRankerUsesDocumentedActivityFactors(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)

	result, err := NewRanker().Rank(Input{
		Statistics: []statistics.Statistics{
			airportStatistics("UBBB", windowStart, windowEnd, 100, 5, 50, 25, 1, 1, 5),
			airportStatistics("UBBA", windowStart, windowEnd, 50, 10, 100, 50, 1, 1, 5),
		},
	})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	first := result.Airports[0]
	second := result.Airports[1]
	assertFloatEqual(t, first.ActivityScore, 87.5)
	assertFloatEqual(t, second.ActivityScore, 62.5)
	assertFloatEqual(t, first.MovementsComponent, 50)
	assertFloatEqual(t, first.RoutesComponent, 100)
	assertFloatEqual(t, first.ObservationsComponent, 100)
	assertFloatEqual(t, first.IntensityComponent, 100)
}

func TestRankerNormalizesCustomWeights(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)

	result, err := NewRanker().Rank(Input{
		Statistics: []statistics.Statistics{
			airportStatistics("UBBB", windowStart, windowEnd, 10, 2, 10, 10, 1, 1, 1),
		},
		ActivityWeights:   ActivityWeights{Movements: 2, Routes: 1, Observations: 1},
		ConfidenceWeights: ConfidenceWeights{Coverage: 3, Freshness: 1},
	})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	assertFloatEqual(t, result.ActivityWeights.Movements, 0.50)
	assertFloatEqual(t, result.ActivityWeights.Routes, 0.25)
	assertFloatEqual(t, result.ActivityWeights.Observations, 0.25)
	assertFloatEqual(t, result.ActivityWeights.Intensity, 0)
	assertFloatEqual(t, result.ConfidenceWeights.Coverage, 0.75)
	assertFloatEqual(t, result.ConfidenceWeights.Freshness, 0.25)
}

func TestRankerSupportsNoObservedActivity(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)

	result, err := NewRanker().Rank(Input{
		Statistics: []statistics.Statistics{
			airportStatistics("UBBB", windowStart, windowEnd, 0, 0, 0, 0, 0.8, 0.6, 0),
		},
	})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	assertFloatEqual(t, result.Airports[0].ActivityScore, 0)
	assertFloatEqual(t, result.Airports[0].DataConfidence, 70)
}

func TestRankerUsesDeterministicTieBreaking(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)

	result, err := NewRanker().Rank(Input{
		Statistics: []statistics.Statistics{
			airportStatistics("UBBB", windowStart, windowEnd, 10, 2, 10, 10, 1, 1, 1),
			airportStatistics("UBBA", windowStart, windowEnd, 10, 2, 10, 10, 0.2, 0.2, 1),
		},
	})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	if result.Airports[0].ICAOCode != "UBBA" || result.Airports[1].ICAOCode != "UBBB" {
		t.Fatalf("tie order = %q, %q, want UBBA, UBBB", result.Airports[0].ICAOCode, result.Airports[1].ICAOCode)
	}
}

func TestRankerDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)
	items := []statistics.Statistics{
		airportStatistics(" ubbb ", windowStart, windowEnd, 10, 2, 10, 10, 1, 1, 1),
	}

	_, err := NewRanker().Rank(Input{Statistics: items})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}
	if items[0].ICAOCode != " ubbb " {
		t.Fatalf("input ICAO = %q, want original value", items[0].ICAOCode)
	}
}

func TestRankerRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)
	valid := airportStatistics("UBBB", windowStart, windowEnd, 10, 2, 10, 10, 1, 1, 1)

	tests := []struct {
		name  string
		input Input
		want  error
	}{
		{name: "empty", input: Input{}, want: ErrInvalidInput},
		{name: "duplicate", input: Input{Statistics: []statistics.Statistics{valid, valid}}, want: ErrDuplicateAirport},
		{name: "different window", input: Input{Statistics: []statistics.Statistics{valid, airportStatistics("UBBA", windowStart, windowEnd.Add(time.Hour), 10, 2, 10, 10, 1, 1, 1)}}, want: ErrIncomparableWindow},
		{name: "negative routes", input: Input{Statistics: []statistics.Statistics{withActiveRoutes(valid, -1)}}, want: ErrInvalidInput},
		{name: "invalid coverage", input: Input{Statistics: []statistics.Statistics{withCoverage(valid, 1.1)}}, want: ErrInvalidInput},
		{name: "invalid activity weights", input: Input{Statistics: []statistics.Statistics{valid}, ActivityWeights: ActivityWeights{Movements: -1, Routes: 1}}, want: ErrInvalidConfiguration},
		{name: "invalid confidence weights", input: Input{Statistics: []statistics.Statistics{valid}, ConfidenceWeights: ConfidenceWeights{Coverage: math.NaN(), Freshness: 1}}, want: ErrInvalidConfiguration},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewRanker().Rank(test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("Rank() error = %v, want %v", err, test.want)
			}
		})
	}
}

func airportStatistics(
	icao string,
	windowStart time.Time,
	windowEnd time.Time,
	totalMovements int,
	activeRoutes int,
	observedSamples int,
	movementsPerHour float64,
	coverageScore float64,
	freshnessScore float64,
	activeAircraft int,
) statistics.Statistics {
	return statistics.Statistics{
		ICAOCode:         icao,
		WindowStart:      windowStart,
		WindowEnd:        windowEnd,
		TotalMovements:   totalMovements,
		ActiveRoutes:     activeRoutes,
		ObservedSamples:  observedSamples,
		ExpectedSamples:  100,
		MovementsPerHour: movementsPerHour,
		CoverageScore:    coverageScore,
		FreshnessScore:   freshnessScore,
		ActiveAircraft:   activeAircraft,
	}
}

func withActiveRoutes(value statistics.Statistics, activeRoutes int) statistics.Statistics {
	value.ActiveRoutes = activeRoutes
	return value
}

func withCoverage(value statistics.Statistics, coverage float64) statistics.Statistics {
	value.CoverageScore = coverage
	return value
}

func assertFloatEqual(t *testing.T, actual, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("value = %v, want %v", actual, expected)
	}
}
