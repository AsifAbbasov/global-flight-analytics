package airplaneslive

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/readsbcompat"

type altitudeReading = readsbcompat.AltitudeReading

func knotsToMetersPerSecond(value float64) float64 {
	return readsbcompat.KnotsToMetersPerSecond(value)
}

func feetPerMinuteToMetersPerSecond(value float64) float64 {
	return readsbcompat.FeetPerMinuteToMetersPerSecond(value)
}

func feetToMeters(value float64) float64 {
	return readsbcompat.FeetToMeters(value)
}

func barometricAltitudeReading(
	value BarometricAltitude,
) altitudeReading {
	return readsbcompat.BarometricAltitudeReading(value)
}

func geometricAltitudeReading(
	altitudeFeet *float64,
) altitudeReading {
	return readsbcompat.GeometricAltitudeReading(altitudeFeet)
}
