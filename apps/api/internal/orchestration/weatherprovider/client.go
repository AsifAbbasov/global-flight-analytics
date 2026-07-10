package weatherprovider

import (
	"context"
	"errors"
	"fmt"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrClientRequired = errors.New(
		"weather provider client is required",
	)

	ErrExecutorRequired = errors.New(
		"weather provider executor is required",
	)
)

type ExecutionValue struct {
	Snapshot domainweather.CurrentSnapshot
}

func (ExecutionValue) RequestCoalescingValue() {}

type Delegate interface {
	GetCurrentWeather(
		ctx context.Context,
		request openmeteo.CurrentWeatherRequest,
	) (domainweather.CurrentSnapshot, error)
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
	Client   Delegate
	Executor Executor
}

type Client struct {
	delegate Delegate
	executor Executor
}

func New(
	config Config,
) (*Client, error) {
	if config.Client == nil {
		return nil, ErrClientRequired
	}

	if config.Executor == nil {
		return nil, ErrExecutorRequired
	}

	return &Client{
		delegate: config.Client,
		executor: config.Executor,
	}, nil
}

func (
	client *Client,
) GetCurrentWeather(
	ctx context.Context,
	request openmeteo.CurrentWeatherRequest,
) (domainweather.CurrentSnapshot, error) {
	result, err := client.executor.Execute(
		ctx,
		providerpolicy.ProviderOpenMeteo,
		currentWeatherRequestKey(
			request,
		),
		func(
			operationContext context.Context,
		) (ExecutionValue, error) {
			snapshot, err := client.delegate.GetCurrentWeather(
				operationContext,
				request,
			)
			if err != nil {
				return ExecutionValue{},
					err
			}

			return ExecutionValue{
				Snapshot: snapshot,
			}, nil
		},
	)
	if err != nil {
		return domainweather.CurrentSnapshot{},
			fmt.Errorf(
				"execute orchestrated weather request: %w",
				err,
			)
	}

	return result.Value.Snapshot,
		nil
}

func currentWeatherRequestKey(
	request openmeteo.CurrentWeatherRequest,
) string {
	return fmt.Sprintf(
		"current:%g:%g",
		request.Latitude,
		request.Longitude,
	)
}
