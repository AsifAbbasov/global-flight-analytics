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

	writeFingerprintBool(digest, evidence.continuation.known)
	writeFingerprintInt(digest, evidence.continuation.sampleCount)
	writeFingerprintInt(digest, evidence.continuation.pairCount)
	writeFingerprintInt(digest, evidence.continuation.comparisonCount)
	writeFingerprintFloat(digest, evidence.continuation.horizonSeconds)
	writeFingerprintInt(digest, len(evidence.continuation.vectors))
	for _, vector := range evidence.continuation.vectors {
		writeFingerprintString(digest, vector.trajectoryID)
		writeFingerprintInt(digest, vector.sampleIndex)
		writeFingerprintFloat(digest, vector.elapsedS)
		writeFingerprintFloat(digest, vector.anchorLatitude)
		writeFingerprintFloat(digest, vector.anchorLongitude)
		writeFingerprintFloat(digest, vector.endpointLatitude)
		writeFingerprintFloat(digest, vector.endpointLongitude)
		writeFingerprintFloat(digest, vector.eastM)
		writeFingerprintFloat(digest, vector.northM)
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
	writeFingerprintInt(digest, config.ContinuationAgreementSampleCount)
	writeFingerprintFloat(digest, config.ContinuationDivergenceNormalizationMPS)
	writeFingerprintFloat(digest, config.MaximumContinuationDivergenceMPS)
	writeFingerprintFloat(digest, config.SimilarityStrengthWeight)
	writeFingerprintFloat(digest, config.SupportWeight)
	writeFingerprintFloat(digest, config.SimilarityConsistencyWeight)
	writeFingerprintFloat(digest, config.AnchorProximityWeight)
	writeFingerprintFloat(digest, config.ContinuationAgreementWeight)

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
