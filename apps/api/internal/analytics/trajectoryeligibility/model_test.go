package trajectoryeligibility

import "testing"

func TestCapabilitiesReturnsStableIndependentOrder(
	t *testing.T,
) {
	first := Capabilities()
	second := Capabilities()

	expected := []Capability{
		CapabilityTrafficMetrics,
		CapabilityAirportActivity,
		CapabilityRouteInference,
		CapabilityHistoricalAggregation,
		CapabilityProjection,
	}

	if len(first) != len(expected) {
		t.Fatalf(
			"expected %d capabilities, got %d",
			len(expected),
			len(first),
		)
	}

	for index := range expected {
		if first[index] != expected[index] {
			t.Fatalf(
				"expected capability %s at index %d, got %s",
				expected[index],
				index,
				first[index],
			)
		}
	}

	first[0] = Capability("mutated")

	if second[0] != CapabilityTrafficMetrics {
		t.Fatal("expected capability slices to be independent")
	}
}

func TestEvaluationDecisionReturnsIndependentReasonSlice(
	t *testing.T,
) {
	evaluation := Evaluation{
		Decisions: []Decision{
			{
				Capability: CapabilityRouteInference,
				Allowed:    false,
				Reasons: []ReasonCode{
					ReasonLowQualityScore,
				},
			},
		},
	}

	decision, exists := evaluation.Decision(
		CapabilityRouteInference,
	)
	if !exists {
		t.Fatal("expected route inference decision")
	}

	decision.Reasons[0] =
		ReasonMissingIdentity

	original, exists := evaluation.Decision(
		CapabilityRouteInference,
	)
	if !exists {
		t.Fatal("expected original route inference decision")
	}

	if original.Reasons[0] !=
		ReasonLowQualityScore {
		t.Fatal("expected original reason slice to remain unchanged")
	}
}

func TestPermissionFlagsAllowedRejectsUnknownCapability(
	t *testing.T,
) {
	flags := PermissionFlags{
		AllowTrafficMetrics: true,
	}

	if !flags.Allowed(
		CapabilityTrafficMetrics,
	) {
		t.Fatal("expected traffic metrics permission")
	}

	if flags.Allowed(
		Capability("unknown"),
	) {
		t.Fatal("expected unknown capability to be denied")
	}
}
