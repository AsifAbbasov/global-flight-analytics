package airplaneslive

const (
	knotToMetersPerSecond       = 0.5144444444444445
	feetPerMinuteToMetersPerSec = 0.00508
	feetToMetersRatio           = 0.3048
)

func knotsToMetersPerSecond(value float64) float64 {
	return value * knotToMetersPerSecond
}

func feetPerMinuteToMetersPerSecond(value float64) float64 {
	return value * feetPerMinuteToMetersPerSec
}

func feetToMeters(value float64) float64 {
	return value * feetToMetersRatio
}

func barometricAltitudeMeters(value any) float64 {
	switch altitude := value.(type) {
	case float64:
		return feetToMeters(altitude)
	case string:
		if altitude == "ground" {
			return 0
		}
		return 0
	default:
		return 0
	}
}
