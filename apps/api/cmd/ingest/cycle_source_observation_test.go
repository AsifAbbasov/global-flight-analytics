package main

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

type sourceObservationRecorderStub struct {
	provider providerpolicy.Provider
	calls    int
}

func (recorder *sourceObservationRecorderStub) RecordObservationEvidence(
	provider providerpolicy.Provider,
	received int64,
	accepted int64,
	rejected int64,
) error {
	recorder.provider = provider
	recorder.calls++
	return nil
}

func TestIngestionCycleAttributesObservationEvidenceToSelectedSource(
	t *testing.T,
) {
	recorder := &sourceObservationRecorderStub{}
	cycle := &ingestionCycle{
		providerID:          providerpolicy.ProviderAirplanesLive,
		observationRecorder: recorder,
	}

	cycle.observeProviderEvidence(
		trafficingestion.LoadAndProcessResult{
			SourceName: string(
				providerpolicy.ProviderOpenSky,
			),
			ProcessingResult: trafficapplication.ProcessAndStoreResult{
				ProcessingResult: processor.ProcessingResult{
					ProcessedAt: time.Date(
						2026,
						time.July,
						18,
						0,
						0,
						0,
						0,
						time.UTC,
					),
					Stats: processor.ProcessingStats{
						ReceivedCount: 4,
						UsableCount:   3,
						InvalidCount:  1,
					},
				},
			},
		},
	)

	if recorder.calls != 1 {
		t.Fatalf("recorder calls = %d, want 1", recorder.calls)
	}
	if recorder.provider != providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"provider = %q, want %q",
			recorder.provider,
			providerpolicy.ProviderOpenSky,
		)
	}
}
