package weatherprovider

import (
	"context"
	"testing"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type delegateStub struct {
	callCount int
}

func (
	stub *delegateStub,
) GetCurrentWeather(
	_ context.Context,
	request openmeteo.CurrentWeatherRequest,
) (domainweather.CurrentSnapshot, error) {
	stub.callCount++

	return domainweather.CurrentSnapshot{
		Provider:  domainweather.ProviderOpenMeteo,
		Latitude:  request.Latitude,
		Longitude: request.Longitude,
	}, nil
}

type executorStub struct {
	callCount  int
	provider   providerpolicy.Provider
	requestKey string
}

func (
	stub *executorStub,
) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function ingestionorchestrator.Function[ExecutionValue],
) (ingestionorchestrator.ExecuteResult[ExecutionValue], error) {
	stub.callCount++
	stub.provider = provider
	stub.requestKey = requestKey

	value, err := function(
		ctx,
	)
	if err != nil {
		return ingestionorchestrator.ExecuteResult[ExecutionValue]{},
			err
	}

	return ingestionorchestrator.ExecuteResult[ExecutionValue]{
		Provider:   provider,
		RequestKey: requestKey,
		Value:      value,
	}, nil
}

func TestGetCurrentWeatherExecutesThroughTypedOrchestrator(
	t *testing.T,
) {
	delegate := &delegateStub{}
	executor := &executorStub{}

	client, err := New(
		Config{
			Client:   delegate,
			Executor: executor,
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrated weather client: %v",
			err,
		)
	}

	request := openmeteo.CurrentWeatherRequest{
		Latitude:  40.4093,
		Longitude: 49.8671,
	}

	snapshot, err := client.GetCurrentWeather(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf(
			"get orchestrated current weather: %v",
			err,
		)
	}

	if executor.callCount != 1 {
		t.Fatalf(
			"expected one orchestrator execution, got %d",
			executor.callCount,
		)
	}

	if delegate.callCount != 1 {
		t.Fatalf(
			"expected one provider execution, got %d",
			delegate.callCount,
		)
	}

	if executor.provider != providerpolicy.ProviderOpenMeteo {
		t.Fatalf(
			"unexpected provider: %s",
			executor.provider,
		)
	}

	expectedRequestKey := "current:40.4093:49.8671"

	if executor.requestKey != expectedRequestKey {
		t.Fatalf(
			"expected request key %q, got %q",
			expectedRequestKey,
			executor.requestKey,
		)
	}

	if snapshot.Provider != domainweather.ProviderOpenMeteo {
		t.Fatalf(
			"unexpected snapshot provider: %s",
			snapshot.Provider,
		)
	}

	if snapshot.Latitude != request.Latitude {
		t.Fatalf(
			"unexpected latitude: %f",
			snapshot.Latitude,
		)
	}

	if snapshot.Longitude != request.Longitude {
		t.Fatalf(
			"unexpected longitude: %f",
			snapshot.Longitude,
		)
	}
}
