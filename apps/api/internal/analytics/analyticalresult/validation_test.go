package analyticalresult

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func TestResultValidationRejectsInvalidStateCombinations(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	allowed := allowedEligibility(calculatedAt)
	denied := deniedEligibility(calculatedAt)
	confidence := highConfidence()

	testCases := []struct {
		name     string
		result   Result[int]
		expected error
	}{
		{
			name: "unknown status",
			result: Result[int]{
				Status:       Status("unknown"),
				CalculatedAt: calculatedAt,
				Confidence:   NoneConfidence(),
			},
			expected: ErrStatusInvalid,
		},
		{
			name: "missing calculation time",
			result: Result[int]{
				Status:     StatusComplete,
				HasValue:   true,
				Confidence: confidence,
			},
			expected: ErrCalculatedAtMissing,
		},
		{
			name: "complete without value",
			result: Result[int]{
				Status:       StatusComplete,
				Confidence:   confidence,
				CalculatedAt: calculatedAt,
			},
			expected: ErrValueRequired,
		},
		{
			name: "complete without confidence",
			result: Result[int]{
				Status:       StatusComplete,
				HasValue:     true,
				Confidence:   NoneConfidence(),
				CalculatedAt: calculatedAt,
			},
			expected: ErrConfidenceRequired,
		},
		{
			name: "complete with limitation",
			result: Result[int]{
				Status:       StatusComplete,
				HasValue:     true,
				Confidence:   confidence,
				CalculatedAt: calculatedAt,
				Limitations: []Notice{{
					Code:    "coverage_limited",
					Message: "Coverage is limited.",
				}},
			},
			expected: ErrLimitedExplanationRequired,
		},
		{
			name: "limited without explanation",
			result: Result[int]{
				Status:       StatusLimited,
				HasValue:     true,
				Confidence:   confidence,
				CalculatedAt: calculatedAt,
			},
			expected: ErrLimitedExplanationRequired,
		},
		{
			name: "denied with value",
			result: Result[int]{
				Status:       StatusDenied,
				HasValue:     true,
				Confidence:   NoneConfidence(),
				Eligibility:  &denied,
				CalculatedAt: calculatedAt,
			},
			expected: ErrValueForbidden,
		},
		{
			name: "denied without eligibility",
			result: Result[int]{
				Status:       StatusDenied,
				Confidence:   NoneConfidence(),
				CalculatedAt: calculatedAt,
			},
			expected: ErrDeniedEligibilityRequired,
		},
		{
			name: "failed without failure metadata",
			result: Result[int]{
				Status:       StatusFailed,
				Confidence:   NoneConfidence(),
				Eligibility:  &allowed,
				CalculatedAt: calculatedAt,
			},
			expected: ErrFailureRequired,
		},
		{
			name: "failed with denied eligibility",
			result: Result[int]{
				Status:       StatusFailed,
				Confidence:   NoneConfidence(),
				Eligibility:  &denied,
				CalculatedAt: calculatedAt,
				Failure: &Failure{
					Code:    "calculation_failed",
					Message: "Calculation failed.",
				},
			},
			expected: ErrDeniedEligibilityForNonDeniedStatus,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.result.Validate()
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, err)
			}
		})
	}
}

func TestConfidenceValidationRejectsInvalidValues(t *testing.T) {
	testCases := []struct {
		name       string
		confidence Confidence
		expected   error
	}{
		{
			name: "unknown level",
			confidence: Confidence{
				Level: ConfidenceLevel("unknown"),
				Score: 0.5,
			},
			expected: ErrConfidenceLevelInvalid,
		},
		{
			name: "not a number score",
			confidence: Confidence{
				Level: ConfidenceLevelLow,
				Score: math.NaN(),
			},
			expected: ErrConfidenceScoreInvalid,
		},
		{
			name: "score above one",
			confidence: Confidence{
				Level: ConfidenceLevelHigh,
				Score: 1.01,
			},
			expected: ErrConfidenceScoreInvalid,
		},
		{
			name: "none with score",
			confidence: Confidence{
				Level: ConfidenceLevelNone,
				Score: 0.1,
			},
			expected: ErrConfidenceNoneScoreInvalid,
		},
		{
			name: "non-none with zero score",
			confidence: Confidence{
				Level: ConfidenceLevelLow,
				Score: 0,
			},
			expected: ErrConfidenceScoreInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.confidence.Validate()
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, err)
			}
		})
	}
}

func TestEligibilityValidationEnforcesDecisionSemantics(t *testing.T) {
	calculatedAt := analyticalResultTestTime()

	allowedWithReasons := Eligibility{
		Capability: trajectoryeligibility.CapabilityTrafficMetrics,
		Allowed:    true,
		Reasons: []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.ReasonLowQualityScore,
		},
		EvaluatedAt: calculatedAt,
	}
	if !errors.Is(
		allowedWithReasons.Validate(),
		ErrAllowedEligibilityReasonsPresent,
	) {
		t.Fatal("expected allowed eligibility reason rejection")
	}

	deniedWithoutReasons := Eligibility{
		Capability:  trajectoryeligibility.CapabilityRouteInference,
		Allowed:     false,
		EvaluatedAt: calculatedAt,
	}
	if !errors.Is(
		deniedWithoutReasons.Validate(),
		ErrDeniedEligibilityReasonsMissing,
	) {
		t.Fatal("expected denied eligibility reason requirement")
	}
}

func TestSourceValidationEnforcesObservationWindow(t *testing.T) {
	now := analyticalResultTestTime()

	incomplete := Source{
		Name:         "airplanes.live",
		Role:         SourceRoleObservation,
		ObservedFrom: now,
	}
	if !errors.Is(
		incomplete.Validate(),
		ErrSourceObservationRangeIncomplete,
	) {
		t.Fatal("expected incomplete observation range rejection")
	}

	invalid := Source{
		Name:         "airplanes.live",
		Role:         SourceRoleObservation,
		ObservedFrom: now,
		ObservedTo:   now.Add(-time.Minute),
	}
	if !errors.Is(
		invalid.Validate(),
		ErrSourceObservationRangeInvalid,
	) {
		t.Fatal("expected invalid observation range rejection")
	}
}

func TestNoticeValidationRejectsDuplicateCodes(t *testing.T) {
	result := Result[int]{
		Status:       StatusLimited,
		HasValue:     true,
		Confidence:   highConfidence(),
		CalculatedAt: analyticalResultTestTime(),
		Warnings: []Notice{
			{
				Code:    "coverage_partial",
				Message: "First warning.",
			},
			{
				Code:    "coverage_partial",
				Message: "Second warning.",
			},
		},
	}

	if !errors.Is(
		result.Validate(),
		ErrDuplicateNoticeCode,
	) {
		t.Fatal("expected duplicate notice code rejection")
	}
}
