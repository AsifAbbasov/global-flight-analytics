package projectionbaseline

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"reflect"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/trajectoryeligibility"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var (
	errProjectionEligibilityDecisionCount = errors.New(
		"projection eligibility must return exactly one projection decision",
	)
	errProjectionEligibilityAllowedReasons = errors.New(
		"allowed projection eligibility decision must not contain denial reasons",
	)
	errProjectionEligibilityDeniedReasons = errors.New(
		"denied projection eligibility decision must contain at least one reason",
	)
	errProjectionEligibilityReasonInvalid = errors.New(
		"projection eligibility decision contains an unknown or duplicate reason",
	)
)

type projectionEligibilityPolicyProvider interface {
	ProjectionPolicyIdentity() trajectoryeligibility.PolicyIdentity
}

func validateProjectionEligibilityEvaluation(
	evaluation trajectoryeligibility.Evaluation,
) (trajectoryeligibility.Decision, error) {
	projectionDecisions := make(
		[]trajectoryeligibility.Decision,
		0,
		1,
	)
	for _, decision := range evaluation.Decisions {
		if decision.Capability == trajectoryeligibility.CapabilityProjection {
			projectionDecisions = append(projectionDecisions, decision)
		}
	}
	if len(projectionDecisions) != 1 {
		return trajectoryeligibility.Decision{}, fmt.Errorf(
			"%w: count=%d",
			errProjectionEligibilityDecisionCount,
			len(projectionDecisions),
		)
	}

	decision := projectionDecisions[0]
	if decision.Allowed && len(decision.Reasons) > 0 {
		return trajectoryeligibility.Decision{}, errProjectionEligibilityAllowedReasons
	}
	if !decision.Allowed && len(decision.Reasons) == 0 {
		return trajectoryeligibility.Decision{}, errProjectionEligibilityDeniedReasons
	}

	seen := make(map[trajectoryeligibility.ReasonCode]struct{}, len(decision.Reasons))
	for _, reason := range decision.Reasons {
		if !reason.IsKnown() {
			return trajectoryeligibility.Decision{}, fmt.Errorf(
				"%w: %q",
				errProjectionEligibilityReasonInvalid,
				reason,
			)
		}
		if _, exists := seen[reason]; exists {
			return trajectoryeligibility.Decision{}, fmt.Errorf(
				"%w: duplicate=%q",
				errProjectionEligibilityReasonInvalid,
				reason,
			)
		}
		seen[reason] = struct{}{}
	}

	decision.Reasons = append(
		[]trajectoryeligibility.ReasonCode(nil),
		decision.Reasons...,
	)
	return decision, nil
}

func eligibilityPolicyIdentity(config Config) trajectoryeligibility.PolicyIdentity {
	if provider, ok := config.EligibilityEvaluator.(projectionEligibilityPolicyProvider); ok {
		identity := provider.ProjectionPolicyIdentity()
		if validEligibilityPolicyIdentity(identity) {
			return identity
		}
	}

	typeName := "unknown"
	if config.EligibilityEvaluator != nil {
		typeName = reflect.TypeOf(config.EligibilityEvaluator).String()
	}
	digest := sha256.Sum256(
		[]byte("projection-baseline-custom-eligibility-v1|" + typeName),
	)
	return trajectoryeligibility.PolicyIdentity{
		Name:        "custom_projection_eligibility",
		Version:     "unversioned",
		Fingerprint: "sha256:" + hex.EncodeToString(digest[:]),
	}
}

func validEligibilityPolicyIdentity(
	identity trajectoryeligibility.PolicyIdentity,
) bool {
	return strings.TrimSpace(identity.Name) == identity.Name &&
		identity.Name != "" &&
		strings.TrimSpace(identity.Version) == identity.Version &&
		identity.Version != "" &&
		len(identity.Fingerprint) == len("sha256:")+64 &&
		strings.HasPrefix(identity.Fingerprint, "sha256:")
}

func eligibilityPolicyInputReference(
	config Config,
	observedAt time.Time,
) projectioncontract.InputReference {
	identity := eligibilityPolicyIdentity(config)
	return projectioncontract.InputReference{
		Name: "eligibility_policy",
		Classification: projectioncontract.
			InputClassificationDerived,
		SourceName: identity.Name + "@" + identity.Version,
		ObservedAt: observedAt.UTC(),
		Limitation: "Eligibility policy fingerprint: " + identity.Fingerprint + ".",
	}
}

func writeEligibilityPolicyFingerprintEvidence(
	digest hash.Hash,
	config Config,
) {
	identity := eligibilityPolicyIdentity(config)
	writeFingerprintString(digest, identity.Name)
	writeFingerprintString(digest, identity.Version)
	writeFingerprintString(digest, identity.Fingerprint)
}
