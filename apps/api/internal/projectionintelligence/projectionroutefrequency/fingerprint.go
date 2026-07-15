package projectionroutefrequency

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const fingerprintPrefix = "sha256:"

func routeFrequencyFingerprint(
	route routecontract.Result,
	history HistorySummary,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FingerprintVersion,
	)
	writeFingerprintString(
		digest,
		route.Provenance.InputFingerprint,
	)
	writeFingerprintString(
		digest,
		history.InputFingerprint,
	)
	writeFingerprintString(
		digest,
		history.RouteKey,
	)
	writeFingerprintTime(
		digest,
		history.AsOfTime,
	)
	writeFingerprintInt(
		digest,
		history.ObservationCount,
	)
	writeFingerprintInt(
		digest,
		history.DistinctFlightCount,
	)
	writeFingerprintInt(
		digest,
		history.DistinctDayCount,
	)
	writeFingerprintInt(
		digest,
		history.RecentObservationCount,
	)
	writeFingerprintTime(
		digest,
		history.LastObservedAt,
	)
	writeFingerprintInt(
		digest,
		config.MinimumObservationCount,
	)
	writeFingerprintInt(
		digest,
		config.TargetObservationCount,
	)
	writeFingerprintInt(
		digest,
		config.MinimumDistinctDayCount,
	)
	writeFingerprintInt(
		digest,
		config.TargetDistinctDayCount,
	)
	writeFingerprintDuration(
		digest,
		config.RecentWindow,
	)
	writeFingerprintInt(
		digest,
		config.MinimumRecentObservationCount,
	)
	writeFingerprintInt(
		digest,
		config.TargetRecentObservationCount,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumLatestObservationAge,
	)
	writeFingerprintFloat(
		digest,
		config.MinimumRouteConfidenceScore,
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
		config.ObservationCountWeight,
	)
	writeFingerprintFloat(
		digest,
		config.DistinctDayWeight,
	)
	writeFingerprintFloat(
		digest,
		config.RecentObservationWeight,
	)
	writeFingerprintFloat(
		digest,
		config.LatestObservationWeight,
	)
	writeFingerprintFloat(
		digest,
		config.RouteConfidenceWeight,
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
