package projectionbaseline

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type altitudeReference string

const (
	altitudeReferenceUnavailable altitudeReference = "unavailable"
	altitudeReferenceGeometric   altitudeReference = "geometric"
	altitudeReferenceBarometric  altitudeReference = "barometric"
)

type altitudeSelection struct {
	ValueM    float64
	Reference altitudeReference
	Available bool
}

func selectAltitude(point trajectory.TrackPoint4D) altitudeSelection {
	if availableAltitudeOutOfBounds(point) {
		return altitudeSelection{Reference: altitudeReferenceUnavailable}
	}

	geometricStatus := flightstate.ResolveAltitudeStatus(
		point.GeometricAltitudeM,
		point.GeometricAltitudeStatus,
	)
	if usableAltitudeStatus(geometricStatus) &&
		altitudeWithinPhysicalBounds(point.GeometricAltitudeM) {
		return altitudeSelection{
			ValueM:    point.GeometricAltitudeM,
			Reference: altitudeReferenceGeometric,
			Available: true,
		}
	}

	barometricStatus := flightstate.ResolveAltitudeStatus(
		point.BarometricAltitudeM,
		point.BarometricAltitudeStatus,
	)
	if usableAltitudeStatus(barometricStatus) &&
		altitudeWithinPhysicalBounds(point.BarometricAltitudeM) {
		return altitudeSelection{
			ValueM:    point.BarometricAltitudeM,
			Reference: altitudeReferenceBarometric,
			Available: true,
		}
	}

	return altitudeSelection{
		Reference: altitudeReferenceUnavailable,
	}
}

func (selection altitudeSelection) provenanceLimitation() string {
	switch selection.Reference {
	case altitudeReferenceGeometric:
		return "Selected altitude reference is geometric."
	case altitudeReferenceBarometric:
		return "Selected altitude reference is barometric."
	default:
		return "Altitude reference is unavailable."
	}
}

func (selection altitudeSelection) normalizedReference() string {
	return strings.TrimSpace(string(selection.Reference))
}

func usableAltitudeStatus(status flightstate.AltitudeStatus) bool {
	return status == flightstate.AltitudeStatusObserved ||
		status == flightstate.AltitudeStatusGround
}
