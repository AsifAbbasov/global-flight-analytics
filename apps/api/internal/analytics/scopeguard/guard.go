package scopeguard

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var (
	ErrEvaluatorRequired = errors.New("trajectory eligibility evaluator is required")
	ErrCapabilityUnknown = errors.New("analytics capability is unknown")
	ErrDecisionMissing   = errors.New("trajectory eligibility decision is missing")
	ErrOperationRequired = errors.New("analytical operation is required")
	ErrDenied            = errors.New("analytical scope denied")
)

type Clock func() time.Time

type Evaluator interface {
	Evaluate(
		item trajectory.FlightTrajectory,
		now time.Time,
	) trajectoryeligibility.Evaluation
}

type Operation func(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) error

type Config struct {
	Evaluator Evaluator
	Now       Clock
}

type Guard struct {
	evaluator Evaluator
	now       Clock
}

func New(config Config) (*Guard, error) {
	if isNilEvaluator(config.Evaluator) {
		return nil, ErrEvaluatorRequired
	}

	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	return &Guard{
		evaluator: config.Evaluator,
		now:       now,
	}, nil
}

func NewDefault() *Guard {
	guard, err := New(Config{
		Evaluator: trajectoryeligibility.NewDefault(),
	})
	if err != nil {
		panic(fmt.Sprintf("default analytical scope guard is invalid: %v", err))
	}
	return guard
}

func (guard *Guard) Check(
	item trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
) (Decision, error) {
	if err := validateCapability(capability); err != nil {
		return Decision{}, err
	}

	evaluatedAt := guard.now().UTC()
	return guard.checkAt(item, capability, evaluatedAt)
}

func (guard *Guard) Require(
	item trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
) (Decision, error) {
	decision, err := guard.Check(item, capability)
	if err != nil {
		return Decision{}, err
	}
	if !decision.Allowed {
		return decision, newDeniedError(item, decision)
	}
	return decision, nil
}

func (guard *Guard) Run(
	ctx context.Context,
	item trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
	operation Operation,
) (Decision, error) {
	if operation == nil {
		return Decision{}, ErrOperationRequired
	}

	decision, err := guard.Require(item, capability)
	if err != nil {
		return decision, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := operation(ctx, item); err != nil {
		return decision, fmt.Errorf(
			"execute %s analytical operation: %w",
			capability,
			err,
		)
	}

	return decision, nil
}

func (guard *Guard) Filter(
	items []trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
) (FilterResult, error) {
	if err := validateCapability(capability); err != nil {
		return FilterResult{}, err
	}

	evaluatedAt := guard.now().UTC()
	result := FilterResult{
		Allowed:     make([]trajectory.FlightTrajectory, 0, len(items)),
		Denied:      make([]DeniedTrajectory, 0),
		EvaluatedAt: evaluatedAt,
	}

	for _, item := range items {
		decision, err := guard.checkAt(item, capability, evaluatedAt)
		if err != nil {
			return FilterResult{}, err
		}

		if decision.Allowed {
			result.Allowed = append(result.Allowed, item)
			continue
		}

		result.Denied = append(result.Denied, DeniedTrajectory{
			Trajectory: item,
			Decision:   decision,
		})
	}

	return result, nil
}

func (guard *Guard) checkAt(
	item trajectory.FlightTrajectory,
	capability trajectoryeligibility.Capability,
	evaluatedAt time.Time,
) (Decision, error) {
	evaluation := guard.evaluator.Evaluate(item, evaluatedAt)
	eligibilityDecision, exists := evaluation.Decision(capability)
	if !exists {
		return Decision{}, fmt.Errorf("%w: %s", ErrDecisionMissing, capability)
	}

	return Decision{
		Capability: eligibilityDecision.Capability,
		Allowed:    eligibilityDecision.Allowed,
		Reasons: append(
			[]trajectoryeligibility.ReasonCode(nil),
			eligibilityDecision.Reasons...,
		),
		EvaluatedAt: evaluatedAt,
	}, nil
}

func validateCapability(capability trajectoryeligibility.Capability) error {
	switch capability {
	case trajectoryeligibility.CapabilityTrafficMetrics,
		trajectoryeligibility.CapabilityAirportActivity,
		trajectoryeligibility.CapabilityRouteInference,
		trajectoryeligibility.CapabilityHistoricalAggregation,
		trajectoryeligibility.CapabilityProjection:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrCapabilityUnknown, capability)
	}
}

func isNilEvaluator(evaluator Evaluator) bool {
	if evaluator == nil {
		return true
	}

	value := reflect.ValueOf(evaluator)
	switch value.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
