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

func TestServiceExposesExecutor(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	if service.Executor() == nil {
		t.Fatal("expected analytics executor")
	}

	var nilService *Service
	if nilService.Executor() != nil {
		t.Fatal("expected nil executor from nil service")
	}
}
