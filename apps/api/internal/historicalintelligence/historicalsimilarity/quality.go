package historicalsimilarity

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const (
	declaredQualityWeight    = 0.30
	segmentQualityWeight     = 0.20
	coverageContinuityWeight = 0.25
	observationCadenceWeight = 0.15
	pointRetentionWeight     = 0.10
)

func assessEvidenceQuality(
	item trajectory.FlightTrajectory,
	points []geoPoint,
	excludedPointCount int,
	equalTimestampPointCount int,
) (EvidenceQuality, []Notice, error) {
	if !ratio(item.QualityScore) {
		return EvidenceQuality{}, nil,
			ErrTrajectoryQualityInvalid
	}

	windowStart := points[0].observedAt
	windowEnd := points[len(points)-1].observedAt

	segmentScore, relevantSegmentCount,
		nonObservedSegmentCount,
		invalidSegmentCount,
		segmentLimitations, err :=
		segmentEvidence(
			item.Segments,
			item.QualityScore,
			windowStart,
			windowEnd,
		)
	if err != nil {
		return EvidenceQuality{}, nil, err
	}

	coverageScore, gapCount,
		gapLimitations, err :=
		coverageEvidence(
			item,
			points,
			windowStart,
			windowEnd,
		)
	if err != nil {
		return EvidenceQuality{}, nil, err
	}

	cadenceScore := observationCadenceScore(
		points,
	)
	retentionScore := float64(len(points)) /
		float64(len(item.Points))
	retentionScore = clampRatio(retentionScore)

	score := item.QualityScore*
		declaredQualityWeight +
		segmentScore*
			segmentQualityWeight +
		coverageScore*
			coverageContinuityWeight +
		cadenceScore*
			observationCadenceWeight +
		retentionScore*
			pointRetentionWeight
	score = clampRatio(score)

	quality := EvidenceQuality{
		Score: score,

		DeclaredQualityScore:     item.QualityScore,
		SegmentQualityScore:      segmentScore,
		CoverageContinuityScore:  coverageScore,
		ObservationCadenceScore:  cadenceScore,
		PointRetentionScore:      retentionScore,
		InputPointCount:          len(item.Points),
		UsablePointCount:         len(points),
		ExcludedPointCount:       excludedPointCount,
		EqualTimestampPointCount: equalTimestampPointCount,
		CoverageGapCount:         gapCount,
		RelevantSegmentCount:     relevantSegmentCount,
		NonObservedSegmentCount:  nonObservedSegmentCount,
		InvalidSegmentCount:      invalidSegmentCount,
		SourceName: strings.TrimSpace(
			item.SourceName,
		),
	}

	limitations := append(
		segmentLimitations,
		gapLimitations...,
	)
	if cadenceScore < 0.6 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_observation_cadence_irregular",
				Message: fmt.Sprintf(
					"Observation cadence regularity score is %.6f because the longest usable interval materially exceeds the median interval.",
					cadenceScore,
				),
			},
		)
	}
	if item.QualityScore < 0.6 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_declared_quality_low",
				Message: fmt.Sprintf(
					"Declared trajectory quality score is %.6f.",
					item.QualityScore,
				),
			},
		)
	}
	if strings.TrimSpace(item.SourceName) == "" {
		limitations = append(
			limitations,
			Notice{
				Code:    "trajectory_source_unidentified",
				Message: "The trajectory does not identify its source; provider reliability cannot be inferred from the available domain contract.",
			},
		)
	}

	return quality,
		normalizeNotices(limitations),
		nil
}

func segmentEvidence(
	segments []trajectory.TrajectorySegment,
	declaredQuality float64,
	windowStart time.Time,
	windowEnd time.Time,
) (
	float64,
	int,
	int,
	int,
	[]Notice,
	error,
) {
	totalWeight := 0.0
	weightedScore := 0.0
	relevantCount := 0
	nonObservedCount := 0
	invalidCount := 0

	for _, segment := range segments {
		if !timeRangesOverlap(
			segment.StartTime,
			segment.EndTime,
			windowStart,
			windowEnd,
		) {
			continue
		}
		if segment.StartTime.IsZero() ||
			segment.EndTime.IsZero() ||
			segment.EndTime.Before(
				segment.StartTime,
			) ||
			!ratio(segment.QualityScore) {
			return 0, 0, 0, 0, nil,
				ErrTrajectorySegmentInvalid
		}

		statusFactor, known :=
			segmentStatusFactor(
				segment.Status,
			)
		if !known {
			return 0, 0, 0, 0, nil,
				ErrTrajectorySegmentInvalid
		}

		relevantCount++
		if segment.Status !=
			trajectory.SegmentStatusObserved {
			nonObservedCount++
		}
		if segment.Status ==
			trajectory.SegmentStatusInvalid {
			invalidCount++
		}

		weight := float64(
			maximumInt(segment.PointCount, 1),
		)
		totalWeight += weight
		weightedScore +=
			segment.QualityScore *
				statusFactor *
				weight
	}

	if relevantCount == 0 {
		return declaredQuality,
			0,
			0,
			0,
			[]Notice{
				{
					Code:    "trajectory_segment_quality_unavailable",
					Message: "No segment-level quality evidence overlaps the compared trajectory window; declared trajectory quality is used as the segment-quality fallback.",
				},
			},
			nil
	}

	score := 0.0
	if totalWeight > 0 {
		score = weightedScore / totalWeight
	}
	score = clampRatio(score)

	limitations := []Notice{}
	if nonObservedCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_non_observed_segments_present",
				Message: fmt.Sprintf(
					"%d of %d relevant trajectory segments are interpolated, estimated, or invalid.",
					nonObservedCount,
					relevantCount,
				),
			},
		)
	}
	if invalidCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_invalid_segments_present",
				Message: fmt.Sprintf(
					"%d relevant trajectory segments are explicitly invalid.",
					invalidCount,
				),
			},
		)
	}

	return score,
		relevantCount,
		nonObservedCount,
		invalidCount,
		limitations,
		nil
}

func segmentStatusFactor(
	status trajectory.SegmentStatus,
) (float64, bool) {
	switch status {
	case trajectory.SegmentStatusObserved:
		return 1, true
	case trajectory.SegmentStatusInterpolated:
		return 0.75, true
	case trajectory.SegmentStatusEstimated:
		return 0.5, true
	case trajectory.SegmentStatusInvalid:
		return 0, true
	default:
		return 0, false
	}
}

func coverageEvidence(
	item trajectory.FlightTrajectory,
	points []geoPoint,
	windowStart time.Time,
	windowEnd time.Time,
) (float64, int, []Notice, error) {
	gapCount := 0
	gapDurationSeconds := 0.0

	if len(item.CoverageGaps) > 0 {
		for _, gap := range item.CoverageGaps {
			if !timeRangesOverlap(
				gap.StartTime,
				gap.EndTime,
				windowStart,
				windowEnd,
			) {
				continue
			}
			if gap.StartTime.IsZero() ||
				gap.EndTime.IsZero() ||
				gap.EndTime.Before(
					gap.StartTime,
				) ||
				!finite(gap.DistanceKm) ||
				gap.DistanceKm < 0 {
				return 0, 0, nil,
					ErrTrajectoryGapInvalid
			}

			gapCount++
			start := latestTime(
				gap.StartTime.UTC(),
				windowStart,
			)
			end := earliestTime(
				gap.EndTime.UTC(),
				windowEnd,
			)
			if end.After(start) {
				gapDurationSeconds +=
					end.Sub(start).Seconds()
			}
		}
	} else {
		if item.CoverageGapCount < 0 {
			return 0, 0, nil,
				ErrTrajectoryGapInvalid
		}
		gapCount = item.CoverageGapCount
	}

	possibleIntervals := maximumInt(
		len(points)-1,
		1,
	)
	countPenalty := clampRatio(
		float64(gapCount) /
			float64(possibleIntervals),
	)
	duration := windowEnd.Sub(
		windowStart,
	).Seconds()
	durationPenalty := 0.0
	if duration > 0 {
		durationPenalty = clampRatio(
			gapDurationSeconds / duration,
		)
	}
	penalty := math.Max(
		countPenalty,
		durationPenalty,
	)
	score := 1 - penalty

	limitations := []Notice{}
	if gapCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_coverage_gaps_present",
				Message: fmt.Sprintf(
					"%d coverage gaps overlap the compared trajectory window; continuity score is %.6f.",
					gapCount,
					score,
				),
			},
		)
	}

	return score,
		gapCount,
		limitations,
		nil
}

func observationCadenceScore(
	points []geoPoint,
) float64 {
	intervals := make(
		[]float64,
		0,
		len(points)-1,
	)
	maximum := 0.0
	for index := 1; index < len(points); index++ {
		seconds := points[index].
			observedAt.Sub(
			points[index-1].
				observedAt,
		).Seconds()
		if seconds < 0 || !finite(seconds) {
			return 0
		}
		intervals = append(
			intervals,
			seconds,
		)
		if seconds > maximum {
			maximum = seconds
		}
	}
	if len(intervals) == 0 ||
		maximum <= 0 {
		return 0
	}

	sort.Float64s(intervals)
	median := intervals[len(intervals)/2]
	if len(intervals)%2 == 0 {
		median = (intervals[len(intervals)/2-1] +
			intervals[len(intervals)/2]) / 2
	}
	return clampRatio(median / maximum)
}

func comparisonConfidence(
	reference EvidenceQuality,
	candidate EvidenceQuality,
) EvidenceConfidence {
	score := math.Min(
		reference.Score,
		candidate.Score,
	)
	return EvidenceConfidence{
		Score:     score,
		Level:     ConfidenceLevelForScore(score),
		Reference: reference,
		Candidate: candidate,
		Reasons: []Notice{
			{
				Code: "comparison_confidence_uses_weaker_trajectory_evidence",
				Message: fmt.Sprintf(
					"Comparison confidence uses the weaker trajectory evidence score: reference=%.6f candidate=%.6f.",
					reference.Score,
					candidate.Score,
				),
			},
		},
	}
}

func timeRangesOverlap(
	leftStart time.Time,
	leftEnd time.Time,
	rightStart time.Time,
	rightEnd time.Time,
) bool {
	if leftStart.IsZero() ||
		leftEnd.IsZero() ||
		rightStart.IsZero() ||
		rightEnd.IsZero() {
		return false
	}
	return !leftEnd.Before(rightStart) &&
		!leftStart.After(rightEnd)
}

func latestTime(
	left time.Time,
	right time.Time,
) time.Time {
	if right.After(left) {
		return right
	}
	return left
}

func earliestTime(
	left time.Time,
	right time.Time,
) time.Time {
	if right.Before(left) {
		return right
	}
	return left
}

func maximumInt(
	left int,
	right int,
) int {
	if right > left {
		return right
	}
	return left
}
