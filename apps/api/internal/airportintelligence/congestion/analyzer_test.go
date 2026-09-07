package congestion

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

func TestAnalyzerDerivesRelativeObservedActivity(t *testing.T) {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(4 * 24 * time.Hour)
	result, err := NewAnalyzer().Analyze(Input{
		History: history.History{
			ICAOCode:    "ubbb",
			WindowStart: start,
			WindowEnd:   end,
			Entries: []statistics.Statistics{
				entry("UBBB", start, 2, 1, 0.95),
				entry("UBBB", start.Add(24*time.Hour), 4, 1, 0.90),
				entry("UBBB", start.Add(3*24*time.Hour), 3, 1, 0.80),
			},
			GeneratedAt: end,
		},
		WindowStart: start,
		WindowEnd:   end,
		GeneratedAt: end,
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Status != StatusAvailable || !result.CongestionScoreKnown {
		t.Fatalf("expected available score, got status=%s known=%v", result.Status, result.CongestionScoreKnown)
	}
	assertClose(t, result.BaselineMedianMovementsPerHour, 3)
	assertClose(t, result.PriorPeakMovementsPerHour, 4)
	assertClose(t, result.CurrentToBaselineRatio, 1)
	assertClose(t, result.CurrentToPriorPeakRatio, 0.75)
	assertClose(t, result.CongestionScore, 0.75)
	assertClose(t, result.EvidenceCoverage, 0.75)
	assertClose(t, result.EvidenceSupport, 0.75)
	if result.GapWindowCount != 1 || result.TrailingGapWindowCount != 0 || !result.CurrentWindowIsLatestExpected {
		t.Fatalf("unexpected gap evidence: %+v", result)
	}
	if result.ObservedWindowCount != 3 || result.ExpectedWindowCount != 4 {
		t.Fatalf("unexpected evidence window counts: %+v", result)
	}
	if result.ExceedsPriorObservedActivityPeak {
		t.Fatal("current activity should not exceed the prior observed peak")
	}
	if result.ScopeGuard != ScopeGuard {
		t.Fatalf("scope guard = %q", result.ScopeGuard)
	}
}

func TestAnalyzerPreservesNewPeakMagnitudeWhileCappingScore(t *testing.T) {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(3 * 24 * time.Hour)
	result, err := NewAnalyzer().Analyze(Input{
		History: history.History{
			ICAOCode:    "UBBB",
			WindowStart: start,
			WindowEnd:   end,
			Entries: []statistics.Statistics{
				entry("UBBB", start, 1, 1, 1),
				entry("UBBB", start.Add(24*time.Hour), 2, 1, 1),
				entry("UBBB", start.Add(2*24*time.Hour), 3, 1, 1),
			},
		},
		WindowStart: start,
		WindowEnd:   end,
		GeneratedAt: end,
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	assertClose(t, result.CurrentToPriorPeakRatio, 1.5)
	assertClose(t, result.CongestionScore, 1)
	if !result.ExceedsPriorObservedActivityPeak {
		t.Fatal("expected new observed activity peak")
	}
}

func TestAnalyzerDoesNotSubstituteStaleDayForMissingLatestWindow(t *testing.T) {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	requestedEnd := start.Add(4 * 24 * time.Hour)
	observedEnd := start.Add(3 * 24 * time.Hour)
	result, err := NewAnalyzer().Analyze(Input{
		History: history.History{
			ICAOCode:    "UBBB",
			WindowStart: start,
			WindowEnd:   observedEnd,
			Entries: []statistics.Statistics{
				entry("UBBB", start, 1, 1, 1),
				entry("UBBB", start.Add(24*time.Hour), 2, 1, 1),
				entry("UBBB", start.Add(2*24*time.Hour), 3, 1, 1),
			},
		},
		WindowStart: start,
		WindowEnd:   requestedEnd,
		GeneratedAt: requestedEnd,
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Status != StatusUnavailable || result.CongestionScoreKnown {
		t.Fatalf("missing latest completed day must keep congestion unavailable: %+v", result)
	}
	if result.TrailingGapWindowCount != 1 || result.CurrentWindowIsLatestExpected {
		t.Fatalf("expected one explicit trailing gap: %+v", result)
	}
	assertClose(t, result.EvidenceCoverage, 0.75)
	if !result.CurrentToPriorPeakRatioKnown {
		t.Fatal("historical ratio should remain available as descriptive evidence")
	}
}

func TestAnalyzerDoesNotInventScoreWithoutNonZeroPriorPeak(t *testing.T) {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(3 * 24 * time.Hour)
	result, err := NewAnalyzer().Analyze(Input{
		History: history.History{
			ICAOCode:    "UBBB",
			WindowStart: start,
			WindowEnd:   end,
			Entries: []statistics.Statistics{
				entry("UBBB", start, 0, 1, 1),
				entry("UBBB", start.Add(24*time.Hour), 0, 1, 1),
				entry("UBBB", start.Add(2*24*time.Hour), 2, 1, 1),
			},
		},
		WindowStart: start,
		WindowEnd:   end,
		GeneratedAt: end,
	})
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Status != StatusUnavailable || result.CongestionScoreKnown || result.CurrentToPriorPeakRatioKnown || result.CurrentToBaselineRatioKnown {
		t.Fatalf("zero prior activity must keep ratios and score unavailable: %+v", result)
	}
	if !result.ExceedsPriorObservedActivityPeak {
		t.Fatal("current activity should still be identified as exceeding the zero prior peak")
	}
}

func TestAnalyzerRejectsInsufficientHistory(t *testing.T) {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * 24 * time.Hour)
	_, err := NewAnalyzer().Analyze(Input{
		History: history.History{
			ICAOCode:    "UBBB",
			WindowStart: start,
			WindowEnd:   start.Add(24 * time.Hour),
			Entries:     []statistics.Statistics{entry("UBBB", start, 1, 1, 1)},
		},
		WindowStart: start,
		WindowEnd:   end,
		GeneratedAt: end,
	})
	if !errors.Is(err, ErrInsufficientHistory) {
		t.Fatalf("expected ErrInsufficientHistory, got %v", err)
	}
}

func TestAnalyzerRejectsInvalidEvidenceScores(t *testing.T) {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * 24 * time.Hour)
	_, err := NewAnalyzer().Analyze(Input{
		History: history.History{
			ICAOCode:    "UBBB",
			WindowStart: start,
			WindowEnd:   end,
			Entries: []statistics.Statistics{
				entry("UBBB", start, 1, 1.1, 1),
				entry("UBBB", start.Add(24*time.Hour), 2, 1, 1),
			},
		},
		WindowStart: start,
		WindowEnd:   end,
		GeneratedAt: end,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func entry(icao string, start time.Time, movementsPerHour float64, coverage float64, freshness float64) statistics.Statistics {
	return statistics.Statistics{
		ICAOCode:            icao,
		WindowStart:         start,
		WindowEnd:           start.Add(24 * time.Hour),
		MovementsPerHour:    movementsPerHour,
		CoverageScore:       coverage,
		FreshnessScore:      freshness,
		LatestObservationAt: start.Add(23 * time.Hour),
		GeneratedAt:         start.Add(24 * time.Hour),
	}
}

func assertClose(t *testing.T, actual float64, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("value = %v, want %v", actual, expected)
	}
}
