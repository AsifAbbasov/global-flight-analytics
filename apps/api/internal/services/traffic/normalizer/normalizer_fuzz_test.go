package normalizer

import (
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func FuzzNormalizeFlightStateIsDeterministic(
	f *testing.F,
) {
	f.Add(
		" a0b1c2 ",
		" az123 ",
		" OpenSky ",
	)
	f.Add(
		"",
		"",
		"",
	)

	f.Fuzz(
		func(
			t *testing.T,
			icao24 string,
			callsign string,
			sourceName string,
		) {
			const maximumFuzzStringLength = 4096
			if len(icao24) > maximumFuzzStringLength ||
				len(callsign) > maximumFuzzStringLength ||
				len(sourceName) > maximumFuzzStringLength {
				t.Skip()
			}

			input := flightstate.FlightState{
				ICAO24:     icao24,
				Callsign:   callsign,
				SourceName: sourceName,
			}

			first := NormalizeFlightState(
				input,
			)
			second := NormalizeFlightState(
				first,
			)

			if first != second {
				t.Fatalf(
					"normalization is not idempotent: first=%+v second=%+v",
					first,
					second,
				)
			}
			if first.ICAO24 != strings.ToUpper(
				strings.TrimSpace(
					icao24,
				),
			) {
				t.Fatalf(
					"unexpected ICAO24 normalization: %q",
					first.ICAO24,
				)
			}
			if first.Callsign != strings.ToUpper(
				strings.TrimSpace(
					callsign,
				),
			) {
				t.Fatalf(
					"unexpected callsign normalization: %q",
					first.Callsign,
				)
			}
			if first.SourceName != strings.ToLower(
				strings.TrimSpace(
					sourceName,
				),
			) {
				t.Fatalf(
					"unexpected source normalization: %q",
					first.SourceName,
				)
			}
		},
	)
}
