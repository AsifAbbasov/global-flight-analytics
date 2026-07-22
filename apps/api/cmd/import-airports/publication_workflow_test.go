package main

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/ourairports"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestExecuteAirportPublicationImportsReservedPublication(t *testing.T) {
	executor := &airportPublicationExecutorStub{}
	repository := &airportImportRepositoryStub{reconciledCount: 2}

	outcome, err := executeAirportPublication(
		context.Background(),
		executor,
		repository,
		ourairports.LoadResult{
			PublicationID: "sha256:publication-a",
			Airports: []airport.ImportRecord{
				{SourceIdent: "UBBB"},
				{SourceIdent: "UGTB"},
			},
		},
	)
	if err != nil {
		t.Fatalf("execute airport publication: %v", err)
	}
	if outcome.ReconciledCount != 2 || outcome.AlreadyProcessed {
		t.Fatalf("unexpected publication outcome: %+v", outcome)
	}
	if repository.calls != 1 || len(repository.records) != 2 {
		t.Fatalf("unexpected repository calls=%d records=%d", repository.calls, len(repository.records))
	}
	if executor.provider != providerpolicy.ProviderOurAirports ||
		executor.requestKey != ourAirportsPublicationRequestKey ||
		executor.publicationID != "sha256:publication-a" {
		t.Fatalf("unexpected executor context: %+v", executor)
	}
}

func TestExecuteAirportPublicationSkipsCommittedPublication(t *testing.T) {
	executor := &airportPublicationExecutorStub{
		err: &ingestionorchestrator.AccessDeniedError{
			Provider: providerpolicy.ProviderOurAirports,
			Reason: providerbudget.
				DecisionReasonPublicationAlreadyProcessed,
		},
	}
	repository := &airportImportRepositoryStub{}

	outcome, err := executeAirportPublication(
		context.Background(),
		executor,
		repository,
		ourairports.LoadResult{PublicationID: "sha256:committed"},
	)
	if err != nil {
		t.Fatalf("execute committed publication: %v", err)
	}
	if !outcome.AlreadyProcessed {
		t.Fatal("expected committed publication outcome")
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", repository.calls)
	}
}

func TestExecuteAirportPublicationPreservesInProgressDenial(t *testing.T) {
	expectedErr := &ingestionorchestrator.AccessDeniedError{
		Provider: providerpolicy.ProviderOurAirports,
		Reason: providerbudget.
			DecisionReasonPublicationInProgress,
	}
	executor := &airportPublicationExecutorStub{err: expectedErr}

	_, err := executeAirportPublication(
		context.Background(),
		executor,
		&airportImportRepositoryStub{},
		ourairports.LoadResult{PublicationID: "sha256:active"},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected in-progress denial, got %v", err)
	}
}

func TestExecuteAirportPublicationReturnsImportFailure(t *testing.T) {
	importErr := errors.New("database unavailable")
	executor := &airportPublicationExecutorStub{}
	repository := &airportImportRepositoryStub{err: importErr}

	_, err := executeAirportPublication(
		context.Background(),
		executor,
		repository,
		ourairports.LoadResult{PublicationID: "sha256:retry"},
	)
	if !errors.Is(err, importErr) {
		t.Fatalf("expected import error, got %v", err)
	}
}

type airportPublicationExecutorStub struct {
	provider      providerpolicy.Provider
	requestKey    string
	publicationID string
	err           error
}

func (stub *airportPublicationExecutorStub) ExecutePublication(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
	function ingestionorchestrator.Function[airportImportExecutionValue],
) (ingestionorchestrator.ExecuteResult[airportImportExecutionValue], error) {
	stub.provider = provider
	stub.requestKey = requestKey
	stub.publicationID = publicationID
	if stub.err != nil {
		return ingestionorchestrator.ExecuteResult[airportImportExecutionValue]{}, stub.err
	}
	value, err := function(ctx)
	if err != nil {
		return ingestionorchestrator.ExecuteResult[airportImportExecutionValue]{}, err
	}
	return ingestionorchestrator.ExecuteResult[airportImportExecutionValue]{
		Provider:   provider,
		RequestKey: requestKey,
		Value:      value,
	}, nil
}

type airportImportRepositoryStub struct {
	reconciledCount int64
	err             error
	calls           int
	records         []airport.ImportRecord
}

func (stub *airportImportRepositoryStub) UpsertImported(
	_ context.Context,
	records []airport.ImportRecord,
) (int64, error) {
	stub.calls++
	stub.records = append([]airport.ImportRecord(nil), records...)
	return stub.reconciledCount, stub.err
}
