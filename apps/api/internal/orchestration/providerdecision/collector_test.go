package providerdecision

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestCollectorRecordsAllowedAndDeniedBudgetDecisions(
	t *testing.T,
) {
	decidedAt := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	collector := New(
		func() time.Time {
			return decidedAt
		},
	)

	collector.RecordBudgetDecision(
		providerpolicy.ProviderAirplanesLive,
		"point:40.4:49.8:250",
		"",
		providerbudget.Decision{
			Provider: providerpolicy.ProviderAirplanesLive,
			Allowed:  true,
			Reason:   providerbudget.DecisionReasonAllowed,
		},
	)

	retryAt := decidedAt.Add(
		time.Second,
	)

	collector.RecordBudgetDecision(
		providerpolicy.ProviderAirplanesLive,
		"point:40.4:49.8:250",
		"",
		providerbudget.Decision{
			Provider: providerpolicy.ProviderAirplanesLive,
			Allowed:  false,
			Reason:   providerbudget.DecisionReasonFixedWindowExhausted,
			RetryAt:  retryAt,
		},
	)

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"snapshot provider decisions: %v",
			err,
		)
	}

	if snapshot.DecisionsTotal != 2 {
		t.Fatalf(
			"expected 2 decisions, got %d",
			snapshot.DecisionsTotal,
		)
	}

	if snapshot.AllowedTotal != 1 {
		t.Fatalf(
			"expected 1 allowed decision, got %d",
			snapshot.AllowedTotal,
		)
	}

	if snapshot.DeniedTotal != 1 {
		t.Fatalf(
			"expected 1 denied decision, got %d",
			snapshot.DeniedTotal,
		)
	}

	if snapshot.ReasonCounts[providerbudget.DecisionReasonAllowed] != 1 {
		t.Fatal(
			"expected one allowed reason",
		)
	}

	if snapshot.ReasonCounts[providerbudget.DecisionReasonFixedWindowExhausted] != 1 {
		t.Fatal(
			"expected one fixed-window denial reason",
		)
	}

	if snapshot.Latest.Allowed {
		t.Fatal(
			"expected latest decision to be denied",
		)
	}

	if !snapshot.Latest.RetryAt.Equal(
		retryAt,
	) {
		t.Fatalf(
			"expected retry at %s, got %s",
			retryAt,
			snapshot.Latest.RetryAt,
		)
	}

	if !snapshot.Latest.DecidedAt.Equal(
		decidedAt,
	) {
		t.Fatalf(
			"expected decided at %s, got %s",
			decidedAt,
			snapshot.Latest.DecidedAt,
		)
	}
}

func TestCollectorRecordsPublicationDecisionContext(
	t *testing.T,
) {
	collector := New(nil)

	collector.RecordBudgetDecision(
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"publication-2026-07-12",
		providerbudget.Decision{
			Provider: providerpolicy.ProviderOurAirports,
			Allowed:  true,
			Reason:   providerbudget.DecisionReasonAllowed,
		},
	)

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderOurAirports,
	)
	if err != nil {
		t.Fatalf(
			"snapshot publication decision: %v",
			err,
		)
	}

	if snapshot.Latest.PublicationID !=
		"publication-2026-07-12" {
		t.Fatalf(
			"unexpected publication identifier: %s",
			snapshot.Latest.PublicationID,
		)
	}
}

func TestCollectorRejectsSnapshotWithoutEvidence(
	t *testing.T,
) {
	collector := New(nil)

	_, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if !errors.Is(
		err,
		ErrNoDecisionEvidence,
	) {
		t.Fatalf(
			"expected ErrNoDecisionEvidence, got %v",
			err,
		)
	}
}

func TestCollectorIsSafeForConcurrentRecording(
	t *testing.T,
) {
	collector := New(nil)

	const goroutineCount = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for index := 0; index < goroutineCount; index++ {
		go func() {
			defer waitGroup.Done()

			collector.RecordBudgetDecision(
				providerpolicy.ProviderAirplanesLive,
				"point:40.4:49.8:250",
				"",
				providerbudget.Decision{
					Provider: providerpolicy.ProviderAirplanesLive,
					Allowed:  true,
					Reason:   providerbudget.DecisionReasonAllowed,
				},
			)
		}()
	}

	waitGroup.Wait()

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"snapshot concurrent decisions: %v",
			err,
		)
	}

	if snapshot.DecisionsTotal != goroutineCount {
		t.Fatalf(
			"expected %d decisions, got %d",
			goroutineCount,
			snapshot.DecisionsTotal,
		)
	}
}
