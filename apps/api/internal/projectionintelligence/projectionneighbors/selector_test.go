package projectionneighbors

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
)

type similarityEngineStub struct {
	scores map[string]float64
	calls  []string
}

func (
	stub *similarityEngineStub,
) Compare(
	reference trajectory.FlightTrajectory,
	candidate trajectory.FlightTrajectory,
) (historicalsimilarity.Result, error) {
	candidateID := strings.TrimSuffix(
		candidate.ID,
		"#projection-prefix",
	)
	stub.calls = append(
		stub.calls,
		reference.ID+"->"+candidateID,
	)

	score, exists := stub.scores[candidateID]
	if !exists {
		return historicalsimilarity.Result{},
			errors.New(
				"candidate similarity unavailable",
			)
	}

	return historicalsimilarity.Result{
		Version:               historicalsimilarity.Version,
		ReferenceTrajectoryID: reference.ID,
		CandidateTrajectoryID: candidate.ID,
		Score:                 score,
		Level: historicalsimilarity.
			LevelForScore(score),
		ReferencePointCount: len(reference.Points),
		CandidatePointCount: len(candidate.Points),
		SampleCount:         4,
		InputFingerprint: "sha256:" +
			strings.Repeat("a", 64),
	}, nil
}

func TestSelectIntegratesHistoricalSimilarityEngine(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.SimilarityEngine =
		historicalsimilarity.NewDefault()
	config.SelectionLimit = 1
	config.MinimumSimilarityScore = 0

	selector, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := selectorTestRequest()
	request.Candidates =
		[]trajectory.FlightTrajectory{
			historicalCandidate(
				"historical-integration",
				request.AsOfTime.Add(
					-24*time.Hour,
				),
			),
		}

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf(
			"Select() error = %v",
			err,
		)
	}
	if len(result.Neighbors) != 1 ||
		result.Neighbors[0].
			TrajectoryID !=
			"historical-integration" {
		t.Fatalf(
			"unexpected integration result: %#v",
			result,
		)
	}
}

func TestSelectRanksHistoricalNeighborsDeterministically(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()

	first, err := selector.Select(request)
	if err != nil {
		t.Fatalf(
			"Select() error = %v",
			err,
		)
	}

	request.Candidates[0],
		request.Candidates[2] =
		request.Candidates[2],
		request.Candidates[0]
	second, err := selector.Select(request)
	if err != nil {
		t.Fatalf(
			"Select() after reorder error = %v",
			err,
		)
	}

	if first.Status != StatusComplete {
		t.Fatalf(
			"status = %q, want complete",
			first.Status,
		)
	}
	if len(first.Neighbors) != 2 {
		t.Fatalf(
			"neighbor count = %d, want 2",
			len(first.Neighbors),
		)
	}
	if first.Neighbors[0].TrajectoryID !=
		"historical-a" ||
		first.Neighbors[1].TrajectoryID !=
			"historical-b" {
		t.Fatalf(
			"unexpected neighbor order: %#v",
			first.Neighbors,
		)
	}
	if first.QualifiedCandidateCount != 3 ||
		first.RejectedCandidateCount != 2 ||
		first.CheckedCandidateCount != 5 {
		t.Fatalf(
			"unexpected counts: %#v",
			first,
		)
	}
	if first.Neighbors[0].AnchorPointIndex != 4 ||
		first.Neighbors[0].PrefixPointCount != 5 ||
		first.Neighbors[0].
			ContinuationPointCount != 2 {
		t.Fatalf(
			"unexpected anchor evidence: %#v",
			first.Neighbors[0],
		)
	}
	if first.InputFingerprint !=
		second.InputFingerprint {
		t.Fatal(
			"candidate input order changed the fingerprint",
		)
	}
	if first.Neighbors[0].TrajectoryID !=
		second.Neighbors[0].TrajectoryID ||
		first.Neighbors[1].TrajectoryID !=
			second.Neighbors[1].TrajectoryID {
		t.Fatal(
			"candidate input order changed selection",
		)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf(
			"result validation error = %v",
			err,
		)
	}
}

func TestSelectExcludesFuturePointsWithoutFingerprintLeakage(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()

	withoutFuture, err := selector.Select(
		request,
	)
	if err != nil {
		t.Fatalf(
			"Select() error = %v",
			err,
		)
	}

	futureCurrentPoint :=
		request.CurrentTrajectory.Points[len(request.CurrentTrajectory.Points)-1]
	futureCurrentPoint.ID =
		"current-future"
	futureCurrentPoint.ObservedAt =
		request.AsOfTime.Add(time.Minute)
	futureCurrentPoint.Latitude = 80
	request.CurrentTrajectory.Points = append(
		request.CurrentTrajectory.Points,
		futureCurrentPoint,
	)

	futureCandidatePoint :=
		request.Candidates[0].Points[len(request.Candidates[0].Points)-1]
	futureCandidatePoint.ID =
		"candidate-future"
	futureCandidatePoint.ObservedAt =
		request.AsOfTime.Add(time.Minute)
	futureCandidatePoint.Longitude = 100
	request.Candidates[0].Points = append(
		request.Candidates[0].Points,
		futureCandidatePoint,
	)

	withFuture, err := selector.Select(request)
	if err != nil {
		t.Fatalf(
			"Select() with future points error = %v",
			err,
		)
	}

	if withoutFuture.InputFingerprint !=
		withFuture.InputFingerprint {
		t.Fatal(
			"future observations changed the as-of input fingerprint",
		)
	}
	if !hasNotice(
		withFuture.Limitations,
		"future_current_points_excluded",
	) {
		t.Fatalf(
			"future current-point limitation missing: %#v",
			withFuture.Limitations,
		)
	}
	if withFuture.Neighbors[0].
		TrajectoryID !=
		withoutFuture.Neighbors[0].
			TrajectoryID {
		t.Fatal(
			"future observations changed selected neighbor",
		)
	}
}

func TestSelectRejectsNonHistoricalAndStaleCandidates(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	config.MaximumCandidateAge =
		36 * time.Hour

	selector, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	request := selectorTestRequest()
	request.Candidates = []trajectory.FlightTrajectory{
		historicalCandidate(
			"overlapping",
			request.AsOfTime.Add(
				-2*time.Minute,
			),
		),
		historicalCandidate(
			"stale",
			request.AsOfTime.Add(
				-72*time.Hour,
			),
		),
	}

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf(
			"Select() error = %v",
			err,
		)
	}

	if result.Status != StatusUnavailable ||
		len(result.Neighbors) != 0 ||
		!hasRejection(
			result.Rejections,
			RejectionNotHistorical,
		) ||
		!hasRejection(
			result.Rejections,
			RejectionTooOld,
		) {
		t.Fatalf(
			"unexpected rejection result: %#v",
			result,
		)
	}
}

func TestSelectRejectsInvalidRequest(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()

	request.CurrentTrajectory.ID = ""
	_, err := selector.Select(request)
	if !errors.Is(
		err,
		ErrCurrentTrajectoryIDRequired,
	) {
		t.Fatalf(
			"trajectory id error = %v",
			err,
		)
	}

	request = selectorTestRequest()
	request.AsOfTime = time.Time{}
	_, err = selector.Select(request)
	if !errors.Is(
		err,
		ErrAsOfTimeRequired,
	) {
		t.Fatalf(
			"as-of error = %v",
			err,
		)
	}

	request = selectorTestRequest()
	request.RequiredContinuationDuration = 0
	_, err = selector.Select(request)
	if !errors.Is(
		err,
		ErrContinuationDurationInvalid,
	) {
		t.Fatalf(
			"duration error = %v",
			err,
		)
	}
}

func TestResultCloneDoesNotShareSlices(
	t *testing.T,
) {
	selector := newTestSelector(t)
	result, err := selector.Select(
		selectorTestRequest(),
	)
	if err != nil {
		t.Fatalf(
			"Select() error = %v",
			err,
		)
	}

	cloned := result.Clone()
	cloned.Neighbors[0].TrajectoryID =
		"changed"
	cloned.Rejections[0].Code =
		"changed"
	cloned.Limitations[0].Code =
		"changed"

	if result.Neighbors[0].TrajectoryID ==
		"changed" ||
		result.Rejections[0].Code ==
			"changed" ||
		result.Limitations[0].Code ==
			"changed" {
		t.Fatal(
			"Result.Clone() shared mutable slices",
		)
	}
}

func newTestSelector(
	t *testing.T,
) *Selector {
	t.Helper()

	selector, err := New(
		validSelectorConfig(),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return selector
}

func validSelectorConfig() Config {
	return Config{
		SimilarityEngine: &similarityEngineStub{
			scores: map[string]float64{
				"historical-a":   0.90,
				"historical-b":   0.80,
				"historical-c":   0.70,
				"historical-low": 0.40,
				"overlapping":    0.90,
				"stale":          0.90,
			},
		},
		SimilarityPolicyKey: historicalsimilarity.Version +
			":selector-test-policy",

		MinimumCurrentPointCount: 4,
		MaximumCandidateCount:    20,
		SelectionLimit:           2,

		MinimumSimilarityScore:  0.60,
		MaximumAnchorDistanceKM: 50,
		MaximumCandidateAge:     7 * 24 * time.Hour,
	}
}

func selectorTestRequest() Request {
	asOfTime := time.Date(
		2026,
		time.July,
		15,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	return Request{
		CurrentTrajectory: currentTrajectory(asOfTime),
		Candidates: []trajectory.FlightTrajectory{
			historicalCandidate(
				"historical-c",
				asOfTime.Add(
					-24*time.Hour,
				),
			),
			historicalCandidate(
				"historical-a",
				asOfTime.Add(
					-24*time.Hour,
				),
			),
			historicalCandidate(
				"historical-b",
				asOfTime.Add(
					-24*time.Hour,
				),
			),
			historicalCandidate(
				"historical-low",
				asOfTime.Add(
					-24*time.Hour,
				),
			),
			currentTrajectory(asOfTime),
		},
		AsOfTime:                     asOfTime,
		RequiredContinuationDuration: 2 * time.Minute,
	}
}

func currentTrajectory(
	asOfTime time.Time,
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
				ID: "current-point-" +
					string(
						rune('0'+index),
					),
				Latitude: 40 +
					float64(index)*0.01,
				Longitude: 49 +
					float64(index)*0.01,
				ObservedAt: asOfTime.Add(
					time.Duration(
						index-4,
					) * time.Minute,
				),
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:           "current",
		StartTime:    points[0].ObservedAt,
		EndTime:      points[len(points)-1].ObservedAt,
		PointCount:   len(points),
		Points:       points,
		QualityScore: 0.9,
	}
}

func historicalCandidate(
	id string,
	endTime time.Time,
) trajectory.FlightTrajectory {
	startTime := endTime.Add(
		-7 * time.Minute,
	)
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		8,
	)
	for index := 0; index < 8; index++ {
		latitude := 39.7 +
			float64(index)*0.085
		longitude := 48.7 +
			float64(index)*0.085
		if index == 4 {
			latitude = 40.04
			longitude = 49.04
		}

		points = append(
			points,
			trajectory.TrackPoint4D{
				ID: id + "-point-" +
					string(
						rune('0'+index),
					),
				Latitude:  latitude,
				Longitude: longitude,
				ObservedAt: startTime.Add(
					time.Duration(index) *
						time.Minute,
				),
			},
		)
	}

	return trajectory.FlightTrajectory{
		ID:           id,
		StartTime:    points[0].ObservedAt,
		EndTime:      points[len(points)-1].ObservedAt,
		PointCount:   len(points),
		Points:       points,
		QualityScore: 0.85,
	}
}

func hasNotice(
	items []Notice,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}

func hasRejection(
	items []Rejection,
	code RejectionCode,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}
