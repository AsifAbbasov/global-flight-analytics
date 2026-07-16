package failureexplanation

import (
	"fmt"
	"strings"
)

func ValidateResult(result Result, policy Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if result.SchemaVersion != SchemaVersionV1 ||
		(result.Status != ResultStatusLimited && result.Status != ResultStatusComplete) ||
		strings.TrimSpace(result.SubjectID) == "" || strings.TrimSpace(result.SubjectType) == "" ||
		result.ScopeGuard != ScopeGuardResearchOnly || result.EvaluatedAt.IsZero() ||
		result.Provenance.PolicyVersion != policy.Version ||
		!strings.HasPrefix(result.Provenance.InputFingerprint, "sha256:") ||
		result.Provenance.InputFingerprint != resultFingerprint(result) ||
		!unitInterval(result.Confidence.Score) || result.Confidence.Level == "" ||
		result.Metrics.FailureCount != len(result.Failures) ||
		result.Metrics.SignalCount != len(result.Provenance.SignalFingerprints) {
		return fmt.Errorf("invalid failure explanation result")
	}
	if len(result.Failures) > 0 && result.PrimaryCode != result.Failures[0].Code {
		return fmt.Errorf("invalid primary failure")
	}
	for index, failure := range result.Failures {
		if failure.Rank != index+1 || failure.Code == "" || failure.Summary == "" || failure.Source == "" ||
			!failure.Category.IsKnown() || !failure.Severity.IsKnown() || !failure.Classification.IsKnown() ||
			!unitInterval(failure.PriorityScore) {
			return fmt.Errorf("invalid failure at index %d", index)
		}
		if index > 0 && result.Failures[index-1].PriorityScore < failure.PriorityScore {
			return fmt.Errorf("failure ranking is not descending")
		}
	}
	return nil
}
