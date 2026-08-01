package main

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/gofiber/fiber/v2"
)

type historicalVerificationContextContractReader struct {
	calls int
}

func (
	reader *historicalVerificationContextContractReader,
) Get(
	context.Context,
	projectionread.Request,
) (projectionproduction.Result, error) {
	reader.calls++
	return projectionproduction.Result{}, nil
}

func TestVerifyHistoricalServiceRejectsNilContextBeforeRead(t *testing.T) {
	reader := &historicalVerificationContextContractReader{}

	result, err := verifyHistoricalService(
		nil,
		reader,
		verificationSchedule{},
	)

	if !errors.Is(err, ErrVerificationContextRequired) {
		t.Fatalf("error = %v, want verification context required", err)
	}
	if reader.calls != 0 {
		t.Fatalf("service calls = %d, want 0", reader.calls)
	}
	if result.Version != "" || result.InputFingerprint != "" {
		t.Fatalf("result = %#v, want empty result", result)
	}
}

func TestVerifyHistoricalEndpointRejectsNilContextBeforeRequest(t *testing.T) {
	app := fiber.New()

	_, err := verifyHistoricalEndpoint(
		nil,
		app,
		verificationSchedule{},
	)

	if !errors.Is(err, ErrVerificationContextRequired) {
		t.Fatalf("error = %v, want verification context required", err)
	}
}
