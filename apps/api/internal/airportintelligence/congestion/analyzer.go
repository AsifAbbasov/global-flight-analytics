package congestion

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

const scoreSemantics = "current completed-day movements_per_hour divided by the prior observed peak movements_per_hour, capped at 1"

type Analyzer struct{}

func NewAnalyzer() Analyzer { return Analyzer{} }

func (Analyzer) Analyze(input Input) (Result, error) {
	historyValue := input.History
	icaoCode, err := normalizeICAO(historyValue.ICAOCode)
	if err != nil {
		return Result{}, err
	}
	windowStart := input.WindowStart.UTC()
	windowEnd := input.WindowEnd.UTC()
	if windowStart.IsZero() || windowEnd.IsZero() || !windowEnd.After(windowStart) {
		return Result{}, fmt.Errorf("%w: requested window must be valid", ErrInvalidInput)
	}
	windowDuration := windowEnd.Sub(windowStart)
	if windowDuration%(24*time.Hour) != 0 {
		return Result{}, fmt.Errorf("%w: requested window must contain whole UTC days", ErrInvalidInput)
	}
	expectedWindowCount := int(windowDuration / (24 * time.Hour))
	if expectedWindowCount < 2 {
		return Result{}, fmt.Errorf("%w: at least two completed requested windows are required", ErrInsufficientHistory)
	}
	if historyValue.WindowStart.IsZero() || historyValue.WindowEnd.IsZero() || !historyValue.WindowEnd.After(historyValue.WindowStart) {
		return Result{}, fmt.Errorf("%w: observed history window must be valid", ErrInvalidInput)
	}
	if historyValue.WindowStart.Before(windowStart) || historyValue.WindowEnd.After(windowEnd) {
		return Result{}, fmt.Errorf("%w: observed history must stay inside the requested window", ErrInvalidInput)
	}
	if input.GeneratedAt.IsZero() || input.GeneratedAt.Before(windowEnd) {
		return Result{}, fmt.Errorf("%w: generated time must be at or after the requested window", ErrInvalidInput)
	}
	if len(historyValue.Entries) < 2 {
		return Result{}, fmt.Errorf("%w: current plus at least one prior observed completed window are required", ErrInsufficientHistory)
	}
	if len(historyValue.Entries) > expectedWindowCount {
		return Result{}, fmt.Errorf("%w: observed history exceeds the requested window count", ErrInvalidInput)
	}

	entries := append([]statistics.Statistics(nil), historyValue.Entries...)
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].WindowStart.Before(entries[right].WindowStart)
	})

	minimumEvidenceSupport := 1.0
	for index, entry := range entries {
		if err := validateEntry(entry, icaoCode, windowStart, windowEnd); err != nil {
			return Result{}, err
		}
		minimumEvidenceSupport = math.Min(minimumEvidenceSupport, math.Min(entry.CoverageScore, entry.FreshnessScore))
		if index > 0 && entry.WindowStart.Before(entries[index-1].WindowEnd) {
			return Result{}, fmt.Errorf("%w: history entries must not overlap", ErrInvalidInput)
		}
	}

	current := entries[len(entries)-1]
	baselineEntries := entries[:len(entries)-1]
	baselineValues := make([]float64, 0, len(baselineEntries))
	priorPeak := 0.0
	for _, entry := range baselineEntries {
		baselineValues = append(baselineValues, entry.MovementsPerHour)
		if entry.MovementsPerHour > priorPeak {
			priorPeak = entry.MovementsPerHour
		}
	}
	baselineMedian := median(baselineValues)

	observedWindowCount := len(entries)
	gapWindowCount := expectedWindowCount - observedWindowCount
	trailingGapWindowCount := int(windowEnd.Sub(current.WindowEnd) / (24 * time.Hour))
	currentWindowIsLatestExpected := trailingGapWindowCount == 0
	evidenceCoverage := float64(observedWindowCount) / float64(expectedWindowCount)
	evidenceSupport := math.Min(evidenceCoverage, minimumEvidenceSupport)

	result := Result{
		Version:                          Version,
		Status:                           StatusUnavailable,
		ICAOCode:                         icaoCode,
		WindowStart:                      windowStart,
		WindowEnd:                        windowEnd,
		Current:                          current,
		ObservedWindowCount:              observedWindowCount,
		ExpectedWindowCount:              expectedWindowCount,
		GapWindowCount:                   gapWindowCount,
		TrailingGapWindowCount:           trailingGapWindowCount,
		CurrentWindowIsLatestExpected:    currentWindowIsLatestExpected,
		EvidenceCoverage:                 evidenceCoverage,
		EvidenceSupport:                  evidenceSupport,
		BaselineWindowCount:              len(baselineEntries),
		BaselineMedianMovementsPerHour:   baselineMedian,
		PriorPeakMovementsPerHour:        priorPeak,
		ExceedsPriorObservedActivityPeak: current.MovementsPerHour > priorPeak,
		ScoreSemantics:                   scoreSemantics,
		ScopeGuard:                       ScopeGuard,
		Explanation:                      "A current observed completed day and a non-zero prior observed activity peak are required before a bounded congestion score can be reported. No airport capacity, queue, slot, runway occupancy, or delay evidence is inferred.",
		GeneratedAt:                      input.GeneratedAt.UTC(),
	}

	if baselineMedian > 0 {
		result.CurrentToBaselineRatio = current.MovementsPerHour / baselineMedian
		result.CurrentToBaselineRatioKnown = true
	}
	if priorPeak > 0 {
		result.CurrentToPriorPeakRatio = current.MovementsPerHour / priorPeak
		result.CurrentToPriorPeakRatioKnown = true
	}
	if currentWindowIsLatestExpected && priorPeak > 0 {
		result.CongestionScore = math.Min(1, result.CurrentToPriorPeakRatio)
		result.CongestionScoreKnown = true
		result.Status = StatusAvailable
		result.Explanation = "The congestion score is a relative observed-activity proxy: the latest expected completed-day movements per hour divided by the prior observed peak, capped at 1. It is not an airport capacity, queue, slot, runway occupancy, delay, or official congestion measure."
	} else if !currentWindowIsLatestExpected {
		result.Explanation = "The latest expected completed day has no Airport Intelligence observation, so no stale congestion score is substituted. Historical activity ratios remain descriptive evidence only."
	}

	return result, nil
}

func validateEntry(entry statistics.Statistics, icaoCode string, parentStart time.Time, parentEnd time.Time) error {
	entryICAO, err := normalizeICAO(entry.ICAOCode)
	if err != nil || entryICAO != icaoCode {
		return fmt.Errorf("%w: history entry ICAO does not match the requested airport", ErrInvalidInput)
	}
	if entry.WindowStart.IsZero() || entry.WindowEnd.IsZero() || entry.WindowEnd.Sub(entry.WindowStart) != 24*time.Hour {
		return fmt.Errorf("%w: history entries must represent one completed day", ErrInvalidInput)
	}
	if entry.WindowStart.Before(parentStart) || entry.WindowEnd.After(parentEnd) {
		return fmt.Errorf("%w: history entry is outside the requested window", ErrInvalidInput)
	}
	if !finiteNonNegative(entry.MovementsPerHour) {
		return fmt.Errorf("%w: movements per hour must be finite and non-negative", ErrInvalidInput)
	}
	if !unitInterval(entry.CoverageScore) || !unitInterval(entry.FreshnessScore) {
		return fmt.Errorf("%w: evidence scores must be between zero and one", ErrInvalidInput)
	}
	return nil
}

func normalizeICAO(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if len(normalized) != 4 {
		return "", fmt.Errorf("%w: ICAO code must contain four letters", ErrInvalidInput)
	}
	for _, character := range normalized {
		if character < 'A' || character > 'Z' {
			return "", fmt.Errorf("%w: ICAO code must contain four Latin letters", ErrInvalidInput)
		}
	}
	return normalized, nil
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	middle := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[middle]
	}
	return (ordered[middle-1] + ordered[middle]) / 2
}

func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func unitInterval(value float64) bool {
	return finiteNonNegative(value) && value <= 1
}
