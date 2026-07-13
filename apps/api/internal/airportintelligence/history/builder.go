package history

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

const floatingPointTolerance = 1e-9

type Builder struct{}

func NewBuilder() Builder {
	return Builder{}
}

func (Builder) Build(input Input) (History, error) {
	icaoCode := strings.ToUpper(strings.TrimSpace(input.ICAOCode))
	if icaoCode == "" {
		return History{}, fmt.Errorf("%w: ICAO code is required", ErrInvalidIdentity)
	}
	if input.GeneratedAt.IsZero() {
		return History{}, fmt.Errorf("%w: history generated time is required", ErrInvalidTime)
	}
	if len(input.Entries) == 0 {
		return History{}, ErrEmptyHistory
	}

	generatedAt := input.GeneratedAt.UTC()
	entries := make([]statistics.Statistics, len(input.Entries))
	copy(entries, input.Entries)

	for index := range entries {
		normalized, err := normalizeEntry(entries[index], icaoCode, generatedAt)
		if err != nil {
			return History{}, fmt.Errorf("entry %d: %w", index, err)
		}
		entries[index] = normalized
	}

	sort.Slice(entries, func(left, right int) bool {
		if entries[left].WindowStart.Equal(entries[right].WindowStart) {
			return entries[left].WindowEnd.Before(entries[right].WindowEnd)
		}
		return entries[left].WindowStart.Before(entries[right].WindowStart)
	})

	for index := 1; index < len(entries); index++ {
		previous := entries[index-1]
		current := entries[index]

		if current.WindowStart.Equal(previous.WindowStart) && current.WindowEnd.Equal(previous.WindowEnd) {
			return History{}, fmt.Errorf(
				"%w: %s to %s",
				ErrDuplicateWindow,
				current.WindowStart.Format(time.RFC3339),
				current.WindowEnd.Format(time.RFC3339),
			)
		}
		if current.WindowStart.Before(previous.WindowEnd) {
			return History{}, fmt.Errorf(
				"%w: previous ends at %s, current starts at %s",
				ErrOverlappingWindow,
				previous.WindowEnd.Format(time.RFC3339),
				current.WindowStart.Format(time.RFC3339),
			)
		}
	}

	return History{
		ICAOCode:    icaoCode,
		WindowStart: entries[0].WindowStart,
		WindowEnd:   entries[len(entries)-1].WindowEnd,
		Entries:     entries,
		GeneratedAt: generatedAt,
	}, nil
}

func normalizeEntry(
	entry statistics.Statistics,
	expectedICAO string,
	historyGeneratedAt time.Time,
) (statistics.Statistics, error) {
	icaoCode := strings.ToUpper(strings.TrimSpace(entry.ICAOCode))
	if icaoCode == "" {
		return statistics.Statistics{}, fmt.Errorf("%w: entry ICAO code is required", ErrInvalidEntry)
	}
	if icaoCode != expectedICAO {
		return statistics.Statistics{}, fmt.Errorf(
			"%w: expected %s, got %s",
			ErrAirportMismatch,
			expectedICAO,
			icaoCode,
		)
	}

	if entry.WindowStart.IsZero() || entry.WindowEnd.IsZero() || !entry.WindowEnd.After(entry.WindowStart) {
		return statistics.Statistics{}, fmt.Errorf("%w: valid statistics window is required", ErrInvalidEntry)
	}
	if entry.LatestObservationAt.IsZero() || entry.GeneratedAt.IsZero() {
		return statistics.Statistics{}, fmt.Errorf("%w: observation and generation times are required", ErrInvalidTime)
	}

	entry.ICAOCode = icaoCode
	entry.WindowStart = entry.WindowStart.UTC()
	entry.WindowEnd = entry.WindowEnd.UTC()
	entry.LatestObservationAt = entry.LatestObservationAt.UTC()
	entry.GeneratedAt = entry.GeneratedAt.UTC()

	if entry.LatestObservationAt.Before(entry.WindowStart) || entry.LatestObservationAt.After(entry.WindowEnd) {
		return statistics.Statistics{}, fmt.Errorf("%w: latest observation must be inside the statistics window", ErrInvalidTime)
	}
	if entry.LatestObservationAt.After(entry.GeneratedAt) {
		return statistics.Statistics{}, fmt.Errorf("%w: latest observation cannot follow entry generation", ErrInvalidTime)
	}
	if entry.GeneratedAt.After(historyGeneratedAt) {
		return statistics.Statistics{}, fmt.Errorf("%w: history cannot predate an entry", ErrInvalidTime)
	}

	if entry.Arrivals < 0 || entry.Departures < 0 || entry.ActiveAircraft < 0 || entry.ActiveRoutes < 0 {
		return statistics.Statistics{}, fmt.Errorf("%w: operational counters cannot be negative", ErrInvalidEntry)
	}
	if entry.ObservedSamples < 0 || entry.ExpectedSamples <= 0 {
		return statistics.Statistics{}, fmt.Errorf("%w: invalid sample counters", ErrInvalidEntry)
	}

	expectedTotalMovements := entry.Arrivals + entry.Departures
	if entry.TotalMovements != expectedTotalMovements {
		return statistics.Statistics{}, fmt.Errorf(
			"%w: total movements must equal arrivals plus departures",
			ErrInvalidEntry,
		)
	}

	windowHours := entry.WindowEnd.Sub(entry.WindowStart).Hours()
	expectedMovementsPerHour := float64(entry.TotalMovements) / windowHours
	if !approximatelyEqual(entry.MovementsPerHour, expectedMovementsPerHour) {
		return statistics.Statistics{}, fmt.Errorf("%w: inconsistent movements per hour", ErrInvalidEntry)
	}

	if !scoreInRange(entry.CoverageScore) || !scoreInRange(entry.FreshnessScore) {
		return statistics.Statistics{}, fmt.Errorf("%w: quality scores must be between 0 and 1", ErrInvalidEntry)
	}

	if entry.TotalMovements == 0 {
		if !approximatelyEqual(entry.ArrivalShare, 0) || !approximatelyEqual(entry.DepartureShare, 0) {
			return statistics.Statistics{}, fmt.Errorf("%w: movement shares must be zero without movements", ErrInvalidEntry)
		}
		return entry, nil
	}

	expectedArrivalShare := float64(entry.Arrivals) / float64(entry.TotalMovements)
	expectedDepartureShare := float64(entry.Departures) / float64(entry.TotalMovements)
	if !approximatelyEqual(entry.ArrivalShare, expectedArrivalShare) ||
		!approximatelyEqual(entry.DepartureShare, expectedDepartureShare) {
		return statistics.Statistics{}, fmt.Errorf("%w: inconsistent movement shares", ErrInvalidEntry)
	}

	return entry, nil
}

func scoreInRange(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func approximatelyEqual(left, right float64) bool {
	return math.Abs(left-right) <= floatingPointTolerance
}
