package projectioncontinuation

import (
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type interpolatedPoint struct {
	latitude  float64
	longitude float64
	altitudeM *float64
}

func trajectorySnapshotAt(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) trajectory.FlightTrajectory {
	type indexedPoint struct {
		point trajectory.TrackPoint4D
		index int
	}

	indexed := make(
		[]indexedPoint,
		0,
		len(item.Points),
	)
	for index, point := range item.Points {
		if point.ObservedAt.IsZero() ||
			point.ObservedAt.UTC().After(
				asOfTime.UTC(),
			) ||
			!validLatitude(point.Latitude) ||
			!validLongitude(point.Longitude) {
			continue
		}

		point.ObservedAt =
			point.ObservedAt.UTC()
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

			return leftTime.Before(
				rightTime,
			)
		},
	)

	points := make(
		[]trajectory.TrackPoint4D,
		len(indexed),
	)
	for index, item := range indexed {
		points[index] = item.point
	}

	snapshot := item
	snapshot.Points = points
	snapshot.PointCount = len(points)

	if len(points) == 0 {
		snapshot.StartTime = time.Time{}
		snapshot.EndTime = time.Time{}
		snapshot.DurationSeconds = 0
		return snapshot
	}

	snapshot.StartTime =
		points[0].ObservedAt.UTC()
	snapshot.EndTime =
		points[len(points)-1].
			ObservedAt.UTC()
	snapshot.DurationSeconds = int64(
		snapshot.EndTime.Sub(
			snapshot.StartTime,
		).Seconds(),
	)

	return snapshot
}

func interpolateTrajectoryPoint(
	points []trajectory.TrackPoint4D,
	targetTime time.Time,
) (interpolatedPoint, bool) {
	if len(points) == 0 ||
		targetTime.IsZero() {
		return interpolatedPoint{}, false
	}

	targetTime = targetTime.UTC()
	if targetTime.Before(
		points[0].ObservedAt.UTC(),
	) ||
		targetTime.After(
			points[len(points)-1].
				ObservedAt.UTC(),
		) {
		return interpolatedPoint{}, false
	}

	for index, point := range points {
		pointTime := point.ObservedAt.UTC()
		if pointTime.Equal(targetTime) {
			return interpolatedFromPoint(
				point,
			), true
		}
		if pointTime.After(targetTime) {
			if index == 0 {
				return interpolatedPoint{},
					false
			}

			return interpolateBetween(
				points[index-1],
				point,
				targetTime,
			)
		}
	}

	return interpolatedFromPoint(
		points[len(points)-1],
	), true
}

func interpolateBetween(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
	targetTime time.Time,
) (interpolatedPoint, bool) {
	leftTime := left.ObservedAt.UTC()
	rightTime := right.ObservedAt.UTC()
	duration := rightTime.Sub(leftTime)
	if duration <= 0 {
		return interpolatedPoint{}, false
	}

	fraction := float64(
		targetTime.Sub(leftTime),
	) / float64(duration)
	if fraction < 0 ||
		fraction > 1 ||
		!finite(fraction) {
		return interpolatedPoint{}, false
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
		return interpolatedPoint{}, false
	}

	result := interpolatedPoint{
		latitude:  latitude,
		longitude: longitude,
	}

	leftAltitude, leftAvailable :=
		usableAltitude(left)
	rightAltitude, rightAvailable :=
		usableAltitude(right)
	if leftAvailable && rightAvailable {
		altitude := leftAltitude +
			(rightAltitude-leftAltitude)*
				fraction
		if finite(altitude) {
			result.altitudeM =
				float64Pointer(altitude)
		}
	}

	return result, true
}

func interpolatedFromPoint(
	point trajectory.TrackPoint4D,
) interpolatedPoint {
	result := interpolatedPoint{
		latitude:  point.Latitude,
		longitude: point.Longitude,
	}
	if altitude, available :=
		usableAltitude(point); available {
		result.altitudeM =
			float64Pointer(altitude)
	}

	return result
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
