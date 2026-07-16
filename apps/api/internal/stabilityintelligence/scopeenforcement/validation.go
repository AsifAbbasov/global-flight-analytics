package scopeenforcement

import (
	"fmt"
	"strings"
)

func ValidateResult(result Result, policy Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if result.SchemaVersion != SchemaVersionV1 || (result.Status != ResultStatusLimited && result.Status != ResultStatusComplete) || result.SubjectID == "" ||
		(result.Decision != DecisionAllowed && result.Decision != DecisionLimited && result.Decision != DecisionBlocked) || result.ScopeGuard != ScopeGuardResearchOnly || result.EvaluatedAt.IsZero() ||
		result.Provenance.PolicyVersion != policy.Version || !strings.HasPrefix(result.Provenance.InputFingerprint, "sha256:") || result.Provenance.InputFingerprint != resultFingerprint(result) ||
		result.Metrics.ClaimCount != len(result.Claims) || result.Metrics.GuardCount != len(result.Provenance.DeclaredGuards) {
		return fmt.Errorf("invalid scope enforcement result")
	}
	if result.Decision == DecisionAllowed && result.Status != ResultStatusComplete {
		return fmt.Errorf("allowed result must be complete")
	}
	if result.Decision != DecisionAllowed && result.Status != ResultStatusLimited {
		return fmt.Errorf("limited or blocked result must be limited")
	}
	return nil
}
