package validator

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const scoreTolerance = 1e-12

func TestEvaluateFlightStateValidComplete(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusValid {
		t.Fatalf(
			"expected valid status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelComplete {
		t.Fatalf(
			"expected complete completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelHigh {
		t.Fatalf(
			"expected high confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		1.0,
	)

	if len(result.MissingFields) != 0 {
		t.Fatalf(
			"expected no missing fields, got %v",
			result.MissingFields,
		)
	}

	if len(result.Warnings) != 0 {
		t.Fatalf(
			"expected no warnings, got %v",
			result.Warnings,
		)
	}
}

func TestEvaluateFlightStateMissingCallsignUsesElevenOfTwelvePassRatio(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.Callsign = "   "

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected partial status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelPartial {
		t.Fatalf(
			"expected partial completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelMedium {
		t.Fatalf(
			"expected medium confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		11.0/12.0,
	)

	assertStringSliceEqual(
		t,
		result.MissingFields,
		[]string{
			"callsign",
		},
	)

	if len(result.Warnings) != 0 {
		t.Fatalf(
			"expected no warnings, got %v",
			result.Warnings,
		)
	}
}

func TestEvaluateFlightStateInvalidBarometricAltitudeUsesElevenOfTwelvePassRatio(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.BarometricAltitudeM = -1

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected partial status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelPartial {
		t.Fatalf(
			"expected partial completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelMedium {
		t.Fatalf(
			"expected medium confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		11.0/12.0,
	)

	assertWarningCodesEqual(
		t,
		result.Warnings,
		[]string{
			"invalid_barometric_altitude",
		},
	)
}

func TestEvaluateFlightStateInvalidVelocityUsesElevenOfTwelvePassRatio(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.VelocityMPS = -1

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected partial status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelPositionOnly {
		t.Fatalf(
			"expected position-only completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelLow {
		t.Fatalf(
			"expected low confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		11.0/12.0,
	)

	assertWarningCodesEqual(
		t,
		result.Warnings,
		[]string{
			"invalid_velocity",
		},
	)

	if !IsValidFlightState(
		item,
		now,
	) {
		t.Fatal(
			"expected position-only flight state to remain usable",
		)
	}
}

func TestEvaluateFlightStateInvalidVelocityAndHeadingUseTenOfTwelvePassRatio(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.VelocityMPS = -1
	item.HeadingDegrees = 720

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected partial status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelPositionOnly {
		t.Fatalf(
			"expected position-only completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelLow {
		t.Fatalf(
			"expected low confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		10.0/12.0,
	)

	assertWarningCodesEqual(
		t,
		result.Warnings,
		[]string{
			"invalid_velocity",
			"invalid_heading",
		},
	)
}

func TestEvaluateFlightStateInvalidICAO24ZeroesScore(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.ICAO24 = "INVALID"

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf(
			"expected invalid status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelInsufficient {
		t.Fatalf(
			"expected insufficient completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelNone {
		t.Fatalf(
			"expected no confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		0,
	)

	assertWarningCodesEqual(
		t,
		result.Warnings,
		[]string{
			"invalid_icao24",
		},
	)

	if IsValidFlightState(
		item,
		now,
	) {
		t.Fatal(
			"expected invalid flight state",
		)
	}
}

func TestEvaluateFlightStateFutureObservedAtZeroesScore(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.ObservedAt = now.Add(
		time.Minute,
	)

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf(
			"expected invalid status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelInsufficient {
		t.Fatalf(
			"expected insufficient completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelNone {
		t.Fatalf(
			"expected no confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		0,
	)

	assertWarningCodesEqual(
		t,
		result.Warnings,
		[]string{
			"future_observed_at",
		},
	)
}

func TestFilterValidFlightStatesKeepsPartialAndDropsInvalid(
	t *testing.T,
) {
	now := fixedValidatorTestTime()

	validItem := makeValidFlightState(now)

	partialItem := makeValidFlightState(now)
	partialItem.ICAO24 = "DEF456"
	partialItem.VelocityMPS = -1

	invalidItem := makeValidFlightState(now)
	invalidItem.ICAO24 = "BAD"

	result := FilterValidFlightStates(
		[]flightstate.FlightState{
			validItem,
			partialItem,
			invalidItem,
		},
		now,
	)

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 usable flight states, got %d",
			len(result),
		)
	}
}

func TestEvaluateFlightStateRejectsInvalidCoordinatesAndZeroesScore(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.Latitude = math.NaN()

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf(
			"expected invalid status, got %s",
			result.ValidationStatus,
		)
	}

	if result.Completeness != dataquality.CompletenessLevelInsufficient {
		t.Fatalf(
			"expected insufficient completeness, got %s",
			result.Completeness,
		)
	}

	if result.Confidence != dataquality.ConfidenceLevelNone {
		t.Fatalf(
			"expected no confidence, got %s",
			result.Confidence,
		)
	}

	assertScoreClose(
		t,
		result.Score,
		0,
	)

	assertWarningCodesEqual(
		t,
		result.Warnings,
		[]string{
			"invalid_latitude",
		},
	)
}

func TestAssessmentCounterPassRatioReturnsZeroWithoutAssessments(
	t *testing.T,
) {
	counter := assessmentCounter{}

	assertScoreClose(
		t,
		counter.PassRatio(),
		0,
	)
}

func TestAssessmentCounterPassRatioReturnsPassedFraction(
	t *testing.T,
) {
	counter := assessmentCounter{}

	counter.Assess(true)
	counter.Assess(false)
	counter.Assess(true)
	counter.Assess(true)

	assertScoreClose(
		t,
		counter.PassRatio(),
		3.0/4.0,
	)
}

func fixedValidatorTestTime() time.Time {
	return time.Date(
		2026,
		7,
		2,
		10,
		0,
		0,
		0,
		time.UTC,
	)
}

func makeValidFlightState(
	now time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                  "state-1",
		FlightID:            "flight-1",
		AircraftID:          "aircraft-1",
		ICAO24:              "ABC123",
		Callsign:            "AHY101",
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 10000,
		GeometricAltitudeM:  10050,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		VerticalRateMPS:     0,
		OnGround:            false,
		OriginCountry:       "Azerbaijan",
		ObservedAt: now.Add(
			-30 * time.Second,
		),
		SourceName: "test",
	}
}

func assertScoreClose(
	t *testing.T,
	actual float64,
	expected float64,
) {
	t.Helper()

	if math.Abs(actual-expected) > scoreTolerance {
		t.Fatalf(
			"unexpected score: got %.15f, want %.15f",
			actual,
			expected,
		)
	}
}

func assertStringSliceEqual(
	t *testing.T,
	actual []string,
	expected []string,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf(
			"unexpected string slice length: got %d, want %d; values=%v",
			len(actual),
			len(expected),
			actual,
		)
	}

	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf(
				"unexpected string at index %d: got %q, want %q",
				index,
				actual[index],
				expected[index],
			)
		}
	}
}

func assertWarningCodesEqual(
	t *testing.T,
	actual []dataquality.Warning,
	expected []string,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf(
			"unexpected warning count: got %d, want %d; warnings=%v",
			len(actual),
			len(expected),
			actual,
		)
	}

	for index := range expected {
		if actual[index].Code != expected[index] {
			t.Fatalf(
				"unexpected warning code at index %d: got %q, want %q",
				index,
				actual[index].Code,
				expected[index],
			)
		}
	}
}
