package overview

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/passport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/ranking"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/statistics"
)

type Assembler struct{}

func NewAssembler() Assembler {
	return Assembler{}
}

func (Assembler) Assemble(input Input) (Overview, error) {
	passportValue, passportICAO, err := normalizePassport(input.Passport)
	if err != nil {
		return Overview{}, err
	}

	statisticsValue, statisticsICAO, err := normalizeStatistics(input.Statistics)
	if err != nil {
		return Overview{}, err
	}

	if passportICAO != statisticsICAO {
		return Overview{}, fmt.Errorf(
			"%w: passport=%s statistics=%s",
			ErrAirportMismatch,
			passportICAO,
			statisticsICAO,
		)
	}

	if input.GeneratedAt.IsZero() {
		return Overview{}, fmt.Errorf("%w: generated time is required", ErrInvalidTime)
	}
	generatedAt := input.GeneratedAt.UTC()
	if generatedAt.Before(passportValue.GeneratedAt) {
		return Overview{}, fmt.Errorf("%w: overview cannot predate passport", ErrInvalidTime)
	}
	if generatedAt.Before(statisticsValue.GeneratedAt) {
		return Overview{}, fmt.Errorf("%w: overview cannot predate statistics", ErrInvalidTime)
	}

	rankingSummary, err := selectRanking(
		input.RankingResult,
		passportICAO,
		statisticsValue.WindowStart,
		statisticsValue.WindowEnd,
	)
	if err != nil {
		return Overview{}, err
	}

	return Overview{
		Passport:    passportValue,
		Statistics:  statisticsValue,
		Ranking:     rankingSummary,
		GeneratedAt: generatedAt,
	}, nil
}

func normalizePassport(value passport.Passport) (passport.Passport, string, error) {
	icao := strings.ToUpper(strings.TrimSpace(value.Identity.ICAOCode))
	if icao == "" {
		return passport.Passport{}, "", fmt.Errorf("%w: passport ICAO code is required", ErrInvalidInput)
	}
	if value.GeneratedAt.IsZero() {
		return passport.Passport{}, "", fmt.Errorf("%w: passport generated time is required", ErrInvalidTime)
	}
	if value.DataQuality.ObservedAt.IsZero() {
		return passport.Passport{}, "", fmt.Errorf("%w: passport observed time is required", ErrInvalidTime)
	}
	if value.DataQuality.ObservedAt.After(value.GeneratedAt) {
		return passport.Passport{}, "", fmt.Errorf("%w: passport observed time cannot follow generated time", ErrInvalidTime)
	}

	value.Identity.ICAOCode = icao
	value.GeneratedAt = value.GeneratedAt.UTC()
	value.DataQuality.ObservedAt = value.DataQuality.ObservedAt.UTC()

	return value, icao, nil
}

func normalizeStatistics(value statistics.Statistics) (statistics.Statistics, string, error) {
	icao := strings.ToUpper(strings.TrimSpace(value.ICAOCode))
	if icao == "" {
		return statistics.Statistics{}, "", fmt.Errorf("%w: statistics ICAO code is required", ErrInvalidInput)
	}
	if value.WindowStart.IsZero() || value.WindowEnd.IsZero() || !value.WindowEnd.After(value.WindowStart) {
		return statistics.Statistics{}, "", fmt.Errorf("%w: valid statistics window is required", ErrInvalidInput)
	}
	if value.GeneratedAt.IsZero() {
		return statistics.Statistics{}, "", fmt.Errorf("%w: statistics generated time is required", ErrInvalidTime)
	}
	if value.LatestObservationAt.IsZero() {
		return statistics.Statistics{}, "", fmt.Errorf("%w: latest observation time is required", ErrInvalidTime)
	}
	if value.LatestObservationAt.After(value.GeneratedAt) {
		return statistics.Statistics{}, "", fmt.Errorf("%w: latest observation cannot follow statistics generation", ErrInvalidTime)
	}

	value.ICAOCode = icao
	value.WindowStart = value.WindowStart.UTC()
	value.WindowEnd = value.WindowEnd.UTC()
	value.LatestObservationAt = value.LatestObservationAt.UTC()
	value.GeneratedAt = value.GeneratedAt.UTC()

	return value, icao, nil
}

func selectRanking(
	result ranking.Result,
	icao string,
	windowStart time.Time,
	windowEnd time.Time,
) (RankingSummary, error) {
	if !result.WindowStart.Equal(windowStart) || !result.WindowEnd.Equal(windowEnd) {
		return RankingSummary{}, fmt.Errorf("%w: ranking and statistics windows differ", ErrIncomparableWindow)
	}
	if len(result.Airports) == 0 {
		return RankingSummary{}, fmt.Errorf("%w: %s", ErrRankingNotFound, icao)
	}

	matches := 0
	var selected ranking.AirportRank
	for _, item := range result.Airports {
		itemICAO := strings.ToUpper(strings.TrimSpace(item.ICAOCode))
		if itemICAO != icao {
			continue
		}
		matches++
		selected = item
	}
	if matches == 0 {
		return RankingSummary{}, fmt.Errorf("%w: %s", ErrRankingNotFound, icao)
	}
	if matches > 1 {
		return RankingSummary{}, fmt.Errorf("%w: duplicate ranking entry for %s", ErrInvalidInput, icao)
	}
	if selected.Position <= 0 || selected.Position > len(result.Airports) {
		return RankingSummary{}, fmt.Errorf("%w: invalid ranking position for %s", ErrInvalidInput, icao)
	}
	if !scoreFromZeroToHundred(selected.ActivityScore) ||
		!scoreFromZeroToHundred(selected.DataConfidence) ||
		!scoreFromZeroToHundred(selected.MovementsComponent) ||
		!scoreFromZeroToHundred(selected.RoutesComponent) ||
		!scoreFromZeroToHundred(selected.ObservationsComponent) ||
		!scoreFromZeroToHundred(selected.IntensityComponent) {
		return RankingSummary{}, fmt.Errorf("%w: activity and confidence scores must be between 0 and 100", ErrInvalidInput)
	}
	if !scoreFromZeroToOne(selected.CoverageScore) ||
		!scoreFromZeroToOne(selected.FreshnessScore) {
		return RankingSummary{}, fmt.Errorf("%w: quality scores must be between 0 and 1", ErrInvalidInput)
	}
	if selected.TotalMovements < 0 || selected.ActiveRoutes < 0 ||
		selected.ObservedSamples < 0 || selected.ExpectedSamples <= 0 ||
		selected.MovementsPerHour < 0 || selected.ActiveAircraft < 0 {
		return RankingSummary{}, fmt.Errorf("%w: ranking counters are invalid", ErrInvalidInput)
	}

	return RankingSummary{
		Position:              selected.Position,
		TotalAirports:         len(result.Airports),
		ActivityScore:         selected.ActivityScore,
		DataConfidence:        selected.DataConfidence,
		MovementsComponent:    selected.MovementsComponent,
		RoutesComponent:       selected.RoutesComponent,
		ObservationsComponent: selected.ObservationsComponent,
		IntensityComponent:    selected.IntensityComponent,
		CoverageScore:         selected.CoverageScore,
		FreshnessScore:        selected.FreshnessScore,
		TotalMovements:        selected.TotalMovements,
		ActiveRoutes:          selected.ActiveRoutes,
		ObservedSamples:       selected.ObservedSamples,
		ExpectedSamples:       selected.ExpectedSamples,
		MovementsPerHour:      selected.MovementsPerHour,
		ActiveAircraft:        selected.ActiveAircraft,
	}, nil
}

func scoreFromZeroToHundred(value float64) bool {
	return value >= 0 && value <= 100
}

func scoreFromZeroToOne(value float64) bool {
	return value >= 0 && value <= 1
}
