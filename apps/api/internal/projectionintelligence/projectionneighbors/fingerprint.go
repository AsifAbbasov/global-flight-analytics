package projectionneighbors

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const fingerprintPrefix = "sha256:"

func selectionFingerprint(
	current trajectory.FlightTrajectory,
	candidates []trajectory.FlightTrajectory,
	asOfTime time.Time,
	requiredContinuationDuration time.Duration,
	config Config,
) string {
	digest := sha256.New()

	writeFingerprintString(
		digest,
		FingerprintVersion,
	)
	writeFingerprintString(
		digest,
		config.SimilarityPolicyKey,
	)
	writeFingerprintTime(
		digest,
		asOfTime,
	)
	writeFingerprintDuration(
		digest,
		requiredContinuationDuration,
	)
	writeFingerprintInt(
		digest,
		config.MinimumCurrentPointCount,
	)
	writeFingerprintInt(
		digest,
		config.MaximumCandidateCount,
	)
	writeFingerprintInt(
		digest,
		config.SelectionLimit,
	)
	writeFingerprintFloat(
		digest,
		config.MinimumSimilarityScore,
	)
	writeFingerprintFloat(
		digest,
		config.MaximumAnchorDistanceKM,
	)
	writeFingerprintDuration(
		digest,
		config.MaximumCandidateAge,
	)

	writeTrajectoryFingerprint(
		digest,
		current,
	)

	sortedCandidates := append(
		[]trajectory.FlightTrajectory(nil),
		candidates...,
	)
	sort.SliceStable(
		sortedCandidates,
		func(left int, right int) bool {
			leftID := strings.TrimSpace(
				sortedCandidates[left].ID,
			)
			rightID := strings.TrimSpace(
				sortedCandidates[right].ID,
			)
			return leftID < rightID
		},
	)

	for _, candidate := range sortedCandidates {
		snapshot, _ := snapshotAt(
			candidate,
			asOfTime,
		)
		writeTrajectoryFingerprint(
			digest,
			snapshot,
		)
	}

	return fingerprintPrefix +
		hex.EncodeToString(
			digest.Sum(nil),
		)
}

func writeTrajectoryFingerprint(
	digest hash.Hash,
	item trajectory.FlightTrajectory,
) {
	writeFingerprintString(
		digest,
		strings.TrimSpace(item.ID),
	)
	writeFingerprintString(
		digest,
		strings.TrimSpace(item.FlightID),
	)
	writeFingerprintString(
		digest,
		strings.TrimSpace(item.AircraftID),
	)
	writeFingerprintString(
		digest,
		strings.TrimSpace(item.ICAO24),
	)
	writeFingerprintString(
		digest,
		strings.TrimSpace(item.Callsign),
	)
	writeFingerprintFloat(
		digest,
		item.QualityScore,
	)
	writeFingerprintTime(
		digest,
		item.StartTime,
	)
	writeFingerprintTime(
		digest,
		item.EndTime,
	)

	points := append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	sort.SliceStable(
		points,
		func(left int, right int) bool {
			leftTime := points[left].
				ObservedAt.UTC()
			rightTime := points[right].
				ObservedAt.UTC()
			if !leftTime.Equal(rightTime) {
				return leftTime.Before(
					rightTime,
				)
			}

			return points[left].ID <
				points[right].ID
		},
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
			point.BarometricAltitudeM,
		)
		writeFingerprintFloat(
			digest,
			point.GeometricAltitudeM,
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
	}
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
