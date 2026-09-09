package readsbcompat

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestPositionSourceMapsReadsbEvidence(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want flightstate.PositionSource
	}{
		{name: "adsb icao", raw: "adsb_icao", want: flightstate.PositionSourceADSB},
		{name: "adsb non transponder", raw: "adsb_icao_nt", want: flightstate.PositionSourceADSB},
		{name: "adsb other", raw: "adsb_other", want: flightstate.PositionSourceADSB},
		{name: "adsr icao", raw: "adsr_icao", want: flightstate.PositionSourceADSR},
		{name: "adsr other", raw: "adsr_other", want: flightstate.PositionSourceADSR},
		{name: "tisb icao", raw: "tisb_icao", want: flightstate.PositionSourceTISB},
		{name: "tisb other", raw: "tisb_other", want: flightstate.PositionSourceTISB},
		{name: "tisb trackfile", raw: "tisb_trackfile", want: flightstate.PositionSourceTISB},
		{name: "adsc", raw: "adsc", want: flightstate.PositionSourceADSC},
		{name: "mlat", raw: "mlat", want: flightstate.PositionSourceMLAT},
		{name: "mode s has no position method", raw: "mode_s", want: flightstate.PositionSourceUnknown},
		{name: "other quality unknown", raw: "other", want: flightstate.PositionSourceUnknown},
		{name: "unrecognized", raw: "future_type", want: flightstate.PositionSourceUnknown},
		{name: "empty", raw: "", want: flightstate.PositionSourceUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := PositionSource(test.raw); got != test.want {
				t.Fatalf("PositionSource(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

func TestMapAircraftPreservesPositionSourceEvidence(t *testing.T) {
	mapped := MapAircraft(
		"adsb.lol",
		AircraftItem{
			Hex:       "4b1801",
			Latitude:  40.4,
			Longitude: 49.8,
			Type:      "mlat",
		},
		time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC),
	)

	if mapped.PositionSource != flightstate.PositionSourceMLAT {
		t.Fatalf("position source = %q, want %q", mapped.PositionSource, flightstate.PositionSourceMLAT)
	}
	if mapped.SourceName != "adsb.lol" {
		t.Fatalf("source name = %q, want adsb.lol", mapped.SourceName)
	}
}
