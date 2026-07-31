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
	projectionReport := projectioncontract.Validate(
		request.Projection,
	)
	if projectionReport.Status !=
		projectioncontract.ValidationStatusValid {
		return validatedArrivalRequest{}, fmt.Errorf(
			"%w: %#v",
			ErrProjectionContractInvalid,
			projectionReport.Issues,
		)
	}

	routeReport := routecontract.Validate(request.Route)
	if routeReport.Status != routecontract.ValidationStatusValid {
		return validatedArrivalRequest{}, fmt.Errorf(
			"%w: %#v",
			ErrRouteContractInvalid,
			routeReport.Issues,
		)
	}

	currentTrajectoryID := strings.TrimSpace(
		request.CurrentTrajectory.ID,
	)
	if currentTrajectoryID == "" {
		return validatedArrivalRequest{},
			ErrCurrentTrajectoryIDRequired
	}
	if request.Route.TrajectoryID !=
		request.Projection.TrajectoryID ||
		currentTrajectoryID !=
			request.CurrentTrajectory.ID ||
		currentTrajectoryID !=
			request.Projection.TrajectoryID {
		return validatedArrivalRequest{},
			ErrTrajectoryMismatch
	}

	projectionAsOf := request.Projection.Horizon.AsOfTime.UTC()
	routeAsOf := request.Route.Window.AsOfTime.UTC()
	if routeAsOf.After(projectionAsOf) {
		return validatedArrivalRequest{},
			ErrFutureRouteEvidence
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(request.Projection.GeneratedAt.UTC()) ||
		generatedAt.Before(request.Route.GeneratedAt.UTC()) ||
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
	evidence positionEvidence,
) (projectioncontract.Result, error) {
	result := request.Projection.Clone()
	arrivalConfidence := estimator.arrivalConfidence(
		request.Projection,
		destination.Confidence.Score,
		computation,
	)

	result.Arrival = &projectioncontract.ArrivalEstimate{
		AirportICAOCode: strings.TrimSpace(
			destination.Airport.ICAOCode,
		),
		EarliestTime:  computation.earliestTime.UTC(),
		EstimatedTime: computation.estimatedTime.UTC(),
		LatestTime:    computation.latestTime.UTC(),
		Confidence:    arrivalConfidence,
		Limitations: arrivalLimitations(
			computation.mode,
			request.Route.Status,
		),
	}

	if result.Status == projectioncontract.ResultStatusComplete &&
		(computation.mode == EstimateModeExtrapolated ||
			request.Route.Status != routecontract.RouteStatusComplete) {
		result.Status = projectioncontract.ResultStatusLimited
	}

	result.Confidence = estimator.combinedConfidence(
		result.Confidence,
		arrivalConfidence,
	)
	result.Limitations = normalizeLimitations(
		append(
			result.Limitations,
			projectioncontract.Limitation{
				Code:    "estimated_arrival_boundary_attached",
				Message: "Projection includes an estimated airport-radius arrival interval.",
				Scope:   "arrival",
			},
		),
	)
	result.Explanations = normalizeExplanations(
		append(
			result.Explanations,
			projectioncontract.Explanation{
				Code:    MethodName,
				Message: "Estimated arrival uses destination inference, canonical position evidence, signed closing speed, bounded physical ground speed, and a complete maximum-duration interval.",
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
			Name: "projected_arrival_closing_speed_profile",
			Classification: projectioncontract.
				InputClassificationEstimated,
			SourceName:  "projectionarrival",
			ObservedAt:  validated.projectionAsOf,
			RetrievedAt: validated.generatedAt,
			Limitation:  "Signed closing speed is derived from observed and estimated positions and is not an official flight-plan or operational Estimated Arrival source.",
		},
	)
	if evidence.currentEndpointPresent {
		result.Provenance.Inputs = append(
			result.Provenance.Inputs,
			projectioncontract.InputReference{
				Name: "current_trajectory_arrival_endpoint",
				Classification: projectioncontract.
					InputClassificationObserved,
				SourceName:  evidence.currentEndpoint.sourceName,
				ObservedAt:  evidence.currentEndpoint.timeValue.UTC(),
				RetrievedAt: validated.generatedAt,
			},
		)
	}
	result.Provenance.Inputs = normalizeInputs(
		result.Provenance.Inputs,
	)
	result.Provenance.LatestInputObservedAt =
		latestInputObservedAt(result.Provenance.Inputs)
	result.Provenance.InputFingerprint = arrivalFingerprint(
		request.Projection,
		request.Route,
		computation,
		evidence.samples,
		estimator.config,
	)
	result.GeneratedAt = validated.generatedAt

	return validateResult(result)
}
