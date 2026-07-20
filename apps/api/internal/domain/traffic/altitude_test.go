package traffic

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestResolveCurrentAltitudePreservesObservedGeometricZero(
	t *testing.T,
) {
	geometric := 0.0
	barometric := 1200.0

	value, status, source := ResolveCurrentAltitude(
		false,
		&geometric,
		flightstate.AltitudeStatusObserved,
		&barometric,
		flightstate.AltitudeStatusObserved,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		true,
		0,
		flightstate.AltitudeStatusObserved,
		AltitudeSourceGeometric,
	)
}

func TestResolveCurrentAltitudeFallsBackToObservedBarometric(
	t *testing.T,
) {
	barometric := 2400.0

	value, status, source := ResolveCurrentAltitude(
		false,
		nil,
		flightstate.AltitudeStatusUnavailable,
		&barometric,
		flightstate.AltitudeStatusObserved,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		true,
		2400,
		flightstate.AltitudeStatusObserved,
		AltitudeSourceBarometric,
	)
}

func TestResolveCurrentAltitudePublishesGroundSeparately(
	t *testing.T,
) {
	value, status, source := ResolveCurrentAltitude(
		true,
		nil,
		flightstate.AltitudeStatusUnavailable,
		nil,
		flightstate.AltitudeStatusUnavailable,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		true,
		0,
		flightstate.AltitudeStatusGround,
		AltitudeSourceGround,
	)
}

func TestResolveCurrentAltitudePublishesUnknownWithoutFakeZero(
	t *testing.T,
) {
	value, status, source := ResolveCurrentAltitude(
		false,
		nil,
		flightstate.AltitudeStatusUnknown,
		nil,
		flightstate.AltitudeStatusUnavailable,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		false,
		0,
		flightstate.AltitudeStatusUnknown,
		AltitudeSourceNone,
	)
}

func TestResolveCurrentAltitudePublishesUnavailableWithoutFakeZero(
	t *testing.T,
) {
	value, status, source := ResolveCurrentAltitude(
		false,
		nil,
		flightstate.AltitudeStatusUnavailable,
		nil,
		flightstate.AltitudeStatusUnavailable,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		false,
		0,
		flightstate.AltitudeStatusUnavailable,
		AltitudeSourceNone,
	)
}

func TestResolveCurrentAltitudeRejectsInAirGroundStatus(
	t *testing.T,
) {
	groundValue := 0.0

	value, status, source := ResolveCurrentAltitude(
		false,
		&groundValue,
		flightstate.AltitudeStatusGround,
		nil,
		flightstate.AltitudeStatusUnavailable,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		false,
		0,
		flightstate.AltitudeStatusInvalid,
		AltitudeSourceNone,
	)
}

func TestResolveCurrentAltitudeFallsBackAfterInvalidGeometricValue(
	t *testing.T,
) {
	geometric := math.NaN()
	barometric := -20.0

	value, status, source := ResolveCurrentAltitude(
		false,
		&geometric,
		flightstate.AltitudeStatusObserved,
		&barometric,
		flightstate.AltitudeStatusObserved,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		true,
		-20,
		flightstate.AltitudeStatusObserved,
		AltitudeSourceBarometric,
	)
}

func TestResolveCurrentAltitudeRejectsValueForUnavailableStatus(
	t *testing.T,
) {
	geometric := 800.0

	value, status, source := ResolveCurrentAltitude(
		false,
		&geometric,
		flightstate.AltitudeStatusUnavailable,
		nil,
		flightstate.AltitudeStatusUnavailable,
	)

	assertCurrentAltitude(
		t,
		value,
		status,
		source,
		false,
		0,
		flightstate.AltitudeStatusInvalid,
		AltitudeSourceNone,
	)
}

func assertCurrentAltitude(
	t *testing.T,
	actualValue *float64,
	actualStatus flightstate.AltitudeStatus,
	actualSource AltitudeSource,
	expectedValuePresent bool,
	expectedValue float64,
	expectedStatus flightstate.AltitudeStatus,
	expectedSource AltitudeSource,
) {
	t.Helper()

	if (actualValue != nil) != expectedValuePresent {
		t.Fatalf(
			"altitude value presence = %v, want %v",
			actualValue != nil,
			expectedValuePresent,
		)
	}

	if expectedValuePresent && *actualValue != expectedValue {
		t.Fatalf(
			"altitude value = %v, want %v",
			*actualValue,
			expectedValue,
		)
	}

	if actualStatus != expectedStatus {
		t.Fatalf(
			"altitude status = %q, want %q",
			actualStatus,
			expectedStatus,
		)
	}

	if actualSource != expectedSource {
		t.Fatalf(
			"altitude source = %q, want %q",
			actualSource,
			expectedSource,
		)
	}
}
