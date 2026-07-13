package statistics

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

type Calculator struct {
	airportActivity metrics.AirportActivity
	coverageScore   metrics.CoverageScoreMetric
	dataFreshness   metrics.DataFreshnessMetric
}

func NewCalculator(maxDataAge time.Duration) (Calculator, error) {
	if maxDataAge <= 0 {
		return Calculator{}, fmt.Errorf("%w: maximum data age must be greater than zero", ErrInvalidConfiguration)
	}

	return Calculator{
		airportActivity: metrics.AirportActivity{},
		coverageScore:   metrics.CoverageScoreMetric{},
		dataFreshness: metrics.DataFreshnessMetric{
			MaxAge: maxDataAge,
		},
	}, nil
}

func (calculator Calculator) Calculate(input Input) (Statistics, error) {
	icaoCode := strings.ToUpper(strings.TrimSpace(input.ICAOCode))
	if icaoCode == "" {
		return Statistics{}, fmt.Errorf("%w: ICAO code is required", ErrInvalidIdentity)
	}

	if input.WindowStart.IsZero() || input.WindowEnd.IsZero() {
		return Statistics{}, fmt.Errorf("%w: start and end times are required", ErrInvalidWindow)
	}
	if !input.WindowEnd.After(input.WindowStart) {
		return Statistics{}, fmt.Errorf("%w: end time must be after start time", ErrInvalidWindow)
	}

	if input.Arrivals < 0 || input.Departures < 0 || input.ActiveAircraft < 0 || input.ActiveRoutes < 0 {
		return Statistics{}, fmt.Errorf("%w: operational counters cannot be negative", ErrInvalidCounters)
	}
	if input.ObservedSamples < 0 {
		return Statistics{}, fmt.Errorf("%w: observed sample count cannot be negative", ErrInvalidCounters)
	}
	if input.ExpectedSamples <= 0 {
		return Statistics{}, fmt.Errorf("%w: expected sample count must be greater than zero", ErrInvalidCounters)
	}

	if input.LatestObservationAt.IsZero() || input.GeneratedAt.IsZero() {
		return Statistics{}, fmt.Errorf("%w: observation and generation times are required", ErrInvalidTime)
	}
	if input.LatestObservationAt.Before(input.WindowStart) || input.LatestObservationAt.After(input.WindowEnd) {
		return Statistics{}, fmt.Errorf("%w: latest observation must be inside the statistics window", ErrInvalidTime)
	}
	if input.LatestObservationAt.After(input.GeneratedAt) {
		return Statistics{}, fmt.Errorf("%w: latest observation cannot be after generation time", ErrInvalidTime)
	}

	data := snapshot.Snapshot{
		Time:            input.LatestObservationAt,
		ObservedSamples: input.ObservedSamples,
		ExpectedSamples: input.ExpectedSamples,
	}

	coverageScore, err := calculator.coverageScore.Calculate(data)
	if err != nil {
		return Statistics{}, fmt.Errorf("calculate coverage score: %w", err)
	}

	freshnessScore, err := calculator.dataFreshness.Calculate(data, input.GeneratedAt)
	if err != nil {
		return Statistics{}, fmt.Errorf("calculate freshness score: %w", err)
	}

	totalMovements := calculator.airportActivity.Calculate(input.Arrivals, input.Departures)
	arrivalShare := 0.0
	departureShare := 0.0
	if totalMovements > 0 {
		arrivalShare = float64(input.Arrivals) / float64(totalMovements)
		departureShare = float64(input.Departures) / float64(totalMovements)
	}

	windowHours := input.WindowEnd.Sub(input.WindowStart).Hours()

	return Statistics{
		ICAOCode:            icaoCode,
		WindowStart:         input.WindowStart.UTC(),
		WindowEnd:           input.WindowEnd.UTC(),
		Arrivals:            input.Arrivals,
		Departures:          input.Departures,
		TotalMovements:      totalMovements,
		ArrivalShare:        arrivalShare,
		DepartureShare:      departureShare,
		MovementsPerHour:    float64(totalMovements) / windowHours,
		ActiveAircraft:      input.ActiveAircraft,
		ActiveRoutes:        input.ActiveRoutes,
		ObservedSamples:     input.ObservedSamples,
		ExpectedSamples:     input.ExpectedSamples,
		CoverageScore:       coverageScore,
		FreshnessScore:      freshnessScore,
		LatestObservationAt: input.LatestObservationAt.UTC(),
		GeneratedAt:         input.GeneratedAt.UTC(),
	}, nil
}
