package providerdecision

import (
	"sync"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestCollectorRecordsFallbackDecision(
	t *testing.T,
) {
	collector := New(nil)

	consideredProviders := []providerpolicy.Provider{
		providerpolicy.ProviderAirplanesLive,
		providerpolicy.ProviderOpenSky,
	}

	collector.RecordFallbackDecision(
		providerfallback.Decision{
			PrimaryProvider: providerpolicy.
				ProviderAirplanesLive,
			SelectedProvider: providerpolicy.
				ProviderOpenSky,
			UsedFallback: true,
			Outcome: providerfallback.
				OutcomeFallbackSelected,
			TriggerReason: providerbudget.
				DecisionReasonFixedWindowExhausted,
			ConsideredProviders: consideredProviders,
			DecidedAt: time.Date(
				2026,
				time.July,
				12,
				19,
				0,
				0,
				0,
				time.UTC,
			),
		},
	)

	consideredProviders[0] =
		providerpolicy.ProviderOpenMeteo

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"snapshot fallback decision: %v",
			err,
		)
	}

	if !snapshot.FallbackObserved {
		t.Fatal(
			"expected observed fallback evidence",
		)
	}

	if snapshot.FallbackDecisionsTotal != 1 {
		t.Fatalf(
			"expected one fallback decision, got %d",
			snapshot.FallbackDecisionsTotal,
		)
	}

	if snapshot.FallbackSelectedTotal != 1 {
		t.Fatalf(
			"expected one fallback selection, got %d",
			snapshot.FallbackSelectedTotal,
		)
	}

	if snapshot.LatestFallback.
		ConsideredProviders[0] !=
		providerpolicy.ProviderAirplanesLive {
		t.Fatal(
			"expected stored considered providers to be immutable",
		)
	}

	if containsLimitation(
		snapshot.Limitations,
		LimitationFallbackNotObserved,
	) {
		t.Fatal(
			"did not expect fallback-not-observed limitation",
		)
	}
}

func TestCollectorReportsFallbackNotObserved(
	t *testing.T,
) {
	collector := New(nil)

	collector.RecordBudgetDecision(
		providerpolicy.ProviderAirplanesLive,
		"traffic:primary",
		"",
		providerbudget.Decision{
			Provider: providerpolicy.
				ProviderAirplanesLive,
			Allowed: true,
			Reason: providerbudget.
				DecisionReasonAllowed,
		},
	)

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"snapshot provider decision: %v",
			err,
		)
	}

	if !containsLimitation(
		snapshot.Limitations,
		LimitationFallbackNotObserved,
	) {
		t.Fatal(
			"expected fallback-not-observed limitation",
		)
	}
}

func TestCollectorRecordsFallbackDecisionsConcurrently(
	t *testing.T,
) {
	collector := New(nil)

	const goroutineCount = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for index := 0; index < goroutineCount; index++ {
		go func() {
			defer waitGroup.Done()

			collector.RecordFallbackDecision(
				providerfallback.Decision{
					PrimaryProvider: providerpolicy.
						ProviderAirplanesLive,
					SelectedProvider: providerpolicy.
						ProviderOpenSky,
					UsedFallback: true,
					Outcome: providerfallback.
						OutcomeFallbackSelected,
					TriggerReason: providerbudget.
						DecisionReasonFixedWindowExhausted,
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
			"snapshot concurrent fallback decisions: %v",
			err,
		)
	}

	if snapshot.FallbackDecisionsTotal !=
		goroutineCount {
		t.Fatalf(
			"expected %d fallback decisions, got %d",
			goroutineCount,
			snapshot.FallbackDecisionsTotal,
		)
	}
}

func containsLimitation(
	limitations []string,
	expected string,
) bool {
	for _, limitation := range limitations {
		if limitation == expected {
			return true
		}
	}

	return false
}
