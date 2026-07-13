package analyticalresult

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func EligibilityFromScopeDecision(
	decision scopeguard.Decision,
) Eligibility {
	return Eligibility{
		Capability: decision.Capability,
		Allowed:    decision.Allowed,
		Reasons: append(
			[]trajectoryeligibility.ReasonCode(nil),
			decision.Reasons...,
		),
		EvaluatedAt: decision.EvaluatedAt.UTC(),
	}
}

func NewComplete[T any](
	value T,
	confidence Confidence,
	eligibility *Eligibility,
	sources []Source,
	calculatedAt time.Time,
) (Result[T], error) {
	result := Result[T]{
		Status:       StatusComplete,
		Value:        value,
		HasValue:     true,
		Confidence:   cloneConfidence(confidence),
		Eligibility:  cloneEligibility(eligibility),
		Sources:      cloneSources(sources),
		CalculatedAt: calculatedAt.UTC(),
	}

	if err := result.Validate(); err != nil {
		return Result[T]{}, err
	}

	return result, nil
}

func NewLimited[T any](
	value T,
	confidence Confidence,
	eligibility *Eligibility,
	sources []Source,
	warnings []Notice,
	limitations []Notice,
	calculatedAt time.Time,
) (Result[T], error) {
	result := Result[T]{
		Status:       StatusLimited,
		Value:        value,
		HasValue:     true,
		Confidence:   cloneConfidence(confidence),
		Eligibility:  cloneEligibility(eligibility),
		Sources:      cloneSources(sources),
		Warnings:     cloneNotices(warnings),
		Limitations:  cloneNotices(limitations),
		CalculatedAt: calculatedAt.UTC(),
	}

	if err := result.Validate(); err != nil {
		return Result[T]{}, err
	}

	return result, nil
}

func NewDenied[T any](
	decision scopeguard.Decision,
	sources []Source,
) (Result[T], error) {
	eligibility := EligibilityFromScopeDecision(
		decision,
	)

	result := Result[T]{
		Status:       StatusDenied,
		Confidence:   NoneConfidence(),
		Eligibility:  &eligibility,
		Sources:      cloneSources(sources),
		CalculatedAt: decision.EvaluatedAt.UTC(),
	}

	if err := result.Validate(); err != nil {
		return Result[T]{}, err
	}

	return result, nil
}

func NewFailed[T any](
	failure Failure,
	eligibility *Eligibility,
	sources []Source,
	warnings []Notice,
	limitations []Notice,
	calculatedAt time.Time,
) (Result[T], error) {
	failureCopy := failure

	result := Result[T]{
		Status:       StatusFailed,
		Confidence:   NoneConfidence(),
		Eligibility:  cloneEligibility(eligibility),
		Sources:      cloneSources(sources),
		Warnings:     cloneNotices(warnings),
		Limitations:  cloneNotices(limitations),
		CalculatedAt: calculatedAt.UTC(),
		Failure:      &failureCopy,
	}

	if err := result.Validate(); err != nil {
		return Result[T]{}, err
	}

	return result, nil
}
