package trajectoryquality

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const scoreComparisonTolerance = 1e-9

func TestSegmentScoreFromAggregateReturnsZeroForEmptyInput(
	t *testing.T,
) {
	t.Parallel()

	actual := SegmentScoreFromAggregate(
		0,
		0,
	)

	if actual != 0 {
		t.Fatalf(
			"expected empty segment score to be 0, got %f",
			actual,
		)
	}
}

func TestSegmentScoreFromAggregateReturnsZeroForNegativePointCount(
	t *testing.T,
) {
	t.Parallel()

	actual := SegmentScoreFromAggregate(
		1.8,
		-3,
	)

	if actual != 0 {
		t.Fatalf(
			"expected negative point count score to be 0, got %f",
			actual,
		)
	}
}

func TestSegmentScoreFromAggregateReturnsArithmeticMean(
	t *testing.T,
) {
	t.Parallel()

	actual := SegmentScoreFromAggregate(
		1.8,
		3,
	)

	expected := 0.6

	assertScoreClose(
		t,
		expected,
		actual,
	)
}

func TestTrajectoryScoreReturnsZeroForEmptyInput(
	t *testing.T,
) {
	t.Parallel()

	actual := TrajectoryScore(nil)

	if actual != 0 {
		t.Fatalf(
			"expected empty trajectory score to be 0, got %f",
			actual,
		)
	}
}

func TestTrajectoryScoreReturnsPointWeightedMean(
	t *testing.T,
) {
	t.Parallel()

	segments := []trajectory.TrajectorySegment{
		{
			QualityScore: 0.8,
			PointCount:   2,
		},
		{
			QualityScore: 0.5,
			PointCount:   4,
		},
	}

	actual := TrajectoryScore(segments)

	expected := 0.6

	assertScoreClose(
		t,
		expected,
		actual,
	)
}

func TestTrajectoryScoreIgnoresSegmentsWithoutPoints(
	t *testing.T,
) {
	t.Parallel()

	segments := []trajectory.TrajectorySegment{
		{
			QualityScore: 1.0,
			PointCount:   0,
		},
		{
			QualityScore: 0.75,
			PointCount:   4,
		},
	}

	actual := TrajectoryScore(segments)

	expected := 0.75

	assertScoreClose(
		t,
		expected,
		actual,
	)
}

func TestTrajectoryScoreReturnsZeroWhenNoSegmentHasPoints(
	t *testing.T,
) {
	t.Parallel()

	segments := []trajectory.TrajectorySegment{
		{
			QualityScore: 0.9,
			PointCount:   0,
		},
		{
			QualityScore: 0.4,
			PointCount:   0,
		},
	}

	actual := TrajectoryScore(segments)

	if actual != 0 {
		t.Fatalf(
			"expected trajectory score to be 0, got %f",
			actual,
		)
	}
}

func assertScoreClose(
	t *testing.T,
	expected float64,
	actual float64,
) {
	t.Helper()

	difference := math.Abs(
		expected - actual,
	)

	if difference > scoreComparisonTolerance {
		t.Fatalf(
			"expected score %.12f, got %.12f",
			expected,
			actual,
		)
	}
}
