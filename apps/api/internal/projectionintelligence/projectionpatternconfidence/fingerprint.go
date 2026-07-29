package projectionpatternconfidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

const fingerprintPrefix = "sha256:"

func inputFingerprint(
	selection projectionneighbors.Result,
	config Config,
	evidence patternEvidence,
) string {
	digest := sha256.New()

	writeFingerprintString(digest, FingerprintVersion)
	writeFingerprintString(digest, selection.Version)
	writeFingerprintString(digest, string(selection.Status))
	writeFingerprintString(digest, selection.InputFingerprint)
	writeFingerprintBool(digest, selection.CandidateEvaluationTruncated)
	writeFingerprintBool(digest, selection.QualifiedSelectionLimited)

	writeFingerprintInt(digest, len(evidence.neighbors))
	for _, neighbor := range evidence.neighbors {
		writeFingerprintString(digest, neighbor.trajectoryID)
		writeFingerprintFloat(digest, neighbor.similarityScore)
		writeFingerprintString(digest, neighbor.similarityInputFingerprint)
		writeFingerprintFloat(digest, neighbor.anchorDistanceKM)
	}

	writeFingerprintInt(digest, len(evidence.limitations))
	for _, limitation := range evidence.limitations {
		writeFingerprintString(digest, limitation.Code)
		writeFingerprintString(digest, limitation.Message)
	}

	writeFingerprintInt(digest, config.MinimumNeighborCount)
	writeFingerprintInt(digest, config.TargetNeighborCount)
	writeFingerprintFloat(digest, config.MinimumSimilarityScore)
	writeFingerprintFloat(digest, config.MaximumSimilarityStandardDeviation)
	writeFingerprintFloat(digest, config.AnchorDistanceNormalizationKM)
	writeFingerprintFloat(digest, config.MinimumUsableScore)
	writeFingerprintFloat(digest, config.MediumConfidenceMinimum)
	writeFingerprintFloat(digest, config.HighConfidenceMinimum)
	writeFingerprintFloat(digest, config.SimilarityStrengthWeight)
	writeFingerprintFloat(digest, config.SupportWeight)
	writeFingerprintFloat(digest, config.SimilarityConsistencyWeight)
	writeFingerprintFloat(digest, config.AnchorProximityWeight)

	return fingerprintPrefix + hex.EncodeToString(digest.Sum(nil))
}

func writeFingerprintString(digest hash.Hash, value string) {
	_, _ = fmt.Fprintf(digest, "%d:%s|", len(value), value)
}

func writeFingerprintInt(digest hash.Hash, value int) {
	_, _ = fmt.Fprintf(digest, "%d|", value)
}

func writeFingerprintFloat(digest hash.Hash, value float64) {
	_, _ = fmt.Fprintf(digest, "%.17g|", value)
}

func writeFingerprintBool(digest hash.Hash, value bool) {
	if value {
		writeFingerprintString(digest, "true")
		return
	}
	writeFingerprintString(digest, "false")
}
