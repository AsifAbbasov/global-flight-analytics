package projectionfreshness

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

func TestEvaluateBlocksUnusablePatternConfidence(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{24 * time.Hour, 48 * time.Hour, 72 * time.Hour})
	makePatternUnusable(&pattern)

	result, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Decision != DecisionBlocked || result.Usable ||
		!hasFreshnessNotice(result.Limitations, "pattern_confidence_unusable") {
		t.Fatalf("unusable pattern confidence did not block freshness: %#v", result)
	}
}

func TestEvaluateRejectsSourceSelectionFingerprintMismatch(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{24 * time.Hour, 48 * time.Hour, 72 * time.Hour})
	pattern.SourceSelectionFingerprint = "sha256:" + strings.Repeat("f", 64)

	_, err := evaluator.Evaluate(selection, pattern)
	if !errors.Is(err, ErrPatternSelectionMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrPatternSelectionMismatch)
	}
}

func TestEvaluateFingerprintIncludesUpstreamStatus(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{24 * time.Hour, 48 * time.Hour, 72 * time.Hour})
	complete, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("complete Evaluate() error = %v", err)
	}
	selection.Status = projectionneighbors.StatusPartial
	partial, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("partial Evaluate() error = %v", err)
	}
	if complete.InputFingerprint == partial.InputFingerprint {
		t.Fatal("selection status change did not change freshness fingerprint")
	}
}

func TestMeanDurationAvoidsInt64Overflow(t *testing.T) {
	values := []time.Duration{time.Duration(math.MaxInt64), time.Duration(math.MaxInt64)}
	if got := meanDuration(values); got != time.Duration(math.MaxInt64) {
		t.Fatalf("meanDuration() = %d, want %d", got, int64(math.MaxInt64))
	}
}

func TestEvaluateRejectsInconsistentCandidateAgeEvidence(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{24 * time.Hour, 48 * time.Hour, 72 * time.Hour})
	selection.Neighbors[0].CandidateAge = time.Second
	_, err := evaluator.Evaluate(selection, pattern)
	if !errors.Is(err, ErrNeighborSelectionInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrNeighborSelectionInvalid)
	}
}

func TestEvaluateReportsAllHardViolations(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{
		40 * 24 * time.Hour,
		50 * 24 * time.Hour,
		60 * 24 * time.Hour,
	})
	result, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	for _, code := range []string{
		"newest_historical_neighbor_too_old",
		"mean_historical_neighbor_age_too_old",
		"oldest_historical_neighbor_too_old",
		"recent_historical_neighbor_support_insufficient",
		"pattern_freshness_score_below_minimum",
	} {
		if !hasFreshnessNotice(result.Limitations, code) {
			t.Fatalf("missing limitation %q in %#v", code, result.Limitations)
		}
	}
}

func TestConfigValidateRejectsThresholdOrderAndZeroScores(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Config)
		wantError error
	}{
		{
			name: "age threshold order",
			mutate: func(config *Config) {
				config.MaximumNewestNeighborAge = 20 * 24 * time.Hour
				config.MaximumMeanNeighborAge = 10 * 24 * time.Hour
			},
			wantError: ErrFreshnessAgeThresholdOrderInvalid,
		},
		{
			name: "recent age beyond oldest maximum",
			mutate: func(config *Config) {
				config.RecentNeighborAgeLimit = config.MaximumOldestNeighborAge + time.Second
			},
			wantError: ErrRecentNeighborAgeLimitInvalid,
		},
		{
			name: "zero minimum score",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 0
			},
			wantError: ErrFreshnessScoreThresholdInvalid,
		},
		{
			name: "zero complete score",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 0
				config.CompleteScoreMinimum = 0
			},
			wantError: ErrFreshnessScoreThresholdInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validFreshnessConfig()
			test.mutate(&config)
			if err := config.Validate(); !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestResultValidateRejectsCrossFieldMutations(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{24 * time.Hour, 48 * time.Hour, 72 * time.Hour})
	valid, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Result)
	}{
		{name: "weighted score", mutate: func(result *Result) { result.Score -= 0.1 }},
		{name: "allowed limitation", mutate: func(result *Result) {
			result.Limitations = []Notice{{Code: "unexpected", Message: "Unexpected limitation."}}
		}},
		{name: "unnormalized identifier", mutate: func(result *Result) {
			result.SelectedTrajectoryIDs[0] = " " + result.SelectedTrajectoryIDs[0]
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := valid.Clone()
			test.mutate(&result)
			if err := result.Validate(); err == nil {
				t.Fatal("Validate() accepted an inconsistent result")
			}
		})
	}
}

func makePatternUnusable(pattern *projectionpatternconfidence.Result) {
	pattern.Status = projectionpatternconfidence.StatusUnavailable
	pattern.Usable = false
	pattern.Level = projectioncontract.ConfidenceLevelNone
	pattern.ContinuationAgreementKnown = false
	pattern.ContinuationAgreementSampleCount = 0
	pattern.ContinuationAgreementPairCount = 0
	pattern.ContinuationComparisonCount = 0
	pattern.ContinuationHorizonSeconds = 0
	pattern.MeanContinuationSpreadM = 0
	pattern.MaximumContinuationSpreadM = 0
	pattern.MeanContinuationDivergenceMPS = 0
	pattern.MaximumContinuationDivergenceMPS = 0
	pattern.Components[4].Score = 0
	pattern.Score = 0
	for _, component := range pattern.Components {
		pattern.Score += component.Score * component.Weight
	}
	pattern.Limitations = []projectionpatternconfidence.Notice{
		{
			Code:    "pattern_continuation_agreement_unavailable",
			Message: "Historical continuation agreement is unavailable for the selected neighbors.",
		},
	}
}
