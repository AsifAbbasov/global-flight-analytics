package history

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

func TestBuilderBuildsSortedHistoryWithoutMutatingInput(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("UTC+4", 4*60*60)
	firstStart := time.Date(2026, time.July, 12, 0, 0, 0, 0, location)
	secondStart := firstStart.Add(24 * time.Hour)
	generatedAt := secondStart.Add(25 * time.Hour)

	input := Input{
		ICAOCode: " ubbb ",
		Entries: []statistics.Statistics{
			validStatistics("ubbb", secondStart, secondStart.Add(24*time.Hour), 12, 8),
			validStatistics("UBBB", firstStart, firstStart.Add(24*time.Hour), 10, 10),
		},
		GeneratedAt: generatedAt,
	}
	originalEntries := append([]statistics.Statistics(nil), input.Entries...)

	result, err := NewBuilder().Build(input)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result.ICAOCode != "UBBB" {
		t.Fatalf("ICAOCode = %q, want UBBB", result.ICAOCode)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(result.Entries))
	}
	if !result.Entries[0].WindowStart.Equal(firstStart) {
		t.Fatalf("first WindowStart = %v, want %v", result.Entries[0].WindowStart, firstStart)
	}
	if !result.Entries[1].WindowStart.Equal(secondStart) {
		t.Fatalf("second WindowStart = %v, want %v", result.Entries[1].WindowStart, secondStart)
	}
	if result.WindowStart.Location() != time.UTC || result.WindowEnd.Location() != time.UTC || result.GeneratedAt.Location() != time.UTC {
		t.Fatal("history times must be normalized to UTC")
	}
	if !reflect.DeepEqual(input.Entries, originalEntries) {
		t.Fatal("Build() mutated input entries")
	}
}

func TestBuilderAllowsGapsAndAdjacentWindows(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	entries := []statistics.Statistics{
		validStatistics("UBBB", start, start.Add(time.Hour), 2, 2),
		validStatistics("UBBB", start.Add(time.Hour), start.Add(2*time.Hour), 3, 1),
		validStatistics("UBBB", start.Add(4*time.Hour), start.Add(5*time.Hour), 1, 1),
	}

	result, err := NewBuilder().Build(Input{
		ICAOCode:    "UBBB",
		Entries:     entries,
		GeneratedAt: start.Add(6 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(result.Entries) != 3 {
		t.Fatalf("len(Entries) = %d, want 3", len(result.Entries))
	}
}

func TestBuilderRejectsEmptyHistory(t *testing.T) {
	t.Parallel()

	_, err := NewBuilder().Build(Input{
		ICAOCode:    "UBBB",
		GeneratedAt: time.Now(),
	})
	if !errors.Is(err, ErrEmptyHistory) {
		t.Fatalf("error = %v, want ErrEmptyHistory", err)
	}
}

func TestBuilderRejectsAirportMismatch(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	_, err := NewBuilder().Build(Input{
		ICAOCode: "UBBB",
		Entries: []statistics.Statistics{
			validStatistics("UGTB", start, start.Add(time.Hour), 2, 2),
		},
		GeneratedAt: start.Add(2 * time.Hour),
	})
	if !errors.Is(err, ErrAirportMismatch) {
		t.Fatalf("error = %v, want ErrAirportMismatch", err)
	}
}

func TestBuilderRejectsDuplicateWindow(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	entry := validStatistics("UBBB", start, start.Add(time.Hour), 2, 2)

	_, err := NewBuilder().Build(Input{
		ICAOCode:    "UBBB",
		Entries:     []statistics.Statistics{entry, entry},
		GeneratedAt: start.Add(2 * time.Hour),
	})
	if !errors.Is(err, ErrDuplicateWindow) {
		t.Fatalf("error = %v, want ErrDuplicateWindow", err)
	}
}

func TestBuilderRejectsOverlappingWindows(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	_, err := NewBuilder().Build(Input{
		ICAOCode: "UBBB",
		Entries: []statistics.Statistics{
			validStatistics("UBBB", start, start.Add(2*time.Hour), 2, 2),
			validStatistics("UBBB", start.Add(time.Hour), start.Add(3*time.Hour), 3, 1),
		},
		GeneratedAt: start.Add(4 * time.Hour),
	})
	if !errors.Is(err, ErrOverlappingWindow) {
		t.Fatalf("error = %v, want ErrOverlappingWindow", err)
	}
}

func TestBuilderRejectsInconsistentDerivedValues(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	entry := validStatistics("UBBB", start, start.Add(time.Hour), 2, 2)
	entry.TotalMovements++

	_, err := NewBuilder().Build(Input{
		ICAOCode:    "UBBB",
		Entries:     []statistics.Statistics{entry},
		GeneratedAt: start.Add(2 * time.Hour),
	})
	if !errors.Is(err, ErrInvalidEntry) {
		t.Fatalf("error = %v, want ErrInvalidEntry", err)
	}
}

func TestBuilderRejectsHistoryGeneratedBeforeEntry(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	entry := validStatistics("UBBB", start, start.Add(time.Hour), 2, 2)

	_, err := NewBuilder().Build(Input{
		ICAOCode:    "UBBB",
		Entries:     []statistics.Statistics{entry},
		GeneratedAt: entry.GeneratedAt.Add(-time.Minute),
	})
	if !errors.Is(err, ErrInvalidTime) {
		t.Fatalf("error = %v, want ErrInvalidTime", err)
	}
}

func validStatistics(
	icaoCode string,
	windowStart time.Time,
	windowEnd time.Time,
	arrivals int,
	departures int,
) statistics.Statistics {
	totalMovements := arrivals + departures
	arrivalShare := 0.0
	departureShare := 0.0
	if totalMovements > 0 {
		arrivalShare = float64(arrivals) / float64(totalMovements)
		departureShare = float64(departures) / float64(totalMovements)
	}

	latestObservationAt := windowEnd.Add(-time.Minute)
	generatedAt := windowEnd.Add(time.Minute)

	return statistics.Statistics{
		ICAOCode:            icaoCode,
		WindowStart:         windowStart,
		WindowEnd:           windowEnd,
		Arrivals:            arrivals,
		Departures:          departures,
		TotalMovements:      totalMovements,
		ArrivalShare:        arrivalShare,
		DepartureShare:      departureShare,
		MovementsPerHour:    float64(totalMovements) / windowEnd.Sub(windowStart).Hours(),
		ActiveAircraft:      3,
		ObservedSamples:     90,
		ExpectedSamples:     100,
		CoverageScore:       0.9,
		FreshnessScore:      0.8,
		LatestObservationAt: latestObservationAt,
		GeneratedAt:         generatedAt,
	}
}
