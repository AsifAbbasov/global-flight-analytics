package projectionarrival

import (
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type validatedArrivalRequest struct {
	projectionAsOf time.Time
	routeAsOf      time.Time
	generatedAt    time.Time
}

func validateArrivalRequest(
	request Request,
) (validatedArrivalRequest, error) {
	projectionReport :=
		projectioncontract.Validate(
			request.Projection,
		)
	if projectionReport.Status !=
		projectioncontract.
			ValidationStatusValid {
		return validatedArrivalRequest{},
			fmt.Errorf(
				"%w: %#v",
				ErrProjectionContractInvalid,
				projectionReport.Issues,
			)
	}

	routeReport := routecontract.Validate(
		request.Route,
	)
	if routeReport.Status !=
		routecontract.ValidationStatusValid {
		return validatedArrivalRequest{},
			fmt.Errorf(
				"%w: %#v",
				ErrRouteContractInvalid,
				routeReport.Issues,
			)
	}

	if request.Route.TrajectoryID !=
		request.Projection.TrajectoryID ||
		(strings.TrimSpace(
			request.CurrentTrajectory.ID,
		) != "" &&
			request.CurrentTrajectory.ID !=
				request.Projection.
					TrajectoryID) {
		return validatedArrivalRequest{},
			ErrTrajectoryMismatch
	}

	projectionAsOf := request.Projection.
		Horizon.AsOfTime.UTC()
	routeAsOf :=
		request.Route.Window.AsOfTime.UTC()
	if routeAsOf.After(projectionAsOf) {
		return validatedArrivalRequest{},
			ErrFutureRouteEvidence
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(
			request.Projection.GeneratedAt.UTC(),
		) ||
		generatedAt.Before(
			request.Route.GeneratedAt.UTC(),
		) ||
		generatedAt.Before(projectionAsOf) {
		return validatedArrivalRequest{},
			ErrGeneratedAtInvalid
	}

	return validatedArrivalRequest{
		projectionAsOf: projectionAsOf,
		routeAsOf:      routeAsOf,
		generatedAt:    generatedAt,
	}, nil
}

func (
	estimator *Estimator,
) attachArrivalResult(
	request Request,
	validated validatedArrivalRequest,
	destination *routecontract.EndpointInference,
	computation arrivalComputation,
) (projectioncontract.Result, error) {
	result := request.Projection.Clone()
	arrivalConfidence :=
		estimator.arrivalConfidence(
			request.Projection,
			destination.Confidence.Score,
			computation,
		)

	result.Arrival =
		&projectioncontract.ArrivalEstimate{
			AirportICAOCode: strings.TrimSpace(
				destination.Airport.ICAOCode,
			),
			EarliestTime: computation.
				earliestTime.UTC(),
			EstimatedTime: computation.
				estimatedTime.UTC(),
			LatestTime: computation.
				latestTime.UTC(),
			Confidence: arrivalConfidence,
			Limitations: arrivalLimitations(
				computation.mode,
				request.Route.Status,
			),
		}

	if result.Status ==
		projectioncontract.
			ResultStatusComplete &&
		(computation.mode ==
			EstimateModeExtrapolated ||
			request.Route.Status !=
				routecontract.
					RouteStatusComplete) {
		result.Status =
			projectioncontract.
				ResultStatusLimited
	}

	result.Confidence =
		estimator.combinedConfidence(
			result.Confidence,
			arrivalConfidence,
		)
	result.Limitations =
		normalizeLimitations(
			append(
				result.Limitations,
				projectioncontract.Limitation{
					Code:    "estimated_arrival_boundary_attached",
					Message: "Projection includes an estimated airport-radius arrival interval.",
					Scope:   "arrival",
				},
			),
		)
	result.Explanations =
		normalizeExplanations(
			append(
				result.Explanations,
				projectioncontract.Explanation{
					Code:    MethodName,
					Message: "Estimated arrival is based on destination inference, projected position samples, and a bounded projected ground-speed profile.",
				},
			),
		)

	result.Provenance.Inputs = append(
		result.Provenance.Inputs,
		projectioncontract.InputReference{
			Name: "route_destination_inference",
			Classification: projectioncontract.
				InputClassificationDerived,
			SourceName:  "routeintelligence",
			ObservedAt:  validated.routeAsOf,
			RetrievedAt: validated.generatedAt,
		},
		projectioncontract.InputReference{
			Name: "projected_arrival_speed_profile",
			Classification: projectioncontract.
				InputClassificationEstimated,
			SourceName:  "projectionarrival",
			ObservedAt:  validated.projectionAsOf,
			RetrievedAt: validated.generatedAt,
			Limitation:  "Ground speed is derived from estimated projection points and is not an official flight-plan speed.",
		},
	)
	result.Provenance.Inputs =
		normalizeInputs(
			result.Provenance.Inputs,
		)
	result.Provenance.
		LatestInputObservedAt =
		latestInputObservedAt(
			result.Provenance.Inputs,
		)
	result.Provenance.InputFingerprint =
		arrivalFingerprint(
			request.Projection,
			request.Route,
			computation,
			estimator.config,
		)
	result.GeneratedAt = validated.generatedAt

	return validateResult(result)
}
