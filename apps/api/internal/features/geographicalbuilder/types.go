package geographicalbuilder

import (
	"fmt"
	"math"
)

type coordinate struct {
	latitude  float64
	longitude float64
}

func (value coordinate) equal(other coordinate) bool {
	return value.latitude == other.latitude &&
		value.longitude == other.longitude
}

func (value coordinate) cellKey(precision int) string {
	scale := math.Pow10(precision)

	return fmt.Sprintf(
		"%d:%d",
		int64(math.Round(value.latitude*scale)),
		int64(math.Round(value.longitude*scale)),
	)
}
