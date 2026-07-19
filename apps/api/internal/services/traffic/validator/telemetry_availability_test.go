package validator

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestEvaluateFlightStateReportsUnavailableKinematicsAsMissing(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		20,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	item := validTelemetryAvailabilityFixture(
		now,
	)
	item.TelemetryAvailabilityKnown = true

	quality := EvaluateFlightState(
		item,
		now,
	)

	for _, field := range []string{
		"velocity_mps",
		"heading_degrees",
		"vertical_rate_mps",
		"on_ground",
	} {
		if !containsMissingField(
			quality.MissingFields,
			field,
		) {
			t.Fatalf(
				"missing telemetry field %q was not reported: %#v",
				field,
				quality.MissingFields,
			)
		}
	}
}

func TestEvaluateFlightStatePreservesAvailableZeroKinematics(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		20,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	item := validTelemetryAvailabilityFixture(
		now,
	)
	item.TelemetryAvailabilityKnown = true
	item.VelocityAvailable = true
	item.HeadingAvailable = true
	item.VerticalRateAvailable = true
	item.OnGroundAvailable = true

	quality := EvaluateFlightState(
		item,
		now,
	)

	for _, field := range []string{
		"velocity_mps",
		"heading_degrees",
		"vertical_rate_mps",
		"on_ground",
	} {
		if containsMissingField(
			quality.MissingFields,
			field,
		) {
			t.Fatalf(
				"available zero telemetry field %q was reported missing: %#v",
				field,
				quality.MissingFields,
			)
		}
	}
}

func TestEvaluateFlightStateOnGroundAvailabilityPreservesScoreContract(
	t *testing.T,
) {
	now := fixedValidatorTestTime()
	item := makeValidFlightState(now)
	item.TelemetryAvailabilityKnown = true
	item.VelocityAvailable = true
	item.HeadingAvailable = true
	item.VerticalRateAvailable = true
	item.OnGroundAvailable = false

	quality := EvaluateFlightState(
		item,
		now,
	)

	assertScoreClose(
		t,
		quality.Score,
		1.0,
	)
	if quality.Completeness !=
		dataquality.CompletenessLevelPositionOnly {
		t.Fatalf(
			"completeness = %q, want %q",
			quality.Completeness,
			dataquality.CompletenessLevelPositionOnly,
		)
	}
	if !containsMissingField(
		quality.MissingFields,
		"on_ground",
	) {
		t.Fatalf(
			"on_ground was not reported missing: %#v",
			quality.MissingFields,
		)
	}
}

func validTelemetryAvailabilityFixture(
	now time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ICAO24:                   "ABC123",
		Latitude:                 40.4093,
		Longitude:                49.8671,
		BarometricAltitudeStatus: flightstate.AltitudeStatusUnavailable,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusUnavailable,
		ObservedAt:               now.Add(-time.Second),
		SourceName:               "opensky",
	}
}

func containsMissingField(
	fields []string,
	target string,
) bool {
	for _, field := range fields {
		if field == target {
			return true
		}
	}
	return false
}
