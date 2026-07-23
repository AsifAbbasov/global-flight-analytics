package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

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
	primary      trafficProviderSelection
	secondary    trafficProviderSelection
	selector     *providerfallback.Selector
	recorder     providerfallback.DecisionRecorder
	healthSource trafficProviderHealthSource
}

func newTrafficFallbackProvider(
	primary trafficProviderSelection,
	secondary trafficProviderSelection,
	selector *providerfallback.Selector,
	recorder providerfallback.DecisionRecorder,
	healthSources ...trafficProviderHealthSource,
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

	var healthSource trafficProviderHealthSource
	if len(healthSources) > 0 {
		healthSource = healthSources[0]
	}

	return &trafficFallbackProvider{
		primary:      primary,
		secondary:    secondary,
		selector:     selector,
		recorder:     recorder,
		healthSource: healthSource,
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

	configuredSelections := []trafficProviderSelection{
		provider.primary,
		provider.secondary,
	}
	selections, healthOrder := orderTrafficProviderSelections(
		configuredSelections,
		provider.healthSource,
	)
	healthAwareSelectAndRecord := provider.selectAndRecord
	selectAndRecord := func(
		candidates []providerfallback.Candidate,
	) (providerfallback.Decision, error) {
		return healthAwareSelectAndRecord(
			candidates,
			healthOrder,
		)
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
					Provider:         selection.ProviderID,
					Allowed:          true,
					Outcome:          providerfallback.AttemptOutcomeSuccess,
					RequestAttempted: true,
				},
			)

			decision, selectErr := selectAndRecord(
				candidates,
			)
			if selectErr != nil {
				return trafficingestion.LoadResult{},
					fmt.Errorf(
						"select successful traffic provider: %w",
						selectErr,
					)
			}

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
			if len(candidates) > 0 {
				candidates = append(
					candidates,
					terminalTrafficFallbackCandidate(
						selection.ProviderID,
						ctx.Err(),
						providerfallback.AttemptErrorClassCancelled,
						true,
					),
				)
				if _, selectErr := selectAndRecord(
					candidates,
				); selectErr != nil {
					return trafficingestion.LoadResult{
							SourceName: string(selection.ProviderID),
						}, errors.Join(
							fmt.Errorf(
								"traffic provider operation context ended: %w",
								ctx.Err(),
							),
							fmt.Errorf(
								"record cancelled fallback chain: %w",
								selectErr,
							),
						)
				}
			}

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
		if recoverable {
			candidates = append(
				candidates,
				candidate,
			)
			continue
		}

		candidates = append(
			candidates,
			terminalTrafficFallbackCandidate(
				selection.ProviderID,
				err,
				classifyTerminalTrafficError(err),
				externalRequestAttemptedByError(err),
			),
		)
		_, selectErr := selectAndRecord(
			candidates,
		)
		operationErr := fmt.Errorf(
			"execute traffic provider %s: %w",
			selection.ProviderID,
			err,
		)
		if selectErr != nil {
			return trafficingestion.LoadResult{
					SourceName: string(selection.ProviderID),
				}, errors.Join(
					operationErr,
					fmt.Errorf(
						"record terminal fallback chain: %w",
						selectErr,
					),
				)
		}

		return trafficingestion.LoadResult{
			SourceName: string(selection.ProviderID),
		}, operationErr
	}

	decision, err := selectAndRecord(
		candidates,
	)
	if err != nil {
		return trafficingestion.LoadResult{},
			fmt.Errorf(
				"select unavailable traffic providers: %w",
				err,
			)
	}

	return trafficingestion.LoadResult{
			SourceName: string(
				decision.PrimaryProvider,
			),
		}, &providerfallback.NoProviderAvailableError{
			Decision: decision,
		}
}

func (
	provider *trafficFallbackProvider,
) selectAndRecord(
	candidates []providerfallback.Candidate,
	healthOrders ...trafficProviderHealthOrder,
) (providerfallback.Decision, error) {
	decision, err := provider.selector.Select(
		candidates,
	)
	if err != nil {
		return providerfallback.Decision{}, err
	}

	if len(healthOrders) > 0 {
		decision = decorateTrafficFallbackDecision(
			decision,
			healthOrders[0],
		)
	}

	provider.recorder.RecordFallbackDecision(
		decision,
	)

	return decision, nil
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
			Provider:         providerID,
			Allowed:          false,
			DenialReason:     accessDeniedError.Reason,
			RetryAt:          accessDeniedError.RetryAt,
			Outcome:          providerfallback.AttemptOutcomeDenied,
			ErrorClass:       providerfallback.AttemptErrorClassAccessDenied,
			RequestAttempted: false,
		}, true
	}

	switch {
	case errors.Is(
		requestErr,
		integrationcommon.ErrProviderRateLimited,
	):
		return providerfallback.Candidate{
			Provider:         providerID,
			Allowed:          false,
			DenialReason:     providerbudget.DecisionReasonProviderBudgetExhausted,
			Outcome:          providerfallback.AttemptOutcomeFailed,
			ErrorClass:       providerfallback.AttemptErrorClassRateLimited,
			RequestAttempted: true,
		}, true

	case errors.Is(
		requestErr,
		opensky.ErrPollingTooSoon,
	):
		return providerfallback.Candidate{
			Provider:         providerID,
			Allowed:          false,
			DenialReason:     providerbudget.DecisionReasonProviderCooldown,
			RetryAt:          retryAtFromError(requestErr),
			Outcome:          providerfallback.AttemptOutcomeDenied,
			ErrorClass:       providerfallback.AttemptErrorClassPollingCooldown,
			RequestAttempted: false,
		}, true

	case errors.Is(
		requestErr,
		integrationcommon.ErrProviderServer,
	):
		return unavailableTrafficProviderCandidate(
			providerID,
			providerfallback.AttemptErrorClassProviderServer,
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
		errorClass := providerfallback.AttemptErrorClassNetwork
		if networkError.Timeout() {
			errorClass = providerfallback.AttemptErrorClassTimeout
		}
		return unavailableTrafficProviderCandidate(
			providerID,
			errorClass,
		), true
	}

	return providerfallback.Candidate{}, false
}

func unavailableTrafficProviderCandidate(
	providerID providerpolicy.Provider,
	errorClass providerfallback.AttemptErrorClass,
) providerfallback.Candidate {
	return providerfallback.Candidate{
		Provider:         providerID,
		Allowed:          false,
		DenialReason:     providerbudget.DecisionReasonProviderUnavailable,
		Outcome:          providerfallback.AttemptOutcomeFailed,
		ErrorClass:       errorClass,
		RequestAttempted: true,
	}
}

func terminalTrafficFallbackCandidate(
	providerID providerpolicy.Provider,
	requestErr error,
	errorClass providerfallback.AttemptErrorClass,
	requestAttempted bool,
) providerfallback.Candidate {
	return providerfallback.Candidate{
		Provider:         providerID,
		Allowed:          false,
		DenialReason:     providerbudget.DecisionReasonProviderUnavailable,
		RetryAt:          retryAtFromError(requestErr),
		Outcome:          providerfallback.AttemptOutcomeTerminalFailure,
		ErrorClass:       errorClass,
		RequestAttempted: requestAttempted,
	}
}

func classifyTerminalTrafficError(
	requestErr error,
) providerfallback.AttemptErrorClass {
	switch {
	case errors.Is(
		requestErr,
		integrationcommon.ErrProviderUnauthorized,
	):
		return providerfallback.AttemptErrorClassUnauthorized
	case errors.Is(
		requestErr,
		integrationcommon.ErrProviderResponseTooLarge,
	):
		return providerfallback.AttemptErrorClassResponseTooLarge
	case errors.Is(
		requestErr,
		context.DeadlineExceeded,
	):
		return providerfallback.AttemptErrorClassTimeout
	case errors.Is(
		requestErr,
		context.Canceled,
	):
		return providerfallback.AttemptErrorClassCancelled
	default:
		return providerfallback.AttemptErrorClassUnknown
	}
}

func retryAtFromError(
	err error,
) time.Time {
	var evidence interface {
		RetryAtTime() time.Time
	}
	if errors.As(err, &evidence) {
		return evidence.RetryAtTime().UTC()
	}

	return time.Time{}
}

func externalRequestAttemptedByError(
	err error,
) bool {
	var evidence interface {
		ExternalRequestAttempted() bool
	}
	if errors.As(err, &evidence) {
		return evidence.ExternalRequestAttempted()
	}

	return true
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
