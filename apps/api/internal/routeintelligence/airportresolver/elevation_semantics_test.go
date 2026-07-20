package airportresolver

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

func TestNormalizeAirportPreservesUnknownAndObservedZeroElevation(t *testing.T) {
	base := airport.Airport{
		ICAOCode:  "TEST",
		Name:      "Test Airport",
		Latitude:  1,
		Longitude: 2,
	}

	unknown, _, valid := normalizeAirport(base)
	if !valid || unknown.ElevationAvailable || unknown.ElevationM != 0 {
		t.Fatalf("unknown elevation was not preserved: %#v", unknown)
	}

	base.ElevationAvailable = true
	observed, _, valid := normalizeAirport(base)
	if !valid || !observed.ElevationAvailable || observed.ElevationM != 0 {
		t.Fatalf("observed sea-level elevation was not preserved: %#v", observed)
	}
}
