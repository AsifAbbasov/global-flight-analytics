package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type fallbackHealthSourceStub struct {
	snapshots map[providerpolicy.Provider]providerhealthdomain.Snapshot
	errors    map[providerpolicy.Provider]error
}

func (source *fallbackHealthSourceStub) Snapshot(
	provider providerpolicy.Provider,
) (providerhealthdomain.Snapshot, error) {
	if source != nil && source.errors != nil {
		if err := source.errors[provider]; err != nil {
			return providerhealthdomain.Snapshot{}, err
		}
	}
	if source != nil && source.snapshots != nil {
		if snapshot, exists := source.snapshots[provider]; exists {
			return snapshot, nil
		}
	}

	return providerhealthdomain.Snapshot{
		ProviderName: string(provider),
		Status:       providerhealthdomain.StatusUnknown,
	}, nil
}

func TestTrafficFallbackProviderPrefersHealthySecondaryOverUnavailablePrimary(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		states: []flightstate.FlightState{
			{ICAO24: "PRIMARY"},
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		states: []flightstate.FlightState{
			{ICAO24: "HEALTHY"},
		},
	}
	recorder := &fallbackDecisionRecorderStub{}
	healthSource := &fallbackHealthSourceStub{
		snapshots: map[providerpolicy.Provider]providerhealthdomain.Snapshot{
			providerpolicy.ProviderAirplanesLive: {
				ProviderName: "airplanes.live",
				Status:       providerhealthdomain.StatusUnavailable,
			},
			providerpolicy.ProviderOpenSky: {
				ProviderName: "opensky",
				Status:       providerhealthdomain.StatusHealthy,
			},
		},
	}

	provider, err := newTrafficFallbackProvider(
		trafficProviderSelection{
			Provider:   primaryProvider,
			ProviderID: providerpolicy.ProviderAirplanesLive,
		},
		trafficProviderSelection{
			Provider:   secondaryProvider,
			ProviderID: providerpolicy.ProviderOpenSky,
		},
		providerfallback.New(nil),
		recorder,
		healthSource,
	)
	if err != nil {
		t.Fatalf("create health-aware fallback provider: %v", err)
	}

	result, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load health-preferred provider: %v", err)
	}

	if result.SourceName != "opensky" {
		t.Fatalf(
			"source = %q, want opensky",
			result.SourceName,
		)
	}
	if primaryProvider.calls != 0 ||
		secondaryProvider.calls != 1 {
		t.Fatalf(
			"calls primary=%d secondary=%d, want 0 and 1",
			primaryProvider.calls,
			secondaryProvider.calls,
		)
	}
	if len(recorder.decisions) != 1 {
		t.Fatalf(
			"decision count = %d, want 1",
			len(recorder.decisions),
		)
	}

	decision := recorder.decisions[0]
	if decision.PrimaryProvider !=
		providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"primary provider = %s, want airplanes.live",
			decision.PrimaryProvider,
		)
	}
	if decision.SelectedProvider !=
		providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"selected provider = %s, want opensky",
			decision.SelectedProvider,
		)
	}
	if decision.Outcome !=
		providerfallback.OutcomeFallbackSelected ||
		!decision.UsedFallback ||
		!decision.HealthAware ||
		!decision.HealthReordered {
		t.Fatalf(
			"unexpected health-aware decision: %+v",
			decision,
		)
	}
	if decision.TriggerReason !=
		providerbudget.DecisionReasonProviderUnavailable {
		t.Fatalf(
			"trigger reason = %s, want provider unavailable",
			decision.TriggerReason,
		)
	}
	if decision.PrimaryHealthStatus !=
		providerhealthdomain.StatusUnavailable ||
		decision.SelectedHealthStatus !=
			providerhealthdomain.StatusHealthy {
		t.Fatalf(
			"unexpected health statuses: primary=%s selected=%s",
			decision.PrimaryHealthStatus,
			decision.SelectedHealthStatus,
		)
	}
}

func TestTrafficFallbackProviderPreservesConfiguredOrderForEqualEvidence(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		states: []flightstate.FlightState{
			{ICAO24: "PRIMARY"},
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		states: []flightstate.FlightState{
			{ICAO24: "SECONDARY"},
		},
	}
	recorder := &fallbackDecisionRecorderStub{}
	healthSource := &fallbackHealthSourceStub{
		snapshots: map[providerpolicy.Provider]providerhealthdomain.Snapshot{
			providerpolicy.ProviderAirplanesLive: {
				ProviderName: "airplanes.live",
				Status:       providerhealthdomain.StatusDegraded,
			},
			providerpolicy.ProviderOpenSky: {
				ProviderName: "opensky",
				Status:       providerhealthdomain.StatusUnknown,
			},
		},
	}

	provider, err := newTrafficFallbackProvider(
		trafficProviderSelection{
			Provider:   primaryProvider,
			ProviderID: providerpolicy.ProviderAirplanesLive,
		},
		trafficProviderSelection{
			Provider:   secondaryProvider,
			ProviderID: providerpolicy.ProviderOpenSky,
		},
		providerfallback.New(nil),
		recorder,
		healthSource,
	)
	if err != nil {
		t.Fatalf("create fallback provider: %v", err)
	}

	result, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load configured primary: %v", err)
	}

	if result.SourceName != "airplanes.live" ||
		primaryProvider.calls != 1 ||
		secondaryProvider.calls != 0 {
		t.Fatalf(
			"source=%q calls primary=%d secondary=%d",
			result.SourceName,
			primaryProvider.calls,
			secondaryProvider.calls,
		)
	}
	if recorder.decisions[0].HealthReordered {
		t.Fatalf(
			"expected configured order, got %+v",
			recorder.decisions[0],
		)
	}
}

func TestTrafficFallbackProviderSchedulesConfiguredPrimaryRecoveryProbe(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		states: []flightstate.FlightState{
			{ICAO24: "RECOVERED"},
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		states: []flightstate.FlightState{
			{ICAO24: "HEALTHY"},
		},
	}
	recorder := &fallbackDecisionRecorderStub{}
	lastRequestAgeSeconds := int64(
		trafficProviderRecoveryProbeAfter / time.Second,
	)
	healthSource := &fallbackHealthSourceStub{
		snapshots: map[providerpolicy.Provider]providerhealthdomain.Snapshot{
			providerpolicy.ProviderAirplanesLive: {
				ProviderName:          "airplanes.live",
				Status:                providerhealthdomain.StatusUnavailable,
				LastRequestAgeSeconds: &lastRequestAgeSeconds,
			},
			providerpolicy.ProviderOpenSky: {
				ProviderName: "opensky",
				Status:       providerhealthdomain.StatusHealthy,
			},
		},
	}

	provider, err := newTrafficFallbackProvider(
		trafficProviderSelection{
			Provider:   primaryProvider,
			ProviderID: providerpolicy.ProviderAirplanesLive,
		},
		trafficProviderSelection{
			Provider:   secondaryProvider,
			ProviderID: providerpolicy.ProviderOpenSky,
		},
		providerfallback.New(nil),
		recorder,
		healthSource,
	)
	if err != nil {
		t.Fatalf("create fallback provider: %v", err)
	}

	result, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load recovery probe provider: %v", err)
	}

	if result.SourceName != "airplanes.live" ||
		primaryProvider.calls != 1 ||
		secondaryProvider.calls != 0 {
		t.Fatalf(
			"source=%q calls primary=%d secondary=%d",
			result.SourceName,
			primaryProvider.calls,
			secondaryProvider.calls,
		)
	}
	decision := recorder.decisions[0]
	if decision.HealthOrderingReason !=
		trafficHealthReasonRecoveryProbe ||
		decision.HealthReordered {
		t.Fatalf(
			"unexpected recovery decision: %+v",
			decision,
		)
	}
}

func TestTrafficFallbackProviderFailsOpenWhenHealthSnapshotFails(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		states: []flightstate.FlightState{
			{ICAO24: "PRIMARY"},
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
	}
	recorder := &fallbackDecisionRecorderStub{}
	healthSource := &fallbackHealthSourceStub{
		errors: map[providerpolicy.Provider]error{
			providerpolicy.ProviderAirplanesLive: errors.New("health collector unavailable"),
		},
	}

	provider, err := newTrafficFallbackProvider(
		trafficProviderSelection{
			Provider:   primaryProvider,
			ProviderID: providerpolicy.ProviderAirplanesLive,
		},
		trafficProviderSelection{
			Provider:   secondaryProvider,
			ProviderID: providerpolicy.ProviderOpenSky,
		},
		providerfallback.New(nil),
		recorder,
		healthSource,
	)
	if err != nil {
		t.Fatalf("create fallback provider: %v", err)
	}

	_, err = provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load fail-open primary: %v", err)
	}

	if primaryProvider.calls != 1 ||
		secondaryProvider.calls != 0 {
		t.Fatalf(
			"calls primary=%d secondary=%d, want 1 and 0",
			primaryProvider.calls,
			secondaryProvider.calls,
		)
	}
	decision := recorder.decisions[0]
	if decision.HealthOrderingReason !=
		trafficHealthReasonSnapshotUnavailable {
		t.Fatalf(
			"health reason = %q, want snapshot unavailable",
			decision.HealthOrderingReason,
		)
	}
}
