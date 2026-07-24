package metricexecution

import (
	"errors"
	"testing"
)

func TestNewRequiresExecutor(
	t *testing.T,
) {
	service, err := New(nil)

	if service != nil {
		t.Fatal("expected nil service")
	}

	if !errors.Is(
		err,
		ErrExecutorRequired,
	) {
		t.Fatalf(
			"expected executor requirement, got %v",
			err,
		)
	}
}

func TestServiceStoresNarrowExecutorDependency(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	if service.executor == nil {
		t.Fatal(
			"expected internal analytics executor behavior",
		)
	}
}
