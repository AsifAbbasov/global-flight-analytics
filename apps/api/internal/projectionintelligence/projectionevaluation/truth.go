package projectionevaluation

import (
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type truthPoint struct {
	source    ActualPointSource
	timeValue time.Time

	latitude  float64
	longitude float64
	altitudeM *float64
}

func normalizeTruthPoints(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
	evaluatedAt time.Time,
) (
	[]trajectory.TrackPoint4D,
	int,
) {
	type indexedPoint struct {
		point trajectory.TrackPoint4D
		index int
	}

	indexed := make(
		[]indexedPoint,
		0,
		len(item.Points),
	)
	excludedAfterEvaluation := 0
	for index, point := range item.Points {
		observedAt :=
			point.ObservedAt.UTC()
		if point.ObservedAt.IsZero() ||
			observedAt.Before(
				asOfTime.UTC(),
			) ||
			!validLatitude(point.Latitude) ||
			!validLongitude(point.Longitude) {
			continue
		}
		if observedAt.After(
			evaluatedAt.UTC(),
		) {
			excludedAfterEvaluation++
			continue
		}

		point.ObservedAt = observedAt
		indexed = append(
			indexed,
			indexedPoint{
				point: point,
				index: index,
			},
		)
	}

	sort.SliceStable(
		indexed,
		func(left int, right int) bool {
			leftTime :=
				indexed[left].
					point.ObservedAt
			rightTime :=
				indexed[right].
					point.ObservedAt
			if leftTime.Equal(rightTime) {
				return indexed[left].index <
					indexed[right].index
			}

			return leftTime.Before(rightTime)
		},
	)

	result := make(
		[]trajectory.TrackPoint4D,
		0,
		len(indexed),
	)
	for _, indexedPoint := range indexed {
		if len(result) > 0 &&
			indexedPoint.point.
				ObservedAt.Equal(
				result[len(result)-1].
					ObservedAt,
			) {
			result[len(result)-1] =
				indexedPoint.point
			continue
		}
		result = append(
			result,
			indexedPoint.point,
		)
	}

	return result,
		excludedAfterEvaluation
}

func truthAt(
	points []trajectory.TrackPoint4D,
	targetTime time.Time,
	maximumGap time.Duration,
) (truthPoint, bool) {
	if len(points) == 0 ||
		targetTime.IsZero() {
		return truthPoint{}, false
	}

	targetTime = targetTime.UTC()
	if targetTime.Before(
		points[0].ObservedAt,
	) ||
		targetTime.After(
			points[len(points)-1].
				ObservedAt,
		) {
		return truthPoint{}, false
	}

	for index, point := range points {
		if point.ObservedAt.Equal(
			targetTime,
		) {
			return truthFromObserved(point),
				true
		}
		if point.ObservedAt.After(
			targetTime,
		) {
			if index == 0 {
				return truthPoint{}, false
			}

			left := points[index-1]
			right := point
			if right.ObservedAt.Sub(
				left.ObservedAt,
			) > maximumGap {
				return truthPoint{}, false
			}

			return interpolateTruth(
				left,
				right,
				targetTime,
			)
		}
	}

	return truthPoint{}, false
}

func truthFromObserved(
	point trajectory.TrackPoint4D,
) truthPoint {
	result := truthPoint{
		source:    ActualPointSourceObserved,
		timeValue: point.ObservedAt.UTC(),
		latitude:  point.Latitude,
		longitude: point.Longitude,
	}

	if altitudeM, available :=
		usableAltitude(point); available {
		result.altitudeM =
			float64Pointer(altitudeM)
	}

	return result
}

func interpolateTruth(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
	targetTime time.Time,
) (truthPoint, bool) {
	duration :=
		right.ObservedAt.Sub(
			left.ObservedAt,
		)
	if duration <= 0 {
		return truthPoint{}, false
	}

	fraction := float64(
		targetTime.Sub(
			left.ObservedAt,
		),
	) / float64(duration)
	if !unitInterval(fraction) {
		return truthPoint{}, false
	}

	distanceM := greatCircleDistanceM(
		left.Latitude,
		left.Longitude,
		right.Latitude,
		right.Longitude,
	)
	bearing := initialBearingDegrees(
		left.Latitude,
		left.Longitude,
		right.Latitude,
		right.Longitude,
	)
	latitude, longitude, valid :=
		destinationPoint(
			left.Latitude,
			left.Longitude,
			bearing,
			distanceM*fraction,
		)
	if !valid {
		return truthPoint{}, false
	}

	result := truthPoint{
		source:    ActualPointSourceInterpolated,
		timeValue: targetTime.UTC(),
		latitude:  latitude,
		longitude: longitude,
	}

	leftAltitudeM, leftAvailable :=
		usableAltitude(left)
	rightAltitudeM, rightAvailable :=
		usableAltitude(right)
	if leftAvailable && rightAvailable {
		altitudeM :=
			leftAltitudeM +
				(rightAltitudeM-
					leftAltitudeM)*
					fraction
		if finite(altitudeM) {
			result.altitudeM =
				float64Pointer(altitudeM)
		}
	}

	return result, true
}

func usableAltitude(
	point trajectory.TrackPoint4D,
) (float64, bool) {
	geometricStatus :=
		flightstate.ResolveAltitudeStatus(
			point.GeometricAltitudeM,
			point.GeometricAltitudeStatus,
		)
	if usableAltitudeStatus(
		geometricStatus,
	) &&
		finite(point.GeometricAltitudeM) {
		return point.GeometricAltitudeM,
			true
	}

	barometricStatus :=
		flightstate.ResolveAltitudeStatus(
			point.BarometricAltitudeM,
			point.BarometricAltitudeStatus,
		)
	if usableAltitudeStatus(
		barometricStatus,
	) &&
		finite(point.BarometricAltitudeM) {
		return point.BarometricAltitudeM,
			true
	}

	return 0, false
}

func usableAltitudeStatus(
	status flightstate.AltitudeStatus,
) bool {
	return status ==
		flightstate.AltitudeStatusObserved ||
		status ==
			flightstate.AltitudeStatusGround
}

func float64Pointer(value float64) *float64 {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
