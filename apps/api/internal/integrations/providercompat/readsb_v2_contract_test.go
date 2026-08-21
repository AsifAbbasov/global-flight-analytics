package providercompat_test

import (
	"encoding/json"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/adsblol"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/airplaneslive"
)

const overlappingV2Fixture = `{
  "now": 1787274000000,
  "messages": 12,
  "total": 1,
  "ac": [
    {
      "hex": "4ca1b2",
      "flight": " AZAL123 ",
      "lat": 40.4093,
      "lon": 49.8671,
      "alt_baro": 32000,
      "alt_geom": 32400,
      "gs": 430,
      "track": 92.5,
      "baro_rate": -640,
      "seen": 1.25,
      "type": "adsb_icao",
      "r": "4K-A1B2",
      "t": "A320",
      "squawk": "1234"
    }
  ]
}`

func normalizeSource(
	state flightstate.FlightState,
) flightstate.FlightState {
	state.SourceName = ""
	return state
}

func TestADSBLOLAndAirplanesLivePreserveOverlappingV2CanonicalSemantics(
	t *testing.T,
) {
	var adsbResponse adsblol.StateResponse
	if err := json.Unmarshal(
		[]byte(overlappingV2Fixture),
		&adsbResponse,
	); err != nil {
		t.Fatalf("decode ADSB.lol fixture: %v", err)
	}

	var airplanesResponse airplaneslive.StateResponse
	if err := json.Unmarshal(
		[]byte(overlappingV2Fixture),
		&airplanesResponse,
	); err != nil {
		t.Fatalf("decode airplanes.live fixture: %v", err)
	}

	adsbStates, adsbEvidence, err :=
		adsblol.MapStateResponseWithEvidence(&adsbResponse)
	if err != nil {
		t.Fatalf("map ADSB.lol fixture: %v", err)
	}

	airplanesStates, airplanesEvidence, err :=
		airplaneslive.MapStateResponseWithEvidence(&airplanesResponse)
	if err != nil {
		t.Fatalf("map airplanes.live fixture: %v", err)
	}

	if adsbEvidence != airplanesEvidence {
		t.Fatalf(
			"batch evidence drift: adsb.lol=%+v airplanes.live=%+v",
			adsbEvidence,
			airplanesEvidence,
		)
	}
	if len(adsbStates) != 1 || len(airplanesStates) != 1 {
		t.Fatalf(
			"unexpected state counts: adsb.lol=%d airplanes.live=%d",
			len(adsbStates),
			len(airplanesStates),
		)
	}

	if adsbStates[0].SourceName != "adsb.lol" {
		t.Fatalf(
			"ADSB.lol source = %q",
			adsbStates[0].SourceName,
		)
	}
	if airplanesStates[0].SourceName != "airplanes.live" {
		t.Fatalf(
			"airplanes.live source = %q",
			airplanesStates[0].SourceName,
		)
	}

	if normalizeSource(adsbStates[0]) != normalizeSource(airplanesStates[0]) {
		t.Fatalf(
			"overlapping v2 canonical mapping drift:\nadsb.lol=%+v\nairplanes.live=%+v",
			adsbStates[0],
			airplanesStates[0],
		)
	}
}
