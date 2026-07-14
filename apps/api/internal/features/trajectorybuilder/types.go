package trajectorybuilder

import "time"

type coordinate struct {
	latitude  float64
	longitude float64
}

func (value coordinate) equal(other coordinate) bool {
	return value.latitude == other.latitude &&
		value.longitude == other.longitude
}

type timeInterval struct {
	start time.Time
	end   time.Time
}
