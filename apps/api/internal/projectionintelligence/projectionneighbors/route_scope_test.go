package projectionneighbors

import (
	"errors"
	"testing"
)

func TestSelectRejectsMissingRouteScope(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()
	request.RouteScope = RouteScope{}

	_, err := selector.Select(request)
	if !errors.Is(err, ErrRouteScopeInvalid) {
		t.Fatalf(
			"Select() error = %v, want %v",
			err,
			ErrRouteScopeInvalid,
		)
	}
}

func TestSelectionFingerprintIncludesRouteScope(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()

	first, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() first error = %v", err)
	}

	request.RouteScope = UniformRouteScope(
		"UBBB",
		"UGTB",
		"selector-test-route-scope-v1",
	)
	second, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() second error = %v", err)
	}

	if first.InputFingerprint == second.InputFingerprint {
		t.Fatal("route scope did not change the selection fingerprint")
	}
}

func TestSelectRejectsCrossRouteCandidateBeforeSimilarity(
	t *testing.T,
) {
	config := validSelectorConfig()
	config.SelectionLimit = 1
	stub := config.SimilarityEngine.(*similarityEngineStub)
	selector, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := selectorTestRequest()
	request.Candidates = request.Candidates[:1]
	candidateID := request.Candidates[0].ID
	request.RouteScope = ExplicitRouteScope(
		"UBBB",
		"LTFM",
		"selector-test-explicit-route-scope-v1",
		[]CandidateRouteEvidence{
			CandidateRoute(
				candidateID,
				"UBBB",
				"UGTB",
				"candidate-route-store-v1",
			),
		},
	)

	result, err := selector.Select(request)
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if result.Status != StatusUnavailable ||
		!hasRejection(result.Rejections, RejectionCrossRoute) {
		t.Fatalf("unexpected cross-route result: %#v", result)
	}
	if len(stub.calls) != 0 {
		t.Fatalf("similarity calls = %v, want none", stub.calls)
	}
}

func TestSelectRejectsIncompleteExplicitRouteScope(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()
	request.RouteScope = ExplicitRouteScope(
		"UBBB",
		"LTFM",
		"selector-test-explicit-route-scope-v1",
		[]CandidateRouteEvidence{
			CandidateRoute(
				request.Candidates[0].ID,
				"UBBB",
				"LTFM",
				"candidate-route-store-v1",
			),
		},
	)

	_, err := selector.Select(request)
	if !errors.Is(err, ErrRouteScopeInvalid) {
		t.Fatalf(
			"Select() error = %v, want %v",
			err,
			ErrRouteScopeInvalid,
		)
	}
}

func TestSelectRejectsTamperedRouteScopeFingerprint(
	t *testing.T,
) {
	selector := newTestSelector(t)
	request := selectorTestRequest()
	request.RouteScope.SourceName = "tampered-source"

	_, err := selector.Select(request)
	if !errors.Is(err, ErrRouteScopeInvalid) {
		t.Fatalf(
			"Select() error = %v, want %v",
			err,
			ErrRouteScopeInvalid,
		)
	}
}
