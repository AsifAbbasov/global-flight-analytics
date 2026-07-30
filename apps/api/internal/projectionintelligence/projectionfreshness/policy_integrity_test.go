package projectionfreshness

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

func TestEvaluatePublishesPolicyAndUpstreamSnapshot(t *testing.T) {
	config := validFreshnessConfig()
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	selection, pattern := freshnessFixtures([]time.Duration{
		24 * time.Hour,
		48 * time.Hour,
		72 * time.Hour,
	})

	result, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Policy != config.policySnapshot() {
		t.Fatalf("policy snapshot = %#v, want %#v", result.Policy, config.policySnapshot())
	}
	if result.SelectionStatus != selection.Status ||
		result.PatternStatus != pattern.Status ||
		result.PatternUsable != pattern.Usable ||
		result.SourceSelectionFingerprint != selection.InputFingerprint ||
		result.SourcePatternFingerprint != pattern.InputFingerprint {
		t.Fatalf("unexpected upstream snapshot: %#v", result)
	}
}

func TestEvaluateRejectsPatternSelectionStatusMismatch(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{
		24 * time.Hour,
		48 * time.Hour,
		72 * time.Hour,
	})
	selection.Status = projectionneighbors.StatusPartial

	_, err := evaluator.Evaluate(selection, pattern)
	if !errors.Is(err, ErrPatternSelectionMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrPatternSelectionMismatch)
	}
}

func TestResultValidateReconstructsPolicyDecision(t *testing.T) {
	evaluator := newFreshnessEvaluator(t)
	selection, pattern := freshnessFixtures([]time.Duration{
		24 * time.Hour,
		48 * time.Hour,
		72 * time.Hour,
	})
	valid, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Result)
	}{
		{
			name: "policy threshold",
			mutate: func(result *Result) {
				result.Policy.CompleteScoreMinimum = 0.99
			},
		},
		{
			name: "coordinated component and score",
			mutate: func(result *Result) {
				result.Components[0].Score -= 0.1
				result.Score = weightedComponentScore(result.Components)
			},
		},
		{
			name: "decision and limitation",
			mutate: func(result *Result) {
				result.Decision = DecisionLimited
				result.Limitations = []Notice{
					{
						Code:    "fabricated_limitation",
						Message: "Fabricated limitation.",
					},
				}
			},
		},
		{
			name: "pattern status",
			mutate: func(result *Result) {
				result.PatternStatus = projectionpatternconfidence.StatusLimited
			},
		},
		{
			name: "pattern usability",
			mutate: func(result *Result) {
				result.PatternUsable = false
			},
		},
		{
			name: "source pattern fingerprint",
			mutate: func(result *Result) {
				result.SourcePatternFingerprint = ""
			},
		},
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

func TestResultValidateAcceptsReconstructedLimitedDecision(t *testing.T) {
	config := validFreshnessConfig()
	config.CompleteScoreMinimum = 0.99
	evaluator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	selection, pattern := freshnessFixtures([]time.Duration{
		24 * time.Hour,
		48 * time.Hour,
		72 * time.Hour,
	})
	result, err := evaluator.Evaluate(selection, pattern)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Decision != DecisionLimited {
		t.Fatalf("decision = %q, want %q", result.Decision, DecisionLimited)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
