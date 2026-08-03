package main

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerdecision"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/ingestdaemon"
	providerhealthservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/providerhealth"
)

func newIngestionCycleObserver(
	metricsRegistry *observability.Registry,
	trafficSelection trafficProviderSelection,
	providerHealthCollector *providerhealthservice.Collector,
	providerDecisionCollector *providerdecision.Collector,
) ingestdaemon.Observer {
	return func(
		result ingestdaemon.CycleResult,
	) {
		status := "success"
		lastError := ""

		if result.Err != nil {
			status = "failed"
			lastError = result.Err.Error()
		}

		metricsRegistry.ObserveIngestionCycle(
			status,
			result.FinishedAt.Sub(result.StartedAt),
			result.ConsecutiveFailures,
			result.NextDelay,
		)

		fmt.Printf(
			"ingest_cycle=%d status=%s started_at=%s finished_at=%s duration=%s consecutive_failures=%d retry_at=%s next_delay=%s error=%q\n",
			result.Number,
			status,
			result.StartedAt.Format(time.RFC3339Nano),
			result.FinishedAt.Format(time.RFC3339Nano),
			result.FinishedAt.Sub(result.StartedAt),
			result.ConsecutiveFailures,
			result.RetryAt.Format(time.RFC3339Nano),
			result.NextDelay,
			lastError,
		)

		printTrafficProviderEvidence(
			trafficSelection.ProviderIDs,
			providerHealthCollector,
			providerDecisionCollector,
		)
	}
}
