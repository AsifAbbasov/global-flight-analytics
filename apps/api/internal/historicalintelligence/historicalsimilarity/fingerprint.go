package historicalsimilarity

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strconv"
	"strings"
	"time"
)

func comparisonFingerprint(
	reference preparedTrajectory,
	candidate preparedTrajectory,
	config Config,
) string {
	records := []string{
		FingerprintVersion,
		"config.minimum_point_count=" +
			strconv.Itoa(
				config.MinimumPointCount,
			),
		"config.sample_count=" +
			strconv.Itoa(
				config.SampleCount,
			),
		"config.geometry_scale=" +
			floatBits(
				config.GeometryScoreScaleKM,
			),
		"config.endpoint_scale=" +
			floatBits(
				config.EndpointScoreScaleKM,
			),
		"config.geometry_weight=" +
			floatBits(
				config.GeometryWeight,
			),
		"config.endpoints_weight=" +
			floatBits(
				config.EndpointsWeight,
			),
		"config.path_length_weight=" +
			floatBits(
				config.PathLengthWeight,
			),
		"config.duration_weight=" +
			floatBits(
				config.DurationWeight,
			),
	}
	records = appendPreparedFingerprint(
		records,
		"reference",
		reference,
	)
	records = appendPreparedFingerprint(
		records,
		"candidate",
		candidate,
	)

	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" +
		hex.EncodeToString(sum[:])
}

func appendPreparedFingerprint(
	records []string,
	role string,
	item preparedTrajectory,
) []string {
	records = append(
		records,
		role+".id="+item.id,
		role+".path_length="+
			floatBits(item.pathLengthKM),
		role+".duration="+
			floatBits(item.durationSeconds),
	)
	records = appendEvidenceFingerprint(
		records,
		role+".quality",
		item.quality,
	)

	for index, point := range item.points {
		prefix := role + ".point." +
			strconv.Itoa(index)
		records = append(
			records,
			prefix+".source_id="+
				point.sourceID,
			prefix+".latitude="+
				floatBits(point.latitude),
			prefix+".longitude="+
				floatBits(point.longitude),
			prefix+".observed_at="+
				formatTime(point.observedAt),
		)
	}
	for index, sample := range item.samples {
		prefix := role + ".sample." +
			strconv.Itoa(index)
		records = append(
			records,
			prefix+".latitude="+
				floatBits(sample.latitude),
			prefix+".longitude="+
				floatBits(sample.longitude),
			prefix+".observed_at="+
				formatTime(sample.observedAt),
		)
	}
	for index, limitation := range item.limitations {
		prefix := role + ".limitation." +
			strconv.Itoa(index)
		records = append(
			records,
			prefix+".code="+
				limitation.Code,
			prefix+".message="+
				limitation.Message,
		)
	}
	return records
}

func appendEvidenceFingerprint(
	records []string,
	prefix string,
	quality EvidenceQuality,
) []string {
	return append(
		records,
		prefix+".score="+
			floatBits(quality.Score),
		prefix+".declared="+
			floatBits(
				quality.DeclaredQualityScore,
			),
		prefix+".segment="+
			floatBits(
				quality.SegmentQualityScore,
			),
		prefix+".coverage="+
			floatBits(
				quality.CoverageContinuityScore,
			),
		prefix+".cadence="+
			floatBits(
				quality.ObservationCadenceScore,
			),
		prefix+".retention="+
			floatBits(
				quality.PointRetentionScore,
			),
		prefix+".input_points="+
			strconv.Itoa(
				quality.InputPointCount,
			),
		prefix+".usable_points="+
			strconv.Itoa(
				quality.UsablePointCount,
			),
		prefix+".excluded_points="+
			strconv.Itoa(
				quality.ExcludedPointCount,
			),
		prefix+".equal_timestamps="+
			strconv.Itoa(
				quality.EqualTimestampPointCount,
			),
		prefix+".gaps="+
			strconv.Itoa(
				quality.CoverageGapCount,
			),
		prefix+".segments="+
			strconv.Itoa(
				quality.RelevantSegmentCount,
			),
		prefix+".non_observed_segments="+
			strconv.Itoa(
				quality.NonObservedSegmentCount,
			),
		prefix+".invalid_segments="+
			strconv.Itoa(
				quality.InvalidSegmentCount,
			),
		prefix+".source="+quality.SourceName,
	)
}

func floatBits(value float64) string {
	return strconv.FormatUint(
		math.Float64bits(value),
		16,
	)
}

func formatTime(value time.Time) string {
	return value.UTC().
		Format(time.RFC3339Nano)
}
