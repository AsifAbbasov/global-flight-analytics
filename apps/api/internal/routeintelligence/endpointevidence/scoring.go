package endpointevidence

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
)

const (
	proximityWeight         = 0.50
	trajectoryQualityWeight = 0.25
	segmentStatusWeight     = 0.15
	pointEvidenceWeight     = 0.10
)

type scoreBreakdown struct {
	ProximityContribution         float64
	TrajectoryQualityContribution float64
	SegmentStatusContribution     float64
	PointEvidenceContribution     float64
	Total                         float64
}

func scoreCandidate(
	candidate airportresolver.Candidate,
	trajectoryQuality float64,
	segmentStatus trajectory.SegmentStatus,
	segmentPointCount int,
) scoreBreakdown {
	pointEvidence := clamp01(
		float64(segmentPointCount) / 5,
	)
	breakdown := scoreBreakdown{
		ProximityContribution: proximityWeight *
			clamp01(candidate.ProximityScore),
		TrajectoryQualityContribution: trajectoryQualityWeight *
			clamp01(trajectoryQuality),
		SegmentStatusContribution: segmentStatusWeight *
			segmentStatusScore(segmentStatus),
		PointEvidenceContribution: pointEvidenceWeight * pointEvidence,
	}
	breakdown.Total = clamp01(
		breakdown.ProximityContribution +
			breakdown.TrajectoryQualityContribution +
			breakdown.SegmentStatusContribution +
			breakdown.PointEvidenceContribution,
	)

	return breakdown
}

func segmentStatusScore(
	status trajectory.SegmentStatus,
) float64 {
	switch status {
	case trajectory.SegmentStatusObserved:
		return 1
	case trajectory.SegmentStatusInterpolated:
		return 0.7
	case trajectory.SegmentStatusEstimated:
		return 0.45
	default:
		return 0
	}
}

func clamp01(value float64) float64 {
	switch {
	case math.IsNaN(value),
		math.IsInf(value, 0),
		value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func finiteRatio(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}
