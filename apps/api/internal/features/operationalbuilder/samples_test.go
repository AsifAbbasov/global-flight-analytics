package operationalbuilder

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestAltitudeValueUsesExplicitStatuses(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		status      flightstate.AltitudeStatus
		wantValue   float64
		wantUsable  bool
		wantInvalid bool
	}{
		{
			name:       "observed",
			value:      1234,
			status:     flightstate.AltitudeStatusObserved,
			wantValue:  1234,
			wantUsable: true,
		},
		{
			name:       "ground",
			value:      999,
			status:     flightstate.AltitudeStatusGround,
			wantValue:  0,
			wantUsable: true,
		},
		{
			name:   "unavailable",
			value:  0,
			status: flightstate.AltitudeStatusUnavailable,
		},
		{
			name:   "unknown",
			value:  0,
			status: flightstate.AltitudeStatusUnknown,
		},
		{
			name:        "invalid status",
			value:       100,
			status:      flightstate.AltitudeStatusInvalid,
			wantInvalid: true,
		},
		{
			name:        "unsupported status",
			value:       100,
			status:      "future",
			wantInvalid: true,
		},
		{
			name:        "non-finite observed value",
			value:       math.NaN(),
			status:      flightstate.AltitudeStatusObserved,
			wantInvalid: true,
		},
		{
			name:       "implicit observed",
			value:      100,
			wantValue:  100,
			wantUsable: true,
		},
		{
			name:  "implicit unavailable zero",
			value: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, usable, invalid := altitudeValue(
				test.value,
				test.status,
			)

			if value != test.wantValue ||
				usable != test.wantUsable ||
				invalid != test.wantInvalid {
				t.Fatalf(
					"altitudeValue() = (%v, %v, %v), want (%v, %v, %v)",
					value,
					usable,
					invalid,
					test.wantValue,
					test.wantUsable,
					test.wantInvalid,
				)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	minimum, maximum, mean := summarize(
		[]float64{-10, 0, 20},
	)

	if minimum != -10 ||
		maximum != 20 ||
		mean != 10.0/3.0 {
		t.Fatalf(
			"summarize() = (%v, %v, %v)",
			minimum,
			maximum,
			mean,
		)
	}
}

func TestCumulativeHeadingChange(t *testing.T) {
	tests := []struct {
		name     string
		headings []float64
		want     float64
	}{
		{
			name:     "empty",
			headings: nil,
			want:     0,
		},
		{
			name:     "single",
			headings: []float64{90},
			want:     0,
		},
		{
			name:     "shortest antimeridian-like turn",
			headings: []float64{350, 10, 20},
			want:     30,
		},
		{
			name:     "half turn",
			headings: []float64{0, 180},
			want:     180,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := cumulativeHeadingChange(
				test.headings,
			); got != test.want {
				t.Fatalf(
					"cumulativeHeadingChange() = %v, want %v",
					got,
					test.want,
				)
			}
		})
	}
}

func TestNormalizeHeading(t *testing.T) {
	tests := []struct {
		value float64
		want  float64
	}{
		{value: -10, want: 350},
		{value: 0, want: 0},
		{value: 360, want: 0},
		{value: 370, want: 10},
		{value: 720, want: 0},
	}

	for _, test := range tests {
		if got := normalizeHeading(test.value); got != test.want {
			t.Fatalf(
				"normalizeHeading(%v) = %v, want %v",
				test.value,
				got,
				test.want,
			)
		}
	}
}
