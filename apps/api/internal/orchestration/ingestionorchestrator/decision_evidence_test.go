package ingestionorchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type decisionRecorderStub struct {
	records []decisionRecord
}

type decisionRecord struct {
	provider      providerpolicy.Provider
	requestKey    string
	publicationID string
	decision      providerbudget.Decision
}

func (
	stub *decisionRecorderStub,
) RecordBudgetDecision(
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
	decision providerbudget.Decision,
) {
	stub.records = append(
		stub.records,
		decisionRecord{
			provider:      provider,
			requestKey:    requestKey,
			publicationID: publicationID,
			decision:      decision,
		},
	)
}

func TestExecuteRecordsAllowedBudgetDecision(
	t *testing.T,
) {
	recorder := &decisionRecorderStub{}

	orchestrator, err := New(
		Config[orchestrationTestValue]{
			BudgetManager:    &budgetManagerStub{},
			Coalescer:        &coalescerStub{},
			DecisionRecorder: recorder,
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrator: %v",
			err,
		)
	}

	_, err = orchestrator.Execute(
		context.Background(),
		providerpolicy.ProviderAirplanesLive,
		"traffic:regional-snapshot",
		func(
			_ context.Context,
		) (orchestrationTestValue, error) {
			return orchestrationTestValue(
				"snapshot",
			), nil
		},
	)
	if err != nil {
		t.Fatalf(
			"execute orchestration: %v",
			err,
		)
	}

	if len(recorder.records) != 1 {
		t.Fatalf(
			"expected one recorded decision, got %d",
			len(recorder.records),
		)
	}

	record := recorder.records[0]

	if record.provider !=
		providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"unexpected provider: %s",
			record.provider,
		)
	}

	if record.requestKey !=
		"traffic:regional-snapshot" {
		t.Fatalf(
			"unexpected request key: %s",
			record.requestKey,
		)
	}

	if !record.decision.Allowed {
		t.Fatal(
			"expected allowed decision",
		)
	}
}

func TestExecuteRecordsDeniedBudgetDecision(
	t *testing.T,
) {
	retryAt := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		1,
		0,
		time.UTC,
	)

	recorder := &decisionRecorderStub{}

	orchestrator, err := New(
		Config[orchestrationTestValue]{
			BudgetManager: &budgetManagerStub{
				acquireFunction: func(
					provider providerpolicy.Provider,
				) (providerbudget.Decision, error) {
					return providerbudget.Decision{
						Provider: provider,
						Allowed:  false,
						Reason: providerbudget.
							DecisionReasonFixedWindowExhausted,
						RetryAt: retryAt,
					}, nil
				},
			},
			Coalescer:        &coalescerStub{},
			DecisionRecorder: recorder,
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrator: %v",
			err,
		)
	}

	_, err = orchestrator.Execute(
		context.Background(),
		providerpolicy.ProviderAirplanesLive,
		"traffic:regional-snapshot",
		func(
			_ context.Context,
		) (orchestrationTestValue, error) {
			return orchestrationTestValue(
				"snapshot",
			), nil
		},
	)
	if err == nil {
		t.Fatal(
			"expected access denied error",
		)
	}

	if len(recorder.records) != 1 {
		t.Fatalf(
			"expected one recorded decision, got %d",
			len(recorder.records),
		)
	}

	record := recorder.records[0]

	if record.decision.Allowed {
		t.Fatal(
			"expected denied decision",
		)
	}

	if !record.decision.RetryAt.Equal(
		retryAt,
	) {
		t.Fatalf(
			"expected retry at %s, got %s",
			retryAt,
			record.decision.RetryAt,
		)
	}
}

func TestExecutePublicationRecordsPublicationContext(
	t *testing.T,
) {
	recorder := &decisionRecorderStub{}

	orchestrator, err := New(
		Config[orchestrationTestValue]{
			BudgetManager:    &budgetManagerStub{},
			Coalescer:        &coalescerStub{},
			DecisionRecorder: recorder,
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrator: %v",
			err,
		)
	}

	_, err = orchestrator.ExecutePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"publication-2026-07-12",
		func(
			_ context.Context,
		) (orchestrationTestValue, error) {
			return orchestrationTestValue(
				"import-result",
			), nil
		},
	)
	if err != nil {
		t.Fatalf(
			"execute publication orchestration: %v",
			err,
		)
	}

	if len(recorder.records) != 1 {
		t.Fatalf(
			"expected one recorded publication decision, got %d",
			len(recorder.records),
		)
	}

	record := recorder.records[0]

	if record.publicationID !=
		"publication-2026-07-12" {
		t.Fatalf(
			"unexpected publication identifier: %s",
			record.publicationID,
		)
	}
}
