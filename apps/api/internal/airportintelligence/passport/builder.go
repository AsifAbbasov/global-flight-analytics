package passport

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

type Builder struct {
	airportActivity metrics.AirportActivity
}

func NewBuilder() Builder {
	return Builder{
		airportActivity: metrics.AirportActivity{},
	}
}

func (builder Builder) Build(
	source airport.Airport,
	analytics AnalyticsInput,
	generatedAt time.Time,
) (Passport, error) {
	identity := Identity{
		ICAOCode: strings.ToUpper(strings.TrimSpace(source.ICAOCode)),
		IATACode: strings.ToUpper(strings.TrimSpace(source.IATACode)),
		Name:     strings.TrimSpace(source.Name),
	}
	if identity.ICAOCode == "" {
		return Passport{}, fmt.Errorf("%w: ICAO code is required", ErrInvalidIdentity)
	}
	if identity.Name == "" {
		return Passport{}, fmt.Errorf("%w: name is required", ErrInvalidIdentity)
	}
	if source.Latitude < -90 || source.Latitude > 90 {
		return Passport{}, fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidCoordinates)
	}
	if source.Longitude < -180 || source.Longitude > 180 {
		return Passport{}, fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidCoordinates)
	}
	if analytics.Arrivals < 0 || analytics.Departures < 0 || analytics.ActiveAircraft < 0 {
		return Passport{}, fmt.Errorf("%w: movement counters cannot be negative", ErrInvalidOperations)
	}
	if analytics.FreshnessScore < 0 || analytics.FreshnessScore > 1 {
		return Passport{}, fmt.Errorf("%w: freshness score must be between 0 and 1", ErrInvalidDataQuality)
	}
	if analytics.CoverageScore < 0 || analytics.CoverageScore > 1 {
		return Passport{}, fmt.Errorf("%w: coverage score must be between 0 and 1", ErrInvalidDataQuality)
	}
	if generatedAt.IsZero() {
		return Passport{}, fmt.Errorf("%w: generated time is required", ErrInvalidTime)
	}
	if analytics.ObservedAt.IsZero() {
		return Passport{}, fmt.Errorf("%w: observed time is required", ErrInvalidTime)
	}
	if analytics.ObservedAt.After(generatedAt) {
		return Passport{}, fmt.Errorf("%w: observed time cannot be after generated time", ErrInvalidTime)
	}

	elevationM, _, elevationAvailable := airport.ResolveElevation(
		source.ElevationM,
		source.ElevationAvailable,
	)

	return Passport{
		Identity: identity,
		Location: Location{
			City:               strings.TrimSpace(source.City),
			Country:            strings.TrimSpace(source.Country),
			Latitude:           source.Latitude,
			Longitude:          source.Longitude,
			ElevationM:         elevationM,
			ElevationAvailable: elevationAvailable,
			Timezone:           strings.TrimSpace(source.Timezone),
		},
		Operations: Operations{
			Arrivals:       analytics.Arrivals,
			Departures:     analytics.Departures,
			Activity:       builder.airportActivity.Calculate(analytics.Arrivals, analytics.Departures),
			ActiveAircraft: analytics.ActiveAircraft,
		},
		DataQuality: DataQuality{
			FreshnessScore: analytics.FreshnessScore,
			CoverageScore:  analytics.CoverageScore,
			ObservedAt:     analytics.ObservedAt.UTC(),
		},
		Description: strings.TrimSpace(source.Description),
		GeneratedAt: generatedAt.UTC(),
	}, nil
}
