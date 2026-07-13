package analyticalresult

import (
	"reflect"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
)

func TestCloneCopiesContractOwnedMetadata(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	eligibility := deniedEligibility(calculatedAt)

	original := Result[string]{
		Status:      StatusFailed,
		Confidence:  NoneConfidence(),
		Eligibility: &eligibility,
		Sources: []Source{{
			Name: "airplanes.live",
			Role: SourceRoleObservation,
			Limitations: []Notice{{
				Code:    "provider_rate_limited",
				Message: "Provider rate limit affected coverage.",
			}},
		}},
		Warnings: []Notice{{
			Code:    "partial_input",
			Message: "Only partial input was available.",
		}},
		Limitations: []Notice{{
			Code:    "calculation_unavailable",
			Message: "Calculation could not be completed.",
		}},
		CalculatedAt: calculatedAt,
		Failure: &Failure{
			Code:    "calculation_failed",
			Message: "Calculation failed.",
		},
	}

	clone := original.Clone()
	clone.Eligibility.Reasons[0] =
		trajectoryeligibility.ReasonLowQualityScore
	clone.Sources[0].Limitations[0].Code =
		"mutated_source_limitation"
	clone.Warnings[0].Code = "mutated_warning"
	clone.Limitations[0].Code = "mutated_limitation"
	clone.Failure.Code = "mutated_failure"

	if original.Eligibility.Reasons[0] !=
		trajectoryeligibility.ReasonMissingIdentity {
		t.Fatal("expected eligibility reasons to be copied")
	}
	if original.Sources[0].Limitations[0].Code !=
		"provider_rate_limited" {
		t.Fatal("expected source limitations to be copied")
	}
	if original.Warnings[0].Code != "partial_input" {
		t.Fatal("expected warnings to be copied")
	}
	if original.Limitations[0].Code !=
		"calculation_unavailable" {
		t.Fatal("expected limitations to be copied")
	}
	if original.Failure.Code != "calculation_failed" {
		t.Fatal("expected failure metadata to be copied")
	}
}

func TestConstructorsCopyInputSlices(t *testing.T) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)
	confidence := highConfidence()
	sources := validSources(calculatedAt)

	result, err := NewComplete(
		42,
		confidence,
		&eligibility,
		sources,
		calculatedAt,
	)
	if err != nil {
		t.Fatalf("expected complete result, got %v", err)
	}

	confidence.Reasons[0].Code = "mutated_confidence"
	eligibility.Capability =
		trajectoryeligibility.CapabilityProjection
	sources[0].Name = "mutated-source"

	if result.Confidence.Reasons[0].Code !=
		"source_consistency" {
		t.Fatal("expected confidence reasons to be copied")
	}
	if result.Eligibility.Capability !=
		trajectoryeligibility.CapabilityTrafficMetrics {
		t.Fatal("expected eligibility to be copied")
	}
	if result.Sources[0].Name != "airplanes.live" {
		t.Fatal("expected sources to be copied")
	}
}

func TestStatusAndConfidenceLevelKnowledge(t *testing.T) {
	for _, status := range []Status{
		StatusComplete,
		StatusLimited,
		StatusDenied,
		StatusFailed,
	} {
		if !status.IsKnown() {
			t.Fatalf("expected known status %s", status)
		}
	}

	for _, level := range []ConfidenceLevel{
		ConfidenceLevelNone,
		ConfidenceLevelLow,
		ConfidenceLevelMedium,
		ConfidenceLevelHigh,
	} {
		if !level.IsKnown() {
			t.Fatalf("expected known confidence level %s", level)
		}
	}

	if Status("unknown").IsKnown() {
		t.Fatal("expected unknown status rejection")
	}
	if ConfidenceLevel("unknown").IsKnown() {
		t.Fatal("expected unknown confidence level rejection")
	}
}

func TestClonePreservesResultValue(t *testing.T) {
	original := Result[[]string]{
		Status:       StatusComplete,
		Value:        []string{"one", "two"},
		HasValue:     true,
		Confidence:   highConfidence(),
		CalculatedAt: analyticalResultTestTime(),
	}

	clone := original.Clone()
	if !reflect.DeepEqual(original.Value, clone.Value) {
		t.Fatalf("expected result value assignment copy, got %#v", clone.Value)
	}
}
