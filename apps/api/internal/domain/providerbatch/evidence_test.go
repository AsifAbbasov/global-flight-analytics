package providerbatch

import (
	"errors"
	"testing"
)

func TestResolveAcceptedOnlyCompatibility(t *testing.T) {
	evidence, err := Resolve(Evidence{}, 3)
	if err != nil {
		t.Fatalf("resolve evidence: %v", err)
	}
	if evidence != (Evidence{
		Received: 3,
		Accepted: 3,
	}) {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestEvidenceRequiresCompleteAccounting(t *testing.T) {
	err := (Evidence{
		Received:          3,
		Accepted:          1,
		RejectedMalformed: 1,
	}).Validate()
	if !errors.Is(err, ErrEvidenceInvalid) {
		t.Fatalf("expected invalid evidence, got %v", err)
	}
}

func TestAllItemsRejectedErrorPreservesEvidence(t *testing.T) {
	expected := Evidence{
		Received:          2,
		RejectedMalformed: 1,
		RejectedUnusable:  1,
	}
	err := NewAllItemsRejectedError("opensky", expected)
	if !errors.Is(err, ErrAllItemsRejected) {
		t.Fatalf("expected all-items-rejected error, got %v", err)
	}
	actual, ok := FromError(err)
	if !ok || actual != expected {
		t.Fatalf("evidence=%+v ok=%t, want %+v", actual, ok, expected)
	}
}
