package providerhealth

import (
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
)

func TestCollectorBuildsHealthySnapshotAfterMinimumEvidence(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	collector := New(func() time.Time { return currentTime })

	for range 5 {
		err := collector.RecordHTTPResponse(
			providerresponse.Observation{
				Provider:   providerpolicy.ProviderAirplanesLive,
				StatusCode: http.StatusOK,
			},
			100*time.Millisecond,
		)
		if err != nil {
			t.Fatalf("RecordHTTPResponse() error = %v", err)
		}
	}

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if snapshot.Status != providerhealthdomain.StatusHealthy {
		t.Fatalf(
			"status = %q, want %q",
			snapshot.Status,
			providerhealthdomain.StatusHealthy,
		)
	}
	if snapshot.RequestsTotal != 5 {
		t.Fatalf("requests total = %d, want 5", snapshot.RequestsTotal)
	}
	if snapshot.SuccessRatio != 1 {
		t.Fatalf("success ratio = %v, want 1", snapshot.SuccessRatio)
	}
	if snapshot.AverageLatency != 100*time.Millisecond {
		t.Fatalf(
			"average latency = %s, want %s",
			snapshot.AverageLatency,
			100*time.Millisecond,
		)
	}
	if slices.Contains(
		snapshot.Limitations,
		"provider_request_latency_not_observed",
	) {
		t.Fatalf("limitations = %v", snapshot.Limitations)
	}
}

func TestCollectorMarksRateLimitedProviderUnavailable(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	collector := New(func() time.Time { return currentTime })
	cooldownUntil := currentTime.Add(30 * time.Second)

	err := collector.RecordHTTPResponse(
		providerresponse.Observation{
			Provider:      providerpolicy.ProviderAirplanesLive,
			StatusCode:    http.StatusTooManyRequests,
			CooldownUntil: cooldownUntil,
		},
		250*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("RecordHTTPResponse() error = %v", err)
	}

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if snapshot.Status != providerhealthdomain.StatusUnavailable {
		t.Fatalf(
			"status = %q, want %q",
			snapshot.Status,
			providerhealthdomain.StatusUnavailable,
		)
	}
	if snapshot.LatestOutcome != providerhealthdomain.RequestOutcomeRateLimited {
		t.Fatalf("latest outcome = %q", snapshot.LatestOutcome)
	}
	if snapshot.Budget.State != providerhealthdomain.BudgetStateExhausted {
		t.Fatalf("budget state = %q", snapshot.Budget.State)
	}
	if snapshot.Budget.ResetsAt == nil ||
		!snapshot.Budget.ResetsAt.Equal(cooldownUntil) {
		t.Fatalf("budget reset = %v, want %v", snapshot.Budget.ResetsAt, cooldownUntil)
	}
}

func TestCollectorRecordsObservationEvidence(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	collector := New(func() time.Time { return currentTime })

	err := collector.RecordObservationEvidence(
		providerpolicy.ProviderAirplanesLive,
		100,
		90,
		10,
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

	if snapshot.Observations.Received != 100 ||
		snapshot.Observations.Accepted != 90 ||
		snapshot.Observations.Rejected != 10 {
		t.Fatalf("observations = %+v", snapshot.Observations)
	}
}

func TestCollectorIsConcurrentSafe(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	collector := New(func() time.Time { return currentTime })

	const requestCount = 100
	var waitGroup sync.WaitGroup
	waitGroup.Add(requestCount)

	for range requestCount {
		go func() {
			defer waitGroup.Done()

			err := collector.RecordHTTPResponse(
				providerresponse.Observation{
					Provider:   providerpolicy.ProviderAirplanesLive,
					StatusCode: http.StatusOK,
				},
				10*time.Millisecond,
			)
			if err != nil {
				t.Errorf("RecordHTTPResponse() error = %v", err)
			}
		}()
	}

	waitGroup.Wait()

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if snapshot.RequestsTotal != requestCount {
		t.Fatalf(
			"requests total = %d, want %d",
			snapshot.RequestsTotal,
			requestCount,
		)
	}
}

func TestCollectorRecordsInvalidResponseFailure(
	t *testing.T,
) {
	t.Parallel()

	currentTime := time.Date(
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
			return currentTime
		},
	)

	err := collector.RecordResponseFailure(
		providerpolicy.ProviderAirplanesLive,
		125*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("RecordResponseFailure() error = %v", err)
	}

	snapshot, err := collector.Snapshot(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if snapshot.RequestsTotal != 1 {
		t.Fatalf(
			"requests total = %d, want 1",
			snapshot.RequestsTotal,
		)
	}
	if snapshot.RequestsSuccessful != 0 {
		t.Fatalf(
			"successful requests = %d, want 0",
			snapshot.RequestsSuccessful,
		)
	}
	if snapshot.LatestOutcome !=
		providerhealthdomain.RequestOutcomeInvalidResponse {
		t.Fatalf(
			"latest outcome = %q, want %q",
			snapshot.LatestOutcome,
			providerhealthdomain.RequestOutcomeInvalidResponse,
		)
	}
}
