package trends

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

func TestAnalyzerAnalyzeBuildsIncreasingTrendAndContinuity(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC), time.Hour, 3, 2, 2, 80, 100, 10*time.Minute),
		statisticsEntry(t, time.Date(2026, time.July, 13, 9, 0, 0, 0, time.UTC), time.Hour, 4, 4, 3, 90, 100, 5*time.Minute),
		statisticsEntry(t, time.Date(2026, time.July, 13, 11, 0, 0, 0, time.UTC), time.Hour, 8, 4, 5, 100, 100, 0),
	}
	airportHistory := historyValue(t, entries, time.Date(2026, time.July, 13, 12, 30, 0, 0, time.UTC))
	originalEntries := append([]statistics.Statistics(nil), airportHistory.Entries...)

	result, err := NewAnalyzer().Analyze(Input{
		History:     airportHistory,
		GeneratedAt: time.Date(2026, time.July, 13, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if result.ICAOCode != "UBBB" {
		t.Fatalf("ICAOCode = %q, want UBBB", result.ICAOCode)
	}
	if result.ComparedWindows != 3 {
		t.Fatalf("ComparedWindows = %d, want 3", result.ComparedWindows)
	}
	if result.WindowDuration != time.Hour {
		t.Fatalf("WindowDuration = %s, want 1h", result.WindowDuration)
	}
	if result.Direction != DirectionIncreasing {
		t.Fatalf("Direction = %q, want %q", result.Direction, DirectionIncreasing)
	}
	if result.TotalMovementsChange != 7 {
		t.Fatalf("TotalMovementsChange = %d, want 7", result.TotalMovementsChange)
	}
	assertFloatEqual(t, result.MovementsPerHourChange, 7)
	if !result.MovementsPerHourChangePercentKnown {
		t.Fatal("expected movement change percent to be available")
	}
	assertFloatEqual(t, result.MovementsPerHourChangePercent, 140)
	if result.ActiveRoutesChange != 3 {
		t.Fatalf("ActiveRoutesChange = %d, want 3", result.ActiveRoutesChange)
	}
	if result.GapCount != 1 {
		t.Fatalf("GapCount = %d, want 1", result.GapCount)
	}
	if result.GapDuration != time.Hour {
		t.Fatalf("GapDuration = %s, want 1h", result.GapDuration)
	}
	if result.ObservedDuration != 3*time.Hour {
		t.Fatalf("ObservedDuration = %s, want 3h", result.ObservedDuration)
	}
	assertFloatEqual(t, result.ContinuityScore, 0.75)
	if result.Peak.WindowStart != entries[2].WindowStart {
		t.Fatalf("Peak.WindowStart = %s, want %s", result.Peak.WindowStart, entries[2].WindowStart)
	}
	if !reflect.DeepEqual(airportHistory.Entries, originalEntries) {
		t.Fatal("Analyze() modified history entries")
	}
}

func TestAnalyzerAnalyzeClassifiesDecreasingTrend(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 8, 4, 4, 100, 100, 0),
		statisticsEntry(t, testStart().Add(time.Hour), time.Hour, 3, 2, 2, 100, 100, 0),
	}

	result, err := NewAnalyzer().Analyze(Input{
		History:     historyValue(t, entries, testStart().Add(3*time.Hour)),
		GeneratedAt: testStart().Add(4 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Direction != DirectionDecreasing {
		t.Fatalf("Direction = %q, want %q", result.Direction, DirectionDecreasing)
	}
	assertFloatEqual(t, result.MovementsPerHourChangePercent, -58.333333333333336)
}

func TestAnalyzerAnalyzeClassifiesStableTrend(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 3, 2, 2, 100, 100, 0),
		statisticsEntry(t, testStart().Add(time.Hour), time.Hour, 3, 2, 3, 100, 100, 0),
	}

	result, err := NewAnalyzer().Analyze(Input{
		History:     historyValue(t, entries, testStart().Add(4*time.Hour)),
		GeneratedAt: testStart().Add(5 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Direction != DirectionStable {
		t.Fatalf("Direction = %q, want %q", result.Direction, DirectionStable)
	}
	assertFloatEqual(t, result.MovementsPerHourChange, 0)
	assertFloatEqual(t, result.MovementsPerHourChangePercent, 0)
}

func TestAnalyzerAnalyzeDoesNotInventPercentFromZeroBaseline(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 0, 0, 0, 100, 100, 0),
		statisticsEntry(t, testStart().Add(time.Hour), time.Hour, 2, 1, 1, 100, 100, 0),
	}

	result, err := NewAnalyzer().Analyze(Input{
		History:     historyValue(t, entries, testStart().Add(3*time.Hour)),
		GeneratedAt: testStart().Add(4 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.MovementsPerHourChangePercentKnown {
		t.Fatal("did not expect movement change percent with zero baseline")
	}
	assertFloatEqual(t, result.MovementsPerHourChangePercent, 0)
}

func TestAnalyzerAnalyzeKeepsEarliestPeakWhenRatesAndTotalsAreEqual(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 3, 2, 2, 100, 100, 0),
		statisticsEntry(t, testStart().Add(time.Hour), time.Hour, 3, 2, 3, 100, 100, 0),
	}

	result, err := NewAnalyzer().Analyze(Input{
		History:     historyValue(t, entries, testStart().Add(3*time.Hour)),
		GeneratedAt: testStart().Add(4 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if !result.Peak.WindowStart.Equal(entries[0].WindowStart) {
		t.Fatalf("Peak.WindowStart = %s, want earliest %s", result.Peak.WindowStart, entries[0].WindowStart)
	}
}

func TestAnalyzerAnalyzeRejectsInsufficientHistory(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 3, 2, 2, 100, 100, 0),
	}

	_, err := NewAnalyzer().Analyze(Input{
		History:     historyValue(t, entries, testStart().Add(2*time.Hour)),
		GeneratedAt: testStart().Add(3 * time.Hour),
	})
	if !errors.Is(err, ErrInsufficientHistory) {
		t.Fatalf("Analyze() error = %v, want ErrInsufficientHistory", err)
	}
}

func TestAnalyzerAnalyzeRejectsInconsistentDeclaredHistoryWindow(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 3, 2, 2, 100, 100, 0),
		statisticsEntry(t, testStart().Add(time.Hour), time.Hour, 4, 2, 3, 100, 100, 0),
	}
	airportHistory := historyValue(t, entries, testStart().Add(3*time.Hour))
	airportHistory.WindowEnd = airportHistory.WindowEnd.Add(time.Minute)

	_, err := NewAnalyzer().Analyze(Input{
		History:     airportHistory,
		GeneratedAt: testStart().Add(4 * time.Hour),
	})
	if !errors.Is(err, ErrInvalidHistory) {
		t.Fatalf("Analyze() error = %v, want ErrInvalidHistory", err)
	}
}

func TestAnalyzerAnalyzeRejectsGenerationBeforeHistory(t *testing.T) {
	entries := []statistics.Statistics{
		statisticsEntry(t, testStart(), time.Hour, 3, 2, 2, 100, 100, 0),
		statisticsEntry(t, testStart().Add(time.Hour), time.Hour, 4, 2, 3, 100, 100, 0),
	}
	airportHistory := historyValue(t, entries, testStart().Add(3*time.Hour))

	_, err := NewAnalyzer().Analyze(Input{
		History:     airportHistory,
		GeneratedAt: airportHistory.GeneratedAt.Add(-time.Second),
	})
	if !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("Analyze() error = %v, want ErrInvalidTime", err)
	}
}

func statisticsEntry(
	t *testing.T,
	windowStart time.Time,
	windowDuration time.Duration,
	arrivals int,
	departures int,
	activeRoutes int,
	observedSamples int,
	expectedSamples int,
	observationAge time.Duration,
) statistics.Statistics {
	t.Helper()

	calculator, err := statistics.NewCalculator(time.Hour)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}
	windowEnd := windowStart.Add(windowDuration)
	result, err := calculator.Calculate(statistics.Input{
		ICAOCode:            " ubBB ",
		WindowStart:         windowStart,
		WindowEnd:           windowEnd,
		Arrivals:            arrivals,
		Departures:          departures,
		ActiveAircraft:      arrivals + departures,
		ActiveRoutes:        activeRoutes,
		ObservedSamples:     observedSamples,
		ExpectedSamples:     expectedSamples,
		LatestObservationAt: windowEnd.Add(-observationAge),
		GeneratedAt:         windowEnd,
	})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	return result
}

func historyValue(t *testing.T, entries []statistics.Statistics, generatedAt time.Time) history.History {
	t.Helper()

	result, err := history.NewBuilder().Build(history.Input{
		ICAOCode:    " ubBB ",
		Entries:     entries,
		GeneratedAt: generatedAt,
	})
	if err != nil {
		t.Fatalf("history Build() error = %v", err)
	}

	return result
}

func testStart() time.Time {
	return time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
}

func assertFloatEqual(t *testing.T, actual, expected float64) {
	t.Helper()

	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("value = %v, want %v", actual, expected)
	}
}
