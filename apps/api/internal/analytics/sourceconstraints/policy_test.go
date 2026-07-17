package sourceconstraints

import (
	"errors"
	"testing"
)

func TestOpenSkyCapabilityMatrix(t *testing.T) {
	tests := []struct {
		capability Capability
		level      DecisionLevel
		strength   ClaimStrength
	}{
		{CapabilityRegionalLiveObservation, DecisionLevelAllowed, ClaimStrengthObserved},
		{CapabilityHistoricalFlightObservation, DecisionLevelLimited, ClaimStrengthDerived},
		{CapabilityEstimatedAirportContext, DecisionLevelLimited, ClaimStrengthEstimated},
		{CapabilityExperimentalTrackContext, DecisionLevelLimited, ClaimStrengthDerived},
		{CapabilityGlobalContinuousTracking, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityOceanicContinuousTracking, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityOwnReceiverObservation, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityOfficialSchedule, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityOfficialDelayCause, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityPilotIntent, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityATCInstruction, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityCertifiedSeparation, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityOperationalWeather, DecisionLevelBlocked, ClaimStrengthBlocked},
		{CapabilityCommercialFleetData, DecisionLevelBlocked, ClaimStrengthBlocked},
	}

	for _, test := range tests {
		t.Run(string(test.capability), func(t *testing.T) {
			decision, err := Evaluate(Request{
				Constraints: FixedProjectConstraints(),
				Source:      OpenSkyProfile(),
				Capability:  test.capability,
			})
			if err != nil {
				t.Fatalf("evaluate capability: %v", err)
			}
			if decision.Level != test.level {
				t.Fatalf("level = %q, want %q", decision.Level, test.level)
			}
			if decision.MaximumClaimStrength != test.strength {
				t.Fatalf(
					"claim strength = %q, want %q",
					decision.MaximumClaimStrength,
					test.strength,
				)
			}
			if len(decision.ScopeGuards) == 0 {
				t.Fatal("expected scope guards")
			}
		})
	}
}

func TestPaidSourceIsBlockedByFreeOnlyBoundary(t *testing.T) {
	source := OpenSkyProfile()
	source.ID = "paid-provider"
	source.FreeAccess = false
	source.Commercial = true

	decision, err := Evaluate(Request{
		Constraints: FixedProjectConstraints(),
		Source:      source,
		Capability:  CapabilityRegionalLiveObservation,
	})
	if err != nil {
		t.Fatalf("evaluate capability: %v", err)
	}
	if decision.Level != DecisionLevelBlocked {
		t.Fatalf("level = %q, want blocked", decision.Level)
	}
}

func TestOwnInfrastructureCapabilityRemainsBlocked(t *testing.T) {
	decision, err := Evaluate(Request{
		Constraints: FixedProjectConstraints(),
		Source:      OpenSkyProfile(),
		Capability:  CapabilityOwnReceiverObservation,
	})
	if err != nil {
		t.Fatalf("evaluate capability: %v", err)
	}
	if decision.Usable() {
		t.Fatal("own receiver capability must not be usable")
	}
}

func TestModifiedConstraintSetIsRejected(t *testing.T) {
	constraints := FixedProjectConstraints()
	constraints.HasSatelliteAccess = true

	_, err := Evaluate(Request{
		Constraints: constraints,
		Source:      OpenSkyProfile(),
		Capability:  CapabilityRegionalLiveObservation,
	})
	if !errors.Is(err, ErrConstraintSetInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrConstraintSetInvalid)
	}
}

func TestOpenSkyUsableCapabilitiesCarryAttributionAndUsageObligations(t *testing.T) {
	decision, err := Evaluate(Request{
		Constraints: FixedProjectConstraints(),
		Source:      OpenSkyProfile(),
		Capability:  CapabilityRegionalLiveObservation,
	})
	if err != nil {
		t.Fatalf("evaluate capability: %v", err)
	}

	for _, expected := range []string{
		"OpenSky Network attribution required",
		"non-commercial research use only",
	} {
		if !containsString(decision.RequiredLabels, expected) {
			t.Fatalf("required labels %v do not contain %q", decision.RequiredLabels, expected)
		}
	}

	if !containsString(
		decision.ScopeGuards,
		"Treat provider access from large cloud-hosting IP ranges as non-guaranteed and retain fallback behavior.",
	) {
		t.Fatalf("scope guards do not contain cloud-hosting fallback obligation: %v", decision.ScopeGuards)
	}
}

func TestAttributionRequirementNeedsCitationText(t *testing.T) {
	source := OpenSkyProfile()
	source.AttributionText = ""

	_, err := Evaluate(Request{
		Constraints: FixedProjectConstraints(),
		Source:      source,
		Capability:  CapabilityRegionalLiveObservation,
	})
	if !errors.Is(err, ErrAttributionTextRequired) {
		t.Fatalf("error = %v, want %v", err, ErrAttributionTextRequired)
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
