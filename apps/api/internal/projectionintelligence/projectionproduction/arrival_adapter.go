package projectionproduction

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionarrival"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var (
	ErrArrivalProjectionEstimatorRequired = errors.New("arrival projection estimator is required")
	ErrArrivalProjectionMutation          = errors.New("arrival estimator changed the authorized position projection")
	ErrArrivalOutcomeInvalid              = errors.New("arrival outcome is invalid")
)

type ArrivalOutcomeStatus string

const (
	ArrivalOutcomeAttached ArrivalOutcomeStatus = "attached"
	ArrivalOutcomeWithheld ArrivalOutcomeStatus = "withheld"
)

type ArrivalOutcome struct {
	Status   ArrivalOutcomeStatus
	Estimate *projectioncontract.ArrivalEstimate
	Notices  []Notice
}

func (outcome ArrivalOutcome) Clone() ArrivalOutcome {
	cloned := outcome
	cloned.Estimate = cloneArrivalEstimate(outcome.Estimate)
	cloned.Notices = append([]Notice(nil), outcome.Notices...)
	return cloned
}

func (outcome ArrivalOutcome) Validate() error {
	switch outcome.Status {
	case ArrivalOutcomeAttached:
		if outcome.Estimate == nil {
			return fmt.Errorf("%w: attached status requires an estimate", ErrArrivalOutcomeInvalid)
		}
	case ArrivalOutcomeWithheld:
		if outcome.Estimate != nil {
			return fmt.Errorf("%w: withheld status must not contain an estimate", ErrArrivalOutcomeInvalid)
		}
	default:
		return fmt.Errorf("%w: unknown status %q", ErrArrivalOutcomeInvalid, outcome.Status)
	}
	for _, notice := range outcome.Notices {
		if notice.Code == "" || notice.Message == "" {
			return fmt.Errorf("%w: notice is invalid", ErrArrivalOutcomeInvalid)
		}
	}
	return nil
}

type ArrivalProjectionEstimator interface {
	Estimate(projectionarrival.Request) (projectioncontract.Result, error)
}

type ArrivalAdapter struct {
	estimator ArrivalProjectionEstimator
}

func NewArrivalAdapter(estimator ArrivalProjectionEstimator) (*ArrivalAdapter, error) {
	if estimator == nil {
		return nil, ErrArrivalProjectionEstimatorRequired
	}
	return &ArrivalAdapter{estimator: estimator}, nil
}

func (adapter *ArrivalAdapter) EstimateArrival(
	request projectionarrival.Request,
) (ArrivalOutcome, error) {
	if adapter == nil || adapter.estimator == nil {
		return ArrivalOutcome{}, ErrArrivalProjectionEstimatorRequired
	}

	before := request.Projection.Clone()
	result, err := adapter.estimator.Estimate(request)
	if err != nil {
		return ArrivalOutcome{}, err
	}
	if !sameProjectionExceptArrival(before, result) {
		return ArrivalOutcome{}, ErrArrivalProjectionMutation
	}

	outcome := ArrivalOutcome{Status: ArrivalOutcomeWithheld}
	if result.Arrival != nil {
		outcome.Status = ArrivalOutcomeAttached
		outcome.Estimate = cloneArrivalEstimate(result.Arrival)
	}
	if err := outcome.Validate(); err != nil {
		return ArrivalOutcome{}, err
	}
	return outcome.Clone(), nil
}

func sameProjectionExceptArrival(
	left projectioncontract.Result,
	right projectioncontract.Result,
) bool {
	left = left.Clone()
	right = right.Clone()
	left.Arrival = nil
	right.Arrival = nil
	return reflect.DeepEqual(left, right)
}

func cloneArrivalEstimate(
	estimate *projectioncontract.ArrivalEstimate,
) *projectioncontract.ArrivalEstimate {
	if estimate == nil {
		return nil
	}
	return projectioncontract.Result{Arrival: estimate}.Clone().Arrival
}
