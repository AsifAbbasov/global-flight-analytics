package projectioncontinuation

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestProjectApprovedUsesAuthorizedEvidenceWithoutReevaluation(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	selector := config.NeighborSelector.(*neighborSelectorStub)
	patternEvaluator := config.
		PatternConfidenceEvaluator.(*patternEvaluatorStub)
	fallback := config.
		FallbackProjector.(*fallbackProjectorStub)

	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)

	result, err := baseline.ProjectApproved(
		request,
		ApprovedEvidence{
			Selection: selection,
			Pattern:   pattern,
		},
	)
	if err != nil {
		t.Fatalf("ProjectApproved() error = %v", err)
	}
	if selector.calls != 0 ||
		patternEvaluator.calls != 0 ||
		fallback.calls != 0 {
		t.Fatalf(
			"approved evidence was reevaluated: selector=%d pattern=%d fallback=%d",
			selector.calls,
			patternEvaluator.calls,
			fallback.calls,
		)
	}
	if result.Method.Name != MethodName {
		t.Fatalf(
			"ProjectApproved() method = %q",
			result.Method.Name,
		)
	}
}

func TestProjectApprovedRejectsPatternFromAnotherSelection(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	selector := config.NeighborSelector.(*neighborSelectorStub)
	patternEvaluator := config.
		PatternConfidenceEvaluator.(*patternEvaluatorStub)
	fallback := config.
		FallbackProjector.(*fallbackProjectorStub)

	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)
	pattern.SourceSelectionFingerprint =
		"sha256:" + strings.Repeat("9", 64)

	result, err := baseline.ProjectApproved(
		request,
		ApprovedEvidence{
			Selection: selection,
			Pattern:   pattern,
		},
	)
	if err != nil {
		t.Fatalf("ProjectApproved() error = %v", err)
	}
	if selector.calls != 0 ||
		patternEvaluator.calls != 0 ||
		fallback.calls != 1 {
		t.Fatalf(
			"mismatched lineage calls: selector=%d pattern=%d fallback=%d",
			selector.calls,
			patternEvaluator.calls,
			fallback.calls,
		)
	}
	if result.Method.Name ==
		MethodName {
		t.Fatal(
			"pattern from another selection was accepted",
		)
	}
}

func TestProjectApprovedRejectsAnchorTimestampMismatch(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	selector := config.NeighborSelector.(*neighborSelectorStub)
	patternEvaluator := config.
		PatternConfidenceEvaluator.(*patternEvaluatorStub)
	fallback := config.
		FallbackProjector.(*fallbackProjectorStub)

	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)
	request.Candidates[0].Points[4].ObservedAt =
		request.Candidates[0].Points[4].
			ObservedAt.Add(time.Second)

	result, err := baseline.ProjectApproved(
		request,
		ApprovedEvidence{
			Selection: selection,
			Pattern:   pattern,
		},
	)
	if err != nil {
		t.Fatalf("ProjectApproved() error = %v", err)
	}
	if selector.calls != 0 ||
		patternEvaluator.calls != 0 ||
		fallback.calls != 1 {
		t.Fatalf(
			"anchor mismatch calls: selector=%d pattern=%d fallback=%d",
			selector.calls,
			patternEvaluator.calls,
			fallback.calls,
		)
	}
	if result.Method.Name ==
		MethodName {
		t.Fatal(
			"mismatched anchor timestamp was accepted",
		)
	}
}

func TestProjectApprovedPreservesObservedCandidateProvenance(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := continuationTestRequest()
	selection := continuationTestSelection(request)
	pattern := continuationTestPattern(selection)

	result, err := baseline.ProjectApproved(
		request,
		ApprovedEvidence{
			Selection: selection,
			Pattern:   pattern,
		},
	)
	if err != nil {
		t.Fatalf("ProjectApproved() error = %v", err)
	}

	found := 0
	for _, input := range result.Provenance.Inputs {
		if !strings.HasPrefix(
			input.Name,
			"historical_neighbor:",
		) {
			continue
		}
		found++
		if input.Classification !=
			projectioncontract.
				InputClassificationObserved ||
			input.SourceName !=
				"historical-store" {
			t.Fatalf(
				"historical input provenance = %#v",
				input,
			)
		}
	}
	if found != len(selection.Neighbors) {
		t.Fatalf(
			"observed historical inputs = %d, want %d",
			found,
			len(selection.Neighbors),
		)
	}
}
