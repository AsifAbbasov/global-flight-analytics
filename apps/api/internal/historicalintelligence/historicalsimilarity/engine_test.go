package historicalsimilarity

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestCompareSeparatesSimilarityFromEvidenceConfidence(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		0.2,
	)
	candidate.CoverageGapCount = 2
	candidate.CoverageGaps = []trajectory.CoverageGap{
		similarityGap(
			"gap-one",
			candidate.ID,
			candidate.Points[1].ObservedAt,
			candidate.Points[2].ObservedAt,
		),
		similarityGap(
			"gap-two",
			candidate.ID,
			candidate.Points[2].ObservedAt,
			candidate.Points[3].ObservedAt,
		),
	}
	candidate.Segments[0].Status =
		trajectory.SegmentStatusEstimated
	candidate.Segments[0].QualityScore = 0.2

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare identical geometry with weak evidence: %v",
			err,
		)
	}

	if result.Score != 1 ||
		result.Level != LevelHigh {
		t.Fatalf(
			"similarity score must remain geometric: score=%f level=%s",
			result.Score,
			result.Level,
		)
	}
	if result.Confidence.Score >= 0.6 ||
		result.Confidence.Level ==
			ConfidenceLevelHigh {
		t.Fatalf(
			"weak trajectory evidence must reduce confidence: %#v",
			result.Confidence,
		)
	}
	assertNoticeCode(
		t,
		result.Limitations,
		"candidate_trajectory_coverage_gaps_present",
	)
	assertNoticeCode(
		t,
		result.Limitations,
		"candidate_trajectory_non_observed_segments_present",
	)
}

func TestCompareIdenticalHighQualityTrajectories(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
	)

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare identical paths: %v",
			err,
		)
	}
	if result.Score != 1 ||
		result.Level != LevelHigh {
		t.Fatalf(
			"expected exact high similarity, got score=%f level=%s",
			result.Score,
			result.Level,
		)
	}
	if result.Confidence.Score != 1 ||
		result.Confidence.Level !=
			ConfidenceLevelHigh {
		t.Fatalf(
			"expected high evidence confidence, got %#v",
			result.Confidence,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf(
			"validate result: %v",
			err,
		)
	}
}

func TestConfigRejectsExcessiveSampleCount(
	t *testing.T,
) {
	config := DefaultConfig()
	config.SampleCount =
		MaximumSampleCount + 1

	_, err := New(config)
	if !errors.Is(
		err,
		ErrSampleCountInvalid,
	) {
		t.Fatalf(
			"expected sample-count error, got %v",
			err,
		)
	}
}

func TestCompareRejectsExcessiveInputPoints(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	reference.Points = make(
		[]trajectory.TrackPoint4D,
		MaximumInputPointCount+1,
	)

	_, err := NewDefault().Compare(
		reference,
		similarityTrajectory(
			"candidate",
			start.Add(time.Hour),
			0,
			1,
		),
	)
	if !errors.Is(
		err,
		ErrTrajectoryPointLimitExceeded,
	) {
		t.Fatalf(
			"expected point-limit error, got %v",
			err,
		)
	}
}

func TestCompareCanonicalizesEqualTimestamps(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
	)
	candidate.Points[1].ObservedAt =
		candidate.Points[2].ObservedAt

	first, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"first compare: %v",
			err,
		)
	}

	candidate.Points[1],
		candidate.Points[2] =
		candidate.Points[2],
		candidate.Points[1]
	second, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"second compare: %v",
			err,
		)
	}

	if first.Score != second.Score ||
		first.InputFingerprint !=
			second.InputFingerprint {
		t.Fatalf(
			"canonical equal timestamps must preserve score and fingerprint: first=%#v second=%#v",
			first,
			second,
		)
	}
	assertNoticeCode(
		t,
		first.Limitations,
		"candidate_equal_timestamp_points_canonicalized",
	)
}

func TestFingerprintUsesExactFloatBits(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
	)

	first, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"first compare: %v",
			err,
		)
	}

	candidate.Points[2].Latitude =
		math.Nextafter(
			candidate.Points[2].Latitude,
			math.Inf(1),
		)
	second, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"second compare: %v",
			err,
		)
	}
	if first.InputFingerprint ==
		second.InputFingerprint {
		t.Fatal(
			"one-bit coordinate change must change fingerprint",
		)
	}
}

func TestEndpointComponentUsesWorstEndpoint(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
	)
	candidate.Points[len(candidate.Points)-1].
		Latitude += 1

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare endpoint shift: %v",
			err,
		)
	}
	endpoint := componentByName(
		t,
		result.Components,
		ComponentEndpoints,
	)
	want := math.Max(
		result.StartEndpointDistanceKM,
		result.EndEndpointDistanceKM,
	)
	if !nearlyEqual(
		endpoint.ObservedValue,
		want,
	) {
		t.Fatalf(
			"endpoint observed=%f want=%f",
			endpoint.ObservedValue,
			want,
		)
	}
}

func TestRelativeDifferenceUsesExactZeroScalePolicy(
	t *testing.T,
) {
	if value := relativeDifference(
		0,
		0.5,
	); value != 1 {
		t.Fatalf(
			"relative difference=%f want=1",
			value,
		)
	}
	if value := relativeDifference(
		0,
		0,
	); value != 0 {
		t.Fatalf(
			"zero difference=%f want=0",
			value,
		)
	}
}

func TestCompareHandlesDateLineAndPolarRoutes(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := specialTrajectory(
		"reference",
		start,
		[]coordinate{
			{latitude: 84, longitude: 179},
			{latitude: 85, longitude: -179},
			{latitude: 86, longitude: -175},
			{latitude: 87, longitude: -170},
		},
	)
	candidate := specialTrajectory(
		"candidate",
		start.Add(time.Hour),
		[]coordinate{
			{latitude: 84, longitude: 179},
			{latitude: 85, longitude: -179},
			{latitude: 86, longitude: -175},
			{latitude: 87, longitude: -170},
		},
	)

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare date-line polar routes: %v",
			err,
		)
	}
	if result.Score != 1 ||
		result.MeanDistanceKM != 0 {
		t.Fatalf(
			"unexpected date-line result: %#v",
			result,
		)
	}
}

func TestCompareReportsZeroLengthIndexFallback(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := zeroLengthTrajectory(
		"reference",
		start,
	)
	candidate := zeroLengthTrajectory(
		"candidate",
		start.Add(time.Hour),
	)

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare zero-length routes: %v",
			err,
		)
	}
	assertNoticeCode(
		t,
		result.Limitations,
		"reference_zero_length_path_index_resampling",
	)
	assertNoticeCode(
		t,
		result.Limitations,
		"candidate_zero_length_path_index_resampling",
	)
}

func TestCompareBindsExcludedPointsIntoQualityAndLimitations(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
	)
	candidate.Points = append(
		candidate.Points,
		trajectory.TrackPoint4D{
			ID:         "invalid",
			Latitude:   999,
			Longitude:  49,
			ObservedAt: start,
		},
	)
	candidate.PointCount =
		len(candidate.Points)

	result, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if err != nil {
		t.Fatalf(
			"compare excluded point: %v",
			err,
		)
	}
	if result.Confidence.Candidate.
		PointRetentionScore >= 1 {
		t.Fatalf(
			"expected reduced retention score: %#v",
			result.Confidence.Candidate,
		)
	}
	assertNoticeCode(
		t,
		result.Limitations,
		"candidate_trajectory_points_excluded",
	)
}

func TestResultValidateRejectsUnknownComponent(
	t *testing.T,
) {
	result := validSimilarityResult(t)
	result.Components[0].Name =
		ComponentName("unknown")

	if !errors.Is(
		result.Validate(),
		ErrResultInvalid,
	) {
		t.Fatal(
			"unknown component must fail validation",
		)
	}
}

func TestResultValidateRejectsWeightedScoreMismatch(
	t *testing.T,
) {
	result := validSimilarityResult(t)
	result.Score -= 0.1

	if !errors.Is(
		result.Validate(),
		ErrResultInvalid,
	) {
		t.Fatal(
			"weighted score mismatch must fail validation",
		)
	}
}

func TestResultValidateRejectsMeanAboveMaximum(
	t *testing.T,
) {
	result := validSimilarityResult(t)
	result.MeanDistanceKM =
		result.MaximumDistanceKM + 1

	if !errors.Is(
		result.Validate(),
		ErrResultInvalid,
	) {
		t.Fatal(
			"mean above maximum must fail validation",
		)
	}
}

func TestResultValidateRejectsObservedValueMismatch(
	t *testing.T,
) {
	result := validSimilarityResult(t)
	result.Components[0].ObservedValue += 1

	if !errors.Is(
		result.Validate(),
		ErrResultInvalid,
	) {
		t.Fatal(
			"component observed value mismatch must fail validation",
		)
	}
}

func TestResultValidateRejectsConfidenceMismatch(
	t *testing.T,
) {
	result := validSimilarityResult(t)
	result.Confidence.Score = 1

	if !errors.Is(
		result.Validate(),
		ErrResultInvalid,
	) {
		t.Fatal(
			"confidence mismatch must fail validation",
		)
	}
}

func TestNewDefaultDoesNotPanicAndProducesValidConfig(
	t *testing.T,
) {
	engine := NewDefault()
	if engine == nil {
		t.Fatal("default engine is required")
	}
	if err := engine.config.Validate(); err != nil {
		t.Fatalf(
			"default config invalid: %v",
			err,
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
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
	)
	candidate.Points =
		candidate.Points[:2]
	candidate.PointCount = 2

	_, err := NewDefault().Compare(
		reference,
		candidate,
	)
	if !errors.Is(
		err,
		ErrCandidateNotComparable,
	) {
		t.Fatalf(
			"expected candidate error, got %v",
			err,
		)
	}
}

func TestCompareSortsUniquePointsByObservationTime(
	t *testing.T,
) {
	start := similarityTestTime()
	reference := similarityTrajectory(
		"reference",
		start,
		0,
		1,
	)
	candidate := similarityTrajectory(
		"candidate",
		start.Add(time.Hour),
		0,
		1,
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

func validSimilarityResult(
	t *testing.T,
) Result {
	t.Helper()
	start := similarityTestTime()
	result, err := NewDefault().Compare(
		similarityTrajectory(
			"reference",
			start,
			0,
			1,
		),
		similarityTrajectory(
			"candidate",
			start.Add(time.Hour),
			0.1,
			0.8,
		),
	)
	if err != nil {
		t.Fatalf(
			"build valid result: %v",
			err,
		)
	}
	return result
}

func similarityTrajectory(
	id string,
	start time.Time,
	latitudeOffset float64,
	quality float64,
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
				ID: id + "-point-" +
					string(
						rune('a'+index),
					),
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

	end := start.Add(4 * time.Minute)
	segment := trajectory.TrajectorySegment{
		ID:             id + "-segment",
		TrajectoryID:   id,
		SequenceNumber: 0,
		Status:         trajectory.SegmentStatusObserved,
		QualityScore:   quality,
		StartTime:      start,
		EndTime:        end,
		DurationSeconds: int64(
			end.Sub(start) / time.Second,
		),
		StartLatitude:  points[0].Latitude,
		StartLongitude: points[0].Longitude,
		EndLatitude: points[len(points)-1].
			Latitude,
		EndLongitude: points[len(points)-1].
			Longitude,
		PointCount: len(points),
		SourceName: "test-source",
	}

	return trajectory.FlightTrajectory{
		ID:        id,
		StartTime: start,
		EndTime:   end,
		DurationSeconds: int64(
			end.Sub(start) / time.Second,
		),
		SegmentCount: 1,
		PointCount:   len(points),
		QualityScore: quality,
		SourceName:   "test-source",
		Points:       points,
		Segments: []trajectory.TrajectorySegment{
			segment,
		},
		CoverageGaps: []trajectory.CoverageGap{},
	}
}

func zeroLengthTrajectory(
	id string,
	start time.Time,
) trajectory.FlightTrajectory {
	item := similarityTrajectory(
		id,
		start,
		0,
		1,
	)
	for index := range item.Points {
		item.Points[index].Latitude = 40
		item.Points[index].Longitude = 49
	}
	item.Segments[0].StartLatitude = 40
	item.Segments[0].StartLongitude = 49
	item.Segments[0].EndLatitude = 40
	item.Segments[0].EndLongitude = 49
	return item
}

type coordinate struct {
	latitude  float64
	longitude float64
}

func specialTrajectory(
	id string,
	start time.Time,
	coordinates []coordinate,
) trajectory.FlightTrajectory {
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		len(coordinates),
	)
	for index, coordinate := range coordinates {
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID: id + "-point-" +
					string(
						rune('a'+index),
					),
				Latitude:  coordinate.latitude,
				Longitude: coordinate.longitude,
				ObservedAt: start.Add(
					time.Duration(index) *
						time.Minute,
				),
			},
		)
	}
	end := points[len(points)-1].ObservedAt

	return trajectory.FlightTrajectory{
		ID:        id,
		StartTime: start,
		EndTime:   end,
		DurationSeconds: int64(
			end.Sub(start) / time.Second,
		),
		SegmentCount: 1,
		PointCount:   len(points),
		QualityScore: 1,
		SourceName:   "test-source",
		Points:       points,
		Segments: []trajectory.TrajectorySegment{
			{
				ID:           id + "-segment",
				TrajectoryID: id,
				Status:       trajectory.SegmentStatusObserved,
				QualityScore: 1,
				StartTime:    start,
				EndTime:      end,
				DurationSeconds: int64(
					end.Sub(start) /
						time.Second,
				),
				StartLatitude: points[0].
					Latitude,
				StartLongitude: points[0].
					Longitude,
				EndLatitude: points[len(points)-1].
					Latitude,
				EndLongitude: points[len(points)-1].
					Longitude,
				PointCount: len(points),
				SourceName: "test-source",
			},
		},
		CoverageGaps: []trajectory.CoverageGap{},
	}
}

func similarityGap(
	id string,
	trajectoryID string,
	start time.Time,
	end time.Time,
) trajectory.CoverageGap {
	return trajectory.CoverageGap{
		ID:           id,
		TrajectoryID: trajectoryID,
		StartTime:    start,
		EndTime:      end,
		DurationSeconds: int64(
			end.Sub(start) / time.Second,
		),
		DistanceKm: 1,
		Reason: trajectory.
			CoverageGapReasonTimeGap,
	}
}

func componentByName(
	t *testing.T,
	components []Component,
	name ComponentName,
) Component {
	t.Helper()
	for _, component := range components {
		if component.Name == name {
			return component
		}
	}
	t.Fatalf(
		"component %q is missing",
		name,
	)
	return Component{}
}

func assertNoticeCode(
	t *testing.T,
	notices []Notice,
	code string,
) {
	t.Helper()
	for _, notice := range notices {
		if notice.Code == code {
			return
		}
	}
	t.Fatalf(
		"notice %q is missing: %#v",
		code,
		notices,
	)
}

func similarityTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		28,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func TestPublicRankMethodWasRemoved(
	t *testing.T,
) {
	content := strings.Join(
		[]string{
			"Production ranking belongs to projectionneighbors.",
			"Historical Similarity exposes only Compare.",
		},
		" ",
	)
	if !strings.Contains(
		content,
		"Compare",
	) {
		t.Fatal("comparison boundary description missing")
	}
}
