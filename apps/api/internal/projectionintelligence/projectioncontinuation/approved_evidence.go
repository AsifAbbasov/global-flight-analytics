package projectioncontinuation

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type continuationEvidenceResolution struct {
	selection projectionneighbors.Result
	pattern   projectionpatternconfidence.Result

	fallbackReason       string
	selectionFingerprint string
	patternFingerprint   string
}

func (
	baseline *Baseline,
) resolveContinuationEvidence(
	request Request,
	plan projectionhorizon.Plan,
	approvedEvidence *ApprovedEvidence,
) continuationEvidenceResolution {
	if approvedEvidence != nil {
		evidence := approvedEvidence.Clone()
		return validateApprovedContinuationEvidence(
			request,
			plan,
			evidence.Selection,
			evidence.Pattern,
		)
	}

	selection, err := baseline.config.
		NeighborSelector.Select(
		projectionneighbors.Request{
			CurrentTrajectory: request.CurrentTrajectory,
			Candidates:        request.Candidates,
			RouteScope:        request.RouteScope,
			AsOfTime:          plan.AsOfTime,
			RequiredContinuationDuration: plan.
				EffectiveDuration,
		},
	)
	if err != nil {
		return continuationEvidenceResolution{
			fallbackReason: "historical_neighbor_selection_failed",
		}
	}
	if err := selection.Validate(); err != nil {
		return continuationEvidenceResolution{
			fallbackReason: "historical_neighbor_selection_invalid",
			selectionFingerprint: selection.
				InputFingerprint,
		}
	}

	pattern, err := evaluatePatternConfidence(
		baseline.config.
			PatternConfidenceEvaluator,
		selection,
		request.Candidates,
	)
	if err != nil {
		return continuationEvidenceResolution{
			fallbackReason: "historical_pattern_confidence_failed",
			selectionFingerprint: selection.
				InputFingerprint,
		}
	}

	return validateApprovedContinuationEvidence(
		request,
		plan,
		selection,
		pattern,
	)
}

func validateApprovedContinuationEvidence(
	request Request,
	plan projectionhorizon.Plan,
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) continuationEvidenceResolution {
	result := continuationEvidenceResolution{
		selection: selection.Clone(),
		pattern:   pattern.Clone(),
		selectionFingerprint: selection.
			InputFingerprint,
		patternFingerprint: pattern.
			InputFingerprint,
	}

	if err := selection.Validate(); err != nil {
		result.fallbackReason =
			"historical_neighbor_selection_invalid"
		return result
	}
	if err := pattern.Validate(); err != nil {
		result.fallbackReason =
			"historical_pattern_confidence_invalid"
		return result
	}
	if strings.TrimSpace(
		selection.CurrentTrajectoryID,
	) != strings.TrimSpace(
		request.CurrentTrajectory.ID,
	) ||
		!selection.AsOfTime.UTC().Equal(
			plan.AsOfTime.UTC(),
		) ||
		selection.RequiredContinuationDuration !=
			plan.EffectiveDuration {
		result.fallbackReason =
			"historical_approved_evidence_request_mismatch"
		return result
	}
	if !patternMatchesSelection(
		pattern,
		selection,
	) {
		result.fallbackReason =
			"historical_pattern_selection_mismatch"
		return result
	}
	if !pattern.Usable {
		result.fallbackReason =
			"historical_pattern_not_usable"
		return result
	}

	return result
}

func selectedCandidateEvidenceMatches(
	selection projectionneighbors.Result,
	candidateByID map[string]trajectory.FlightTrajectory,
) bool {
	for _, neighbor := range selection.Neighbors {
		candidate, exists := candidateByID[strings.TrimSpace(
			neighbor.TrajectoryID,
		)]
		if !exists ||
			len(candidate.Points) == 0 ||
			!candidate.StartTime.UTC().Equal(
				neighbor.CandidateStartTime.UTC(),
			) ||
			!candidate.EndTime.UTC().Equal(
				neighbor.CandidateEndTime.UTC(),
			) ||
			neighbor.AnchorPointIndex !=
				neighbor.PrefixPointCount-1 ||
			neighbor.AnchorPointIndex < 0 ||
			neighbor.AnchorPointIndex >=
				len(candidate.Points) {
			return false
		}

		anchor := candidate.Points[neighbor.AnchorPointIndex]
		if !anchor.ObservedAt.UTC().Equal(
			neighbor.AnchorObservedAt.UTC(),
		) {
			return false
		}

		continuationEndIndex :=
			neighbor.AnchorPointIndex +
				neighbor.ContinuationPointCount
		if continuationEndIndex <
			neighbor.AnchorPointIndex ||
			continuationEndIndex >=
				len(candidate.Points) ||
			!candidate.Points[continuationEndIndex].ObservedAt.UTC().Equal(
				neighbor.
					ContinuationEndTime.UTC(),
			) ||
			candidateEvidenceSourceName(
				candidate,
				neighbor.AnchorPointIndex,
			) == "" {
			return false
		}
	}

	return true
}

func candidateEvidenceSourceName(
	candidate trajectory.FlightTrajectory,
	anchorPointIndex int,
) string {
	if sourceName := strings.TrimSpace(
		candidate.SourceName,
	); sourceName != "" {
		return sourceName
	}
	if anchorPointIndex >= 0 &&
		anchorPointIndex <
			len(candidate.Points) {
		if sourceName := strings.TrimSpace(
			candidate.Points[anchorPointIndex].SourceName,
		); sourceName != "" {
			return sourceName
		}
	}
	if len(candidate.Points) > 0 {
		return strings.TrimSpace(
			candidate.Points[len(candidate.Points)-1].SourceName,
		)
	}
	return ""
}
