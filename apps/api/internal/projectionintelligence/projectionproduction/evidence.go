package projectionproduction

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type AuthorizedHistoricalEvidence struct {
	Plan       projectionhorizon.Plan
	Route      routecontract.Result
	History    projectionroutefrequency.HistorySummary
	RouteScope projectionneighbors.RouteScope
	Selection  projectionneighbors.Result
	Pattern    projectionpatternconfidence.Result
	Freshness  projectionfreshness.Result
	Frequency  projectionroutefrequency.Result
}

func (evidence AuthorizedHistoricalEvidence) Clone() AuthorizedHistoricalEvidence {
	return AuthorizedHistoricalEvidence{
		Plan:       evidence.Plan.Clone(),
		Route:      evidence.Route.Clone(),
		History:    evidence.History.Clone(),
		RouteScope: evidence.RouteScope.Clone(),
		Selection:  evidence.Selection.Clone(),
		Pattern:    evidence.Pattern.Clone(),
		Freshness:  evidence.Freshness.Clone(),
		Frequency:  evidence.Frequency.Clone(),
	}
}

func (evidence AuthorizedHistoricalEvidence) Validate(request Request) error {
	if err := evidence.Plan.Validate(); err != nil {
		return fmt.Errorf("authorized plan: %w", err)
	}
	if err := validateRouteBinding(request, evidence.Plan); err != nil {
		return err
	}
	if err := evidence.RouteScope.ValidateForCandidates(request.HistoricalCandidates); err != nil {
		return fmt.Errorf("authorized route scope: %w", err)
	}
	if err := evidence.Selection.Validate(); err != nil {
		return fmt.Errorf("authorized selection: %w", err)
	}
	if strings.TrimSpace(evidence.Selection.CurrentTrajectoryID) != strings.TrimSpace(request.CurrentTrajectory.ID) ||
		!evidence.Selection.AsOfTime.UTC().Equal(evidence.Plan.AsOfTime) ||
		evidence.Selection.RequiredContinuationDuration != evidence.Plan.EffectiveDuration {
		return fmt.Errorf("authorized selection does not match production request and plan")
	}
	selected := selectedTrajectoryIDs(evidence.Selection)
	available := candidateIDs(request.HistoricalCandidates)
	for _, id := range selected {
		if _, ok := available[id]; !ok {
			return fmt.Errorf("authorized selection references missing candidate %q", id)
		}
	}
	if err := evidence.Pattern.Validate(); err != nil {
		return fmt.Errorf("authorized pattern: %w", err)
	}
	if evidence.Pattern.SourceSelectionFingerprint != evidence.Selection.InputFingerprint ||
		!sameStrings(selected, evidence.Pattern.SelectedTrajectoryIDs) {
		return fmt.Errorf("authorized pattern does not match selection")
	}
	if err := evidence.Freshness.Validate(); err != nil {
		return fmt.Errorf("authorized freshness: %w", err)
	}
	if evidence.Freshness.SourceSelectionFingerprint != evidence.Selection.InputFingerprint ||
		evidence.Freshness.SourcePatternFingerprint != evidence.Pattern.InputFingerprint ||
		!evidence.Freshness.AsOfTime.UTC().Equal(evidence.Plan.AsOfTime) ||
		!sameStrings(selected, evidence.Freshness.SelectedTrajectoryIDs) {
		return fmt.Errorf("authorized freshness does not match selection, pattern, or plan")
	}
	if err := evidence.History.Validate(); err != nil {
		return fmt.Errorf("authorized route history: %w", err)
	}
	if err := evidence.Frequency.Validate(); err != nil {
		return fmt.Errorf("authorized route frequency: %w", err)
	}
	expectedRouteKey, err := completeRouteKey(evidence.Route)
	if err != nil {
		return err
	}
	if evidence.History.RouteKey != expectedRouteKey ||
		!evidence.History.AsOfTime.UTC().Equal(evidence.Plan.AsOfTime) ||
		evidence.Frequency.RouteKey != expectedRouteKey ||
		!evidence.Frequency.AsOfTime.UTC().Equal(evidence.Plan.AsOfTime) ||
		evidence.Frequency.HistoryInputFingerprint != evidence.History.InputFingerprint {
		return fmt.Errorf("authorized route-frequency evidence does not match route history and plan")
	}
	return nil
}

func validateRouteBinding(request Request, plan projectionhorizon.Plan) error {
	report := routecontract.Validate(request.Route)
	if report.Status != routecontract.ValidationStatusValid {
		return fmt.Errorf("%w: %#v", ErrRouteContractInvalid, report.Issues)
	}
	if strings.TrimSpace(request.Route.TrajectoryID) != strings.TrimSpace(request.CurrentTrajectory.ID) ||
		request.Route.FlightID != request.CurrentTrajectory.FlightID ||
		request.Route.AircraftID != request.CurrentTrajectory.AircraftID ||
		request.Route.ICAO24 != request.CurrentTrajectory.ICAO24 ||
		request.Route.Callsign != request.CurrentTrajectory.Callsign {
		return fmt.Errorf("%w: route identity does not match current trajectory", ErrRouteContractInvalid)
	}
	if request.Route.Window.AsOfTime.After(plan.AsOfTime) ||
		request.Route.GeneratedAt.After(request.GeneratedAt.UTC()) ||
		request.Route.Provenance.TrajectoryUpdatedAt.After(plan.AsOfTime) {
		return fmt.Errorf("%w: route evidence exceeds production time boundary", ErrRouteContractInvalid)
	}
	return nil
}

func completeRouteKey(route routecontract.Result) (string, error) {
	if route.Status != routecontract.RouteStatusComplete || route.Origin == nil || route.Destination == nil {
		return "", fmt.Errorf("complete route is required")
	}
	return route.Origin.Airport.ICAOCode + ">" + route.Destination.Airport.ICAOCode, nil
}

func selectedTrajectoryIDs(selection projectionneighbors.Result) []string {
	items := make([]string, 0, len(selection.Neighbors))
	for _, neighbor := range selection.Neighbors {
		items = append(items, strings.TrimSpace(neighbor.TrajectoryID))
	}
	sort.Strings(items)
	return items
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
