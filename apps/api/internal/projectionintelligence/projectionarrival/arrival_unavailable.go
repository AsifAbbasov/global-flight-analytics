package projectionarrival

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func (
	estimator *Estimator,
) withUnavailableArrival(
	request Request,
	reason unavailableReason,
	message string,
	evidence positionEvidence,
) (projectioncontract.Result, error) {
	result := request.Projection.Clone()
	result.Arrival = nil

	if result.Status == projectioncontract.ResultStatusComplete {
		result.Status = projectioncontract.ResultStatusLimited
	}

	result.Limitations = normalizeLimitations(
		append(
			result.Limitations,
			projectioncontract.Limitation{
				Code:    "estimated_arrival_unavailable",
				Message: strings.TrimSpace(message),
				Scope:   "arrival",
			},
			projectioncontract.Limitation{
				Code: "estimated_arrival_unavailable_reason",
				Message: "Estimated arrival reason: " +
					strings.TrimSpace(string(reason)) + ".",
				Scope: "arrival",
			},
		),
	)

	if result.Status != projectioncontract.ResultStatusUnavailable {
		result.Explanations = normalizeExplanations(
			append(
				result.Explanations,
				projectioncontract.Explanation{
					Code:    "estimated_arrival_withheld",
					Message: "Estimated arrival was withheld rather than publishing an unsupported interval.",
				},
			),
		)
	}

	routeFingerprint := request.Route.Provenance.InputFingerprint
	result.Provenance.InputFingerprint = unavailableFingerprint(
		result.Provenance.InputFingerprint,
		routeFingerprint,
		string(reason),
		evidence.fingerprint,
		estimator.config,
	)

	if !request.Route.Window.AsOfTime.IsZero() &&
		!request.Route.Window.AsOfTime.After(
			result.Horizon.AsOfTime,
		) {
		result.Provenance.Inputs = append(
			result.Provenance.Inputs,
			projectioncontract.InputReference{
				Name: "route_destination_inference",
				Classification: projectioncontract.
					InputClassificationDerived,
				SourceName:  "routeintelligence",
				ObservedAt:  request.Route.Window.AsOfTime.UTC(),
				RetrievedAt: request.GeneratedAt.UTC(),
				Limitation:  string(reason),
			},
		)
	}
	if evidence.currentEndpointPresent {
		result.Provenance.Inputs = append(
			result.Provenance.Inputs,
			projectioncontract.InputReference{
				Name: "current_trajectory_arrival_endpoint",
				Classification: projectioncontract.
					InputClassificationObserved,
				SourceName:  evidence.currentEndpoint.sourceName,
				ObservedAt:  evidence.currentEndpoint.timeValue.UTC(),
				RetrievedAt: request.GeneratedAt.UTC(),
				Limitation:  string(reason),
			},
		)
	}
	result.Provenance.Inputs = normalizeInputs(
		result.Provenance.Inputs,
	)
	result.Provenance.LatestInputObservedAt =
		latestInputObservedAt(result.Provenance.Inputs)
	result.GeneratedAt = request.GeneratedAt.UTC()

	return validateResult(result)
}
