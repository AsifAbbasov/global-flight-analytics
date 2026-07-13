package overview

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/passport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/ranking"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

func TestAssemblerBuildsAirportOverview(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.FixedZone("AZT", 4*60*60))
	windowEnd := windowStart.Add(4 * time.Hour)
	statisticsGeneratedAt := windowEnd.Add(time.Minute)
	passportGeneratedAt := statisticsGeneratedAt.Add(time.Minute)
	overviewGeneratedAt := passportGeneratedAt.Add(time.Minute)

	result, err := NewAssembler().Assemble(Input{
		Passport:      airportPassport(" ubba ", passportGeneratedAt),
		Statistics:    airportStatistics("UBBA", windowStart, windowEnd, statisticsGeneratedAt),
		RankingResult: airportRanking(windowStart, windowEnd),
		GeneratedAt:   overviewGeneratedAt,
	})
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	if result.Passport.Identity.ICAOCode != "UBBA" {
		t.Fatalf("passport ICAO = %q, want UBBA", result.Passport.Identity.ICAOCode)
	}
	if result.Statistics.ICAOCode != "UBBA" {
		t.Fatalf("statistics ICAO = %q, want UBBA", result.Statistics.ICAOCode)
	}
	if result.Ranking.Position != 2 || result.Ranking.TotalAirports != 3 {
		t.Fatalf("ranking = position %d of %d, want 2 of 3", result.Ranking.Position, result.Ranking.TotalAirports)
	}
	if result.Ranking.ActivityScore != 70 || result.Ranking.DataConfidence != 85 {
		t.Fatalf("ranking scores = activity %.2f confidence %.2f, want 70 and 85", result.Ranking.ActivityScore, result.Ranking.DataConfidence)
	}
	if result.Ranking.ActiveRoutes != 4 || result.Ranking.ObservedSamples != 80 {
		t.Fatalf("ranking evidence = routes %d observations %d, want 4 and 80", result.Ranking.ActiveRoutes, result.Ranking.ObservedSamples)
	}
	if result.GeneratedAt.Location() != time.UTC || result.Statistics.WindowStart.Location() != time.UTC {
		t.Fatal("overview times must be normalized to UTC")
	}
}

func TestAssemblerDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)
	generatedAt := windowEnd.Add(time.Minute)
	passportValue := airportPassport(" ubba ", generatedAt)
	statisticsValue := airportStatistics(" ubba ", windowStart, windowEnd, generatedAt)

	_, err := NewAssembler().Assemble(Input{
		Passport:      passportValue,
		Statistics:    statisticsValue,
		RankingResult: airportRanking(windowStart, windowEnd),
		GeneratedAt:   generatedAt,
	})
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}
	if passportValue.Identity.ICAOCode != " ubba " || statisticsValue.ICAOCode != " ubba " {
		t.Fatal("Assemble() mutated input")
	}
}

func TestAssemblerRejectsInvalidCombinations(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)
	generatedAt := windowEnd.Add(2 * time.Minute)

	tests := []struct {
		name  string
		input Input
		want  error
	}{
		{
			name:  "airport mismatch",
			input: Input{Passport: airportPassport("UBBA", generatedAt), Statistics: airportStatistics("UBBB", windowStart, windowEnd, generatedAt), RankingResult: airportRanking(windowStart, windowEnd), GeneratedAt: generatedAt},
			want:  ErrAirportMismatch,
		},
		{
			name:  "missing ranking",
			input: Input{Passport: airportPassport("UBBA", generatedAt), Statistics: airportStatistics("UBBA", windowStart, windowEnd, generatedAt), RankingResult: ranking.Result{WindowStart: windowStart, WindowEnd: windowEnd, Airports: []ranking.AirportRank{validRank("UBBB", 1)}}, GeneratedAt: generatedAt},
			want:  ErrRankingNotFound,
		},
		{
			name:  "different windows",
			input: Input{Passport: airportPassport("UBBA", generatedAt), Statistics: airportStatistics("UBBA", windowStart, windowEnd, generatedAt), RankingResult: airportRanking(windowStart, windowEnd.Add(time.Hour)), GeneratedAt: generatedAt},
			want:  ErrIncomparableWindow,
		},
		{
			name:  "overview predates sources",
			input: Input{Passport: airportPassport("UBBA", generatedAt), Statistics: airportStatistics("UBBA", windowStart, windowEnd, generatedAt), RankingResult: airportRanking(windowStart, windowEnd), GeneratedAt: generatedAt.Add(-time.Second)},
			want:  ErrInvalidTime,
		},
		{
			name:  "invalid activity score",
			input: inputWithMutatedRank(windowStart, windowEnd, generatedAt, func(value *ranking.AirportRank) { value.ActivityScore = 101 }),
			want:  ErrInvalidInput,
		},
		{
			name:  "invalid confidence score",
			input: inputWithMutatedRank(windowStart, windowEnd, generatedAt, func(value *ranking.AirportRank) { value.DataConfidence = -1 }),
			want:  ErrInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewAssembler().Assemble(test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("Assemble() error = %v, want %v", err, test.want)
			}
		})
	}
}

func airportPassport(icao string, generatedAt time.Time) passport.Passport {
	return passport.Passport{
		Identity:    passport.Identity{ICAOCode: icao, IATACode: "GYD", Name: "Heydar Aliyev International Airport"},
		DataQuality: passport.DataQuality{FreshnessScore: 0.9, CoverageScore: 0.8, ObservedAt: generatedAt.Add(-time.Minute)},
		GeneratedAt: generatedAt,
	}
}

func airportStatistics(icao string, windowStart, windowEnd, generatedAt time.Time) statistics.Statistics {
	return statistics.Statistics{
		ICAOCode: icao, WindowStart: windowStart, WindowEnd: windowEnd,
		Arrivals: 10, Departures: 8, TotalMovements: 18,
		ArrivalShare: 10.0 / 18.0, DepartureShare: 8.0 / 18.0,
		MovementsPerHour: 4.5, ActiveAircraft: 6, ActiveRoutes: 4,
		ObservedSamples: 80, ExpectedSamples: 100,
		CoverageScore: 0.8, FreshnessScore: 0.9,
		LatestObservationAt: generatedAt.Add(-time.Minute), GeneratedAt: generatedAt,
	}
}

func airportRanking(windowStart, windowEnd time.Time) ranking.Result {
	return ranking.Result{
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		Airports: []ranking.AirportRank{
			validRank("UBBB", 1),
			validRank(" ubba ", 2),
			validRank("UGTB", 3),
		},
	}
}

func validRank(icao string, position int) ranking.AirportRank {
	return ranking.AirportRank{
		Position: position, ICAOCode: icao,
		ActivityScore: 70, DataConfidence: 85,
		MovementsComponent: 75, RoutesComponent: 65, ObservationsComponent: 70, IntensityComponent: 70,
		CoverageScore: 0.8, FreshnessScore: 0.9,
		TotalMovements: 18, ActiveRoutes: 4, ObservedSamples: 80, ExpectedSamples: 100,
		MovementsPerHour: 4.5, ActiveAircraft: 6,
	}
}

func inputWithMutatedRank(windowStart, windowEnd, generatedAt time.Time, mutate func(*ranking.AirportRank)) Input {
	result := airportRanking(windowStart, windowEnd)
	mutate(&result.Airports[1])
	return Input{
		Passport:      airportPassport("UBBA", generatedAt),
		Statistics:    airportStatistics("UBBA", windowStart, windowEnd, generatedAt),
		RankingResult: result,
		GeneratedAt:   generatedAt,
	}
}
