package unknownintervention

import (
	"fmt"
	"strings"
)

func ValidateResult(result Result, policy Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if result.SchemaVersion != SchemaVersionV1 || (result.Status != ResultStatusLimited && result.Status != ResultStatusComplete) || result.SubjectID == "" || !result.ClaimKind.IsKnown() ||
		(result.Decision != DecisionAllowedContextOnly && result.Decision != DecisionLimitedContext && result.Decision != DecisionWithheld) || !unit(result.ConfidenceScore) ||
		result.ScopeGuard != ScopeGuardResearchOnly || result.EvaluatedAt.IsZero() || result.Provenance.PolicyVersion != policy.Version || !strings.HasPrefix(result.Provenance.InputFingerprint, "sha256:") || result.Provenance.InputFingerprint != resultFingerprint(result) || result.Metrics.EvidenceCount != len(result.Provenance.EvidenceFingerprints) {
		return fmt.Errorf("invalid unknown intervention result")
	}
	if result.Decision == DecisionAllowedContextOnly && result.Status != ResultStatusComplete {
		return fmt.Errorf("allowed context result must be complete")
	}
	if result.Decision != DecisionAllowedContextOnly && result.Status != ResultStatusLimited {
		return fmt.Errorf("limited or withheld result must be limited")
	}
	return nil
}
