package projectionhorizon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"
)

const (
	FingerprintVersion = "projection-horizon-plan-fingerprint-v1"
	fingerprintPrefix  = "sha256:"
)

func calculatePlanFingerprint(plan Plan) string {
	digest := sha256.New()

	writeFingerprintString(digest, FingerprintVersion)
	writeFingerprintString(digest, plan.Version)
	writeFingerprintString(digest, plan.PolicyName)
	writeFingerprintTime(digest, plan.AsOfTime)
	writeFingerprintTime(digest, plan.EndTime)
	writeFingerprintDuration(digest, plan.Step)
	writeFingerprintDuration(digest, plan.RequestedDuration)
	writeFingerprintDuration(digest, plan.EffectiveDuration)
	writeFingerprintBool(digest, plan.Truncated)
	writeFingerprintString(
		digest,
		string(plan.TruncationReason),
	)
	writeFingerprintInt(digest, len(plan.ForecastTimes))
	for _, forecastTime := range plan.ForecastTimes {
		writeFingerprintTime(digest, forecastTime)
	}

	return fingerprintPrefix + hex.EncodeToString(digest.Sum(nil))
}

func writeFingerprintString(digest hash.Hash, value string) {
	_, _ = fmt.Fprintf(
		digest,
		"s:%d:%s|",
		len(value),
		value,
	)
}

func writeFingerprintTime(digest hash.Hash, value time.Time) {
	writeFingerprintString(
		digest,
		value.UTC().Format(time.RFC3339Nano),
	)
}

func writeFingerprintDuration(
	digest hash.Hash,
	value time.Duration,
) {
	_, _ = fmt.Fprintf(
		digest,
		"d:%d|",
		value.Nanoseconds(),
	)
}

func writeFingerprintBool(digest hash.Hash, value bool) {
	_, _ = fmt.Fprintf(digest, "b:%t|", value)
}

func writeFingerprintInt(digest hash.Hash, value int) {
	_, _ = fmt.Fprintf(digest, "i:%d|", value)
}
