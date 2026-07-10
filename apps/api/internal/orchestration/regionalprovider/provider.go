package regionalprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrProviderRequired = errors.New(
		"regional provider is required",
	)

	ErrExecutorRequired = errors.New(
		"regional provider executor is required",
	)

	ErrProviderIDRequired = errors.New(
		"regional provider identifier is required",
	)
)

type ExecutionValue struct {
	States []flightstate.FlightState
}

func (ExecutionValue) RequestCoalescingValue() {}

type Delegate interface {
	SourceName() string

	LoadByPoint(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) ([]flightstate.FlightState, error)
}

type Executor interface {
	Execute(
		ctx context.Context,
		provider providerpolicy.Provider,
		requestKey string,
		function ingestionorchestrator.Function[ExecutionValue],
	) (ingestionorchestrator.ExecuteResult[ExecutionValue], error)
}

type Config struct {
	Provider   Delegate
	ProviderID providerpolicy.Provider
	Executor   Executor
}

type Provider struct {
	delegate   Delegate
	providerID providerpolicy.Provider
	executor   Executor
}

func New(
	config Config,
) (*Provider, error) {
	if config.Provider == nil {
		return nil, ErrProviderRequired
	}

	if config.ProviderID == "" {
		return nil, ErrProviderIDRequired
	}

	if config.Executor == nil {
		return nil, ErrExecutorRequired
	}

	return &Provider{
		delegate:   config.Provider,
		providerID: config.ProviderID,
		executor:   config.Executor,
	}, nil
}

func (
	provider *Provider,
) SourceName() string {
	return provider.delegate.SourceName()
}

func (
	provider *Provider,
) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	result, err := provider.executor.Execute(
		ctx,
		provider.providerID,
		regionalRequestKey(
			latitude,
			longitude,
			radius,
		),
		func(
			operationContext context.Context,
		) (ExecutionValue, error) {
			states, err := provider.delegate.LoadByPoint(
				operationContext,
				latitude,
				longitude,
				radius,
			)
			if err != nil {
				return ExecutionValue{},
					err
			}

			return ExecutionValue{
				States: states,
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"execute orchestrated regional provider request: %w",
			err,
		)
	}

	return result.Value.States,
		nil
}

func regionalRequestKey(
	latitude float64,
	longitude float64,
	radius int,
) string {
	return fmt.Sprintf(
		"point:%g:%g:%d",
		latitude,
		longitude,
		radius,
	)
}
