package projectionneighbors

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type selectionContext struct {
	currentID                    string
	asOfTime                     time.Time
	requiredContinuationDuration time.Duration

	current            trajectory.FlightTrajectory
	latestCurrentPoint trajectory.TrackPoint4D
	inputCandidates    []trajectory.FlightTrajectory
	routeScope         routeScopeIndex
	candidatePool      preparedCandidatePool

	excludedCurrentPointCount int
}

func (selector *Selector) prepareSelectionContext(
	request Request,
) (selectionContext, error) {
	currentID := strings.TrimSpace(request.CurrentTrajectory.ID)
	if currentID == "" {
		return selectionContext{}, ErrCurrentTrajectoryIDRequired
	}
	if request.AsOfTime.IsZero() {
		return selectionContext{}, ErrAsOfTimeRequired
	}
	if request.RequiredContinuationDuration <= 0 {
		return selectionContext{}, ErrContinuationDurationInvalid
	}

	routeScope, err := prepareRouteScope(
		request.RouteScope,
		request.Candidates,
	)
	if err != nil {
		return selectionContext{}, fmt.Errorf(
			"%w: %v",
			ErrRouteScopeInvalid,
			err,
		)
	}

	asOfTime := request.AsOfTime.UTC()
	current, excludedCurrentPointCount := snapshotAt(
		request.CurrentTrajectory,
		asOfTime,
	)
	if len(current.Points) < selector.config.MinimumCurrentPointCount {
		return selectionContext{}, fmt.Errorf(
			"%w: usable=%d minimum=%d",
			ErrCurrentTrajectoryNotComparable,
			len(current.Points),
			selector.config.MinimumCurrentPointCount,
		)
	}

	latestCurrentPoint := current.Points[len(current.Points)-1]
	currentStartTime := current.Points[0].ObservedAt.UTC()
	candidatePool := prepareCandidatePool(
		request.Candidates,
		currentID,
		currentStartTime,
		asOfTime,
		selector.config,
		routeScope,
	)

	return selectionContext{
		currentID:                    currentID,
		asOfTime:                     asOfTime,
		requiredContinuationDuration: request.RequiredContinuationDuration,
		current:                      current,
		latestCurrentPoint:           latestCurrentPoint,
		inputCandidates: append(
			[]trajectory.FlightTrajectory(nil),
			request.Candidates...,
		),
		routeScope:                routeScope,
		candidatePool:             candidatePool,
		excludedCurrentPointCount: excludedCurrentPointCount,
	}, nil
}
