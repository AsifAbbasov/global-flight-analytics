package trajectoryquality

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func SegmentScoreFromAggregate(
	totalScore float64,
	pointCount int,
) float64 {
	if pointCount <= 0 {
		return 0
	}

	return totalScore / float64(pointCount)
}

func TrajectoryScore(
	segments []trajectory.TrajectorySegment,
) float64 {
	totalPointCount := 0
	weightedScoreTotal := 0.0

	for _, segment := range segments {
		if segment.PointCount <= 0 {
			continue
		}

		totalPointCount += segment.PointCount

		weightedScoreTotal +=
			segment.QualityScore *
				float64(segment.PointCount)
	}

	if totalPointCount == 0 {
		return 0
	}

	return weightedScoreTotal /
		float64(totalPointCount)
}
