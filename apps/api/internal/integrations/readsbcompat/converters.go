package readsbcompat

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const (
	knotToMetersPerSecond       = 0.5144444444444445
	feetPerMinuteToMetersPerSec = 0.00508
	feetToMetersRatio           = 0.3048
)

type AltitudeReading struct {
	Meters float64
	Status flightstate.AltitudeStatus
}

func KnotsToMetersPerSecond(value float64) float64 {
	return value * knotToMetersPerSecond
}

func FeetPerMinuteToMetersPerSecond(value float64) float64 {
	return value * feetPerMinuteToMetersPerSec
}

func FeetToMeters(value float64) float64 {
	return value * feetToMetersRatio
}

func BarometricAltitudeReading(
	value BarometricAltitude,
) AltitudeReading {
	switch value.Kind {
	case BarometricAltitudeKindObserved:
		if !isFiniteFloat64(value.Feet) {
			return invalidAltitudeReading()
		}
		return AltitudeReading{
			Meters: FeetToMeters(value.Feet),
			Status: flightstate.AltitudeStatusObserved,
		}
	case BarometricAltitudeKindGround:
		return AltitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusGround,
		}
	case BarometricAltitudeKindUnknown:
		return AltitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusUnknown,
		}
	case BarometricAltitudeKindUnavailable, "":
		return AltitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusUnavailable,
		}
	case BarometricAltitudeKindInvalid:
		return invalidAltitudeReading()
	default:
		return invalidAltitudeReading()
	}
}

func GeometricAltitudeReading(
	altitudeFeet *float64,
) AltitudeReading {
	if altitudeFeet == nil {
		return AltitudeReading{
			Meters: 0,
			Status: flightstate.AltitudeStatusUnavailable,
		}
	}
	if !isFiniteFloat64(*altitudeFeet) {
		return invalidAltitudeReading()
	}
	return AltitudeReading{
		Meters: FeetToMeters(*altitudeFeet),
		Status: flightstate.AltitudeStatusObserved,
	}
}

func invalidAltitudeReading() AltitudeReading {
	return AltitudeReading{
		Meters: 0,
		Status: flightstate.AltitudeStatusInvalid,
	}
}

func isFiniteFloat64(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
