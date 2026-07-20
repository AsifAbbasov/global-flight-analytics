package dto

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestAirportElevationDTOSeparatesUnknownAndObservedZero(t *testing.T) {
	unknownValue, unknownStatus := ToAirportElevation(0, false)
	if unknownValue != nil || unknownStatus != airport.ElevationStatusUnknown {
		t.Fatalf("unexpected unknown elevation DTO: value=%v status=%q", unknownValue, unknownStatus)
	}

	zeroValue, zeroStatus := ToAirportElevation(0, true)
	if zeroValue == nil || *zeroValue != 0 || zeroStatus != airport.ElevationStatusObserved {
		t.Fatalf("unexpected observed zero DTO: value=%v status=%q", zeroValue, zeroStatus)
	}
}

func TestRouteIntelligenceAirportPublishesUnknownElevationAsNull(t *testing.T) {
	endpoint := toRouteIntelligenceEndpoint(&routecontract.EndpointInference{
		Airport: routecontract.AirportReference{ICAOCode: "TEST"},
	})
	if endpoint == nil || endpoint.Airport.ElevationM != nil || endpoint.Airport.ElevationStatus != airport.ElevationStatusUnknown {
		t.Fatalf("unexpected route intelligence airport DTO: %#v", endpoint)
	}
}
