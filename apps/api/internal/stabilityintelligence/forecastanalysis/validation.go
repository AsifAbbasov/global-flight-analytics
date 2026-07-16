package forecastanalysis

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
		strings.TrimSpace(result.TrajectoryID) == "" ||
		result.ScopeGuard != ScopeGuardResearchOnly ||
		result.EvaluatedAt.IsZero() {
		return fmt.Errorf("%w: identity", ErrInvalidResult)
	}
	if !knownTrend(result.Trend) || !knownHealth(result.Health) {
		return fmt.Errorf("%w: classification", ErrInvalidResult)
	}
	if result.Metrics.VersionCount < 2 ||
		result.Metrics.TransitionCount != len(result.Transitions) ||
		result.Metrics.TransitionCount != result.Metrics.VersionCount-1 ||
		!unitInterval(result.Metrics.StableTransitionShare) ||
		!unitInterval(result.Metrics.ComparableTransitionShare) ||
		!unitInterval(result.Metrics.MaterialChangeShare) ||
		!unitInterval(result.Metrics.MeanStabilityScore) ||
		!unitInterval(result.Confidence.Score) {
		return fmt.Errorf("%w: metrics", ErrInvalidResult)
	}
	if len(result.Limitations) == 0 ||
		len(result.Explanations) == 0 ||
		len(result.Provenance.VersionIDs) != result.Metrics.VersionCount ||
		len(result.Provenance.OutputFingerprints) != result.Metrics.VersionCount ||
		result.Provenance.PolicyVersion != policy.Version {
		return fmt.Errorf("%w: evidence", ErrInvalidResult)
	}
	if result.Provenance.InputFingerprint != resultFingerprint(result) {
		return fmt.Errorf("%w: fingerprint", ErrInvalidResult)
	}
	return nil
}

func knownTrend(value Trend) bool {
	switch value {
	case TrendInsufficient,
		TrendSteady,
		TrendImproving,
		TrendDegrading,
		TrendVolatile:
		return true
	default:
		return false
	}
}

func knownHealth(value Health) bool {
	switch value {
	case HealthInsufficient,
		HealthStable,
		HealthWatch,
		HealthUnstable:
		return true
	default:
		return false
	}
}
