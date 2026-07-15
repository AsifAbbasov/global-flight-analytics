package projectionbaseline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

const fingerprintPrefix = "sha256:"

func inputFingerprint(
	item trajectory.FlightTrajectory,
	point trajectory.TrackPoint4D,
	plan projectionhorizon.Plan,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		Version,
	)
	writeFingerprintString(
		digest,
		item.ID,
	)
	writeFingerprintString(
		digest,
		item.FlightID,
	)
	writeFingerprintString(
		digest,
		item.AircraftID,
	)
	writeFingerprintString(
		digest,
		item.ICAO24,
	)
	writeFingerprintString(
		digest,
		item.Callsign,
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
	writeFingerprintTime(
		digest,
		point.ObservedAt,
	)
	writeFingerprintFloat(
		digest,
		point.Latitude,
	)
	writeFingerprintFloat(
		digest,
		point.Longitude,
	)
	writeFingerprintFloat(
		digest,
		point.BarometricAltitudeM,
	)
	writeFingerprintString(
		digest,
		string(
			point.BarometricAltitudeStatus,
		),
	)
	writeFingerprintFloat(
		digest,
		point.GeometricAltitudeM,
	)
	writeFingerprintString(
		digest,
		string(
			point.GeometricAltitudeStatus,
		),
	)
	writeFingerprintFloat(
		digest,
		point.VelocityMPS,
	)
	writeFingerprintFloat(
		digest,
		point.HeadingDegrees,
	)
	writeFingerprintFloat(
		digest,
		point.VerticalRateMPS,
	)
	writeFingerprintFloat(
		digest,
		item.QualityScore,
	)
	writeFingerprintFloat(
		digest,
		config.
			InitialHorizontalUncertaintyM,
	)
	writeFingerprintFloat(
		digest,
		config.
			HorizontalUncertaintyGrowthMPS,
	)
	writeFingerprintFloat(
		digest,
		config.
			InitialVerticalUncertaintyM,
	)
	writeFingerprintFloat(
		digest,
		config.
			VerticalUncertaintyGrowthMPS,
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
	writeFingerprintString(
		digest,
		fmt.Sprintf(
			"%t",
			config.AllowOnGround,
		),
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
