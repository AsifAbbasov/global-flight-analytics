package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

type cycleTestObservationRecorder struct {
	provider providerpolicy.Provider
	received int64
	accepted int64
	rejected int64
	calls    int
}

func (recorder *cycleTestObservationRecorder) RecordObservationEvidence(
	provider providerpolicy.Provider,
	received int64,
	accepted int64,
	rejected int64,
) error {
	recorder.provider = provider
	recorder.received = received
	recorder.accepted = accepted
	recorder.rejected = rejected
	recorder.calls++

	return nil
}

type cycleTestUnsupportedTrafficProvider struct {
	cycleTestTrafficProvider
}

func (
	*cycleTestUnsupportedTrafficProvider,
) SourceName() string {
	return "unsupported-provider"
}

func TestIngestionCycleRecordsProviderObservationEvidence(
	t *testing.T,
) {
	processedAt := time.Date(
		2026,
		time.July,
		12,
		16,
		0,
		0,
		0,
		time.UTC,
	)

	processingService := &cycleTestProcessingService{
		result: trafficapplication.ProcessAndStoreResult{
			ProcessingResult: processor.ProcessingResult{
				ProcessedAt: processedAt,
				Stats: processor.ProcessingStats{
					ReceivedCount:  10,
					UsableCount:    7,
					InvalidCount:   2,
					DuplicateCount: 1,
				},
			},
			StoredAt: processedAt,
		},
	}

	recorder := &cycleTestObservationRecorder{}

	cycle, err := newIngestionCycle(
		ingestionCycleConfig{
			TrafficProvider:        &cycleTestTrafficProvider{},
			ProcessingService:      processingService,
			IngestionRunRepository: &cycleTestRunRepository{},
			ObservationRecorder:    recorder,
		},
	)
	if err != nil {
		t.Fatalf("newIngestionCycle() error = %v", err)
	}

	if err := cycle.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if recorder.calls != 1 {
		t.Fatalf("recorder calls = %d, want 1", recorder.calls)
	}
	if recorder.provider != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"provider = %q, want %q",
			recorder.provider,
			providerpolicy.ProviderAirplanesLive,
		)
	}
	if recorder.received != 10 ||
		recorder.accepted != 7 ||
		recorder.rejected != 3 {
		t.Fatalf(
			"evidence = received:%d accepted:%d rejected:%d",
			recorder.received,
			recorder.accepted,
			recorder.rejected,
		)
	}
}

func TestIngestionCycleRejectsUnsupportedProviderIdentity(
	t *testing.T,
) {
	cycle, err := newIngestionCycle(
		ingestionCycleConfig{
			TrafficProvider:        &cycleTestUnsupportedTrafficProvider{},
			ProcessingService:      &cycleTestProcessingService{},
			IngestionRunRepository: &cycleTestRunRepository{},
			ObservationRecorder:    &cycleTestObservationRecorder{},
		},
	)

	if cycle != nil {
		t.Fatal("expected nil ingestion cycle")
	}
	if !errors.Is(
		err,
		errIngestionCycleProviderIdentityInvalid,
	) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			errIngestionCycleProviderIdentityInvalid,
		)
	}
}
