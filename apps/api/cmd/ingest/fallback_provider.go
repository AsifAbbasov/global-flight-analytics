package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/opensky"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
)

var (
	errTrafficFallbackPrimaryRequired = errors.New(
		"primary traffic fallback provider is required",
	)
	errTrafficFallbackSecondaryRequired = errors.New(
		"secondary traffic fallback provider is required",
	)
	errTrafficFallbackProviderIdentityRequired = errors.New(
		"traffic fallback provider identity is required",
	)
	errTrafficFallbackDuplicateProvider = errors.New(
		"traffic fallback providers must be different",
	)
	errTrafficFallbackSelectorRequired = errors.New(
		"traffic fallback selector is required",
	)
	errTrafficFallbackRecorderRequired = errors.New(
		"traffic fallback decision recorder is required",
	)
)

type trafficFallbackProvider struct {
	primary   trafficProviderSelection
	secondary trafficProviderSelection
	selector  *providerfallback.Selector
	recorder  providerfallback.DecisionRecorder
}

func newTrafficFallbackProvider(
	primary trafficProviderSelection,
	secondary trafficProviderSelection,
	selector *providerfallback.Selector,
	recorder providerfallback.DecisionRecorder,
) (*trafficFallbackProvider, error) {
	if primary.Provider == nil {
		return nil, errTrafficFallbackPrimaryRequired
	}
	if secondary.Provider == nil {
		return nil, errTrafficFallbackSecondaryRequired
	}
	if primary.ProviderID == "" || secondary.ProviderID == "" {
		return nil, errTrafficFallbackProviderIdentityRequired
	}
	if primary.ProviderID == secondary.ProviderID {
		return nil, errTrafficFallbackDuplicateProvider
	}
	if selector == nil {
		return nil, errTrafficFallbackSelectorRequired
	}
	if recorder == nil {
		return nil, errTrafficFallbackRecorderRequired
	}

	return &trafficFallbackProvider{
		primary:   primary,
		secondary: secondary,
		selector:  selector,
		recorder:  recorder,
	}, nil
}

func (provider *trafficFallbackProvider) SourceName() string {
	if provider == nil {
		return ""
	}
	return string(provider.primary.ProviderID)
}

func (provider *trafficFallbackProvider) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	result, err := provider.LoadByPointWithSource(
		ctx,
		latitude,
		longitude,
		radius,
	)
	return result.States, err
}

func (provider *trafficFallbackProvider) LoadByPointWithSource(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) (trafficingestion.LoadResult, error) {
	if provider == nil {
		return trafficingestion.LoadResult{},
			errTrafficFallbackPrimaryRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}

	selections := []trafficProviderSelection{
		provider.primary,
		provider.secondary,
	}
	candidates := make(
		[]providerfallback.Candidate,
		0,
		len(selections),
	)

	for _, selection := range selections {
		states, err := selection.Provider.LoadByPoint(
			ctx,
			latitude,
			longitude,
			radius,
		)
		if err == nil {
			candidates = append(
				candidates,
				providerfallback.Candidate{
					Provider: selection.ProviderID,
					Allowed:  true,
				},
			)

			decision, selectErr := provider.selector.Select(
				candidates,
			)
			if selectErr != nil {
				return trafficingestion.LoadResult{},
					fmt.Errorf(
						"select successful traffic provider: %w",
						selectErr,
					)
			}

			provider.recorder.RecordFallbackDecision(
				decision,
			)

			sourceName := string(
				decision.SelectedProvider,
			)
			normalizedStates := normalizeTrafficStateSources(
				states,
				sourceName,
			)

			return trafficingestion.LoadResult{
				SourceName: sourceName,
				States:     normalizedStates,
			}, nil
		}

		if ctx.Err() != nil {
			return trafficingestion.LoadResult{
					SourceName: string(selection.ProviderID),
				}, fmt.Errorf(
					"traffic provider operation context ended: %w",
					ctx.Err(),
				)
		}

		candidate, recoverable := trafficFallbackCandidate(
			selection.ProviderID,
			err,
		)
		if !recoverable {
			return trafficingestion.LoadResult{
					SourceName: string(selection.ProviderID),
				}, fmt.Errorf(
					"execute traffic provider %s: %w",
					selection.ProviderID,
					err,
				)
		}
		candidates = append(
			candidates,
			candidate,
		)
	}

	decision, err := provider.selector.Select(
		candidates,
	)
	if err != nil {
		return trafficingestion.LoadResult{},
			fmt.Errorf(
				"select unavailable traffic providers: %w",
				err,
			)
	}

	provider.recorder.RecordFallbackDecision(
		decision,
	)

	return trafficingestion.LoadResult{
			SourceName: string(
				decision.PrimaryProvider,
			),
		}, &providerfallback.NoProviderAvailableError{
			Decision: decision,
		}
}

func trafficFallbackCandidate(
	providerID providerpolicy.Provider,
	requestErr error,
) (providerfallback.Candidate, bool) {
	var accessDeniedError *ingestionorchestrator.AccessDeniedError
	if errors.As(
		requestErr,
		&accessDeniedError,
	) {
		if accessDeniedError.Provider != providerID {
			return providerfallback.Candidate{}, false
		}
		return providerfallback.Candidate{
			Provider:     providerID,
			Allowed:      false,
			DenialReason: accessDeniedError.Reason,
			RetryAt:      accessDeniedError.RetryAt,
		}, true
	}

	switch {
	case errors.Is(
		requestErr,
		integrationcommon.ErrProviderRateLimited,
	):
		return providerfallback.Candidate{
			Provider: providerID,
			Allowed:  false,
			DenialReason: providerbudget.
				DecisionReasonProviderBudgetExhausted,
		}, true

	case errors.Is(
		requestErr,
		opensky.ErrPollingTooSoon,
	):
		return providerfallback.Candidate{
			Provider: providerID,
			Allowed:  false,
			DenialReason: providerbudget.
				DecisionReasonProviderCooldown,
		}, true

	case errors.Is(
		requestErr,
		integrationcommon.ErrProviderServer,
	):
		return unavailableTrafficProviderCandidate(
			providerID,
		), true

	case errors.Is(
		requestErr,
		context.Canceled,
	):
		return providerfallback.Candidate{}, false
	}

	var networkError net.Error
	if errors.As(
		requestErr,
		&networkError,
	) {
		return unavailableTrafficProviderCandidate(
			providerID,
		), true
	}

	return providerfallback.Candidate{}, false
}

func unavailableTrafficProviderCandidate(
	providerID providerpolicy.Provider,
) providerfallback.Candidate {
	return providerfallback.Candidate{
		Provider: providerID,
		Allowed:  false,
		DenialReason: providerbudget.
			DecisionReasonProviderUnavailable,
	}
}

func normalizeTrafficStateSources(
	states []flightstate.FlightState,
	sourceName string,
) []flightstate.FlightState {
	normalizedSourceName := strings.TrimSpace(
		sourceName,
	)
	result := append(
		[]flightstate.FlightState(nil),
		states...,
	)
	for index := range result {
		result[index].SourceName = normalizedSourceName
	}
	return result
}
