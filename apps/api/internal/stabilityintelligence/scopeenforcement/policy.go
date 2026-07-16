package scopeenforcement

import "fmt"

const PolicyVersionV1 = "scope-guard-enforcement-policy-v1"

type Policy struct {
	Version           string
	MaximumClaimCount int
	KnownGuards       map[string]bool
}

func DefaultPolicy() Policy {
	return Policy{Version: PolicyVersionV1, MaximumClaimCount: 100, KnownGuards: map[string]bool{
		"research_only_not_for_operational_use":                                   true,
		"research_only_not_for_operational_forecast_or_decision_use":              true,
		"research_only_not_for_operational_decision_use":                          true,
		"research_only_not_for_operational_separation_or_air_traffic_control_use": true,
		"research_only_not_for_operational_failure_or_causal_decision_use":        true,
		"research_only_no_pilot_intent_atc_instruction_or_exact_cause_claim":      true,
		ScopeGuardResearchOnly: true,
	}}
}
func (policy Policy) Validate() error {
	if policy.Version != PolicyVersionV1 || policy.MaximumClaimCount <= 0 || len(policy.KnownGuards) == 0 {
		return fmt.Errorf("invalid scope enforcement policy")
	}
	return nil
}
