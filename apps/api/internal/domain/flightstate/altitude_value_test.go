package flightstate

import (
	"errors"
	"math"
	"testing"
)

func TestAltitudeValueObjectPreservesExplicitZeroObservation(t *testing.T) {
	altitude, err := NewAltitude(0, AltitudeStatusObserved)
	if err != nil {
		t.Fatalf("NewAltitude() error = %v", err)
	}
	if altitude.Status() != AltitudeStatusObserved ||
		altitude.Meters() != 0 ||
		!altitude.Available() {
		t.Fatalf("unexpected altitude = status:%q meters:%v available:%t",
			altitude.Status(), altitude.Meters(), altitude.Available())
	}
}

func TestAltitudeValueObjectNormalizesUnavailablePlaceholder(t *testing.T) {
	altitude, err := NewAltitude(math.NaN(), AltitudeStatusUnavailable)
	if err != nil {
		t.Fatalf("NewAltitude() error = %v", err)
	}
	if altitude.Status() != AltitudeStatusUnavailable ||
		altitude.Meters() != 0 ||
		altitude.Available() {
		t.Fatalf("unexpected altitude = status:%q meters:%v available:%t",
			altitude.Status(), altitude.Meters(), altitude.Available())
	}
}

func TestAltitudeValueObjectRejectsNonFiniteObservedValue(t *testing.T) {
	_, err := NewAltitude(math.Inf(1), AltitudeStatusObserved)
	if !errors.Is(err, ErrAltitudeValueInvalid) {
		t.Fatalf("NewAltitude() error = %v", err)
	}
}

func TestAltitudeValueObjectRejectsUnavailableFiniteValue(t *testing.T) {
	_, err := NewAltitude(125, AltitudeStatusUnavailable)
	if !errors.Is(err, ErrAltitudeStateConflict) {
		t.Fatalf("NewAltitude() error = %v", err)
	}
}
