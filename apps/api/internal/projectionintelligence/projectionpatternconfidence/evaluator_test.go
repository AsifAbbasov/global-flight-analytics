package projectionpatternconfidence

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

func TestEvaluateProducesCompleteDistributionConfidence(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	result, err := evaluator.Evaluate(confidenceSelection(3))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Status != StatusComplete ||
		!result.Usable ||
		result.NeighborCount != 3 ||
		result.Level != projectioncontract.ConfidenceLevelHigh {
		t.Fatalf("unexpected complete result: %#v", result)
	}
	if math.Abs(result.MeanSimilarityScore-0.8) > scoreComparisonTolerance ||
		math.Abs(result.MinimumSimilarityScore-0.7) > scoreComparisonTolerance ||
		math.Abs(result.SimilarityStandardDeviation-0.0816496580927726) > 1e-12 {
		t.Fatalf("unexpected similarity distribution: %#v", result)
	}
	if result.MeanCandidateAgeSeconds != 0 {
		t.Fatalf("pattern confidence retained freshness evidence: %#v", result)
	}
	if len(result.Components) != 4 || len(result.SelectedTrajectoryIDs) != 3 {
		t.Fatalf("unexpected evidence: %#v", result)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("result validation error = %v", err)
	}
}

func TestEvaluateProducesLimitedUsablePattern(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(2)
	selection.Status = projectionneighbors.StatusPartial
	selection.SelectionLimit = 3
	result, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Status != StatusLimited ||
		!result.Usable ||
		result.Level == projectioncontract.ConfidenceLevelHigh ||
		!hasConfidenceNotice(result.Limitations, "pattern_support_partial") {
		t.Fatalf("unexpected limited result: %#v", result)
	}
}

func TestEvaluateRejectsInsufficientPatternSupport(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(1)
	selection.Status = projectionneighbors.StatusPartial
	selection.SelectionLimit = 3
	result, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Status != StatusUnavailable ||
		result.Usable ||
		result.Level != projectioncontract.ConfidenceLevelNone ||
		!hasConfidenceNotice(
			result.Limitations,
			"insufficient_historical_neighbor_support",
		) {
		t.Fatalf("unexpected unavailable result: %#v", result)
	}
}

func TestEvaluateRejectsWeakSimilarityFloor(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	selection.Neighbors[2].SimilarityScore = 0.54
	selection.Neighbors[2].SimilarityLevel = historicalsimilarity.LevelForScore(0.54)

	result, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Usable || result.Status != StatusUnavailable ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_similarity_floor_below_minimum",
		) {
		t.Fatalf("weak similarity floor was not blocked: %#v", result)
	}
}

func TestEvaluateRejectsSimilarityDispersion(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	scores := []float64{1, 1, 0.55}
	for index, score := range scores {
		selection.Neighbors[index].SimilarityScore = score
		selection.Neighbors[index].SimilarityLevel = historicalsimilarity.LevelForScore(score)
	}

	result, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Usable || result.Status != StatusUnavailable ||
		result.Level != projectioncontract.ConfidenceLevelNone ||
		result.SimilarityStandardDeviation <=
			validConfidenceConfig().MaximumSimilarityStandardDeviation ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_similarity_dispersion_above_maximum",
		) {
		t.Fatalf("dispersed similarity evidence was not blocked: %#v", result)
	}
}

func TestEvaluateFingerprintIsDeterministic(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	first, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	second, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("second Evaluate() error = %v", err)
	}
	if first.InputFingerprint != second.InputFingerprint {
		t.Fatal("equal inputs produced different fingerprints")
	}
}

func TestEvaluateIgnoresFreshnessEvidence(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	first, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	selection.Neighbors[0].CandidateAge += 24 * time.Hour
	selection.Neighbors[0].CandidateStartTime = selection.Neighbors[0].CandidateStartTime.Add(-24 * time.Hour)
	selection.Neighbors[0].CandidateEndTime = selection.Neighbors[0].CandidateEndTime.Add(-24 * time.Hour)
	selection.Neighbors[0].AnchorObservedAt = selection.Neighbors[0].AnchorObservedAt.Add(-24 * time.Hour)
	selection.Neighbors[0].ContinuationEndTime = selection.Neighbors[0].ContinuationEndTime.Add(-24 * time.Hour)

	changed, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("changed Evaluate() error = %v", err)
	}
	if changed.InputFingerprint != first.InputFingerprint ||
		changed.Score != first.Score ||
		changed.MeanSimilarityScore != first.MeanSimilarityScore {
		t.Fatalf("freshness leaked into pattern confidence: first=%#v changed=%#v", first, changed)
	}
}

func TestEvaluateFingerprintChangesWithSimilarityDistribution(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	first, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	selection.Neighbors[0].SimilarityScore = 0.95
	selection.Neighbors[0].SimilarityLevel = historicalsimilarity.LevelForScore(0.95)
	changed, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("changed Evaluate() error = %v", err)
	}
	if changed.InputFingerprint == first.InputFingerprint {
		t.Fatal("changed similarity distribution was ignored by fingerprint")
	}
	if changed.SimilarityStandardDeviation == first.SimilarityStandardDeviation {
		t.Fatal("changed similarity distribution did not change dispersion evidence")
	}
}

func TestEvaluateFingerprintChangesWithAnchorEvidence(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	first, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	selection.Neighbors[0].AnchorDistanceKM += 7
	changed, err := evaluator.Evaluate(selection)
	if err != nil {
		t.Fatalf("changed Evaluate() error = %v", err)
	}
	if changed.InputFingerprint == first.InputFingerprint || changed.Score == first.Score {
		t.Fatal("changed anchor-distance evidence was ignored")
	}
}

func TestResultCloneDoesNotShareSlices(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	result, err := evaluator.Evaluate(confidenceSelection(3))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	cloned := result.Clone()
	cloned.Components[0].Score = 0
	cloned.SelectedTrajectoryIDs[0] = "changed"
	cloned.Limitations = append(
		cloned.Limitations,
		Notice{Code: "changed", Message: "Changed."},
	)
	if result.Components[0].Score == 0 ||
		result.SelectedTrajectoryIDs[0] == "changed" ||
		len(result.Limitations) == len(cloned.Limitations) {
		t.Fatal("Result.Clone() shared mutable slices")
	}
}

func newConfidenceEvaluator(t *testing.T) *Evaluator {
	t.Helper()
	evaluator, err := New(validConfidenceConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return evaluator
}

func confidenceSelection(neighborCount int) projectionneighbors.Result {
	asOfTime := time.Date(2026, time.July, 15, 17, 0, 0, 0, time.UTC)
	neighbors := make([]projectionneighbors.Neighbor, 0, neighborCount)
	for index := 0; index < neighborCount; index++ {
		score := 0.9 - float64(index)*0.1
		age := time.Duration(index+1) * 24 * time.Hour
		anchorTime := asOfTime.Add(-age - 10*time.Minute)
		neighbors = append(neighbors, projectionneighbors.Neighbor{
			TrajectoryID:    "historical-" + string(rune('a'+index)),
			SimilarityScore: score,
			SimilarityLevel: historicalsimilarity.LevelForScore(score),
			SimilarityInputFingerprint: "sha256:" + strings.Repeat(
				string(rune('a'+index)),
				64,
			),
			AnchorPointIndex:       4,
			AnchorObservedAt:       anchorTime,
			AnchorDistanceKM:       float64((index + 1) * 5),
			CandidateStartTime:     anchorTime.Add(-10 * time.Minute),
			CandidateEndTime:       asOfTime.Add(-age),
			CandidateAge:           age,
			PrefixPointCount:       5,
			ContinuationPointCount: 2,
			ContinuationEndTime:    anchorTime.Add(2 * time.Minute),
		})
	}

	status := projectionneighbors.StatusComplete
	limitations := []projectionneighbors.Notice(nil)
	if neighborCount == 0 {
		status = projectionneighbors.StatusUnavailable
		limitations = []projectionneighbors.Notice{{
			Code:    "historical_neighbor_unavailable",
			Message: "No historical neighbor was selected.",
		}}
	}

	return projectionneighbors.Result{
		Version:                      projectionneighbors.Version,
		Status:                       status,
		CurrentTrajectoryID:          "current",
		AsOfTime:                     asOfTime,
		RequiredContinuationDuration: 2 * time.Minute,
		InputCandidateCount:          neighborCount,
		CheckedCandidateCount:        neighborCount,
		QualifiedCandidateCount:      neighborCount,
		RejectedCandidateCount:       0,
		SelectionLimit:               neighborCount,
		Neighbors:                    neighbors,
		Limitations:                  limitations,
		InputFingerprint:             "sha256:" + strings.Repeat("f", 64),
	}
}

func hasConfidenceNotice(items []Notice, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}
