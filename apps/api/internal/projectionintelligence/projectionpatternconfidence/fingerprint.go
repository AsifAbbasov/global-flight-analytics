package projectionpatternconfidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

const fingerprintPrefix = "sha256:"

func inputFingerprint(
	selection projectionneighbors.Result,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FingerprintVersion,
	)
	writeFingerprintString(
		digest,
		selection.InputFingerprint,
	)
	writeFingerprintInt(
		digest,
		config.MinimumNeighborCount,
	)
	writeFingerprintInt(
		digest,
		config.TargetNeighborCount,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumCandidateAge,
	)
	writeFingerprintFloat(
		digest,
		config.MaximumMeanAnchorDistanceKM,
	)
	writeFingerprintFloat(
		digest,
		config.MinimumUsableScore,
	)
	writeFingerprintFloat(
		digest,
		config.MediumConfidenceMinimum,
	)
	writeFingerprintFloat(
		digest,
		config.HighConfidenceMinimum,
	)
	writeFingerprintFloat(
		digest,
		config.SimilarityWeight,
	)
	writeFingerprintFloat(
		digest,
		config.SupportWeight,
	)
	writeFingerprintFloat(
		digest,
		config.FreshnessWeight,
	)
	writeFingerprintFloat(
		digest,
		config.AnchorProximityWeight,
	)

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

func writeFingerprintInt(
	digest hash.Hash,
	value int,
) {
	_, _ = fmt.Fprintf(
		digest,
		"%d|",
		value,
	)
}

func writeFingerprintFloat(
	digest hash.Hash,
	value float64,
) {
	_, _ = fmt.Fprintf(
		digest,
		"%.17g|",
		value,
	)
}

func writeFingerprintDuration(
	digest hash.Hash,
	value time.Duration,
) {
	_, _ = fmt.Fprintf(
		digest,
		"%d|",
		value.Nanoseconds(),
	)
}
