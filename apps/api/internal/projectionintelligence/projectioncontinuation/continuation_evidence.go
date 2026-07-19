package projectioncontinuation

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func buildCandidateIndex(
	items []trajectory.FlightTrajectory,
	asOfTime time.Time,
) map[string]trajectory.FlightTrajectory {
	result := make(
		map[string]trajectory.FlightTrajectory,
		len(items),
	)
	duplicateIDs := make(map[string]bool)

	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, exists := result[id]; exists {
			duplicateIDs[id] = true
			delete(result, id)
			continue
		}
		if duplicateIDs[id] {
			continue
		}

		result[id] =
			trajectorySnapshotAt(
				item,
				asOfTime,
			)
	}

	return result
}

func historicalContinuationLimitations(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) []projectioncontract.Limitation {
	result := []projectioncontract.Limitation{
		{
			Code:    "historical_neighbor_continuation_experimental",
			Message: "Historical-neighbor continuation is project-derived and experimental until calibrated by replay.",
			Scope:   "method",
		},
		{
			Code:    "historical_behavior_not_intent",
			Message: "Historical continuation patterns do not represent official flight plans, Air Traffic Control instructions, pilot intent, or guaranteed future maneuvers.",
			Scope:   "method",
		},
		{
			Code:    "no_weather_adjustment",
			Message: "Weather and wind are not applied by this continuation baseline.",
			Scope:   "method",
		},
		{
			Code:    "research_only",
			Message: "Projection is a research estimate and must not be used for operational aviation decisions.",
			Scope:   "result",
		},
	}

	for _, limitation := range selection.Limitations {
		result = append(
			result,
			projectioncontract.Limitation{
				Code: "neighbor_selection_" +
					limitation.Code,
				Message: limitation.Message,
				Scope:   "selection",
			},
		)
	}
	for _, limitation := range pattern.Limitations {
		result = append(
			result,
			projectioncontract.Limitation{
				Code: "pattern_confidence_" +
					limitation.Code,
				Message: limitation.Message,
				Scope:   "confidence",
			},
		)
	}

	return result
}

func continuationInputs(
	currentEndpoint trajectory.TrackPoint4D,
	selection projectionneighbors.Result,
) []projectioncontract.InputReference {
	result := []projectioncontract.InputReference{
		{
			Name: "current_trajectory_endpoint",
			Classification: projectioncontract.
				InputClassificationObserved,
			SourceName: currentEndpoint.SourceName,
			ObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
		{
			Name: "historical_neighbor_selection",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "projectionneighbors",
			ObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
		{
			Name: "historical_pattern_confidence",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "projectionpatternconfidence",
			ObservedAt: currentEndpoint.
				ObservedAt.UTC(),
		},
	}

	for _, neighbor := range selection.Neighbors {
		result = append(
			result,
			projectioncontract.InputReference{
				Name: "historical_neighbor:" +
					neighbor.
						TrajectoryID,
				Classification: projectioncontract.
					InputClassificationDerived,
				SourceName: "historical_trajectory",
				ObservedAt: neighbor.
					CandidateEndTime.UTC(),
			},
		)
	}

	return result
}

func patternMatchesSelection(
	pattern projectionpatternconfidence.Result,
	selection projectionneighbors.Result,
) bool {
	if pattern.NeighborCount !=
		len(selection.Neighbors) {
		return false
	}

	selected := make(
		[]string,
		0,
		len(selection.Neighbors),
	)
	for _, neighbor := range selection.Neighbors {
		selected = append(
			selected,
			strings.TrimSpace(
				neighbor.TrajectoryID,
			),
		)
	}
	sort.Strings(selected)

	if len(selected) !=
		len(pattern.SelectedTrajectoryIDs) {
		return false
	}
	for index := range selected {
		if selected[index] !=
			pattern.SelectedTrajectoryIDs[index] {
			return false
		}
	}

	return true
}

func minimumPointConfidence(
	points []projectioncontract.ProjectionPoint,
) projectioncontract.Confidence {
	if len(points) == 0 {
		return projectioncontract.Confidence{
			Score: 0,
			Level: projectioncontract.
				ConfidenceLevelNone,
		}
	}

	minimum := points[0].Confidence
	for _, point := range points[1:] {
		if point.Confidence.Score <
			minimum.Score {
			minimum = point.Confidence
		}
	}

	minimum.Reasons =
		[]projectioncontract.ConfidenceReason{
			{
				Code:         "minimum_historical_continuation_point_confidence",
				Message:      "Result confidence equals the lowest confidence across historical-continuation forecast points.",
				Contribution: minimum.Score,
			},
		}

	return minimum
}

func (
	baseline *Baseline,
) confidenceLevel(
	score float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= baseline.config.
		HighConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelHigh
	case score >= baseline.config.
		MediumConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.
			ConfidenceLevelLow
	default:
		return projectioncontract.
			ConfidenceLevelNone
	}
}

func normalizeLimitations(
	items []projectioncontract.Limitation,
) []projectioncontract.Limitation {
	seen := make(
		map[string]projectioncontract.Limitation,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message :=
			strings.TrimSpace(item.Message)
		scope := strings.TrimSpace(item.Scope)
		if code == "" ||
			message == "" ||
			scope == "" {
			continue
		}

		key := code + "\x00" +
			message + "\x00" +
			scope
		seen[key] =
			projectioncontract.Limitation{
				Code:    code,
				Message: message,
				Scope:   scope,
			}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]projectioncontract.Limitation,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}

func validateProjectionResult(
	result projectioncontract.Result,
) (projectioncontract.Result, error) {
	report := projectioncontract.Validate(
		result,
	)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrContinuationContractInvalid,
				report.Issues,
			)
	}

	return result.Clone(), nil
}
