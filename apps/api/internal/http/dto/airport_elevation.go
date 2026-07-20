package dto

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"

func ToAirportElevation(
	value float64,
	available bool,
) (
	*float64,
	airport.ElevationStatus,
) {
	normalized, status, present := airport.ResolveElevation(value, available)
	if !present {
		return nil, status
	}

	result := normalized
	return &result, status
}
