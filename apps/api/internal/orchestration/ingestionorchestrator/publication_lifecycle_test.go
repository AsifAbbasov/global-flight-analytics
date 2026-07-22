package ingestionorchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type publicationTestValue struct {
	Value string
}

func (publicationTestValue) RequestCoalescingValue() {}

func TestExecutePublicationReleasesReservationAfterFunctionFailure(t *testing.T) {
	manager, err := providerbudget.New(nil)
	if err != nil {
		t.Fatalf("create publication manager: %v", err)
	}
	orchestrator, err := NewPublicationOnly[publicationTestValue](manager, nil)
	if err != nil {
		t.Fatalf("create publication orchestrator: %v", err)
	}

	operationErr := errors.New("database import failed")
	_, err = orchestrator.ExecutePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"sha256:retryable",
		func(context.Context) (publicationTestValue, error) {
			return publicationTestValue{}, operationErr
		},
	)
	if !errors.Is(err, operationErr) {
		t.Fatalf("expected operation error, got %v", err)
	}

	result, err := orchestrator.ExecutePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"sha256:retryable",
		func(context.Context) (publicationTestValue, error) {
			return publicationTestValue{Value: "imported"}, nil
		},
	)
	if err != nil {
		t.Fatalf("retry released publication: %v", err)
	}
	if result.Value.Value != "imported" {
		t.Fatalf("unexpected retry result: %+v", result.Value)
	}
}

func TestExecutePublicationCommitsBeforeRejectingDuplicate(t *testing.T) {
	manager, err := providerbudget.New(nil)
	if err != nil {
		t.Fatalf("create publication manager: %v", err)
	}
	orchestrator, err := NewPublicationOnly[publicationTestValue](manager, nil)
	if err != nil {
		t.Fatalf("create publication orchestrator: %v", err)
	}

	_, err = orchestrator.ExecutePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"sha256:committed",
		func(context.Context) (publicationTestValue, error) {
			return publicationTestValue{Value: "imported"}, nil
		},
	)
	if err != nil {
		t.Fatalf("execute first publication: %v", err)
	}

	functionCalled := false
	_, err = orchestrator.ExecutePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"sha256:committed",
		func(context.Context) (publicationTestValue, error) {
			functionCalled = true
			return publicationTestValue{}, nil
		},
	)
	var accessDenied *AccessDeniedError
	if !errors.As(err, &accessDenied) {
		t.Fatalf("expected publication access denial, got %v", err)
	}
	if accessDenied.Reason != providerbudget.DecisionReasonPublicationAlreadyProcessed {
		t.Fatalf("unexpected denial reason: %s", accessDenied.Reason)
	}
	if functionCalled {
		t.Fatal("duplicate publication function must not execute")
	}
}
