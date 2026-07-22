package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/opensky"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestTrafficFallbackProviderRecordsMixedTerminalFailure(
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
			integrationcommon.ErrProviderUnauthorized,
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
	if !errors.Is(
		err,
		integrationcommon.ErrProviderUnauthorized,
	) {
		t.Fatalf(
			"expected unauthorized terminal error, got %v",
			err,
		)
	}
	if len(recorder.decisions) != 1 {
		t.Fatalf(
			"decision count = %d, want 1",
			len(recorder.decisions),
		)
	}
	decision := recorder.decisions[0]
	if decision.Outcome !=
		providerfallback.OutcomeTerminalFailure {
		t.Fatalf(
			"outcome = %s, want terminal_failure",
			decision.Outcome,
		)
	}
	if len(decision.Attempts) != 2 {
		t.Fatalf(
			"attempt count = %d, want 2",
			len(decision.Attempts),
		)
	}
	if decision.Attempts[1].ErrorClass !=
		providerfallback.AttemptErrorClassUnauthorized {
		t.Fatalf(
			"secondary error class = %s",
			decision.Attempts[1].ErrorClass,
		)
	}
}

func TestTrafficFallbackProviderReportsNoExternalRequestWhenAllProvidersAreLocallyDenied(
	t *testing.T,
) {
	retryAt := time.Now().UTC().Add(
		time.Minute,
	)
	primaryProvider := &fallbackTrafficProviderStub{
		sourceName: "airplanes.live",
		err: &ingestionorchestrator.AccessDeniedError{
			Provider: providerpolicy.ProviderAirplanesLive,
			Reason: providerbudget.
				DecisionReasonFixedWindowExhausted,
			RetryAt: retryAt,
		},
	}
	secondaryProvider := &fallbackTrafficProviderStub{
		sourceName: "opensky",
		err: &opensky.PollingTooSoonError{
			RetryAt: retryAt,
		},
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
	var unavailable *providerfallback.NoProviderAvailableError
	if !errors.As(
		err,
		&unavailable,
	) {
		t.Fatalf(
			"expected no provider available error, got %v",
			err,
		)
	}
	if unavailable.ExternalRequestAttempted() {
		t.Fatal(
			"expected no external HTTP attempt",
		)
	}
	if !unavailable.RetryAtTime().Equal(
		retryAt,
	) {
		t.Fatalf(
			"retry at = %s, want %s",
			unavailable.RetryAtTime(),
			retryAt,
		)
	}
}
