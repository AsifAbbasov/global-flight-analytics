package trajectoryeligibility

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	ProjectionPolicyName    = "trajectory_projection_eligibility"
	ProjectionPolicyVersion = "trajectory-projection-eligibility-v1"
)

type PolicyIdentity struct {
	Name        string
	Version     string
	Fingerprint string
}

func (evaluator *Evaluator) ProjectionPolicyIdentity() PolicyIdentity {
	if evaluator == nil {
		return PolicyIdentity{}
	}

	policy := evaluator.config.Projection
	digest := sha256.Sum256(
		[]byte(
			fmt.Sprintf(
				"%s|%s|%d|%.17g|%d|%d|%d|%d|%d|%d|%t|%t|%t",
				ProjectionPolicyName,
				ProjectionPolicyVersion,
				policy.MinimumPointCount,
				policy.MinimumQualityScore,
				policy.MaximumCoverageGapCount,
				policy.MinimumDuration.Nanoseconds(),
				policy.MaximumDuration.Nanoseconds(),
				policy.MaximumObservationAge.Nanoseconds(),
				policy.MaximumFutureObservationSkew.Nanoseconds(),
				policy.MaximumRecentPointGap.Nanoseconds(),
				policy.RequireReliableIdentity,
				policy.RequireCallsign,
				policy.RequireAltitude,
			),
		),
	)
	return PolicyIdentity{
		Name:        ProjectionPolicyName,
		Version:     ProjectionPolicyVersion,
		Fingerprint: "sha256:" + hex.EncodeToString(digest[:]),
	}
}
