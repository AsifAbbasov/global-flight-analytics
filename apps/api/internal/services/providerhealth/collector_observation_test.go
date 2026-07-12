package providerhealth

import (
	"slices"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestCollectorKnowsThatEmptyObservationBatchWasEvaluated(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		12,
		16,
		0,
		0,
		0,
		time.UTC,
	)
	collector := New(
		func() time.Time {
			return currentTime
		},
	)

	err := collector.RecordObservationEvidence(
		providerpolicy.ProviderAirplanesLive,
		0,
		0,
		0,
	)
	if err != nil {
		t.Fatalf("RecordObservationEvidence() error = %v", err)
	}

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if slices.Contains(
		snapshot.Limitations,
		"provider_observation_quality_not_observed",
	) {
		t.Fatalf(
			"limitations incorrectly report missing observation evidence: %v",
			snapshot.Limitations,
		)
	}
}
