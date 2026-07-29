package projectioncontinuation

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

type routeScopeCapturingSelector struct {
	result  projectionneighbors.Result
	request projectionneighbors.Request
	calls   int
}

func (
	selector *routeScopeCapturingSelector,
) Select(
	request projectionneighbors.Request,
) (projectionneighbors.Result, error) {
	selector.calls++
	selector.request = request
	return selector.result.Clone(), nil
}

func TestProjectForwardsRouteScopeToNeighborSelector(
	t *testing.T,
) {
	request := continuationTestRequest()
	selector := &routeScopeCapturingSelector{
		result: continuationTestSelection(request),
	}
	config := validContinuationConfig(t)
	config.NeighborSelector = selector

	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = baseline.Project(request)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if selector.calls != 1 {
		t.Fatalf("selector calls = %d, want 1", selector.calls)
	}
	if selector.request.RouteScope.InputFingerprint !=
		request.RouteScope.InputFingerprint {
		t.Fatalf(
			"route scope fingerprint = %q, want %q",
			selector.request.RouteScope.InputFingerprint,
			request.RouteScope.InputFingerprint,
		)
	}
}
