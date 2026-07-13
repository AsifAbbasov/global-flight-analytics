package trends

import (
	"fmt"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/history"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

const floatingPointTolerance = 1e-9

type Analyzer struct {
	historyBuilder history.Builder
}

func NewAnalyzer() Analyzer {
	return Analyzer{
		historyBuilder: history.NewBuilder(),
	}
}

func (analyzer Analyzer) Analyze(input Input) (Trend, error) {
	normalizedHistory, err := analyzer.normalizeHistory(input.History)
	if err != nil {
		return Trend{}, err
	}
	if len(normalizedHistory.Entries) < 2 {
		return Trend{}, ErrInsufficientHistory
	}
	windowDuration, err := comparableWindowDuration(normalizedHistory.Entries)
	if err != nil {
		return Trend{}, err
	}
	if input.GeneratedAt.IsZero() {
		return Trend{}, fmt.Errorf("%w: trend generated time is required", ErrInvalidTime)
	}

	generatedAt := input.GeneratedAt.UTC()
	if generatedAt.Before(normalizedHistory.GeneratedAt) {
		return Trend{}, fmt.Errorf("%w: trend cannot predate airport history", ErrInvalidTime)
	}

	baselineEntry := normalizedHistory.Entries[0]
	currentEntry := normalizedHistory.Entries[len(normalizedHistory.Entries)-1]
	peakEntry := selectPeak(normalizedHistory.Entries)

	movementsPerHourChange := currentEntry.MovementsPerHour - baselineEntry.MovementsPerHour
	changePercent, changePercentKnown := relativeChangePercent(
		baselineEntry.MovementsPerHour,
		currentEntry.MovementsPerHour,
	)

	gapCount, gapDuration, observedDuration := continuity(normalizedHistory.Entries)
	fullDuration := normalizedHistory.WindowEnd.Sub(normalizedHistory.WindowStart)
	continuityScore := 0.0
	if fullDuration > 0 {
		continuityScore = float64(observedDuration) / float64(fullDuration)
	}

	return Trend{
		ICAOCode:        normalizedHistory.ICAOCode,
		WindowStart:     normalizedHistory.WindowStart,
		WindowEnd:       normalizedHistory.WindowEnd,
		ComparedWindows: len(normalizedHistory.Entries),
		WindowDuration:  windowDuration,
		Direction:       classifyDirection(movementsPerHourChange),
		Baseline:        pointFromStatistics(baselineEntry),
		Current:         pointFromStatistics(currentEntry),
		Peak:            pointFromStatistics(peakEntry),
		TotalMovementsChange: currentEntry.TotalMovements -
			baselineEntry.TotalMovements,
		MovementsPerHourChange:             movementsPerHourChange,
		MovementsPerHourChangePercent:      changePercent,
		MovementsPerHourChangePercentKnown: changePercentKnown,
		ActiveRoutesChange:                 currentEntry.ActiveRoutes - baselineEntry.ActiveRoutes,
		CoverageScoreChange:                currentEntry.CoverageScore - baselineEntry.CoverageScore,
		FreshnessScoreChange:               currentEntry.FreshnessScore - baselineEntry.FreshnessScore,
		GapCount:                           gapCount,
		GapDuration:                        gapDuration,
		ObservedDuration:                   observedDuration,
		ContinuityScore:                    continuityScore,
		GeneratedAt:                        generatedAt,
	}, nil
}

func (analyzer Analyzer) normalizeHistory(value history.History) (history.History, error) {
	normalized, err := analyzer.historyBuilder.Build(history.Input{
		ICAOCode:    value.ICAOCode,
		Entries:     value.Entries,
		GeneratedAt: value.GeneratedAt,
	})
	if err != nil {
		return history.History{}, fmt.Errorf("%w: %v", ErrInvalidHistory, err)
	}

	if value.WindowStart.IsZero() || value.WindowEnd.IsZero() ||
		!value.WindowStart.UTC().Equal(normalized.WindowStart) ||
		!value.WindowEnd.UTC().Equal(normalized.WindowEnd) {
		return history.History{}, fmt.Errorf(
			"%w: declared history window does not match its entries",
			ErrInvalidHistory,
		)
	}

	return normalized, nil
}

func comparableWindowDuration(entries []statistics.Statistics) (time.Duration, error) {
	expected := entries[0].WindowEnd.Sub(entries[0].WindowStart)

	for index, entry := range entries[1:] {
		duration := entry.WindowEnd.Sub(entry.WindowStart)
		if duration != expected {
			return 0, fmt.Errorf(
				"%w: entry %d has %s, expected %s",
				ErrIncomparableWindows,
				index+1,
				duration,
				expected,
			)
		}
	}

	return expected, nil
}

func classifyDirection(change float64) Direction {
	switch {
	case math.Abs(change) <= floatingPointTolerance:
		return DirectionStable
	case change > 0:
		return DirectionIncreasing
	default:
		return DirectionDecreasing
	}
}

func relativeChangePercent(baseline, current float64) (float64, bool) {
	if math.Abs(baseline) <= floatingPointTolerance {
		return 0, false
	}

	return ((current - baseline) / baseline) * 100, true
}

func continuity(entries []statistics.Statistics) (int, time.Duration, time.Duration) {
	gapCount := 0
	gapDuration := time.Duration(0)
	observedDuration := time.Duration(0)

	for index, entry := range entries {
		observedDuration += entry.WindowEnd.Sub(entry.WindowStart)
		if index == 0 {
			continue
		}

		gap := entry.WindowStart.Sub(entries[index-1].WindowEnd)
		if gap > 0 {
			gapCount++
			gapDuration += gap
		}
	}

	return gapCount, gapDuration, observedDuration
}

func selectPeak(entries []statistics.Statistics) statistics.Statistics {
	peak := entries[0]

	for _, entry := range entries[1:] {
		if entry.MovementsPerHour > peak.MovementsPerHour+floatingPointTolerance {
			peak = entry
			continue
		}
		if math.Abs(entry.MovementsPerHour-peak.MovementsPerHour) <= floatingPointTolerance &&
			entry.TotalMovements > peak.TotalMovements {
			peak = entry
		}
	}

	return peak
}

func pointFromStatistics(value statistics.Statistics) Point {
	return Point{
		WindowStart:      value.WindowStart,
		WindowEnd:        value.WindowEnd,
		TotalMovements:   value.TotalMovements,
		MovementsPerHour: value.MovementsPerHour,
		ActiveRoutes:     value.ActiveRoutes,
		CoverageScore:    value.CoverageScore,
		FreshnessScore:   value.FreshnessScore,
	}
}
