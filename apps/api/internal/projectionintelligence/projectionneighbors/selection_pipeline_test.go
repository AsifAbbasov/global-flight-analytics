package projectionneighbors

import (
	"testing"
)

func TestSelectSeparatesEvaluationTruncationFromSelectionLimiting(
	t *testing.T,
) {
	selector := newTestSelector(t)
	result, err := selector.Select(selectorTestRequest())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result.CandidateEvaluationTruncated || result.Truncated {
		t.Fatalf("candidate evaluation unexpectedly truncated: %#v", result)
	}
	if !result.QualifiedSelectionLimited {
		t.Fatalf("qualified selection limit was not reported: %#v", result)
	}
	if result.Status != StatusComplete {
		t.Fatalf("status = %q, want complete", result.Status)
	}
	if !hasNotice(result.Limitations, "qualified_neighbors_limited") ||
		hasNotice(result.Limitations, "candidate_evaluation_truncated") {
		t.Fatalf("unexpected limitations: %#v", result.Limitations)
	}
}

func TestSelectReportsCandidateEvaluationTruncationIndependently(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.MaximumCandidateCount = 1
	config.SelectionLimit = 1
	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := selector.Select(selectorTestRequest())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if !result.CandidateEvaluationTruncated || !result.Truncated {
		t.Fatalf("candidate evaluation truncation missing: %#v", result)
	}
	if result.QualifiedSelectionLimited {
		t.Fatalf("qualified selection was unexpectedly limited: %#v", result)
	}
	if result.Status != StatusPartial {
		t.Fatalf("status = %q, want partial", result.Status)
	}
	if !hasNotice(result.Limitations, "candidate_evaluation_truncated") ||
		hasNotice(result.Limitations, "qualified_neighbors_limited") {
		t.Fatalf("unexpected limitations: %#v", result.Limitations)
	}
}

func TestResultValidateRejectsFalseExplicitLimitFlags(
	t *testing.T,
) {
	selector := newTestSelector(t)
	result, err := selector.Select(selectorTestRequest())
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}

	result.QualifiedSelectionLimited = false
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted missing qualified-selection limit evidence")
	}

	result, err = selector.Select(selectorTestRequest())
	if err != nil {
		t.Fatalf("Select() second error = %v", err)
	}
	result.CandidateEvaluationTruncated = true
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted false candidate-evaluation truncation evidence")
	}
}
