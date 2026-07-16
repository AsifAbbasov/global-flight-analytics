package scopeenforcement

import (
	"testing"
	"time"
)

func TestEnforceAllowsResearchAnalyticalClaim(t *testing.T) {
	guard := "research_only_not_for_operational_forecast_or_decision_use"
	result, err := Enforce(Request{SubjectID: "forecast-1", DeclaredGuards: []string{guard}, Claims: []Claim{{Code: "stability", Text: "Forecast stability decreased.", Capability: "forecast_stability", Scope: ScopeResearchAnalysis, Strength: StrengthAnalytical, SourceGuard: guard}}, EvaluatedAt: time.Now().UTC()}, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != DecisionAllowed || result.Metrics.AllowedCount != 1 {
		t.Fatalf("result=%#v", result)
	}
}

func TestEnforceBlocksOperationalAndDirectiveClaims(t *testing.T) {
	guard := "research_only_not_for_operational_decision_use"
	result, err := Enforce(Request{SubjectID: "forecast-1", DeclaredGuards: []string{guard}, Claims: []Claim{{Code: "directive", Text: "Change course now.", Capability: "projection", Scope: ScopeAirTrafficControl, Strength: StrengthDirective, SourceGuard: guard}}, EvaluatedAt: time.Now().UTC()}, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != DecisionBlocked || result.Metrics.BlockedCount != 1 {
		t.Fatalf("result=%#v", result)
	}
}

func TestEnforceLimitsCausalResearchClaim(t *testing.T) {
	guard := "research_only_no_pilot_intent_atc_instruction_or_exact_cause_claim"
	result, err := Enforce(Request{SubjectID: "forecast-1", DeclaredGuards: []string{guard}, Claims: []Claim{{Code: "cause", Text: "Weather caused the maneuver.", Capability: "weather_context", Scope: ScopeResearchAnalysis, Strength: StrengthCausal, SourceGuard: guard}}, EvaluatedAt: time.Now().UTC()}, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != DecisionLimited || result.Metrics.LimitedCount != 1 {
		t.Fatalf("result=%#v", result)
	}
}

func TestEnforceDeterministicAcrossClaimOrder(t *testing.T) {
	guard := "research_only_not_for_operational_use"
	claims := []Claim{{Code: "b", Text: "B.", Capability: "b", Scope: ScopeResearchAnalysis, Strength: StrengthDescriptive, SourceGuard: guard}, {Code: "a", Text: "A.", Capability: "a", Scope: ScopeResearchVisualization, Strength: StrengthAnalytical, SourceGuard: guard}}
	request := Request{SubjectID: "subject", DeclaredGuards: []string{guard}, Claims: claims, EvaluatedAt: time.Now().UTC()}
	left, err := Enforce(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	request.Claims[0], request.Claims[1] = request.Claims[1], request.Claims[0]
	right, err := Enforce(request, DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if left.Provenance.InputFingerprint != right.Provenance.InputFingerprint {
		t.Fatal("fingerprint changed")
	}
}
