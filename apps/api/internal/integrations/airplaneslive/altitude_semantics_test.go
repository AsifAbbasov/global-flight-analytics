package airplaneslive

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestBarometricAltitudeJSONSemantics(
	t *testing.T,
) {
	testCases := []struct {
		name      string
		rawJSON   string
		wantKind  BarometricAltitudeKind
		wantFeet  float64
		checkFeet bool
	}{
		{
			name:      "observed positive number",
			rawJSON:   `1000`,
			wantKind:  BarometricAltitudeKindObserved,
			wantFeet:  1000,
			checkFeet: true,
		},
		{
			name:      "observed numeric zero",
			rawJSON:   `0`,
			wantKind:  BarometricAltitudeKindObserved,
			wantFeet:  0,
			checkFeet: true,
		},
		{
			name:     "ground marker",
			rawJSON:  `"ground"`,
			wantKind: BarometricAltitudeKindGround,
		},
		{
			name:     "ground marker normalized",
			rawJSON:  `"  GROUND  "`,
			wantKind: BarometricAltitudeKindGround,
		},
		{
			name:     "explicit unknown",
			rawJSON:  `"unknown"`,
			wantKind: BarometricAltitudeKindUnknown,
		},
		{
			name:     "blank string",
			rawJSON:  `"   "`,
			wantKind: BarometricAltitudeKindUnknown,
		},
		{
			name:     "null value",
			rawJSON:  `null`,
			wantKind: BarometricAltitudeKindUnavailable,
		},
		{
			name:     "unsupported string",
			rawJSON:  `"not-a-real-altitude"`,
			wantKind: BarometricAltitudeKindInvalid,
		},
		{
			name:     "unsupported boolean",
			rawJSON:  `true`,
			wantKind: BarometricAltitudeKindInvalid,
		},
		{
			name:     "unsupported object",
			rawJSON:  `{}`,
			wantKind: BarometricAltitudeKindInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var value BarometricAltitude

				if err := json.Unmarshal(
					[]byte(testCase.rawJSON),
					&value,
				); err != nil {
					t.Fatalf(
						"decode barometric altitude: %v",
						err,
					)
				}

				if value.Kind != testCase.wantKind {
					t.Fatalf(
						"expected kind %q, got %q",
						testCase.wantKind,
						value.Kind,
					)
				}

				if testCase.checkFeet &&
					value.Feet != testCase.wantFeet {
					t.Fatalf(
						"expected feet %v, got %v",
						testCase.wantFeet,
						value.Feet,
					)
				}
			},
		)
	}
}

func TestBarometricAltitudeReadingDistinguishesKnownZeroFromGround(
	t *testing.T,
) {
	knownZero := barometricAltitudeReading(
		BarometricAltitude{
			Feet: 0,
			Kind: BarometricAltitudeKindObserved,
		},
	)
	ground := barometricAltitudeReading(
		BarometricAltitude{
			Kind: BarometricAltitudeKindGround,
		},
	)

	if knownZero.Status != flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"expected numeric zero status %q, got %q",
			flightstate.AltitudeStatusObserved,
			knownZero.Status,
		)
	}

	if knownZero.Meters != 0 {
		t.Fatalf(
			"expected numeric zero altitude 0 meters, got %v",
			knownZero.Meters,
		)
	}

	if ground.Status != flightstate.AltitudeStatusGround {
		t.Fatalf(
			"expected ground status %q, got %q",
			flightstate.AltitudeStatusGround,
			ground.Status,
		)
	}

	if ground.Meters != 0 {
		t.Fatalf(
			"expected ground altitude placeholder 0 meters, got %v",
			ground.Meters,
		)
	}

	if knownZero.Status == ground.Status {
		t.Fatal(
			"expected numeric zero and ground to have distinct statuses",
		)
	}
}

func TestBarometricAltitudeReadingPreservesObservedValue(
	t *testing.T,
) {
	reading := barometricAltitudeReading(
		BarometricAltitude{
			Feet: 1000,
			Kind: BarometricAltitudeKindObserved,
		},
	)

	if reading.Status != flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"expected observed status, got %q",
			reading.Status,
		)
	}

	if reading.Meters != 304.8 {
		t.Fatalf(
			"expected 304.8 meters, got %v",
			reading.Meters,
		)
	}
}

func TestBarometricAltitudeReadingDistinguishesUnknownUnavailableAndInvalid(
	t *testing.T,
) {
	testCases := []struct {
		name       string
		input      BarometricAltitude
		wantStatus flightstate.AltitudeStatus
	}{
		{
			name: "explicit unknown",
			input: BarometricAltitude{
				Kind: BarometricAltitudeKindUnknown,
			},
			wantStatus: flightstate.AltitudeStatusUnknown,
		},
		{
			name:       "missing provider value",
			input:      BarometricAltitude{},
			wantStatus: flightstate.AltitudeStatusUnavailable,
		},
		{
			name: "explicit unavailable",
			input: BarometricAltitude{
				Kind: BarometricAltitudeKindUnavailable,
			},
			wantStatus: flightstate.AltitudeStatusUnavailable,
		},
		{
			name: "invalid provider value",
			input: BarometricAltitude{
				Kind: BarometricAltitudeKindInvalid,
			},
			wantStatus: flightstate.AltitudeStatusInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				reading := barometricAltitudeReading(
					testCase.input,
				)

				if reading.Status != testCase.wantStatus {
					t.Fatalf(
						"expected status %q, got %q",
						testCase.wantStatus,
						reading.Status,
					)
				}
			},
		)
	}
}

func TestBarometricAltitudeReadingRejectsNonFiniteObservedNumbers(
	t *testing.T,
) {
	testCases := []struct {
		name string
		feet float64
	}{
		{
			name: "NaN",
			feet: math.NaN(),
		},
		{
			name: "positive infinity",
			feet: math.Inf(1),
		},
		{
			name: "negative infinity",
			feet: math.Inf(-1),
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				reading := barometricAltitudeReading(
					BarometricAltitude{
						Feet: testCase.feet,
						Kind: BarometricAltitudeKindObserved,
					},
				)

				if reading.Status != flightstate.AltitudeStatusInvalid {
					t.Fatalf(
						"expected invalid status, got %q",
						reading.Status,
					)
				}
			},
		)
	}
}

func TestGeometricAltitudeReadingDistinguishesObservedZeroFromUnavailable(
	t *testing.T,
) {
	zeroFeet := float64(0)

	observedZero := geometricAltitudeReading(
		&zeroFeet,
	)
	unavailable := geometricAltitudeReading(
		nil,
	)

	if observedZero.Status != flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"expected observed zero status, got %q",
			observedZero.Status,
		)
	}

	if unavailable.Status != flightstate.AltitudeStatusUnavailable {
		t.Fatalf(
			"expected unavailable status, got %q",
			unavailable.Status,
		)
	}

	if observedZero.Status == unavailable.Status {
		t.Fatal(
			"expected observed zero and unavailable altitude to have distinct statuses",
		)
	}
}

func TestGeometricAltitudeReadingRejectsNonFiniteNumbers(
	t *testing.T,
) {
	values := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for index := range values {
		reading := geometricAltitudeReading(
			&values[index],
		)

		if reading.Status != flightstate.AltitudeStatusInvalid {
			t.Fatalf(
				"expected invalid status for value at index %d, got %q",
				index,
				reading.Status,
			)
		}
	}
}

func TestMapAircraftPropagatesAltitudeStatuses(
	t *testing.T,
) {
	geometricAltitudeFeet := float64(0)

	state := mapAircraft(
		AircraftItem{
			Hex: "abc123",
			AltBaro: BarometricAltitude{
				Kind: BarometricAltitudeKindGround,
			},
			AltGeom: &geometricAltitudeFeet,
		},
		0,
	)

	if state.BarometricAltitudeStatus != flightstate.AltitudeStatusGround {
		t.Fatalf(
			"expected ground barometric status, got %q",
			state.BarometricAltitudeStatus,
		)
	}

	if state.GeometricAltitudeStatus != flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"expected observed geometric status, got %q",
			state.GeometricAltitudeStatus,
		)
	}

	if !state.OnGround {
		t.Fatal(
			"expected ground barometric status to set OnGround",
		)
	}
}
