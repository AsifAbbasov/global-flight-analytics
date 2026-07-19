package projectioncontinuation

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"strings"
)

func (
	baseline *Baseline,
) fallback(
	request Request,
	reason string,
	selectionFingerprint string,
	patternFingerprint string,
) (projectioncontract.Result, error) {
	result, err := baseline.config.
		FallbackProjector.Project(
		projectionbaseline.Request{
			Trajectory:        request.CurrentTrajectory,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
			GeneratedAt:       request.GeneratedAt,
		},
	)
	if err != nil {
		return projectioncontract.Result{},
			fmt.Errorf(
				"%w: %v",
				ErrFallbackProjectionFailed,
				err,
			)
	}

	result.Limitations = append(
		result.Limitations,
		projectioncontract.Limitation{
			Code:    "historical_neighbor_strategy_fallback",
			Message: "Historical-neighbor continuation was not usable; the result was produced by the conservative kinematic baseline.",
			Scope:   "method",
		},
		projectioncontract.Limitation{
			Code: "historical_neighbor_fallback_reason",
			Message: "Fallback reason: " +
				strings.TrimSpace(
					reason,
				) +
				".",
			Scope: "method",
		},
	)
	if result.Status !=
		projectioncontract.
			ResultStatusUnavailable {
		result.Explanations = append(
			result.Explanations,
			projectioncontract.Explanation{
				Code:    "kinematic_fallback_selected",
				Message: "Historical pattern evidence was unavailable or insufficient, so the deterministic kinematic baseline was selected.",
			},
		)
	}

	latestObservedAt :=
		result.Provenance.
			LatestInputObservedAt
	result.Provenance.Inputs = append(
		result.Provenance.Inputs,
		projectioncontract.InputReference{
			Name: "historical_neighbor_strategy_decision",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName: "projectioncontinuation",
			ObservedAt: latestObservedAt,
			Limitation: reason,
		},
	)
	result.Provenance.InputFingerprint =
		fallbackFingerprint(
			result.Provenance.
				InputFingerprint,
			reason,
			selectionFingerprint,
			patternFingerprint,
		)
	result.Limitations =
		normalizeLimitations(
			result.Limitations,
		)

	return validateProjectionResult(
		result,
	)
}
