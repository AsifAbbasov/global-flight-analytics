package postgres

import (
	"math"
	"testing"
)

const feetToMetersTestTolerance = 1e-9

func TestFeetToMeters(
	t *testing.T,
) {
	tests := []struct {
		name     string
		feet     float64
		expected float64
	}{
		{
			name:     "zero feet",
			feet:     0,
			expected: 0,
		},
		{
			name:     "one foot",
			feet:     1,
			expected: 0.3048,
		},
		{
			name:     "one thousand feet",
			feet:     1000,
			expected: 304.8,
		},
		{
			name:     "one thousand six hundred twenty four feet",
			feet:     1624,
			expected: 494.9952,
		},
		{
			name:     "negative one foot",
			feet:     -1,
			expected: -0.3048,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				actual := feetToMeters(
					test.feet,
				)

				difference := math.Abs(
					actual - test.expected,
				)
				if difference > feetToMetersTestTolerance {
					t.Fatalf(
						"unexpected conversion: feet=%v actual=%.12f expected=%.12f difference=%.12f",
						test.feet,
						actual,
						test.expected,
						difference,
					)
				}
			},
		)
	}
}
