package projectionbaseline

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	maximumSupportedGroundSpeedMPS          = 700.0
	maximumSupportedAbsoluteVerticalRateMPS = 100.0
	minimumSupportedAltitudeM               = -1000.0
	maximumSupportedAltitudeM               = 30000.0
	maximumSupportedOnGroundSpeedMPS        = 50.0
	maximumSupportedOnGroundVerticalRateMPS = 2.0
)

type kinematicPolicy struct {
	AllowOnGround bool
}

func (policy kinematicPolicy) validate(
	point trajectory.TrackPoint4D,
) (projectioncontract.Limitation, bool) {
	switch {
	case !finiteLatitude(point.Latitude) ||
		!finiteLongitude(point.Longitude):
		return projectioncontract.Limitation{
			Code:    "projection_position_invalid",
			Message: "Latest trajectory position is invalid.",
			Scope:   "input",
		}, false

	case !nonNegativeFinite(point.VelocityMPS):
		return projectioncontract.Limitation{
			Code:    "projection_velocity_invalid",
			Message: "Latest trajectory velocity is invalid.",
			Scope:   "input",
		}, false

	case point.VelocityMPS > maximumSupportedGroundSpeedMPS:
		return projectioncontract.Limitation{
			Code:    "projection_velocity_out_of_bounds",
			Message: "Latest trajectory ground speed exceeds the supported physical baseline bound.",
			Scope:   "input",
		}, false

	case !finite(point.HeadingDegrees):
		return projectioncontract.Limitation{
			Code:    "projection_heading_invalid",
			Message: "Latest trajectory heading is invalid.",
			Scope:   "input",
		}, false

	case point.HeadingDegrees < 0 || point.HeadingDegrees >= 360:
		return projectioncontract.Limitation{
			Code:    "projection_heading_out_of_bounds",
			Message: "Latest trajectory heading must be within zero inclusive and three hundred sixty exclusive degrees.",
			Scope:   "input",
		}, false

	case !finite(point.VerticalRateMPS):
		return projectioncontract.Limitation{
			Code:    "projection_vertical_rate_invalid",
			Message: "Latest trajectory vertical rate is invalid.",
			Scope:   "input",
		}, false

	case math.Abs(point.VerticalRateMPS) >
		maximumSupportedAbsoluteVerticalRateMPS:
		return projectioncontract.Limitation{
			Code:    "projection_vertical_rate_out_of_bounds",
			Message: "Latest trajectory vertical rate exceeds the supported physical baseline bound.",
			Scope:   "input",
		}, false

	case availableAltitudeOutOfBounds(point):
		return projectioncontract.Limitation{
			Code:    "projection_altitude_out_of_bounds",
			Message: "Latest trajectory altitude exceeds the supported physical baseline bounds.",
			Scope:   "input",
		}, false

	case point.OnGround && !policy.AllowOnGround:
		return projectioncontract.Limitation{
			Code:    "projection_on_ground_not_allowed",
			Message: "Configured projection policy does not allow an on-ground baseline.",
			Scope:   "input",
		}, false

	case point.OnGround &&
		(point.VelocityMPS > maximumSupportedOnGroundSpeedMPS ||
			math.Abs(point.VerticalRateMPS) >
				maximumSupportedOnGroundVerticalRateMPS):
		return projectioncontract.Limitation{
			Code:    "projection_on_ground_motion_out_of_bounds",
			Message: "On-ground baseline motion exceeds the supported taxi-model bounds.",
			Scope:   "input",
		}, false

	default:
		return projectioncontract.Limitation{}, true
	}
}

func availableAltitudeOutOfBounds(point trajectory.TrackPoint4D) bool {
	barometricStatus := flightstate.ResolveAltitudeStatus(
		point.BarometricAltitudeM,
		point.BarometricAltitudeStatus,
	)
	if barometricStatus == flightstate.AltitudeStatusInvalid ||
		(usableAltitudeStatus(barometricStatus) &&
			!altitudeWithinPhysicalBounds(point.BarometricAltitudeM)) {
		return true
	}

	geometricStatus := flightstate.ResolveAltitudeStatus(
		point.GeometricAltitudeM,
		point.GeometricAltitudeStatus,
	)
	return geometricStatus == flightstate.AltitudeStatusInvalid ||
		(usableAltitudeStatus(geometricStatus) &&
			!altitudeWithinPhysicalBounds(point.GeometricAltitudeM))
}

func altitudeWithinPhysicalBounds(value float64) bool {
	return finite(value) &&
		value >= minimumSupportedAltitudeM &&
		value <= maximumSupportedAltitudeM
}
