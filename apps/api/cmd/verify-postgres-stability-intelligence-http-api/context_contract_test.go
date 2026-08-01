package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
)

type runtimeStabilityContextContractService struct {
	calls int
}

func (
	service *runtimeStabilityContextContractService,
) Get(
	context.Context,
	stabilityproduction.Request,
) (stabilityproduction.Result, error) {
	service.calls++
	return stabilityproduction.Result{}, nil
}

func TestRuntimeStabilityReaderRejectsNilContextBeforeServiceRead(t *testing.T) {
	service := &runtimeStabilityContextContractService{}
	reader := runtimeStabilityReader{
		service: service,
		timeout: time.Second,
	}

	result, err := reader.Get(
		nil,
		stabilityproduction.Request{},
	)

	if !errors.Is(err, ErrRuntimeStabilityContextRequired) {
		t.Fatalf("error = %v, want verification context required", err)
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
	if result.Version != "" ||
		result.TrajectoryID != "" ||
		result.InputFingerprint != "" {
		t.Fatalf("result = %#v, want empty result", result)
	}
}
