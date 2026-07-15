package projectionpatternconfidence

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

func TestEvaluateProducesCompletePatternConfidence(
	t *testing.T,
) {
	evaluator := newConfidenceEvaluator(t)

	result, err := evaluator.Evaluate(
		confidenceSelection(3),
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Status != StatusComplete ||
		!result.Usable ||
		result.NeighborCount != 3 ||
		result.Level !=
			projectioncontract.
				ConfidenceLevelHigh {
		t.Fatalf(
			"unexpected complete result: %#v",
			result,
		)
	}
	if result.Score < 0.84 ||
		result.Score > 0.85 {
		t.Fatalf(
			"score = %f, want approximately 0.842857",
			result.Score,
		)
	}
	if len(result.Components) != 4 ||
		len(result.SelectedTrajectoryIDs) != 3 {
		t.Fatalf(
			"unexpected evidence: %#v",
			result,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf(
			"result validation error = %v",
			err,
		)
	}
}

func TestEvaluateProducesLimitedUsablePattern(
	t *testing.T,
) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(2)
	selection.Status =
		projectionneighbors.StatusPartial
	selection.SelectionLimit = 3

	result, err := evaluator.Evaluate(
		selection,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Status != StatusLimited ||
		!result.Usable ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_support_partial",
		) {
		t.Fatalf(
			"unexpected limited result: %#v",
			result,
		)
	}
}

func TestEvaluateRejectsInsufficientPatternSupport(
	t *testing.T,
) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(1)
	selection.Status =
		projectionneighbors.StatusPartial
	selection.SelectionLimit = 3

	result, err := evaluator.Evaluate(
		selection,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Status != StatusUnavailable ||
		result.Usable ||
		!hasConfidenceNotice(
			result.Limitations,
			"insufficient_historical_neighbor_support",
		) {
		t.Fatalf(
			"unexpected unavailable result: %#v",
			result,
		)
	}
}

func TestEvaluateFingerprintIsDeterministicAndSensitive(
	t *testing.T,
) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)

	first, err := evaluator.Evaluate(
		selection,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}
	second, err := evaluator.Evaluate(
		selection,
	)
	if err != nil {
		t.Fatalf(
			"second Evaluate() error = %v",
			err,
		)
	}
	if first.InputFingerprint !=
		second.InputFingerprint {
		t.Fatal(
			"equal inputs produced different fingerprints",
		)
	}

	selection.InputFingerprint =
		"sha256:" +
			strings.Repeat("b", 64)
	changed, err := evaluator.Evaluate(
		selection,
	)
	if err != nil {
		t.Fatalf(
			"changed Evaluate() error = %v",
			err,
		)
	}
	if changed.InputFingerprint ==
		first.InputFingerprint {
		t.Fatal(
			"changed selection fingerprint was ignored",
		)
	}
}

func TestResultCloneDoesNotShareSlices(
	t *testing.T,
) {
	evaluator := newConfidenceEvaluator(t)
	result, err := evaluator.Evaluate(
		confidenceSelection(3),
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	cloned := result.Clone()
	cloned.Components[0].Score = 0
	cloned.SelectedTrajectoryIDs[0] =
		"changed"
	cloned.Limitations = append(
		cloned.Limitations,
		Notice{
			Code:    "changed",
			Message: "Changed.",
		},
	)

	if result.Components[0].Score == 0 ||
		result.SelectedTrajectoryIDs[0] ==
			"changed" ||
		len(result.Limitations) ==
			len(cloned.Limitations) {
		t.Fatal(
			"Result.Clone() shared mutable slices",
		)
	}
}

func newConfidenceEvaluator(
	t *testing.T,
) *Evaluator {
	t.Helper()

	evaluator, err := New(
		validConfidenceConfig(),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return evaluator
}

func confidenceSelection(
	neighborCount int,
) projectionneighbors.Result {
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
	neighbors := make(
		[]projectionneighbors.Neighbor,
		0,
		neighborCount,
	)
	for index := 0; index < neighborCount; index++ {
		score := 0.9 -
			float64(index)*0.1
		age := time.Duration(
			index+1,
		) * 24 * time.Hour
		anchorTime := asOfTime.Add(
			-age - 10*time.Minute,
		)
		neighbors = append(
			neighbors,
			projectionneighbors.Neighbor{
				TrajectoryID: "historical-" +
					string(
						rune('a'+index),
					),
				SimilarityScore: score,
				SimilarityLevel: historicalsimilarity.
					LevelForScore(score),
				SimilarityInputFingerprint: "sha256:" +
					strings.Repeat(
						string(
							rune('a'+index),
						),
						64,
					),
				AnchorPointIndex: 4,
				AnchorObservedAt: anchorTime,
				AnchorDistanceKM: float64(
					(index + 1) * 5,
				),
				CandidateStartTime: anchorTime.Add(
					-10 * time.Minute,
				),
				CandidateEndTime:       asOfTime.Add(-age),
				CandidateAge:           age,
				PrefixPointCount:       5,
				ContinuationPointCount: 2,
				ContinuationEndTime: anchorTime.Add(
					2 * time.Minute,
				),
			},
		)
	}

	status :=
		projectionneighbors.StatusComplete
	if neighborCount == 0 {
		status =
			projectionneighbors.
				StatusUnavailable
	}
	limitations :=
		[]projectionneighbors.Notice(nil)
	if neighborCount == 0 {
		limitations =
			[]projectionneighbors.Notice{
				{
					Code:    "historical_neighbor_unavailable",
					Message: "No historical neighbor was selected.",
				},
			}
	}

	return projectionneighbors.Result{
		Version:                      projectionneighbors.Version,
		Status:                       status,
		CurrentTrajectoryID:          "current",
		AsOfTime:                     asOfTime,
		RequiredContinuationDuration: 2 * time.Minute,

		InputCandidateCount:     neighborCount,
		CheckedCandidateCount:   neighborCount,
		QualifiedCandidateCount: neighborCount,
		RejectedCandidateCount:  0,

		SelectionLimit: neighborCount,
		Neighbors:      neighbors,
		Limitations:    limitations,
		InputFingerprint: "sha256:" +
			strings.Repeat("f", 64),
	}
}

func hasConfidenceNotice(
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
