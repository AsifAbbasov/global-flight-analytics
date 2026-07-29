package projectionpatternconfidence

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

func TestEvaluateWithContinuationsProducesCompleteConfidence(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	result, err := evaluator.EvaluateWithContinuations(
		selection,
		confidenceCandidates(selection),
	)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	if result.Status != StatusComplete ||
		!result.Usable ||
		result.NeighborCount != 3 ||
		result.Level != projectioncontract.ConfidenceLevelHigh ||
		!result.ContinuationAgreementKnown {
		t.Fatalf("unexpected complete result: %#v", result)
	}
	if result.ContinuationAgreementSampleCount != 4 ||
		result.ContinuationAgreementPairCount != 3 ||
		result.ContinuationComparisonCount != 12 ||
		result.MaximumContinuationDivergenceMPS > 1e-9 {
		t.Fatalf("unexpected continuation agreement: %#v", result)
	}
	if len(result.Components) != 5 ||
		result.Components[4].Name != ComponentContinuationAgreement ||
		math.Abs(result.Components[4].Score-1) > scoreComparisonTolerance {
		t.Fatalf("unexpected component catalog: %#v", result.Components)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("result validation error = %v", err)
	}
}

func TestEvaluateWithoutContinuationsCannotAuthorize(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	result, err := evaluator.Evaluate(confidenceSelection(3))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Status != StatusUnavailable ||
		result.Usable ||
		result.Level != projectioncontract.ConfidenceLevelNone ||
		result.ContinuationAgreementKnown ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_continuation_agreement_unavailable",
		) {
		t.Fatalf("legacy evaluation authorized an unverified continuation: %#v", result)
	}
}

func TestEvaluateWithContinuationsRejectsOpposingRoutes(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	candidates := confidenceCandidates(selection)
	setContinuationDelta(&candidates[2], -0.10, -0.10)

	result, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	if result.Status != StatusUnavailable ||
		result.Usable ||
		result.Level != projectioncontract.ConfidenceLevelNone ||
		result.MaximumContinuationDivergenceMPS <=
			result.Policy.MaximumContinuationDivergenceMPS ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_continuation_divergence_above_maximum",
		) {
		t.Fatalf("opposing continuations were not blocked: %#v", result)
	}
}

func TestEvaluateWithContinuationsRejectsIntermediateRouteConflict(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	candidates := confidenceCandidates(selection)
	anchor := candidates[2].Points[4]
	candidates[2].Points[5].Latitude = anchor.Latitude + 0.20
	candidates[2].Points[5].Longitude = anchor.Longitude - 0.20
	// Rejoin the common endpoint so an endpoint-only policy would miss the conflict.
	candidates[2].Points[6].Latitude = anchor.Latitude + 0.02
	candidates[2].Points[6].Longitude = anchor.Longitude + 0.02

	result, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	if result.Usable ||
		result.Status != StatusUnavailable ||
		result.MaximumContinuationDivergenceMPS <=
			result.Policy.MaximumContinuationDivergenceMPS {
		t.Fatalf("intermediate route conflict was not blocked: %#v", result)
	}
}

func TestEvaluateWithContinuationsChangesFingerprintForRouteMutation(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	candidates := confidenceCandidates(selection)
	first, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("first evaluation error = %v", err)
	}

	setContinuationDelta(&candidates[2], 0.011, 0.009)
	changed, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("changed evaluation error = %v", err)
	}
	if changed.InputFingerprint == first.InputFingerprint ||
		changed.Score == first.Score ||
		changed.MeanContinuationDivergenceMPS ==
			first.MeanContinuationDivergenceMPS {
		t.Fatalf("continuation mutation was ignored: first=%#v changed=%#v", first, changed)
	}
}

func TestEvaluateWithContinuationsIgnoresCandidateOrder(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	candidates := confidenceCandidates(selection)
	first, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("first evaluation error = %v", err)
	}
	candidates[0], candidates[2] = candidates[2], candidates[0]
	changed, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("reordered evaluation error = %v", err)
	}
	if changed.InputFingerprint != first.InputFingerprint ||
		changed.Score != first.Score {
		t.Fatalf("candidate order affected deterministic evidence: first=%#v changed=%#v", first, changed)
	}
}

func TestEvaluateWithContinuationsRejectsMissingCandidate(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	_, err := evaluator.EvaluateWithContinuations(
		selection,
		confidenceCandidates(selection)[:2],
	)
	if !errors.Is(err, ErrContinuationCandidatesInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrContinuationCandidatesInvalid)
	}
}

func TestEvaluateWithContinuationsUsesInterpolatedSamples(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	candidates := confidenceCandidates(selection)
	for index := range candidates {
		points := candidates[index].Points
		candidates[index].Points = append(
			append([]trajectory.TrackPoint4D(nil), points[:5]...),
			points[6],
		)
		candidates[index].PointCount = len(candidates[index].Points)
	}

	result, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	if !result.Usable || result.MaximumContinuationDivergenceMPS > 1e-9 {
		t.Fatalf("interpolated continuation samples were inconsistent: %#v", result)
	}
}

func TestEvaluateWithContinuationsRejectsWeakSimilarityFloor(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	selection.Neighbors[2].SimilarityScore = 0.54
	selection.Neighbors[2].SimilarityLevel = historicalsimilarity.LevelForScore(0.54)

	result, err := evaluator.EvaluateWithContinuations(
		selection,
		confidenceCandidates(selection),
	)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	if result.Usable || result.Status != StatusUnavailable ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_similarity_floor_below_minimum",
		) {
		t.Fatalf("weak similarity floor was not blocked: %#v", result)
	}
}

func TestEvaluateWithContinuationsRejectsSimilarityDispersion(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	scores := []float64{1, 1, 0.55}
	for index, score := range scores {
		selection.Neighbors[index].SimilarityScore = score
		selection.Neighbors[index].SimilarityLevel = historicalsimilarity.LevelForScore(score)
	}

	result, err := evaluator.EvaluateWithContinuations(
		selection,
		confidenceCandidates(selection),
	)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	if result.Usable || result.Status != StatusUnavailable ||
		result.SimilarityStandardDeviation <=
			validConfidenceConfig().MaximumSimilarityStandardDeviation ||
		!hasConfidenceNotice(
			result.Limitations,
			"pattern_similarity_dispersion_above_maximum",
		) {
		t.Fatalf("dispersed similarity evidence was not blocked: %#v", result)
	}
}

func TestEvaluateWithContinuationsIgnoresFreshnessEvidence(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	candidates := confidenceCandidates(selection)
	first, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("first evaluation error = %v", err)
	}

	for index := range selection.Neighbors {
		selection.Neighbors[index].CandidateAge += 24 * time.Hour
		selection.Neighbors[index].CandidateStartTime =
			selection.Neighbors[index].CandidateStartTime.Add(-24 * time.Hour)
		selection.Neighbors[index].CandidateEndTime =
			selection.Neighbors[index].CandidateEndTime.Add(-24 * time.Hour)
		selection.Neighbors[index].AnchorObservedAt =
			selection.Neighbors[index].AnchorObservedAt.Add(-24 * time.Hour)
		selection.Neighbors[index].ContinuationEndTime =
			selection.Neighbors[index].ContinuationEndTime.Add(-24 * time.Hour)
		for pointIndex := range candidates[index].Points {
			candidates[index].Points[pointIndex].ObservedAt =
				candidates[index].Points[pointIndex].ObservedAt.Add(-24 * time.Hour)
		}
		candidates[index].StartTime = candidates[index].StartTime.Add(-24 * time.Hour)
		candidates[index].EndTime = candidates[index].EndTime.Add(-24 * time.Hour)
	}

	changed, err := evaluator.EvaluateWithContinuations(selection, candidates)
	if err != nil {
		t.Fatalf("changed evaluation error = %v", err)
	}
	if changed.InputFingerprint != first.InputFingerprint ||
		changed.Score != first.Score {
		t.Fatalf("freshness leaked into pattern confidence: first=%#v changed=%#v", first, changed)
	}
}

func TestResultCloneDoesNotShareSlices(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	selection := confidenceSelection(3)
	result, err := evaluator.EvaluateWithContinuations(
		selection,
		confidenceCandidates(selection),
	)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
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
		anchorTime := asOfTime.Add(-age - 2*time.Minute)
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
			CandidateStartTime:     anchorTime.Add(-4 * time.Minute),
			CandidateEndTime:       anchorTime.Add(2 * time.Minute),
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

func confidenceCandidates(selection projectionneighbors.Result) []trajectory.FlightTrajectory {
	candidates := make([]trajectory.FlightTrajectory, 0, len(selection.Neighbors))
	for _, neighbor := range selection.Neighbors {
		points := make([]trajectory.TrackPoint4D, 0, 7)
		for index := 0; index < 5; index++ {
			points = append(points, trajectory.TrackPoint4D{
				ID:         neighbor.TrajectoryID + "-prefix-" + string(rune('0'+index)),
				Latitude:   40 - 0.01*float64(4-index),
				Longitude:  50 - 0.01*float64(4-index),
				ObservedAt: neighbor.AnchorObservedAt.Add(time.Duration(index-4) * time.Minute),
				SourceName: "test-source",
			})
		}
		points = append(points,
			trajectory.TrackPoint4D{
				ID:         neighbor.TrajectoryID + "-continuation-1",
				Latitude:   40.01,
				Longitude:  50.01,
				ObservedAt: neighbor.AnchorObservedAt.Add(time.Minute),
				SourceName: "test-source",
			},
			trajectory.TrackPoint4D{
				ID:         neighbor.TrajectoryID + "-continuation-2",
				Latitude:   40.02,
				Longitude:  50.02,
				ObservedAt: neighbor.AnchorObservedAt.Add(2 * time.Minute),
				SourceName: "test-source",
			},
		)
		candidates = append(candidates, trajectory.FlightTrajectory{
			ID:         neighbor.TrajectoryID,
			StartTime:  points[0].ObservedAt,
			EndTime:    points[len(points)-1].ObservedAt,
			PointCount: len(points),
			SourceName: "test-source",
			Points:     points,
		})
	}
	return candidates
}

func setContinuationDelta(
	candidate *trajectory.FlightTrajectory,
	latitudeDelta float64,
	longitudeDelta float64,
) {
	anchor := candidate.Points[4]
	candidate.Points[5].Latitude = anchor.Latitude + latitudeDelta
	candidate.Points[5].Longitude = anchor.Longitude + longitudeDelta
	candidate.Points[6].Latitude = anchor.Latitude + 2*latitudeDelta
	candidate.Points[6].Longitude = anchor.Longitude + 2*longitudeDelta
}

func hasConfidenceNotice(items []Notice, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}
