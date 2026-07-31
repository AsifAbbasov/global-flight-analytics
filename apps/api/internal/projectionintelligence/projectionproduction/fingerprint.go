package projectionproduction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const fingerprintPrefix = "sha256:"

func requestInputFingerprint(
	request Request,
	planFingerprint string,
	config Config,
) string {
	snapshot := struct {
		Version                     string
		Request                     Request
		PlanFingerprint             string
		FreshnessLimitedPolicy      LimitedEvidencePolicy
		RouteFrequencyLimitedPolicy LimitedEvidencePolicy
		DependencyFailurePolicy     DependencyFailurePolicy
		ArrivalFailurePolicy        ArrivalFailurePolicy
	}{
		Version:                     RequestFingerprintVersion,
		Request:                     request.Clone(),
		PlanFingerprint:             planFingerprint,
		FreshnessLimitedPolicy:      config.FreshnessLimitedPolicy,
		RouteFrequencyLimitedPolicy: config.RouteFrequencyLimitedPolicy,
		DependencyFailurePolicy:     config.DependencyFailurePolicy,
		ArrivalFailurePolicy:        config.ArrivalFailurePolicy,
	}
	return hashJSON(snapshot)
}

func compositionFingerprint(result Result) string {
	cloned := result.Clone()
	cloned.CompositionFingerprint = ""
	return hashJSON(struct {
		Version string
		Result  Result
	}{
		Version: CompositionFingerprintVersion,
		Result:  cloned,
	})
}

func hashJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(encoded)
	return fingerprintPrefix + hex.EncodeToString(digest[:])
}
