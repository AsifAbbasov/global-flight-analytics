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
	altitude := altitudeSelection{Reference: altitudeReferenceUnavailable}
	inputs := []projectioncontract.InputReference{
		eligibilityPolicyInputReference(
			config,
			plan.AsOfTime,
		),
	}
	latestObservedAt := plan.AsOfTime.UTC()

	if len(item.Points) > 0 {
		selectedPoint, _, selected :=
			selectLatestProjectionPoint(
				item.Points,
			)
		if selected {
			latestPoint = selectedPoint
		}
		if selected &&
			!latestPoint.ObservedAt.IsZero() &&
			strings.TrimSpace(latestPoint.SourceName) != "" {
			altitude = selectAltitude(latestPoint)
			inputs = projectionInputs(
				item,
				latestPoint,
				altitude,
			)
			inputs = append(
				inputs,
				eligibilityPolicyInputReference(
					config,
					latestPoint.ObservedAt,
				),
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
