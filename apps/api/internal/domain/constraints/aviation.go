package constraints

import "math"

const (
	ICAO24HexadecimalLength = 6

	MinimumLatitudeDegrees  = -90.0
	MaximumLatitudeDegrees  = 90.0
	MinimumLongitudeDegrees = -180.0
	MaximumLongitudeDegrees = 180.0

	MinimumHeadingDegrees          = 0.0
	MaximumHeadingDegreesExclusive = 360.0
	MaximumHeadingDegreesInclusive = 360

	MinimumPercentValue = 0
	MaximumPercentValue = 100

	MinimumNonNegativeFloat = 0.0

	EarthRadiusKilometers = 6371.0088
)

func IsFiniteFloat64(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func IsLatitude(value float64) bool {
	return IsFiniteFloat64(value) &&
		value >= MinimumLatitudeDegrees &&
		value <= MaximumLatitudeDegrees
}

func IsLongitude(value float64) bool {
	return IsFiniteFloat64(value) &&
		value >= MinimumLongitudeDegrees &&
		value <= MaximumLongitudeDegrees
}

func IsPercentInt(value int) bool {
	return value >= MinimumPercentValue && value <= MaximumPercentValue
}

func IsNonNegativeFloat64(value float64) bool {
	return IsFiniteFloat64(value) && value >= MinimumNonNegativeFloat
}

func IsPositiveFloat64(value float64) bool {
	return IsFiniteFloat64(value) && value > MinimumNonNegativeFloat
}

func IsHeadingDegreesExclusive(value float64) bool {
	return IsFiniteFloat64(value) &&
		value >= MinimumHeadingDegrees &&
		value < MaximumHeadingDegreesExclusive
}

func IsHeadingDegreesInclusive(value int) bool {
	return value >= int(MinimumHeadingDegrees) &&
		value <= MaximumHeadingDegreesInclusive
}
