package projectioncontinuation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

const fingerprintPrefix = "sha256:"

func continuationFingerprint(
	current trajectory.FlightTrajectory,
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
	plan projectionhorizon.Plan,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FingerprintVersion,
	)
	writeFingerprintString(
		digest,
		current.ID,
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
		plan.AsOfTime,
	)
	writeFingerprintTime(
		digest,
		plan.EndTime,
	)
	writeFingerprintDuration(
		digest,
		plan.Step,
	)
	writeFingerprintInt(
		digest,
		config.MinimumPointSupport,
	)
	writeFingerprintInt(
		digest,
		config.MinimumAltitudeSupport,
	)
	writeFingerprintFloat(
		digest,
		config.InitialHorizontalUncertaintyM,
	)
	writeFingerprintFloat(
		digest,
		config.HorizontalUncertaintyGrowthMPS,
	)
	writeFingerprintFloat(
		digest,
		config.InitialVerticalUncertaintyM,
	)
	writeFingerprintFloat(
		digest,
		config.VerticalUncertaintyGrowthMPS,
	)
	writeFingerprintFloat(
		digest,
		config.NeighborSpreadMultiplier,
	)
	writeFingerprintFloat(
		digest,
		config.MaximumConfidenceLoss,
	)
	writeFingerprintFloat(
		digest,
		config.MediumConfidenceMinimum,
	)
	writeFingerprintFloat(
		digest,
		config.HighConfidenceMinimum,
	)

	return fingerprintPrefix +
		hex.EncodeToString(
			digest.Sum(nil),
		)
}

func fallbackFingerprint(
	existingFingerprint string,
	reason string,
	selectionFingerprint string,
	patternFingerprint string,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FallbackFingerprintVersion,
	)
	writeFingerprintString(
		digest,
		existingFingerprint,
	)
	writeFingerprintString(
		digest,
		reason,
	)
	writeFingerprintString(
		digest,
		selectionFingerprint,
	)
	writeFingerprintString(
		digest,
		patternFingerprint,
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
