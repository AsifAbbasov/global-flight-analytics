package projectionproduction

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"
)

const fingerprintPrefix = "sha256:"

func productionFingerprint(
	result Result,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FingerprintVersion,
	)
	writeFingerprintString(
		digest,
		string(result.Strategy),
	)
	writeFingerprintString(
		digest,
		result.FallbackReason,
	)
	writeFingerprintString(
		digest,
		string(result.ArrivalStatus),
	)
	writeFingerprintString(
		digest,
		result.Projection.Provenance.
			InputFingerprint,
	)
	writeFingerprintString(
		digest,
		string(config.FreshnessLimitedPolicy),
	)
	writeFingerprintString(
		digest,
		string(config.RouteFrequencyLimitedPolicy),
	)
	writeFingerprintString(
		digest,
		string(config.DependencyFailurePolicy),
	)
	writeFingerprintString(
		digest,
		string(config.ArrivalFailurePolicy),
	)
	writeFingerprintTime(
		digest,
		result.GeneratedAt,
	)

	if result.NeighborSelection != nil {
		writeFingerprintString(
			digest,
			result.NeighborSelection.
				InputFingerprint,
		)
	} else {
		writeFingerprintString(
			digest,
			"neighbor-selection:nil",
		)
	}
	if result.PatternConfidence != nil {
		writeFingerprintString(
			digest,
			result.PatternConfidence.
				InputFingerprint,
		)
	} else {
		writeFingerprintString(
			digest,
			"pattern-confidence:nil",
		)
	}
	if result.Freshness != nil {
		writeFingerprintString(
			digest,
			result.Freshness.InputFingerprint,
		)
	} else {
		writeFingerprintString(
			digest,
			"freshness:nil",
		)
	}
	if result.RouteFrequency != nil {
		writeFingerprintString(
			digest,
			result.RouteFrequency.
				InputFingerprint,
		)
	} else {
		writeFingerprintString(
			digest,
			"route-frequency:nil",
		)
	}

	return fingerprintPrefix +
		hex.EncodeToString(
			digest.Sum(nil),
		)
}

func writeFingerprintString(
	digest hash.Hash,
	value string,
) {
	_, _ = fmt.Fprintf(
		digest,
		"%d:%s|",
		len(value),
		value,
	)
}

func writeFingerprintTime(
	digest hash.Hash,
	value time.Time,
) {
	writeFingerprintString(
		digest,
		value.UTC().Format(
			time.RFC3339Nano,
		),
	)
}
