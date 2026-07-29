package projectionneighbors

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestConfigValidateRejectsNegativeContinuationGap(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.MaximumContinuationGap = -time.Second

	err := config.Validate()
	if !errors.Is(err, ErrMaximumContinuationGapInvalid) {
		t.Fatalf(
			"Validate() error = %v, want %v",
			err,
			ErrMaximumContinuationGapInvalid,
		)
	}
}

func TestSelectionFingerprintIncludesContinuationGapPolicy(
	t *testing.T,
) {
	request := selectorTestRequest()

	firstConfig := validSelectorConfig()
	firstConfig.MaximumContinuationGap = 90 * time.Second
	firstSelector, err := New(firstConfig)
	if err != nil {
		t.Fatalf("New() first error = %v", err)
	}
	first, err := firstSelector.Select(request)
	if err != nil {
		t.Fatalf("Select() first error = %v", err)
	}

	secondConfig := validSelectorConfig()
	secondConfig.MaximumContinuationGap = 2 * time.Minute
	secondSelector, err := New(secondConfig)
	if err != nil {
		t.Fatalf("New() second error = %v", err)
	}
	second, err := secondSelector.Select(request)
	if err != nil {
		t.Fatalf("Select() second error = %v", err)
	}

	if first.InputFingerprint == second.InputFingerprint {
		t.Fatal("continuation-gap policy did not change the input fingerprint")
	}
}

func TestSelectRejectsDiscontinuousContinuation(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	config.MaximumContinuationGap = 90 * time.Second
	config.MinimumSimilarityScore = 0
	stub := config.SimilarityEngine.(*similarityEngineStub)
	stub.scores["gapped"] = 0.9

	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := selectorTestRequest()
	candidate := historicalCandidate(
		"gapped",
		request.AsOfTime.Add(-24*time.Hour),
	)
	candidate.Points = append(
		[]trajectory.TrackPoint4D(nil),
		candidate.Points[:6]...,
	)
	candidate.Points[5].ObservedAt = candidate.Points[4].ObservedAt.Add(
		2 * time.Hour,
	)
	candidate.PointCount = len(candidate.Points)
	candidate.EndTime = candidate.Points[len(candidate.Points)-1].ObservedAt
	request.Candidates = []trajectory.FlightTrajectory{candidate}

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result.Status != StatusUnavailable ||
		!hasRejection(
			result.Rejections,
			RejectionContinuationDiscontinuous,
		) {
		t.Fatalf("unexpected discontinuity result: %#v", result)
	}
	if len(stub.calls) != 0 {
		t.Fatalf("similarity engine calls = %v, want none", stub.calls)
	}
}

func TestFindAnchorPublishesContinuousEvidence(
	t *testing.T,
) {
	request := selectorTestRequest()
	candidate := historicalCandidate(
		"continuous",
		request.AsOfTime.Add(-24*time.Hour),
	)
	current := request.CurrentTrajectory.Points[len(request.CurrentTrajectory.Points)-1]

	search := findAnchor(
		current,
		candidate.Points,
		4,
		2*time.Minute,
		90*time.Second,
	)
	if !search.Found() {
		t.Fatalf("findAnchor() failure = %q", search.Failure)
	}
	if search.Evidence.PointIndex != 4 ||
		search.Evidence.ContinuationPointCount != 2 ||
		search.Evidence.MaximumObservedGap != time.Minute {
		t.Fatalf("unexpected anchor evidence: %#v", search.Evidence)
	}
}

func TestFindAnchorHandlesLargeTrajectory(
	t *testing.T,
) {
	const pointCount = 10000
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	points := make([]trajectory.TrackPoint4D, 0, pointCount)
	for index := 0; index < pointCount; index++ {
		points = append(
			points,
			trajectory.TrackPoint4D{
				ID:         fmt.Sprintf("point-%05d", index),
				Latitude:   40 + float64(index)*0.000001,
				Longitude:  49 + float64(index)*0.000001,
				ObservedAt: start.Add(time.Duration(index) * 10 * time.Second),
			},
		)
	}
	current := points[5000]

	search := findAnchor(
		current,
		points,
		4,
		5*time.Minute,
		20*time.Second,
	)
	if !search.Found() {
		t.Fatalf("findAnchor() failure = %q", search.Failure)
	}
	if search.Evidence.PointIndex != 5000 ||
		search.Evidence.ContinuationPointCount != 30 {
		t.Fatalf("unexpected large-input evidence: %#v", search.Evidence)
	}
}

func TestDefaultContinuationGapIsPositive(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.MaximumContinuationGap = 0

	if got := config.effectiveMaximumContinuationGap(); got != DefaultMaximumContinuationGap {
		t.Fatalf(
			"effectiveMaximumContinuationGap() = %s, want %s",
			got,
			DefaultMaximumContinuationGap,
		)
	}
}
