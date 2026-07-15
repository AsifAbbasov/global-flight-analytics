package historicalsimilarity

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestCompareIdenticalTrajectoriesProducesHighSimilarity(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
	)

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf("compare identical paths: %v", err)
	}

	if result.Score != 1 ||
		result.Level != LevelHigh {
		t.Fatalf(
			"expected exact high similarity, got score=%f level=%s",
			result.Score,
			result.Level,
		)
	}
	if result.MeanDistanceKM != 0 ||
		result.MaximumDistanceKM != 0 {
		t.Fatalf(
			"expected zero geometric distance, got %#v",
			result,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("validate similarity result: %v", err)
	}
}

func TestCompareShiftedTrajectoryReducesScore(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		2,
	)

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf("compare shifted path: %v", err)
	}

	if result.Score >= 1 ||
		result.MeanDistanceKM <= 100 {
		t.Fatalf(
			"expected reduced similarity for shifted path, got score=%f mean=%f",
			result.Score,
			result.MeanDistanceKM,
		)
	}
}

func TestRankOrdersByScoreThenIdentifier(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
	)
	nearB := similarityTrajectory(
		"near-b",
		start.Add(time.Hour),
		0.1,
	)
	far := similarityTrajectory(
		"far",
		start.Add(2*time.Hour),
		3,
	)
	nearA := similarityTrajectory(
		"near-a",
		start.Add(3*time.Hour),
		0.1,
	)

	results, err := NewDefault().Rank(
		reference,
		[]trajectory.FlightTrajectory{
			nearB,
			far,
			nearA,
		},
		3,
	)
	if err != nil {
		t.Fatalf("rank trajectories: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf(
			"expected three ranked results, got %d",
			len(results),
		)
	}
	if results[0].CandidateTrajectoryID !=
		"near-a" ||
		results[1].CandidateTrajectoryID !=
			"near-b" ||
		results[2].CandidateTrajectoryID !=
			"far" {
		t.Fatalf(
			"unexpected deterministic ranking: %#v",
			results,
		)
	}
}

func TestCompareRejectsInsufficientPoints(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
	)
	candidate.Points =
		candidate.Points[:2]

	_, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if !errors.Is(
		err,
		ErrCandidateNotComparable,
	) {
		t.Fatalf(
			"expected candidate not comparable error, got %v",
			err,
		)
	}
}

func TestCompareSortsPointsByObservationTime(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
	)

	for left, right := 0, len(candidate.Points)-1; left < right; left, right =
		left+1, right-1 {
		candidate.Points[left],
			candidate.Points[right] =
			candidate.Points[right],
			candidate.Points[left]
	}

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare reordered trajectory: %v",
			err,
		)
	}
	if result.Score != 1 {
		t.Fatalf(
			"expected chronological normalization, got %f",
			result.Score,
		)
	}
}

func similarityTrajectory(
	id string,
	start time.Time,
	latitudeOffset float64,
) trajectory.FlightTrajectory {
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		5,
	)
	for index := 0; index < 5; index++ {
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID: id + "-point",
				Latitude: 40 +
					latitudeOffset +
					float64(index)*0.1,
				Longitude: 49 +
					float64(index)*0.1,
				ObservedAt: start.Add(
					time.Duration(index) *
						time.Minute,
				),
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:        id,
		StartTime: start,
		EndTime: start.Add(
			4 * time.Minute,
		),
		PointCount: len(points),
		Points:     points,
	}
}

func similarityTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}
