package airplaneslive

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const (
	knotToMetersPerSecond       = 0.5144444444444445
	feetPerMinuteToMetersPerSec = 0.00508
	feetToMetersRatio           = 0.3048
)

type altitudeReading struct {
	Meters float64
	Status flightstate.AltitudeStatus
}

func knotsToMetersPerSecond(value float64) float64 {
	return value * knotToMetersPerSecond
}

func feetPerMinuteToMetersPerSecond(value float64) float64 {
	return value * feetPerMinuteToMetersPerSec
}

func feetToMeters(value float64) float64 {
	return value * feetToMetersRatio
}

func barometricAltitudeReading(
	value BarometricAltitude,
) altitudeReading {
	switch value.Kind {
	case BarometricAltitudeKindObserved:
		if !isFiniteFloat64(value.Feet) {
			return invalidAltitudeReading()
		}

		return altitudeReading{
			Meters: feetToMeters(value.Feet),
			Status: flightstate.AltitudeStatusObserved,
		}

	case BarometricAltitudeKindGround:
		return altitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusGround,
		}

	case BarometricAltitudeKindUnknown:
		return altitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusUnknown,
		}

	case BarometricAltitudeKindUnavailable, "":
		return altitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusUnavailable,
		}

	case BarometricAltitudeKindInvalid:
		return invalidAltitudeReading()

	default:
		return invalidAltitudeReading()
	}
}

func geometricAltitudeReading(
	altitudeFeet *float64,
) altitudeReading {
	if altitudeFeet == nil {
		return altitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusUnavailable,
		}
	}

	if !isFiniteFloat64(*altitudeFeet) {
		return invalidAltitudeReading()
	}

	return altitudeReading{
		Meters: feetToMeters(*altitudeFeet),
		Status: flightstate.AltitudeStatusObserved,
	}
}

func invalidAltitudeReading() altitudeReading {
	return altitudeReading{
		Meters: 0,
		Status: flightstate.AltitudeStatusInvalid,
	}
}

func isFiniteFloat64(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
