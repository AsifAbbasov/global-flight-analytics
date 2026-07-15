package projectionarrival

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const fingerprintPrefix = "sha256:"

func arrivalFingerprint(
	projection projectioncontract.Result,
	route routecontract.Result,
	computation arrivalComputation,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FingerprintVersion,
	)
	writeFingerprintString(
		digest,
		projection.Provenance.
			InputFingerprint,
	)
	writeFingerprintString(
		digest,
		route.Provenance.
			InputFingerprint,
	)
	writeFingerprintString(
		digest,
		route.Destination.
			Airport.ICAOCode,
	)
	writeFingerprintString(
		digest,
		string(computation.mode),
	)
	writeFingerprintTime(
		digest,
		computation.earliestTime,
	)
	writeFingerprintTime(
		digest,
		computation.estimatedTime,
	)
	writeFingerprintTime(
		digest,
		computation.latestTime,
	)
	writeFingerprintFloat(
		digest,
		computation.
			estimatedGroundSpeedMPS,
	)
	writeFingerprintFloat(
		digest,
		computation.
			groundSpeedStdDevMPS,
	)
	writeFingerprintInt(
		digest,
		computation.speedSampleCount,
	)
	writeFingerprintFloat(
		digest,
		computation.
			remainingDistanceM,
	)
	writeFingerprintFloat(
		digest,
		computation.
			lastPositionUncertaintyM,
	)
	writeFingerprintDuration(
		digest,
		computation.
			extrapolationDuration,
	)
	writeConfigFingerprint(
		digest,
		config,
	)

	return fingerprintPrefix +
		hex.EncodeToString(
			digest.Sum(nil),
		)
}

func unavailableFingerprint(
	projectionFingerprint string,
	routeFingerprint string,
	reason string,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		UnavailableFingerprintVersion,
	)
	writeFingerprintString(
		digest,
		projectionFingerprint,
	)
	writeFingerprintString(
		digest,
		routeFingerprint,
	)
	writeFingerprintString(
		digest,
		reason,
	)
	writeConfigFingerprint(
		digest,
		config,
	)

	return fingerprintPrefix +
		hex.EncodeToString(
			digest.Sum(nil),
		)
}

func writeConfigFingerprint(
	digest hash.Hash,
	config Config,
) {
	writeFingerprintFloat(
		digest,
		config.ArrivalRadiusM,
	)
	writeFingerprintFloat(
		digest,
		config.
			MinimumDestinationConfidenceScore,
	)
	writeFingerprintInt(
		digest,
		config.MinimumSpeedSampleCount,
	)
	writeFingerprintInt(
		digest,
		config.MaximumSpeedSampleCount,
	)
	writeFingerprintFloat(
		digest,
		config.MinimumGroundSpeedMPS,
	)
	writeFingerprintFloat(
		digest,
		config.SpeedUncertaintyMultiplier,
	)
	writeFingerprintDuration(
		digest,
		config.MinimumArrivalInterval,
	)
	writeFingerprintDuration(
		digest,
		config.
			MaximumEstimatedArrivalDuration,
	)
	writeFingerprintFloat(
		digest,
		config.
			MaximumExtrapolationConfidenceLoss,
	)
	writeFingerprintFloat(
		digest,
		config.
			ProjectionConfidenceWeight,
	)
	writeFingerprintFloat(
		digest,
		config.
			DestinationConfidenceWeight,
	)
	writeFingerprintFloat(
		digest,
		config.SpeedStabilityWeight,
	)
	writeFingerprintFloat(
		digest,
		config.MediumConfidenceMinimum,
	)
	writeFingerprintFloat(
		digest,
		config.HighConfidenceMinimum,
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
