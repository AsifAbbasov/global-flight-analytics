package normalizer

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func NormalizeFlightState(item flightstate.FlightState) flightstate.FlightState {
	item.ICAO24 = strings.ToUpper(strings.TrimSpace(item.ICAO24))
	item.Callsign = strings.ToUpper(strings.TrimSpace(item.Callsign))
	item.SourceName = strings.ToLower(strings.TrimSpace(item.SourceName))

	return item
}

func NormalizeFlightStates(items []flightstate.FlightState) []flightstate.FlightState {
	result := make([]flightstate.FlightState, 0, len(items))

	for _, item := range items {
		result = append(result, NormalizeFlightState(item))
	}

	return result
}
