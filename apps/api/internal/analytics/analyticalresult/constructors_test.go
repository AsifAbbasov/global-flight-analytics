package analyticalresult

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/scopeguard"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func TestNewCompleteCreatesUsableResult(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)

	result, err := NewComplete(
		42,
		highConfidence(),
		&eligibility,
		validSources(calculatedAt),
		calculatedAt,
	)
	if err != nil {
		t.Fatalf("expected complete result, got %v", err)
	}

	if result.Status != StatusComplete || !result.IsUsable() {
		t.Fatalf("expected usable complete result, got %#v", result)
	}
	if !result.HasValue || result.Value != 42 || result.ValueOrZero() != 42 {
		t.Fatalf("expected value 42, got %#v", result)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("expected valid complete result, got %v", err)
	}
}

func TestNewLimitedRequiresExplanation(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)

	_, err := NewLimited(
		42,
		mediumConfidence(),
		&eligibility,
		validSources(calculatedAt),
		nil,
		nil,
		calculatedAt,
	)
	if !errors.Is(err, ErrLimitedExplanationRequired) {
		t.Fatalf("expected limited explanation error, got %v", err)
	}
}

func TestNewLimitedCreatesUsableResult(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)

	result, err := NewLimited(
		42,
		mediumConfidence(),
		&eligibility,
		validSources(calculatedAt),
		[]Notice{{
			Code:    "coverage_partial",
			Message: "Coverage is incomplete near the region boundary.",
		}},
		[]Notice{{
			Code:    "historical_depth_limited",
			Message: "Historical depth is limited to the retained observation window.",
		}},
		calculatedAt,
	)
	if err != nil {
		t.Fatalf("expected limited result, got %v", err)
	}

	if result.Status != StatusLimited || !result.IsUsable() {
		t.Fatalf("expected usable limited result, got %#v", result)
	}
}

func TestNewDeniedUsesScopeDecision(t *testing.T) {
	evaluatedAt := analyticalResultTestTime()
	decision := scopeguard.Decision{
		Capability: trajectoryeligibility.CapabilityRouteInference,
		Allowed:    false,
		Reasons: []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.ReasonMissingIdentity,
			trajectoryeligibility.ReasonLowQualityScore,
		},
		EvaluatedAt: evaluatedAt,
	}

	result, err := NewDenied[string](
		decision,
		validSources(evaluatedAt),
	)
	if err != nil {
		t.Fatalf("expected denied result, got %v", err)
	}

	if result.Status != StatusDenied || result.IsUsable() || result.HasValue {
		t.Fatalf("expected unusable denied result, got %#v", result)
	}
	if result.ValueOrZero() != "" {
		t.Fatalf("expected zero string value, got %q", result.ValueOrZero())
	}
	if result.Eligibility == nil || result.Eligibility.Allowed {
		t.Fatalf("expected denied eligibility, got %#v", result.Eligibility)
	}
	if len(result.Eligibility.Reasons) != 2 {
		t.Fatalf("expected two denial reasons, got %v", result.Eligibility.Reasons)
	}
}

func TestNewDeniedRejectsAllowedScopeDecision(t *testing.T) {
	decision := scopeguard.Decision{
		Capability:  trajectoryeligibility.CapabilityTrafficMetrics,
		Allowed:     true,
		EvaluatedAt: analyticalResultTestTime(),
	}

	_, err := NewDenied[int](decision, nil)
	if !errors.Is(err, ErrDeniedEligibilityRequired) {
		t.Fatalf("expected denied eligibility error, got %v", err)
	}
}

func TestNewFailedCreatesTypedFailure(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)

	result, err := NewFailed[int](
		Failure{
			Code:      "calculation_failed",
			Message:   "The calculator returned an unexpected error.",
			Retriable: true,
		},
		&eligibility,
		validSources(calculatedAt),
		nil,
		nil,
		calculatedAt,
	)
	if err != nil {
		t.Fatalf("expected failed result, got %v", err)
	}

	if result.Status != StatusFailed || result.IsUsable() || result.HasValue {
		t.Fatalf("expected failed result without value, got %#v", result)
	}
	if result.Failure == nil || !result.Failure.Retriable {
		t.Fatalf("expected retriable failure, got %#v", result.Failure)
	}
}

func TestEligibilityFromScopeDecisionCopiesReasons(t *testing.T) {
	decision := scopeguard.Decision{
		Capability: trajectoryeligibility.CapabilityProjection,
		Allowed:    false,
		Reasons: []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.ReasonMissingAltitude,
		},
		EvaluatedAt: analyticalResultTestTime().In(
			time.FixedZone("test", 4*60*60),
		),
	}

	eligibility := EligibilityFromScopeDecision(decision)
	decision.Reasons[0] = trajectoryeligibility.ReasonLowQualityScore

	if eligibility.Reasons[0] != trajectoryeligibility.ReasonMissingAltitude {
		t.Fatal("expected copied eligibility reasons")
	}
	if eligibility.EvaluatedAt.Location() != time.UTC {
		t.Fatalf("expected UTC evaluation time, got %s", eligibility.EvaluatedAt.Location())
	}
}
