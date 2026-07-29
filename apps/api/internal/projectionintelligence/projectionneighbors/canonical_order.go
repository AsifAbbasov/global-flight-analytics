package projectionneighbors

import (
	"math"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func canonicalPointLess(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
) bool {
	leftTime := left.ObservedAt.UTC()
	rightTime := right.ObservedAt.UTC()
	if !leftTime.Equal(rightTime) {
		return leftTime.Before(rightTime)
	}

	leftID := strings.TrimSpace(left.ID)
	rightID := strings.TrimSpace(right.ID)
	if leftID != rightID {
		return leftID < rightID
	}

	leftValues := [...]float64{
		left.Latitude,
		left.Longitude,
		left.BarometricAltitudeM,
		left.GeometricAltitudeM,
		left.VelocityMPS,
		left.HeadingDegrees,
		left.VerticalRateMPS,
	}
	rightValues := [...]float64{
		right.Latitude,
		right.Longitude,
		right.BarometricAltitudeM,
		right.GeometricAltitudeM,
		right.VelocityMPS,
		right.HeadingDegrees,
		right.VerticalRateMPS,
	}
	for index := range leftValues {
		leftBits := math.Float64bits(leftValues[index])
		rightBits := math.Float64bits(rightValues[index])
		if leftBits != rightBits {
			return leftBits < rightBits
		}
	}

	return false
}
