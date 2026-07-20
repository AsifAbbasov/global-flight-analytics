package postgres

import (
	"errors"
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestAltitudeMetersToPostgresIntegerRoundsToNearestWholeMeter(
	t *testing.T,
) {
	t.Parallel()

	testCases := []struct {
		name  string
		value float64
		want  int32
	}{
		{name: "positive below half", value: 9753.49, want: 9753},
		{name: "positive half away from zero", value: 9753.5, want: 9754},
		{name: "negative below half", value: -12.49, want: -12},
		{name: "negative half away from zero", value: -12.5, want: -13},
		{name: "observed zero", value: 0, want: 0},
		{
			name:  "maximum boundary",
			value: float64(maximumPostgresIntegerAltitudeMeters),
			want:  maximumPostgresIntegerAltitudeMeters,
		},
		{
			name:  "minimum boundary",
			value: float64(minimumPostgresIntegerAltitudeMeters),
			want:  minimumPostgresIntegerAltitudeMeters,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := altitudeMetersToPostgresInteger(
				testCase.value,
			)
			if err != nil {
				t.Fatalf("convert altitude: %v", err)
			}
			if got != testCase.want {
				t.Fatalf(
					"altitudeMetersToPostgresInteger(%v) = %d, want %d",
					testCase.value,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestAltitudeMetersToPostgresIntegerRejectsNonFiniteValues(
	t *testing.T,
) {
	t.Parallel()

	for _, value := range []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	} {
		_, err := altitudeMetersToPostgresInteger(value)
		if !errors.Is(err, ErrAltitudeMetersNotFinite) {
			t.Fatalf(
				"value %v: expected ErrAltitudeMetersNotFinite, got %v",
				value,
				err,
			)
		}
	}
}

func TestAltitudeMetersToPostgresIntegerRejectsRoundedOverflow(
	t *testing.T,
) {
	t.Parallel()

	for _, value := range []float64{
		float64(maximumPostgresIntegerAltitudeMeters) + 0.5,
		float64(minimumPostgresIntegerAltitudeMeters) - 0.5,
	} {
		_, err := altitudeMetersToPostgresInteger(value)
		if !errors.Is(
			err,
			ErrAltitudeMetersOutsidePostgresIntegerRange,
		) {
			t.Fatalf(
				"value %v: expected range error, got %v",
				value,
				err,
			)
		}
	}
}

func TestAltitudeDatabaseValueUsesExplicitIntegerPolicy(
	t *testing.T,
) {
	t.Parallel()

	observed, status, err := altitudeDatabaseValue(
		1000.5,
		flightstate.AltitudeStatusObserved,
	)
	if err != nil {
		t.Fatalf("prepare observed altitude: %v", err)
	}
	if !observed.Valid || observed.Int32 != 1001 ||
		status != string(flightstate.AltitudeStatusObserved) {
		t.Fatalf(
			"unexpected observed value: %#v status=%q",
			observed,
			status,
		)
	}

	ground, status, err := altitudeDatabaseValue(
		1234.9,
		flightstate.AltitudeStatusGround,
	)
	if err != nil {
		t.Fatalf("prepare ground altitude: %v", err)
	}
	if !ground.Valid || ground.Int32 != 0 ||
		status != string(flightstate.AltitudeStatusGround) {
		t.Fatalf(
			"unexpected ground value: %#v status=%q",
			ground,
			status,
		)
	}

	unavailable, status, err := altitudeDatabaseValue(
		math.NaN(),
		flightstate.AltitudeStatusUnavailable,
	)
	if err != nil {
		t.Fatalf("prepare unavailable altitude: %v", err)
	}
	if unavailable.Valid ||
		status != string(flightstate.AltitudeStatusUnavailable) {
		t.Fatalf(
			"unexpected unavailable value: %#v status=%q",
			unavailable,
			status,
		)
	}
}
