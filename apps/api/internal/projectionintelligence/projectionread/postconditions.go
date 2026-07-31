package projectionread

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func validateSnapshotPostconditions(
	snapshot Snapshot,
	trajectoryID string,
	asOfTime time.Time,
) error {
	expectedTrajectoryID := strings.TrimSpace(
		trajectoryID,
	)
	asOfTime = asOfTime.UTC()
	current := snapshot.CurrentTrajectory

	if strings.TrimSpace(current.ID) != expectedTrajectoryID ||
		current.StartTime.IsZero() ||
		current.EndTime.IsZero() ||
		current.StartTime.UTC().After(asOfTime) ||
		current.EndTime.UTC().After(asOfTime) ||
		current.EndTime.UTC().Before(
			current.StartTime.UTC(),
		) {
		return fmt.Errorf(
			"%w: current trajectory does not match the requested as-of snapshot",
			ErrSnapshotIdentityMismatch,
		)
	}
	for _, point := range current.Points {
		if point.ObservedAt.IsZero() ||
			point.ObservedAt.UTC().After(asOfTime) {
			return fmt.Errorf(
				"%w: current trajectory contains a point outside the as-of snapshot",
				ErrSnapshotIdentityMismatch,
			)
		}
	}

	if snapshot.Route != nil {
		route := snapshot.Route.Clone()
		report := routecontract.Validate(route)
		if report.Status !=
			routecontract.ValidationStatusValid {
			return fmt.Errorf(
				"%w: route contract is invalid: %#v",
				ErrRouteSnapshotInvalid,
				report.Issues,
			)
		}
		if strings.TrimSpace(route.TrajectoryID) !=
			expectedTrajectoryID ||
			route.Window.AsOfTime.IsZero() ||
			route.Window.AsOfTime.UTC().After(asOfTime) {
			return fmt.Errorf(
				"%w: route identity or time boundary does not match the snapshot",
				ErrRouteSnapshotInvalid,
			)
		}
	}

	for _, candidate := range snapshot.HistoricalCandidates {
		candidateID := strings.TrimSpace(candidate.ID)
		if candidateID == "" ||
			candidateID == expectedTrajectoryID ||
			candidate.EndTime.IsZero() ||
			!candidate.EndTime.UTC().Before(
				current.StartTime.UTC(),
			) ||
			candidate.EndTime.UTC().After(asOfTime) {
			return fmt.Errorf(
				"%w: historical candidate does not belong to the authorized snapshot",
				ErrSnapshotIdentityMismatch,
			)
		}
	}

	if snapshot.RouteHistory != nil {
		history := snapshot.RouteHistory.Clone()
		if err := history.Validate(); err != nil {
			return fmt.Errorf(
				"%w: route history is invalid: %w",
				ErrSnapshotIdentityMismatch,
				err,
			)
		}
		if !history.AsOfTime.UTC().Equal(asOfTime) {
			return fmt.Errorf(
				"%w: route history does not match the snapshot as-of time",
				ErrSnapshotIdentityMismatch,
			)
		}
	}

	return nil
}

func validateComposedResult(
	result projectionproduction.Result,
	current trajectory.FlightTrajectory,
	asOfTime time.Time,
) error {
	if err := result.Validate(); err != nil {
		return fmt.Errorf(
			"%w: %w",
			ErrComposedResultInvalid,
			err,
		)
	}

	projection := result.Projection
	if strings.TrimSpace(projection.TrajectoryID) !=
		strings.TrimSpace(current.ID) ||
		strings.TrimSpace(projection.FlightID) !=
			strings.TrimSpace(current.FlightID) ||
		strings.TrimSpace(projection.AircraftID) !=
			strings.TrimSpace(current.AircraftID) ||
		strings.ToUpper(
			strings.TrimSpace(projection.ICAO24),
		) != strings.ToUpper(
			strings.TrimSpace(current.ICAO24),
		) ||
		strings.TrimSpace(projection.Callsign) !=
			strings.TrimSpace(current.Callsign) {
		return fmt.Errorf(
			"%w: composed projection identity does not match the snapshot trajectory",
			ErrSnapshotIdentityMismatch,
		)
	}
	if !result.HorizonPlan.AsOfTime.UTC().Equal(
		asOfTime.UTC(),
	) {
		return fmt.Errorf(
			"%w: composed projection does not match the requested as-of time",
			ErrSnapshotIdentityMismatch,
		)
	}

	return nil
}
