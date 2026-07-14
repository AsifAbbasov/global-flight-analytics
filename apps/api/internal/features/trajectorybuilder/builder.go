package trajectorybuilder

import (
	"context"
	"fmt"
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var _ extractor.TrajectoryBuilder = (*Builder)(nil)

type Builder struct{}

func New() *Builder {
	return &Builder{}
}

func (builder *Builder) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.TrajectoryFeatures, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}

	pointCount := len(item.Points)
	segmentCount := len(item.Segments)
	coverageGapCount := len(item.CoverageGaps)

	limitations := metadataLimitations(
		item,
		pointCount,
		segmentCount,
		coverageGapCount,
	)

	segmentSummary := summarizeSegments(item.Segments)
	limitations = append(
		limitations,
		segmentSummary.limitations...,
	)

	features := flightfeatures.TrajectoryFeatures{
		Evidence: flightfeatures.GroupEvidence{
			TotalFieldCount:      TrajectoryFeatureFieldCount,
			SupportingPointCount: pointCount,
		},
		PointCount:               pointCount,
		SegmentCount:             segmentCount,
		CoverageGapCount:         coverageGapCount,
		ObservedSegmentCount:     segmentSummary.observedCount,
		InterpolatedSegmentCount: segmentSummary.interpolatedCount,
		EstimatedSegmentCount:    segmentSummary.estimatedCount,
		InvalidSegmentCount:      segmentSummary.invalidCount,
	}

	availableFieldCount := 11

	if segmentCount > 0 {
		denominator := float64(segmentCount)
		features.ObservedSegmentShare =
			float64(segmentSummary.observedCount) / denominator
		features.InterpolatedSegmentShare =
			float64(segmentSummary.interpolatedCount) / denominator
		features.EstimatedSegmentShare =
			float64(segmentSummary.estimatedCount) / denominator
		features.InvalidSegmentShare =
			float64(segmentSummary.invalidCount) / denominator
	}

	if finiteRatio(item.QualityScore) {
		features.TrajectoryQualityScore =
			item.QualityScore
		availableFieldCount++
	} else {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_quality_score_invalid",
				Message: "Persisted trajectory quality score is non-finite or outside the inclusive zero-to-one range.",
			},
		)
	}

	sampling, samplingLimitations :=
		calculateSamplingMetrics(item.Points)
	limitations = append(
		limitations,
		samplingLimitations...,
	)
	if sampling.available {
		features.MeanSamplingIntervalSeconds =
			sampling.meanSeconds
		features.MaximumSamplingGapSeconds =
			sampling.maximumSeconds
		availableFieldCount += 2
	}

	coverage, coverageLimitations :=
		calculateCoverageRatio(item)
	limitations = append(
		limitations,
		coverageLimitations...,
	)
	if coverage.available {
		features.CoverageRatio = coverage.value
		availableFieldCount++
	}

	pathEfficiency, pathLimitations :=
		calculatePathEfficiency(ctx, item)
	if err := ctx.Err(); err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	limitations = append(
		limitations,
		pathLimitations...,
	)
	if pathEfficiency.available {
		features.PathEfficiencyRatio =
			pathEfficiency.value
		availableFieldCount++
	}

	features.Evidence.AvailableFieldCount =
		availableFieldCount
	features.Evidence.Limitations = limitations

	switch {
	case availableFieldCount ==
		TrajectoryFeatureFieldCount:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusAvailable
	case availableFieldCount > 0:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusPartial
	default:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusUnavailable
	}

	if err := ctx.Err(); err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}

	return cloneFeatures(features), nil
}

type segmentStatusSummary struct {
	observedCount     int
	interpolatedCount int
	estimatedCount    int
	invalidCount      int
	limitations       []flightfeatures.FeatureLimitation
}

func summarizeSegments(
	segments []trajectory.TrajectorySegment,
) segmentStatusSummary {
	summary := segmentStatusSummary{}

	for _, segment := range segments {
		switch segment.Status {
		case trajectory.SegmentStatusObserved:
			summary.observedCount++
		case trajectory.SegmentStatusInterpolated:
			summary.interpolatedCount++
		case trajectory.SegmentStatusEstimated:
			summary.estimatedCount++
		case trajectory.SegmentStatusInvalid:
			summary.invalidCount++
		default:
			summary.invalidCount++
			summary.limitations = append(
				summary.limitations,
				flightfeatures.FeatureLimitation{
					Code: "trajectory_segment_status_unknown",
					Message: fmt.Sprintf(
						"Trajectory segment %q has unsupported status %q and was classified as invalid for feature aggregation.",
						segment.ID,
						segment.Status,
					),
				},
			)
		}
	}

	return summary
}

func metadataLimitations(
	item trajectory.FlightTrajectory,
	pointCount int,
	segmentCount int,
	coverageGapCount int,
) []flightfeatures.FeatureLimitation {
	result := make(
		[]flightfeatures.FeatureLimitation,
		0,
		3,
	)

	if item.PointCount != pointCount {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_point_count_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory point-count metadata reports %d points while %d point records are present.",
					item.PointCount,
					pointCount,
				),
			},
		)
	}
	if item.SegmentCount != segmentCount {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_segment_count_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory segment-count metadata reports %d segments while %d segment records are present.",
					item.SegmentCount,
					segmentCount,
				),
			},
		)
	}
	if item.CoverageGapCount != coverageGapCount {
		result = append(
			result,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_coverage_gap_count_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory coverage-gap metadata reports %d gaps while %d gap records are present.",
					item.CoverageGapCount,
					coverageGapCount,
				),
			},
		)
	}

	return result
}

func finiteRatio(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}

func cloneFeatures(
	features flightfeatures.TrajectoryFeatures,
) flightfeatures.TrajectoryFeatures {
	cloned := features
	cloned.Evidence.Limitations = append(
		[]flightfeatures.FeatureLimitation(nil),
		features.Evidence.Limitations...,
	)

	return cloned
}
