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
		return flightfeatures.TrajectoryFeatures{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}

	evidence, err := canonicalizeEvidence(ctx, item)
	if err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	limitations := append([]flightfeatures.FeatureLimitation(nil), evidence.limitations...)

	segmentSummary, err := summarizeSegments(ctx, evidence)
	if err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	limitations = append(limitations, segmentSummary.limitations...)

	features := flightfeatures.TrajectoryFeatures{
		Evidence: flightfeatures.GroupEvidence{
			TotalFieldCount:      TrajectoryFeatureFieldCount,
			SupportingPointCount: evidence.supportingPointCount,
		},
	}
	availableFieldCount := 0

	if evidence.pointCountAvailable {
		features.PointCount = evidence.pointCount
		availableFieldCount++
	}
	if evidence.segmentCountAvailable {
		features.SegmentCount = evidence.segmentCount
		availableFieldCount++
	}
	if evidence.gapCountAvailable {
		features.CoverageGapCount = evidence.gapCount
		availableFieldCount++
	}

	hasTrajectoryEvidence := evidence.pointCountAvailable ||
		evidence.segmentCountAvailable || evidence.gapCountAvailable
	if hasTrajectoryEvidence && finiteRatio(item.QualityScore) {
		features.TrajectoryQualityScore = item.QualityScore
		availableFieldCount++
	} else if hasTrajectoryEvidence {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationQualityScoreInvalid,
			Message: "Persisted trajectory quality score is non-finite or outside the inclusive zero-to-one range.",
		})
	}

	if segmentSummary.available {
		features.ObservedSegmentCount = segmentSummary.observedCount
		features.InterpolatedSegmentCount = segmentSummary.interpolatedCount
		features.EstimatedSegmentCount = segmentSummary.estimatedCount
		features.InvalidSegmentCount = segmentSummary.invalidCount
		availableFieldCount += 4
		if evidence.segmentCount > 0 {
			denominator := float64(evidence.segmentCount)
			features.ObservedSegmentShare = float64(segmentSummary.observedCount) / denominator
			features.InterpolatedSegmentShare = float64(segmentSummary.interpolatedCount) / denominator
			features.EstimatedSegmentShare = float64(segmentSummary.estimatedCount) / denominator
			features.InvalidSegmentShare = float64(segmentSummary.invalidCount) / denominator
			availableFieldCount += 4
		} else {
			limitations = append(limitations, flightfeatures.FeatureLimitation{
				Code:    flightfeatures.TrajectoryLimitationSegmentSharesUndefined,
				Message: "Segment status shares are undefined when segment count is zero and are not counted as available fields.",
			})
		}
	}

	sampling, samplingLimitations, err := calculateSamplingMetrics(ctx, evidence.points)
	if err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	limitations = append(limitations, samplingLimitations...)
	if sampling.available {
		features.MeanSamplingIntervalSeconds = sampling.meanSeconds
		features.MaximumSamplingGapSeconds = sampling.maximumSeconds
		availableFieldCount += 2
	}

	coverage, coverageLimitations, err := calculateCoverageRatio(ctx, evidence, segmentSummary)
	if err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	limitations = append(limitations, coverageLimitations...)
	if coverage.available {
		features.CoverageRatio = coverage.value
		availableFieldCount++
	}

	pathEfficiency, pathLimitations, err := calculatePathEfficiency(ctx, evidence)
	if err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	limitations = append(limitations, pathLimitations...)
	if pathEfficiency.available {
		features.PathEfficiencyRatio = pathEfficiency.value
		availableFieldCount++
	}

	features.Evidence.AvailableFieldCount = availableFieldCount
	features.Evidence.Limitations = limitations
	switch {
	case availableFieldCount == TrajectoryFeatureFieldCount:
		features.Evidence.Status = flightfeatures.AvailabilityStatusAvailable
	case availableFieldCount > 0:
		features.Evidence.Status = flightfeatures.AvailabilityStatusPartial
	default:
		features.Evidence.Status = flightfeatures.AvailabilityStatusUnavailable
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.TrajectoryFeatures{}, err
	}
	return cloneFeatures(features), nil
}

func summarizeSegments(
	ctx context.Context,
	evidence canonicalEvidence,
) (segmentStatusSummary, error) {
	if !evidence.segmentCountAvailable {
		return segmentStatusSummary{}, nil
	}
	summary := segmentStatusSummary{available: true}
	unknownCount := 0
	for index, segment := range evidence.segments {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return segmentStatusSummary{}, err
			}
		}
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
			unknownCount++
		}
	}
	if unknownCount > 0 {
		summary.limitations = append(summary.limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationSegmentStatusUnknown,
			Message: fmt.Sprintf(
				"%d trajectory segments have unsupported statuses and were classified as invalid for feature aggregation.",
				unknownCount,
			),
		})
	}
	return summary, nil
}

func finiteRatio(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func cloneFeatures(features flightfeatures.TrajectoryFeatures) flightfeatures.TrajectoryFeatures {
	cloned := features
	cloned.Evidence.Limitations = append([]flightfeatures.FeatureLimitation(nil), features.Evidence.Limitations...)
	return cloned
}
