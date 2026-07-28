package projectionbaseline

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

func unavailableProvenance(
	item trajectory.FlightTrajectory,
	plan projectionhorizon.Plan,
	config Config,
) projectioncontract.Provenance {
	latestPoint := trajectory.TrackPoint4D{}
	altitudeAvailable := false
	inputs := []projectioncontract.InputReference(nil)
	latestObservedAt := latestPoint.ObservedAt

	if len(item.Points) > 0 {
		latestPoint = item.Points[len(item.Points)-1]
		if !latestPoint.ObservedAt.IsZero() &&
			strings.TrimSpace(latestPoint.SourceName) != "" {
			_, altitudeAvailable = usableAltitude(latestPoint)
			inputs = projectionInputs(
				item,
				latestPoint,
				altitudeAvailable,
			)
			latestObservedAt = latestPoint.ObservedAt.UTC()
		}
	}

	return projectioncontract.Provenance{
		InputFingerprint: inputFingerprint(
			item,
			latestPoint,
			plan,
			config,
		),
		Inputs:                inputs,
		LatestInputObservedAt: latestObservedAt,
	}
}
