package providerdecision

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestFallbackAttemptEvidenceIsCopiedAtCollectorBoundary(
	t *testing.T,
) {
	collector := New(nil)
	decision := providerfallback.Decision{
		PrimaryProvider: providerpolicy.ProviderAirplanesLive,
		Outcome:         providerfallback.OutcomeTerminalFailure,
		Attempts: []providerfallback.AttemptEvidence{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Outcome:  providerfallback.AttemptOutcomeFailed,
			},
		},
	}
	collector.RecordFallbackDecision(
		decision,
	)

	decision.Attempts[0].Outcome =
		providerfallback.AttemptOutcomeSuccess

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"load provider decision snapshot: %v",
			err,
		)
	}
	if snapshot.LatestFallback.Attempts[0].Outcome !=
		providerfallback.AttemptOutcomeFailed {
		t.Fatalf(
			"stored attempt changed through input alias: %+v",
			snapshot.LatestFallback.Attempts[0],
		)
	}

	snapshot.LatestFallback.Attempts[0].Outcome =
		providerfallback.AttemptOutcomeSuccess

	secondSnapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"reload provider decision snapshot: %v",
			err,
		)
	}
	if secondSnapshot.LatestFallback.Attempts[0].Outcome !=
		providerfallback.AttemptOutcomeFailed {
		t.Fatalf(
			"stored attempt changed through output alias: %+v",
			secondSnapshot.LatestFallback.Attempts[0],
		)
	}
}
