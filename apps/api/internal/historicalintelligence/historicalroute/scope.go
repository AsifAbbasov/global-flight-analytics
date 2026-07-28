package historicalroute

import (
	"regexp"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

var airportICAOPattern = regexp.MustCompile(`^[A-Z0-9]{4}$`)

func normalizeScope(
	origin string,
	destination string,
) (historicalcontract.Scope, string, string, error) {
	origin = normalizedAirportCode(origin)
	destination = normalizedAirportCode(destination)

	if origin == "" && destination == "" {
		return historicalcontract.Scope{Type: historicalcontract.ScopeTypeGlobal}, "", "", nil
	}
	if origin == "" || destination == "" {
		return historicalcontract.Scope{}, "", "", ErrRouteScopeIncomplete
	}
	if !airportICAOPattern.MatchString(origin) {
		return historicalcontract.Scope{}, "", "", ErrOriginICAOInvalid
	}
	if !airportICAOPattern.MatchString(destination) {
		return historicalcontract.Scope{}, "", "", ErrDestinationICAOInvalid
	}

	return historicalcontract.Scope{
		Type:                historicalcontract.ScopeTypeRoute,
		OriginICAOCode:      origin,
		DestinationICAOCode: destination,
	}, origin, destination, nil
}

func normalizedAirportCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
