package adsblol

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestMapStateResponseMapsADSBExchangeCompatiblePayload(t *testing.T) {
	var response StateResponse
	if err := json.Unmarshal([]byte(`{
		"ac":[{
			"hex":"a9cee9",
			"flight":"N731BP  ",
			"alt_baro":38000,
			"alt_geom":38275,
			"gs":338.9,
			"track":276.1,
			"baro_rate":0,
			"squawk":"3301",
			"lat":37.358322,
			"lon":-93.374147,
			"seen":0.7
		}],
		"now":1675633671226,
		"total":1
	}`), &response); err != nil {
		t.Fatal(err)
	}

	states, evidence, err := MapStateResponseWithEvidence(&response)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Accepted != 1 || evidence.RejectedMalformed != 0 {
		t.Fatalf("evidence = %+v", evidence)
	}
	if len(states) != 1 {
		t.Fatalf("states = %d", len(states))
	}

	state := states[0]
	if state.ICAO24 != "A9CEE9" {
		t.Fatalf("icao24 = %q", state.ICAO24)
	}
	if state.SourceName != "adsb.lol" {
		t.Fatalf("source = %q", state.SourceName)
	}
	if !state.VelocityAvailable || !state.HeadingAvailable ||
		!state.VerticalRateAvailable {
		t.Fatalf("telemetry availability lost: %+v", state)
	}
	if math.Abs(state.BarometricAltitudeM-11582.4) > 0.001 {
		t.Fatalf("baro altitude = %f", state.BarometricAltitudeM)
	}

	wantObservedAt := time.UnixMilli(1675633671226).
		Add(-700 * time.Millisecond).
		UTC()
	if !state.ObservedAt.Equal(wantObservedAt) {
		t.Fatalf("observed_at = %s want %s", state.ObservedAt, wantObservedAt)
	}
}

func TestMapStateResponsePreservesGroundAltitudeSemantics(t *testing.T) {
	lat := 40.0
	lon := 49.0
	seen := 0.0
	var altitude BarometricAltitude
	if err := altitude.UnmarshalJSON([]byte(`"ground"`)); err != nil {
		t.Fatal(err)
	}

	states, _, err := MapStateResponseWithEvidence(&StateResponse{
		Now: 1675633671226,
		Aircraft: []AircraftItem{{
			Hex:       "4b1801",
			Latitude:  &lat,
			Longitude: &lon,
			AltBaro:   altitude,
			Seen:      &seen,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 {
		t.Fatalf("states = %d", len(states))
	}
	if states[0].BarometricAltitudeStatus != flightstate.AltitudeStatusGround ||
		!states[0].OnGround ||
		!states[0].OnGroundAvailable {
		t.Fatalf("ground semantics lost: %+v", states[0])
	}
}

func TestMapStateResponseRejectsNonICAOAddress(t *testing.T) {
	lat := 40.0
	lon := 49.0
	seen := 0.1
	states, evidence, err := MapStateResponseWithEvidence(&StateResponse{
		Now: 1675633671226,
		Aircraft: []AircraftItem{{
			Hex:       "~abc123",
			Latitude:  &lat,
			Longitude: &lon,
			Seen:      &seen,
		}},
	})
	if err == nil {
		t.Fatal("expected non-ICAO address rejection")
	}
	if len(states) != 0 || evidence.RejectedMalformed != 1 {
		t.Fatalf("unexpected result: states=%d evidence=%+v", len(states), evidence)
	}
}

func TestMapStateResponseRejectsMissingSeen(t *testing.T) {
	lat := 40.0
	lon := 49.0
	states, evidence, err := MapStateResponseWithEvidence(&StateResponse{
		Now: 1675633671226,
		Aircraft: []AircraftItem{{
			Hex:       "4b1801",
			Latitude:  &lat,
			Longitude: &lon,
		}},
	})
	if err == nil {
		t.Fatal("expected missing seen rejection")
	}
	if len(states) != 0 || evidence.RejectedMalformed != 1 {
		t.Fatalf("unexpected result: states=%d evidence=%+v", len(states), evidence)
	}
}

func TestMapStateResponseRejectsMissingPosition(t *testing.T) {
	states, evidence, err := MapStateResponseWithEvidence(&StateResponse{
		Now: 1675633671226,
		Aircraft: []AircraftItem{{
			Hex: "4b1801",
		}},
	})
	if err == nil {
		t.Fatal("expected all-items-rejected error")
	}
	if len(states) != 0 || evidence.RejectedMalformed != 1 {
		t.Fatalf("unexpected result: states=%d evidence=%+v", len(states), evidence)
	}
}
