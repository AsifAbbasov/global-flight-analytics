package projectionfreshness

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

func TestEvaluateAllowsFreshHistoricalPattern(
	t *testing.T,
) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern :=
		freshnessFixtures(
			[]time.Duration{
				24 * time.Hour,
				48 * time.Hour,
				72 * time.Hour,
			},
		)

	first, err := evaluator.Evaluate(
		selection,
		pattern,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}
	second, err := evaluator.Evaluate(
		selection,
		pattern,
	)
	if err != nil {
		t.Fatalf(
			"second Evaluate() error = %v",
			err,
		)
	}

	if first.Decision != DecisionAllowed ||
		!first.Usable ||
		first.RecentNeighborCount != 3 ||
		first.Score <
			validFreshnessConfig().
				CompleteScoreMinimum {
		t.Fatalf(
			"unexpected allowed result: %#v",
			first,
		)
	}
	if first.InputFingerprint !=
		second.InputFingerprint {
		t.Fatal(
			"deterministic freshness input produced different fingerprints",
		)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf(
			"result validation error = %v",
			err,
		)
	}
}

func TestEvaluateBlocksWhenNewestNeighborIsTooOld(
	t *testing.T,
) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern :=
		freshnessFixtures(
			[]time.Duration{
				8 * 24 * time.Hour,
				9 * 24 * time.Hour,
				10 * 24 * time.Hour,
			},
		)

	result, err := evaluator.Evaluate(
		selection,
		pattern,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionBlocked ||
		result.Usable ||
		!hasFreshnessNotice(
			result.Limitations,
			"newest_historical_neighbor_too_old",
		) {
		t.Fatalf(
			"stale newest neighbor did not block: %#v",
			result,
		)
	}
}

func TestEvaluateBlocksInsufficientRecentSupport(
	t *testing.T,
) {
	config := validFreshnessConfig()
	config.MaximumNewestNeighborAge =
		20 * 24 * time.Hour
	config.MaximumMeanNeighborAge =
		30 * 24 * time.Hour
	config.MaximumOldestNeighborAge =
		40 * 24 * time.Hour
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	selection, pattern :=
		freshnessFixtures(
			[]time.Duration{
				5 * 24 * time.Hour,
				15 * 24 * time.Hour,
				20 * 24 * time.Hour,
			},
		)

	result, err := evaluator.Evaluate(
		selection,
		pattern,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionBlocked ||
		result.RecentNeighborCount != 1 ||
		!hasFreshnessNotice(
			result.Limitations,
			"recent_historical_neighbor_support_insufficient",
		) {
		t.Fatalf(
			"recent support did not block: %#v",
			result,
		)
	}
}

func TestEvaluateReturnsLimitedForUsableIncompletePattern(
	t *testing.T,
) {
	config := validFreshnessConfig()
	config.CompleteScoreMinimum = 0.95
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	selection, pattern :=
		freshnessFixtures(
			[]time.Duration{
				24 * time.Hour,
				48 * time.Hour,
				72 * time.Hour,
			},
		)
	pattern.Status =
		projectionpatternconfidence.
			StatusLimited
	pattern.Level =
		projectioncontract.
			ConfidenceLevelMedium
	pattern.TargetNeighborCount =
		pattern.NeighborCount + 1
	pattern.Policy.TargetNeighborCount =
		pattern.TargetNeighborCount
	pattern.Components[1].Score =
		float64(pattern.NeighborCount) /
			float64(pattern.TargetNeighborCount)
	pattern.Score = 0
	for _, component := range pattern.Components {
		pattern.Score += component.Score * component.Weight
	}
	pattern.Limitations =
		[]projectionpatternconfidence.Notice{
			{
				Code:    "pattern_support_partial",
				Message: "Historical pattern is usable but does not satisfy complete target support.",
			},
		}

	result, err := evaluator.Evaluate(
		selection,
		pattern,
	)
	if err != nil {
		t.Fatalf(
			"Evaluate() error = %v",
			err,
		)
	}

	if result.Decision != DecisionLimited ||
		!result.Usable ||
		len(result.Limitations) == 0 {
		t.Fatalf(
			"unexpected limited result: %#v",
			result,
		)
	}
}

func TestEvaluateRejectsPatternSelectionMismatch(
	t *testing.T,
) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern :=
		freshnessFixtures(
			[]time.Duration{
				24 * time.Hour,
				48 * time.Hour,
				72 * time.Hour,
			},
		)
	pattern.SelectedTrajectoryIDs[0] =
		"other-trajectory"
	sort.Strings(pattern.SelectedTrajectoryIDs)

	_, err := evaluator.Evaluate(
		selection,
		pattern,
	)
	if !errors.Is(
		err,
		ErrPatternSelectionMismatch,
	) {
		t.Fatalf(
			"error = %v, want mismatch",
			err,
		)
	}
}

func newFreshnessEvaluator(
	t *testing.T,
) *Evaluator {
	t.Helper()

	evaluator, err := New(
		validFreshnessConfig(),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return evaluator
}

func freshnessFixtures(
	ages []time.Duration,
) (
	projectionneighbors.Result,
	projectionpatternconfidence.Result,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	neighbors := make(
		[]projectionneighbors.Neighbor,
		0,
		len(ages),
	)
	selectedIDs := make(
		[]string,
		0,
		len(ages),
	)
	for index, age := range ages {
		id := "historical-" +
			string(rune('a'+index))
		selectedIDs = append(
			selectedIDs,
			id,
		)
		score := 0.9 -
			float64(index)*0.05
		endTime := asOfTime.Add(-age)
		neighbors = append(
			neighbors,
			projectionneighbors.Neighbor{
				TrajectoryID:    id,
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
				AnchorObservedAt: endTime.Add(
					-2 * time.Minute,
				),
				AnchorDistanceKM: float64(index + 1),
				CandidateStartTime: endTime.Add(
					-10 * time.Minute,
				),
				CandidateEndTime:       endTime,
				CandidateAge:           age,
				PrefixPointCount:       5,
				ContinuationPointCount: 2,
				ContinuationEndTime:    endTime,
			},
		)
	}

	selection := projectionneighbors.Result{
		Version:                      projectionneighbors.Version,
		Status:                       projectionneighbors.StatusComplete,
		CurrentTrajectoryID:          "current",
		AsOfTime:                     asOfTime,
		RequiredContinuationDuration: 2 * time.Minute,
		InputCandidateCount:          len(neighbors),
		CheckedCandidateCount:        len(neighbors),
		QualifiedCandidateCount:      len(neighbors),
		RejectedCandidateCount:       0,
		SelectionLimit:               len(neighbors),
		Neighbors:                    neighbors,
		InputFingerprint: "sha256:" +
			strings.Repeat("d", 64),
	}

	sort.Strings(selectedIDs)
	pattern := projectionpatternconfidence.Result{
		Version: projectionpatternconfidence.Version,
		Status: projectionpatternconfidence.
			StatusComplete,
		Usable:                     true,
		NeighborCount:              len(neighbors),
		TargetNeighborCount:        len(neighbors),
		SelectionStatus:            projectionneighbors.StatusComplete,
		SourceSelectionFingerprint: selection.InputFingerprint,
		Policy: projectionpatternconfidence.Policy{
			MinimumNeighborCount:                   2,
			TargetNeighborCount:                    len(neighbors),
			MinimumSimilarityScore:                 0.5,
			MaximumSimilarityStandardDeviation:     0.2,
			AnchorDistanceNormalizationKM:          50,
			MinimumUsableScore:                     0.5,
			MediumConfidenceMinimum:                0.6,
			HighConfidenceMinimum:                  0.8,
			ContinuationAgreementSampleCount:       4,
			ContinuationDivergenceNormalizationMPS: 100,
			MaximumContinuationDivergenceMPS:       200,
			SimilarityStrengthWeight:               0.2,
			SupportWeight:                          0.2,
			SimilarityConsistencyWeight:            0.2,
			AnchorProximityWeight:                  0.2,
			ContinuationAgreementWeight:            0.2,
		},

		ContinuationAgreementKnown:       true,
		ContinuationAgreementSampleCount: 4,
		ContinuationAgreementPairCount:   (len(neighbors)) * ((len(neighbors)) - 1) / 2,
		ContinuationComparisonCount:      ((len(neighbors)) * ((len(neighbors)) - 1) / 2) * 4,
		ContinuationHorizonSeconds:       120,
		MeanContinuationSpreadM:          750,
		MaximumContinuationSpreadM:       1800,
		MeanContinuationDivergenceMPS:    10,
		MaximumContinuationDivergenceMPS: 15,
		MeanSimilarityScore:              0.85,
		MinimumSimilarityScore:           0.8,
		SimilarityStandardDeviation:      0.04082482904638629,
		MeanAnchorDistanceKM:             2,
		Score:                            0.9256700683814455,
		Level: projectioncontract.
			ConfidenceLevelHigh,
		Components: []projectionpatternconfidence.Component{
			{
				Name: projectionpatternconfidence.
					ComponentSimilarityStrength,
				Score:  0.85,
				Weight: 0.20000000000000001,
			},
			{
				Name: projectionpatternconfidence.
					ComponentSupport,
				Score:  1,
				Weight: 0.20000000000000001,
			},
			{
				Name: projectionpatternconfidence.
					ComponentSimilarityConsistency,
				Score:  0.9183503419072274,
				Weight: 0.20000000000000001,
			},
			{
				Name: projectionpatternconfidence.
					ComponentAnchorProximity,
				Score:  0.95999999999999996,
				Weight: 0.20000000000000001,
			},

			{
				Name: projectionpatternconfidence.
					ComponentContinuationAgreement,
				Score:  0.9,
				Weight: 0.2,
			}},
		SelectedTrajectoryIDs: selectedIDs,
		InputFingerprint: "sha256:" +
			strings.Repeat("e", 64),
	}

	return selection, pattern
}

func hasFreshnessNotice(
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
