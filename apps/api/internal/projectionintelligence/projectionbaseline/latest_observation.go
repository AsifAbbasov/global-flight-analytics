package projectionbaseline

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func selectLatestProjectionPoint(
	points []trajectory.TrackPoint4D,
) (trajectory.TrackPoint4D, projectioncontract.Limitation, bool) {
	if len(points) == 0 {
		return trajectory.TrackPoint4D{}, projectioncontract.Limitation{}, false
	}

	latestTime := points[len(points)-1].ObservedAt.UTC()
	start := len(points) - 1
	for start > 0 && points[start-1].ObservedAt.UTC().Equal(latestTime) {
		start--
	}

	selected := points[start]
	for _, candidate := range points[start+1:] {
		if !equivalentProjectionEvidence(selected, candidate) {
			return trajectory.TrackPoint4D{}, projectioncontract.Limitation{
				Code:    "projection_latest_observation_ambiguous",
				Message: "Multiple observations at the latest timestamp contain conflicting projection-driving telemetry.",
				Scope:   "input",
			}, false
		}
	}

	return selected, projectioncontract.Limitation{}, true
}

func equivalentProjectionEvidence(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
) bool {
	return sameTime(left.ObservedAt, right.ObservedAt) &&
		left.Latitude == right.Latitude &&
		left.Longitude == right.Longitude &&
		left.BarometricAltitudeM == right.BarometricAltitudeM &&
		left.BarometricAltitudeStatus == right.BarometricAltitudeStatus &&
		left.GeometricAltitudeM == right.GeometricAltitudeM &&
		left.GeometricAltitudeStatus == right.GeometricAltitudeStatus &&
		left.VelocityMPS == right.VelocityMPS &&
		left.VelocityAvailable == right.VelocityAvailable &&
		left.HeadingDegrees == right.HeadingDegrees &&
		left.HeadingAvailable == right.HeadingAvailable &&
		left.VerticalRateMPS == right.VerticalRateMPS &&
		left.VerticalRateAvailable == right.VerticalRateAvailable &&
		left.OnGround == right.OnGround &&
		left.OnGroundAvailable == right.OnGroundAvailable &&
		left.TelemetryAvailabilityKnown == right.TelemetryAvailabilityKnown
}

func sameTime(left time.Time, right time.Time) bool {
	return left.UTC().Equal(right.UTC())
}
