package projectionfreshness

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

const fingerprintPrefix = "sha256:"

func freshnessFingerprint(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
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
	writeFingerprintString(
		digest,
		pattern.InputFingerprint,
	)
	writeFingerprintTime(
		digest,
		selection.AsOfTime,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumNewestNeighborAge,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumMeanNeighborAge,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumOldestNeighborAge,
	)
	writeFingerprintDuration(
		digest,
		config.RecentNeighborAgeLimit,
	)
	writeFingerprintInt(
		digest,
		config.MinimumRecentNeighborCount,
	)
	writeFingerprintInt(
		digest,
		config.TargetRecentNeighborCount,
	)
	writeFingerprintFloat(
		digest,
		config.MinimumUsableScore,
	)
	writeFingerprintFloat(
		digest,
		config.CompleteScoreMinimum,
	)
	writeFingerprintFloat(
		digest,
		config.NewestAgeWeight,
	)
	writeFingerprintFloat(
		digest,
		config.MeanAgeWeight,
	)
	writeFingerprintFloat(
		digest,
		config.OldestAgeWeight,
	)
	writeFingerprintFloat(
		digest,
		config.RecentSupportWeight,
	)

	neighbors := append(
		[]projectionneighbors.Neighbor(nil),
		selection.Neighbors...,
	)
	sort.SliceStable(
		neighbors,
		func(left int, right int) bool {
			return neighbors[left].TrajectoryID <
				neighbors[right].TrajectoryID
		},
	)
	for _, neighbor := range neighbors {
		writeFingerprintString(
			digest,
			neighbor.TrajectoryID,
		)
		writeFingerprintDuration(
			digest,
			neighbor.CandidateAge,
		)
		writeFingerprintTime(
			digest,
			neighbor.CandidateEndTime,
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
