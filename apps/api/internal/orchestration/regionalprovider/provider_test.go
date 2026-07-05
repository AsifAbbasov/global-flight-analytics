package regionalprovider

import (
	"context"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type delegateStub struct {
	loadCount int
}

func (stub *delegateStub) SourceName() string {
	return "airplanes.live"
}

func (stub *delegateStub) LoadByPoint(
	_ context.Context,
	_ float64,
	_ float64,
	_ int,
) ([]flightstate.FlightState, error) {
	stub.loadCount++

	return []flightstate.FlightState{
		{
			ICAO24: "4K1234",
		},
	}, nil
}

type executorStub struct {
	executeCount int
	provider     providerpolicy.Provider
	requestKey   string
}

func (stub *executorStub) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function ingestionorchestrator.Function,
) (ingestionorchestrator.ExecuteResult, error) {
	stub.executeCount++
	stub.provider = provider
	stub.requestKey = requestKey

	value, err := function(
		ctx,
	)
	if err != nil {
		return ingestionorchestrator.ExecuteResult{}, err
	}

	return ingestionorchestrator.ExecuteResult{
		Provider:   provider,
		RequestKey: requestKey,
		Value:      value,
	}, nil
}

func TestLoadByPointExecutesThroughOrchestrator(
	t *testing.T,
) {
	delegate := &delegateStub{}
	executor := &executorStub{}

	provider, err := New(
		Config{
			Provider:   delegate,
			ProviderID: providerpolicy.ProviderAirplanesLive,
			Executor:   executor,
		},
	)
	if err != nil {
		t.Fatalf(
			"create regional provider: %v",
			err,
		)
	}

	states, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err != nil {
		t.Fatalf(
			"load orchestrated regional states: %v",
			err,
		)
	}

	if executor.executeCount != 1 {
		t.Fatalf(
			"expected one orchestrator execution, got %d",
			executor.executeCount,
		)
	}

	if delegate.loadCount != 1 {
		t.Fatalf(
			"expected one provider execution, got %d",
			delegate.loadCount,
		)
	}

	if executor.provider != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"unexpected provider identifier: %s",
			executor.provider,
		)
	}

	expectedRequestKey := "point:40.4093:49.8671:250"

	if executor.requestKey != expectedRequestKey {
		t.Fatalf(
			"expected request key %q, got %q",
			expectedRequestKey,
			executor.requestKey,
		)
	}

	if len(states) != 1 {
		t.Fatalf(
			"expected one flight state, got %d",
			len(states),
		)
	}

	if states[0].ICAO24 != "4K1234" {
		t.Fatalf(
			"unexpected ICAO24: %s",
			states[0].ICAO24,
		)
	}
}
