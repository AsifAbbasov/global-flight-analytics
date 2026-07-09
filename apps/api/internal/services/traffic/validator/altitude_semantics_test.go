package validator

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestEvaluateFlightStateAcceptsObservedZeroAltitude(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusValid {
		t.Fatalf(
			"expected observed zero altitude to remain valid, got %s",
			result.ValidationStatus,
		)
	}
}

func TestEvaluateFlightStateAcceptsGroundAltitudeSemantics(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	item.OnGround = true

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusValid {
		t.Fatalf(
			"expected ground altitude semantics to remain valid, got %s",
			result.ValidationStatus,
		)
	}
}

func TestEvaluateFlightStateMarksUnknownAltitudeAsMissingNotInvalid(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusUnknown

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected unknown altitude to produce partial status, got %s",
			result.ValidationStatus,
		)
	}

	if !altitudeSemanticContainsString(
		result.MissingFields,
		"barometric_altitude_m",
	) {
		t.Fatalf(
			"expected barometric altitude to be missing, got %v",
			result.MissingFields,
		)
	}

	if altitudeSemanticHasWarningCode(
		result.Warnings,
		"invalid_barometric_altitude",
	) {
		t.Fatalf(
			"expected unknown altitude not to be classified as invalid, got %v",
			result.Warnings,
		)
	}
}

func TestEvaluateFlightStateMarksUnavailableAltitudeAsMissingNotInvalid(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.GeometricAltitudeM = 0
	item.GeometricAltitudeStatus = flightstate.AltitudeStatusUnavailable

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected unavailable altitude to produce partial status, got %s",
			result.ValidationStatus,
		)
	}

	if !altitudeSemanticContainsString(
		result.MissingFields,
		"geometric_altitude_m",
	) {
		t.Fatalf(
			"expected geometric altitude to be missing, got %v",
			result.MissingFields,
		)
	}

	if altitudeSemanticHasWarningCode(
		result.Warnings,
		"invalid_geometric_altitude",
	) {
		t.Fatalf(
			"expected unavailable altitude not to be classified as invalid, got %v",
			result.Warnings,
		)
	}
}

func TestEvaluateFlightStateRejectsInvalidAltitudeStatus(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.BarometricAltitudeStatus = flightstate.AltitudeStatusInvalid

	result := EvaluateFlightState(
		item,
		now,
	)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf(
			"expected invalid altitude semantics to produce partial status, got %s",
			result.ValidationStatus,
		)
	}

	if !altitudeSemanticHasWarningCode(
		result.Warnings,
		"invalid_barometric_altitude",
	) {
		t.Fatalf(
			"expected invalid barometric altitude warning, got %v",
			result.Warnings,
		)
	}
}

func TestEvaluateFlightStateRejectsNonFiniteObservedAltitude(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.BarometricAltitudeM = math.NaN()
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved

	result := EvaluateFlightState(
		item,
		now,
	)

	if !altitudeSemanticHasWarningCode(
		result.Warnings,
		"invalid_barometric_altitude",
	) {
		t.Fatalf(
			"expected non-finite observed altitude warning, got %v",
			result.Warnings,
		)
	}
}

func TestEvaluateFlightStateRejectsGroundStatusWithoutOnGround(
	t *testing.T,
) {
	now := altitudeSemanticValidatorTestTime()
	item := altitudeSemanticTestState(
		now,
	)

	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	item.OnGround = false

	result := EvaluateFlightState(
		item,
		now,
	)

	if !altitudeSemanticHasWarningCode(
		result.Warnings,
		"invalid_barometric_altitude",
	) {
		t.Fatalf(
			"expected inconsistent ground semantics warning, got %v",
			result.Warnings,
		)
	}
}

func altitudeSemanticValidatorTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		9,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func altitudeSemanticTestState(
	now time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ICAO24:                   "ABC123",
		Callsign:                 "AHY101",
		Latitude:                 40.4093,
		Longitude:                49.8671,
		BarometricAltitudeM:      1000,
		BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
		GeometricAltitudeM:       1050,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
		VelocityMPS:              220,
		HeadingDegrees:           90,
		VerticalRateMPS:          0,
		OnGround:                 false,
		OriginCountry:            "Azerbaijan",
		ObservedAt:               now.Add(-time.Second),
		SourceName:               "airplanes.live",
	}
}

func altitudeSemanticContainsString(
	items []string,
	expected string,
) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}

	return false
}

func altitudeSemanticHasWarningCode(
	warnings []dataquality.Warning,
	expected string,
) bool {
	for _, warning := range warnings {
		if warning.Code == expected {
			return true
		}
	}

	return false
}
