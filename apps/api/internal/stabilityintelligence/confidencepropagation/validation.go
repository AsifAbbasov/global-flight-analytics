package confidencepropagation

import (
	"fmt"
	"strings"
)

func ValidateResult(result Result, policy Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if result.SchemaVersion != SchemaVersionV1 ||
		(result.Status != ResultStatusLimited &&
			result.Status != ResultStatusComplete) ||
		strings.TrimSpace(result.TargetNodeID) == "" ||
		!unitInterval(result.Score) ||
		result.Level != levelFor(result.Score, policy) ||
		result.ScopeGuard != ScopeGuardResearchOnly ||
		result.EvaluatedAt.IsZero() {
		return fmt.Errorf("%w: identity", ErrInvalidResult)
	}
	if len(result.Nodes) == 0 ||
		len(result.Limitations) == 0 ||
		len(result.Provenance.NodeFingerprints) != len(result.Nodes) ||
		result.Provenance.TargetNodeID != result.TargetNodeID ||
		result.Provenance.PolicyVersion != policy.Version {
		return fmt.Errorf("%w: evidence", ErrInvalidResult)
	}

	targetFound := false
	for _, item := range result.Nodes {
		if item.NodeID == result.TargetNodeID {
			targetFound = true
			if item.Score != result.Score {
				return fmt.Errorf(
					"%w: target score",
					ErrInvalidResult,
				)
			}
		}
		if !unitInterval(item.Score) ||
			!unitInterval(item.DependencyScore) ||
			!unitInterval(item.WeakestRequiredScore) ||
			!knownClassification(item.Classification) {
			return fmt.Errorf(
				"%w: node result",
				ErrInvalidResult,
			)
		}
	}
	if !targetFound {
		return fmt.Errorf("%w: target result", ErrInvalidResult)
	}
	if result.Provenance.InputFingerprint != resultFingerprint(result) {
		return fmt.Errorf("%w: fingerprint", ErrInvalidResult)
	}
	return nil
}
