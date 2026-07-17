package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type fallbackTrafficProviderStub struct {
	sourceName string
	states     []flightstate.FlightState
	err        error
	calls      int
}

func (provider *fallbackTrafficProviderStub) SourceName() string {
	return provider.sourceName
}

func (provider *fallbackTrafficProviderStub) LoadByPoint(
	context.Context,
	float64,
	float64,
	int,
) ([]flightstate.FlightState, error) {
	provider.calls++
	return append(
		[]flightstate.FlightState(nil),
		provider.states...,
	), provider.err
}

type fallbackDecisionRecorderStub struct {
	decisions []providerfallback.Decision
}

func (recorder *fallbackDecisionRecorderStub) RecordFallbackDecision(
	decision providerfallback.Decision,
) {
	recorder.decisions = append(
		recorder.decisions,
		decision,
	)
}

func TestTrafficFallbackProviderUsesPrimaryWhenAvailable(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		states: []flightstate.FlightState{
			{ICAO24: "ABC123"},
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
	}
	recorder := &fallbackDecisionRecorderStub{}

	provider := mustTrafficFallbackProvider(
		t,
		primaryProvider,
		secondaryProvider,
		recorder,
	)

	result, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load primary provider: %v", err)
	}

	if result.SourceName != "airplanes.live" {
		t.Fatalf(
			"source = %q, want airplanes.live",
			result.SourceName,
		)
	}
	if primaryProvider.calls != 1 {
		t.Fatalf("primary calls = %d, want 1", primaryProvider.calls)
	}
	if secondaryProvider.calls != 0 {
		t.Fatalf("secondary calls = %d, want 0", secondaryProvider.calls)
	}
	if len(result.States) != 1 ||
		result.States[0].SourceName != "airplanes.live" {
		t.Fatalf("unexpected normalized states: %+v", result.States)
	}
	if len(recorder.decisions) != 1 ||
		recorder.decisions[0].Outcome !=
			providerfallback.OutcomePrimarySelected {
		t.Fatalf("unexpected fallback decisions: %+v", recorder.decisions)
	}
}

func TestTrafficFallbackProviderUsesOpenSkyAfterBudgetDenial(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		err: &ingestionorchestrator.AccessDeniedError{
			Provider: providerpolicy.ProviderAirplanesLive,
			Reason: providerbudget.
				DecisionReasonFixedWindowExhausted,
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		states: []flightstate.FlightState{
			{ICAO24: "DEF456"},
		},
	}
	recorder := &fallbackDecisionRecorderStub{}

	provider := mustTrafficFallbackProvider(
		t,
		primaryProvider,
		secondaryProvider,
		recorder,
	)

	result, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load fallback provider: %v", err)
	}

	if result.SourceName != "opensky" {
		t.Fatalf("source = %q, want opensky", result.SourceName)
	}
	if primaryProvider.calls != 1 || secondaryProvider.calls != 1 {
		t.Fatalf(
			"calls primary=%d secondary=%d, want 1 and 1",
			primaryProvider.calls,
			secondaryProvider.calls,
		)
	}
	if len(recorder.decisions) != 1 {
		t.Fatalf("decisions = %d, want 1", len(recorder.decisions))
	}
	decision := recorder.decisions[0]
	if decision.Outcome != providerfallback.OutcomeFallbackSelected ||
		decision.SelectedProvider != providerpolicy.ProviderOpenSky ||
		!decision.UsedFallback {
		t.Fatalf("unexpected fallback decision: %+v", decision)
	}
	if decision.TriggerReason !=
		providerbudget.DecisionReasonFixedWindowExhausted {
		t.Fatalf(
			"trigger reason = %q",
			decision.TriggerReason,
		)
	}
}

func TestTrafficFallbackProviderUsesOpenSkyAfterPrimaryServerFailure(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		err: fmt.Errorf(
			"primary request: %w",
			integrationcommon.ErrProviderServer,
		),
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		states: []flightstate.FlightState{
			{ICAO24: "FALL01"},
		},
	}
	recorder := &fallbackDecisionRecorderStub{}

	provider := mustTrafficFallbackProvider(
		t,
		primaryProvider,
		secondaryProvider,
		recorder,
	)

	result, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load server-error fallback: %v", err)
	}

	if result.SourceName != "opensky" {
		t.Fatalf("source = %q, want opensky", result.SourceName)
	}
	if recorder.decisions[0].TriggerReason !=
		providerbudget.DecisionReasonProviderUnavailable {
		t.Fatalf(
			"trigger reason = %q, want provider-unavailable",
			recorder.decisions[0].TriggerReason,
		)
	}
}

func TestTrafficFallbackProviderDoesNotHideUnauthorizedConfiguration(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		err: fmt.Errorf(
			"primary request: %w",
			integrationcommon.ErrProviderUnauthorized,
		),
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
	}
	recorder := &fallbackDecisionRecorderStub{}

	provider := mustTrafficFallbackProvider(
		t,
		primaryProvider,
		secondaryProvider,
		recorder,
	)

	_, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(
		err,
		integrationcommon.ErrProviderUnauthorized,
	) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if secondaryProvider.calls != 0 {
		t.Fatalf(
			"secondary calls = %d, want 0",
			secondaryProvider.calls,
		)
	}
	if len(recorder.decisions) != 0 {
		t.Fatalf(
			"decisions = %d, want 0",
			len(recorder.decisions),
		)
	}
}

func TestTrafficFallbackProviderReportsNoProviderAvailable(
	t *testing.T,
) {
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		err: fmt.Errorf(
			"primary request: %w",
			integrationcommon.ErrProviderServer,
		),
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		err: fmt.Errorf(
			"secondary request: %w",
			integrationcommon.ErrProviderRateLimited,
		),
	}
	recorder := &fallbackDecisionRecorderStub{}

	provider := mustTrafficFallbackProvider(
		t,
		primaryProvider,
		secondaryProvider,
		recorder,
	)

	_, err := provider.LoadByPointWithSource(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)

	var unavailableError *providerfallback.NoProviderAvailableError
	if !errors.As(err, &unavailableError) {
		t.Fatalf(
			"expected no-provider-available error, got %v",
			err,
		)
	}
	if unavailableError.Decision.Outcome !=
		providerfallback.OutcomeNoProviderAvailable {
		t.Fatalf(
			"unexpected decision: %+v",
			unavailableError.Decision,
		)
	}
	if len(recorder.decisions) != 1 {
		t.Fatalf("decisions = %d, want 1", len(recorder.decisions))
	}
}

func mustTrafficFallbackProvider(
	t *testing.T,
	primaryProvider *fallbackTrafficProviderStub,
	secondaryProvider *fallbackTrafficProviderStub,
	recorder *fallbackDecisionRecorderStub,
) *trafficFallbackProvider {
	t.Helper()

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
	)
	if err != nil {
		t.Fatalf("create traffic fallback provider: %v", err)
	}
	return provider
}
