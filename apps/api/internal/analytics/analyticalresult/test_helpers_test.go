package analyticalresult

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func analyticalResultTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		13,
		18,
		0,
		0,
		0,
		time.UTC,
	)
}

func highConfidence() Confidence {
	return Confidence{
		Level: ConfidenceLevelHigh,
		Score: 0.95,
		Reasons: []Notice{{
			Code:    "source_consistency",
			Message: "Independent observations are internally consistent.",
		}},
	}
}

func mediumConfidence() Confidence {
	return Confidence{
		Level: ConfidenceLevelMedium,
		Score: 0.65,
		Reasons: []Notice{{
			Code:    "coverage_partial",
			Message: "Coverage is sufficient but incomplete.",
		}},
	}
}

func allowedEligibility(
	evaluatedAt time.Time,
) Eligibility {
	return Eligibility{
		Capability: trajectoryeligibility.
			CapabilityTrafficMetrics,
		Allowed:     true,
		EvaluatedAt: evaluatedAt,
	}
}

func deniedEligibility(
	evaluatedAt time.Time,
) Eligibility {
	return Eligibility{
		Capability: trajectoryeligibility.
			CapabilityRouteInference,
		Allowed: false,
		Reasons: []trajectoryeligibility.ReasonCode{
			trajectoryeligibility.ReasonMissingIdentity,
		},
		EvaluatedAt: evaluatedAt,
	}
}

func validSources(
	calculatedAt time.Time,
) []Source {
	return []Source{{
		Name:         "airplanes.live",
		Role:         SourceRoleObservation,
		ObservedFrom: calculatedAt.Add(-5 * time.Minute),
		ObservedTo:   calculatedAt.Add(-time.Minute),
		RetrievedAt:  calculatedAt,
	}}
}
