package projectionneighbors

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
)

func TestSelectAppliesEvaluationBudgetAfterCheapGuards(t *testing.T) {
	config := validSelectorConfig()
	config.MaximumCandidateCount = 1
	config.SelectionLimit = 1
	stub := config.SimilarityEngine.(*similarityEngineStub)
	stub.scores["z-valid"] = 0.90

	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := selectorTestRequest()
	request.Candidates = []trajectory.FlightTrajectory{
		historicalCandidate(
			"a-stale",
			request.AsOfTime.Add(-30*24*time.Hour),
		),
		historicalCandidate(
			"z-valid",
			request.AsOfTime.Add(-24*time.Hour),
		),
	}

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result.Status != StatusComplete ||
		len(result.Neighbors) != 1 ||
		result.Neighbors[0].TrajectoryID != "z-valid" ||
		!hasRejection(result.Rejections, RejectionTooOld) ||
		result.Truncated {
		t.Fatalf("unexpected budget result: %#v", result)
	}
}

func TestSelectRejectsDuplicatesBeforeEvaluationBudget(t *testing.T) {
	config := validSelectorConfig()
	config.MaximumCandidateCount = 1
	config.SelectionLimit = 1
	stub := config.SimilarityEngine.(*similarityEngineStub)
	stub.scores["z-valid"] = 0.90

	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := selectorTestRequest()
	duplicate := historicalCandidate(
		"a-duplicate",
		request.AsOfTime.Add(-24*time.Hour),
	)
	request.Candidates = []trajectory.FlightTrajectory{
		duplicate,
		duplicate,
		historicalCandidate(
			"z-valid",
			request.AsOfTime.Add(-24*time.Hour),
		),
	}

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result.Status != StatusComplete ||
		len(result.Neighbors) != 1 ||
		result.Neighbors[0].TrajectoryID != "z-valid" ||
		countRejections(result.Rejections, RejectionDuplicateCandidate) != 2 ||
		result.Truncated {
		t.Fatalf("unexpected duplicate result: %#v", result)
	}
}

func TestSelectCanonicalizesEqualTimestampPointOrder(t *testing.T) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	firstRequest := selectorTestRequest()
	firstRequest.Candidates = firstRequest.Candidates[:1]
	latest := firstRequest.CurrentTrajectory.Points[len(firstRequest.CurrentTrajectory.Points)-1]
	equalTimePoint := latest
	equalTimePoint.ID = "current-point-0-equal"
	equalTimePoint.Latitude += 5
	equalTimePoint.Longitude += 5
	firstRequest.CurrentTrajectory.Points = append(
		append(
			[]trajectory.TrackPoint4D(nil),
			firstRequest.CurrentTrajectory.Points...,
		),
		equalTimePoint,
	)
	firstRequest.CurrentTrajectory.PointCount =
		len(firstRequest.CurrentTrajectory.Points)

	secondRequest := firstRequest
	secondRequest.CurrentTrajectory = firstRequest.CurrentTrajectory
	secondRequest.CurrentTrajectory.Points = append(
		[]trajectory.TrackPoint4D(nil),
		firstRequest.CurrentTrajectory.Points...,
	)
	last := len(secondRequest.CurrentTrajectory.Points) - 1
	secondRequest.CurrentTrajectory.Points[last-1],
		secondRequest.CurrentTrajectory.Points[last] =
		secondRequest.CurrentTrajectory.Points[last],
		secondRequest.CurrentTrajectory.Points[last-1]

	first, err := selector.Select(firstRequest)
	if err != nil {
		t.Fatalf("first Select() error = %v", err)
	}
	second, err := selector.Select(secondRequest)
	if err != nil {
		t.Fatalf("second Select() error = %v", err)
	}

	if first.InputFingerprint != second.InputFingerprint ||
		len(first.Neighbors) != 1 ||
		len(second.Neighbors) != 1 ||
		first.Neighbors[0].TrajectoryID != second.Neighbors[0].TrajectoryID ||
		first.Neighbors[0].AnchorPointIndex !=
			second.Neighbors[0].AnchorPointIndex ||
		first.Neighbors[0].AnchorDistanceKM !=
			second.Neighbors[0].AnchorDistanceKM {
		t.Fatalf(
			"equal-timestamp permutation changed selection:\nfirst=%#v\nsecond=%#v",
			first,
			second,
		)
	}
}

func TestSelectReturnsErrorForSimilarityEngineFailure(t *testing.T) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	stub := config.SimilarityEngine.(*similarityEngineStub)
	stub.errors = map[string]error{
		"historical-a": errors.New("similarity subsystem offline"),
	}

	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request := selectorTestRequest()
	request.Candidates = request.Candidates[1:2]

	_, err = selector.Select(request)
	if !errors.Is(err, ErrSimilarityEngineFailed) {
		t.Fatalf("Select() error = %v, want %v", err, ErrSimilarityEngineFailed)
	}
}

func TestSelectKeepsCandidateSpecificSimilarityFailureAsRejection(t *testing.T) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	stub := config.SimilarityEngine.(*similarityEngineStub)
	stub.errors = map[string]error{
		"historical-a": fmt.Errorf(
			"%w: fixture",
			historicalsimilarity.ErrCandidateNotComparable,
		),
	}

	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request := selectorTestRequest()
	request.Candidates = request.Candidates[1:2]

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result.Status != StatusUnavailable ||
		!hasRejection(result.Rejections, RejectionSimilarityUnavailable) {
		t.Fatalf("unexpected candidate failure result: %#v", result)
	}
}

func TestSelectRejectsInvalidSimilarityConsumerEvidence(t *testing.T) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	stub := config.SimilarityEngine.(*similarityEngineStub)
	stub.results = map[string]historicalsimilarity.Result{
		"historical-a": {
			Version:               historicalsimilarity.Version,
			ReferenceTrajectoryID: "current",
			CandidateTrajectoryID: "historical-a#projection-prefix",
			Score:                 0.90,
			Level:                 historicalsimilarity.LevelHigh,
			ReferencePointCount:   5,
			CandidatePointCount:   5,
			SampleCount:           4,
			InputFingerprint:      "invalid",
		},
	}

	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	request := selectorTestRequest()
	request.Candidates = request.Candidates[1:2]

	_, err = selector.Select(request)
	if !errors.Is(err, ErrSimilarityEvidenceInvalid) {
		t.Fatalf(
			"Select() error = %v, want %v",
			err,
			ErrSimilarityEvidenceInvalid,
		)
	}
}

func TestConfigRejectsSelectionLimitAboveCandidateBudget(t *testing.T) {
	config := validSelectorConfig()
	config.MaximumCandidateCount = 1
	config.SelectionLimit = 2

	if err := config.Validate(); !errors.Is(err, ErrSelectionLimitExceedsCandidateBudget) {
		t.Fatalf(
			"Validate() error = %v, want %v",
			err,
			ErrSelectionLimitExceedsCandidateBudget,
		)
	}
}

func TestResultValidationEnforcesCrossFieldEvidence(t *testing.T) {
	selector := newTestSelector(t)
	result, err := selector.Select(selectorTestRequest())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Result)
	}{
		{
			name: "anchor and prefix",
			mutate: func(item *Result) {
				item.Neighbors[0].PrefixPointCount++
			},
		},
		{
			name: "candidate age",
			mutate: func(item *Result) {
				item.Neighbors[0].CandidateAge++
			},
		},
		{
			name: "truncation counts",
			mutate: func(item *Result) {
				item.Truncated = !item.Truncated
			},
		},
		{
			name: "rejection code",
			mutate: func(item *Result) {
				item.Rejections[0].Code = "unknown_rejection"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := result.Clone()
			test.mutate(&changed)
			if err := changed.Validate(); err == nil {
				t.Fatalf("Validate() error = nil for %#v", changed)
			}
		})
	}
}

func countRejections(items []Rejection, code RejectionCode) int {
	count := 0
	for _, item := range items {
		if item.Code == code {
			count++
		}
	}
	return count
}
