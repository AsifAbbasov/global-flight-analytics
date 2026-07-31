package projectionarrival

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrProjectionContractInvalid = errors.New(
		"projection contract is invalid",
	)
	ErrRouteContractInvalid = errors.New(
		"route contract is invalid",
	)
	ErrCurrentTrajectoryIDRequired = errors.New(
		"current trajectory identifier is required",
	)
	ErrTrajectoryMismatch = errors.New(
		"projection, route, and current trajectory identifiers must match",
	)
	ErrFutureRouteEvidence = errors.New(
		"route evidence as-of time must not exceed projection as-of time",
	)
	ErrGeneratedAtInvalid = errors.New(
		"arrival generated-at time must not precede its inputs",
	)
	ErrArrivalContractInvalid = errors.New(
		"generated arrival projection contract is invalid",
	)
)

type Estimator struct {
	config Config
}

func New(
	config Config,
) (*Estimator, error) {
	config = config.normalized()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate estimated arrival config: %w",
			err,
		)
	}

	return &Estimator{config: config}, nil
}

type Request struct {
	Projection        projectioncontract.Result
	Route             routecontract.Result
	CurrentTrajectory trajectory.FlightTrajectory
	GeneratedAt       time.Time
}

func (
	estimator *Estimator,
) Estimate(
	request Request,
) (projectioncontract.Result, error) {
	if estimator == nil {
		return projectioncontract.Result{},
			ErrArrivalContractInvalid
	}

	validated, err := validateArrivalRequest(request)
	if err != nil {
		return projectioncontract.Result{}, err
	}

	if request.Projection.Status ==
		projectioncontract.ResultStatusUnavailable {
		return estimator.withUnavailableArrival(
			request,
			unavailableProjection,
			"Estimated arrival is unavailable because the position projection is unavailable.",
			positionEvidence{},
		)
	}
	if request.Route.Destination == nil {
		return estimator.withUnavailableArrival(
			request,
			unavailableDestination,
			"Estimated arrival is unavailable because Route Intelligence did not resolve a destination airport.",
			positionEvidence{},
		)
	}

	destination := request.Route.Destination
	if destination.Confidence.Score <
		estimator.config.MinimumDestinationConfidenceScore {
		return estimator.withUnavailableArrival(
			request,
			unavailableDestinationConfidence,
			fmt.Sprintf(
				"Estimated arrival is withheld because destination confidence %.6f is below the configured minimum %.6f.",
				destination.Confidence.Score,
				estimator.config.MinimumDestinationConfidenceScore,
			),
			positionEvidence{},
		)
	}

	evidence := buildPositionEvidence(
		request.CurrentTrajectory,
		request.Projection,
	)
	if len(evidence.samples) == 0 {
		return estimator.withUnavailableArrival(
			request,
			unavailablePositionSamples,
			"Estimated arrival is unavailable because no usable current or projected position samples were available.",
			evidence,
		)
	}

	computation, exists := estimator.computeArrival(
		evidence.samples,
		destination.Airport.Latitude,
		destination.Airport.Longitude,
		request.Projection,
	)
	if !exists {
		return estimator.withUnavailableArrival(
			request,
			unavailableSpeedOrDuration,
			"Estimated arrival is unavailable because directional closing-speed evidence or the complete bounded arrival interval was not usable.",
			evidence,
		)
	}

	return estimator.attachArrivalResult(
		request,
		validated,
		destination,
		computation,
		evidence,
	)
}
