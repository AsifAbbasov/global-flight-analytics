package projectionevaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const fingerprintPrefix = "sha256:"

func evaluationFingerprint(
	projection projectioncontract.Result,
	actualTrajectory trajectory.FlightTrajectory,
	actualArrival *ActualArrival,
	evaluatedAt time.Time,
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
		projection.Method.Name,
	)
	writeFingerprintString(
		digest,
		projection.Method.Version,
	)
	writeFingerprintTime(
		digest,
		projection.Horizon.AsOfTime,
	)
	writeFingerprintTime(
		digest,
		projection.Horizon.EndTime,
	)
	writeFingerprintTime(
		digest,
		evaluatedAt,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumInterpolationGap,
	)
	writeFingerprintInt(
		digest,
		config.MinimumEvaluatedPointCount,
	)
	writeFingerprintFloat(
		digest,
		config.MaximumHorizontalErrorM,
	)
	writeFingerprintFloat(
		digest,
		config.MaximumAltitudeErrorM,
	)

	points := make(
		[]trajectory.TrackPoint4D,
		0,
		len(actualTrajectory.Points),
	)
	for _, point := range actualTrajectory.Points {
		if point.ObservedAt.IsZero() ||
			point.ObservedAt.UTC().Before(
				projection.Horizon.
					AsOfTime.UTC(),
			) ||
			point.ObservedAt.UTC().After(
				evaluatedAt.UTC(),
			) {
			continue
		}
		points = append(points, point)
	}
	sort.SliceStable(
		points,
		func(left int, right int) bool {
			leftTime :=
				points[left].ObservedAt.UTC()
			rightTime :=
				points[right].ObservedAt.UTC()
			if !leftTime.Equal(rightTime) {
				return leftTime.Before(rightTime)
			}

			return points[left].ID <
				points[right].ID
		},
	)

	writeFingerprintString(
		digest,
		actualTrajectory.ID,
	)
	writeFingerprintInt(
		digest,
		len(points),
	)
	for _, point := range points {
		writeFingerprintString(
			digest,
			point.ID,
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
			point.GeometricAltitudeM,
		)
		writeFingerprintFloat(
			digest,
			point.BarometricAltitudeM,
		)
	}

	if actualArrival == nil {
		writeFingerprintString(
			digest,
			"actual-arrival:nil",
		)
	} else {
		writeFingerprintString(
			digest,
			actualArrival.AirportICAOCode,
		)
		writeFingerprintTime(
			digest,
			actualArrival.BoundaryTime,
		)
		writeFingerprintString(
			digest,
			actualArrival.SourceName,
		)
		writeFingerprintTime(
			digest,
			actualArrival.ObservedAt,
		)
	}

	return fingerprintPrefix +
		hex.EncodeToString(
			digest.Sum(nil),
		)
}

func aggregateFingerprint(
	results []Result,
	generatedAt time.Time,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		AggregateFingerprintVersion,
	)
	writeFingerprintTime(
		digest,
		generatedAt,
	)

	fingerprints := make(
		[]string,
		0,
		len(results),
	)
	for _, result := range results {
		fingerprints = append(
			fingerprints,
			result.EvaluationInputFingerprint,
		)
	}
	sort.Strings(fingerprints)

	for _, fingerprint := range fingerprints {
		writeFingerprintString(
			digest,
			fingerprint,
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
