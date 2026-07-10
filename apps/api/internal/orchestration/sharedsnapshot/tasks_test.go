package sharedsnapshot

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/weatherprovider"
)

type recordingTrafficSource struct {
	called    bool
	latitude  float64
	longitude float64
	radius    int
	states    []flightstate.FlightState
	err       error
}

func (
	source *recordingTrafficSource,
) LoadByPoint(
	_ context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	source.called = true
	source.latitude = latitude
	source.longitude = longitude
	source.radius = radius

	return source.states,
		source.err
}

type recordingWeatherSource struct {
	called   bool
	request  openmeteo.CurrentWeatherRequest
	snapshot domainweather.CurrentSnapshot
	err      error
}

func (
	source *recordingWeatherSource,
) GetCurrentWeather(
	_ context.Context,
	request openmeteo.CurrentWeatherRequest,
) (domainweather.CurrentSnapshot, error) {
	source.called = true
	source.request = request

	return source.snapshot,
		source.err
}

func TestBuildRegionalTrafficTaskRequiresTrafficSource(
	t *testing.T,
) {
	_, err := BuildRegionalTrafficTask(
		RegionalTrafficTaskConfig{
			Provider: providerpolicy.ProviderAirplanesLive,
		},
	)

	if !errors.Is(
		err,
		ErrRegionalTrafficSourceRequired,
	) {
		t.Fatalf(
			"expected ErrRegionalTrafficSourceRequired, got %v",
			err,
		)
	}
}

func TestBuildRegionalTrafficTaskRequiresProviderIdentity(
	t *testing.T,
) {
	_, err := BuildRegionalTrafficTask(
		RegionalTrafficTaskConfig{
			TrafficSource: &recordingTrafficSource{},
		},
	)

	if !errors.Is(
		err,
		ErrRegionalTrafficProviderRequired,
	) {
		t.Fatalf(
			"expected ErrRegionalTrafficProviderRequired, got %v",
			err,
		)
	}
}

func TestBuildRegionalTrafficTaskPreservesExplicitIdentityAndRequestKey(
	t *testing.T,
) {
	const (
		latitude  = 40.4093
		longitude = 49.8671
		radius    = 250
	)

	task, err := BuildRegionalTrafficTask(
		RegionalTrafficTaskConfig{
			TrafficSource: &recordingTrafficSource{},
			Provider:      providerpolicy.ProviderOpenSky,
			Latitude:      latitude,
			Longitude:     longitude,
			Radius:        radius,
		},
	)
	if err != nil {
		t.Fatalf(
			"build regional traffic task: %v",
			err,
		)
	}

	if task.ID != TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected traffic task identifier: got %q, want %q",
			task.ID,
			TaskIDRegionalTraffic,
		)
	}

	if task.Provider != providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"unexpected traffic provider: got %q, want %q",
			task.Provider,
			providerpolicy.ProviderOpenSky,
		)
	}

	expectedRequestKey := regionalprovider.PointRequestKey(
		latitude,
		longitude,
		radius,
	)

	if task.RequestKey != expectedRequestKey {
		t.Fatalf(
			"unexpected traffic request key: got %q, want %q",
			task.RequestKey,
			expectedRequestKey,
		)
	}
}

func TestBuildRegionalTrafficTaskExecutesExactPointRequest(
	t *testing.T,
) {
	const (
		latitude  = 40.4093
		longitude = 49.8671
		radius    = 250
	)

	source := &recordingTrafficSource{
		states: []flightstate.FlightState{
			{
				ID:     "state-1",
				ICAO24: "abc123",
			},
		},
	}

	task, err := BuildRegionalTrafficTask(
		RegionalTrafficTaskConfig{
			TrafficSource: source,
			Provider:      providerpolicy.ProviderAirplanesLive,
			Latitude:      latitude,
			Longitude:     longitude,
			Radius:        radius,
		},
	)
	if err != nil {
		t.Fatalf(
			"build regional traffic task: %v",
			err,
		)
	}

	value, err := task.Function(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"execute regional traffic task: %v",
			err,
		)
	}

	if !source.called {
		t.Fatal(
			"expected traffic source to be called",
		)
	}

	if source.latitude != latitude ||
		source.longitude != longitude ||
		source.radius != radius {
		t.Fatalf(
			"unexpected point request: latitude=%f longitude=%f radius=%d",
			source.latitude,
			source.longitude,
			source.radius,
		)
	}

	trafficPayload, ok := value.RegionalTraffic()
	if !ok {
		t.Fatalf(
			"expected regional traffic payload, got kind %q",
			value.Kind(),
		)
	}

	if len(trafficPayload.States) != 1 {
		t.Fatalf(
			"unexpected traffic task state count: got %d, want 1",
			len(trafficPayload.States),
		)
	}

	if trafficPayload.States[0].ID != "state-1" {
		t.Fatalf(
			"unexpected traffic state identifier: %q",
			trafficPayload.States[0].ID,
		)
	}
}

func TestBuildRegionalTrafficTaskPropagatesSourceError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"traffic source failed",
	)

	task, err := BuildRegionalTrafficTask(
		RegionalTrafficTaskConfig{
			TrafficSource: &recordingTrafficSource{
				err: expectedErr,
			},
			Provider: providerpolicy.ProviderAirplanesLive,
		},
	)
	if err != nil {
		t.Fatalf(
			"build regional traffic task: %v",
			err,
		)
	}

	_, err = task.Function(
		context.Background(),
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected source error propagation, got %v",
			err,
		)
	}
}

func TestBuildCurrentWeatherTaskRequiresWeatherSource(
	t *testing.T,
) {
	_, err := BuildCurrentWeatherTask(
		CurrentWeatherTaskConfig{
			Provider: providerpolicy.ProviderOpenMeteo,
		},
	)

	if !errors.Is(
		err,
		ErrCurrentWeatherSourceRequired,
	) {
		t.Fatalf(
			"expected ErrCurrentWeatherSourceRequired, got %v",
			err,
		)
	}
}

func TestBuildCurrentWeatherTaskRequiresProviderIdentity(
	t *testing.T,
) {
	_, err := BuildCurrentWeatherTask(
		CurrentWeatherTaskConfig{
			WeatherSource: &recordingWeatherSource{},
		},
	)

	if !errors.Is(
		err,
		ErrCurrentWeatherProviderRequired,
	) {
		t.Fatalf(
			"expected ErrCurrentWeatherProviderRequired, got %v",
			err,
		)
	}
}

func TestBuildCurrentWeatherTaskPreservesExplicitIdentityAndRequestKey(
	t *testing.T,
) {
	const (
		latitude  = 40.4093
		longitude = 49.8671
	)

	task, err := BuildCurrentWeatherTask(
		CurrentWeatherTaskConfig{
			WeatherSource: &recordingWeatherSource{},
			Provider:      providerpolicy.ProviderOurAirports,
			Latitude:      latitude,
			Longitude:     longitude,
		},
	)
	if err != nil {
		t.Fatalf(
			"build current weather task: %v",
			err,
		)
	}

	if task.ID != TaskIDCurrentWeather {
		t.Fatalf(
			"unexpected weather task identifier: got %q, want %q",
			task.ID,
			TaskIDCurrentWeather,
		)
	}

	if task.Provider != providerpolicy.ProviderOurAirports {
		t.Fatalf(
			"unexpected weather provider: got %q, want %q",
			task.Provider,
			providerpolicy.ProviderOurAirports,
		)
	}

	expectedRequestKey := weatherprovider.CurrentWeatherRequestKey(
		openmeteo.CurrentWeatherRequest{
			Latitude:  latitude,
			Longitude: longitude,
		},
	)

	if task.RequestKey != expectedRequestKey {
		t.Fatalf(
			"unexpected weather request key: got %q, want %q",
			task.RequestKey,
			expectedRequestKey,
		)
	}
}

func TestBuildCurrentWeatherTaskExecutesExactRequest(
	t *testing.T,
) {
	const (
		latitude  = 40.4093
		longitude = 49.8671
	)

	expectedSnapshot := domainweather.CurrentSnapshot{
		Provider:  domainweather.ProviderOpenMeteo,
		Latitude:  latitude,
		Longitude: longitude,
	}

	source := &recordingWeatherSource{
		snapshot: expectedSnapshot,
	}

	task, err := BuildCurrentWeatherTask(
		CurrentWeatherTaskConfig{
			WeatherSource: source,
			Provider:      providerpolicy.ProviderOpenMeteo,
			Latitude:      latitude,
			Longitude:     longitude,
		},
	)
	if err != nil {
		t.Fatalf(
			"build current weather task: %v",
			err,
		)
	}

	value, err := task.Function(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"execute current weather task: %v",
			err,
		)
	}

	if !source.called {
		t.Fatal(
			"expected weather source to be called",
		)
	}

	if source.request.Latitude != latitude ||
		source.request.Longitude != longitude {
		t.Fatalf(
			"unexpected weather request: latitude=%f longitude=%f",
			source.request.Latitude,
			source.request.Longitude,
		)
	}

	weatherPayload, ok := value.CurrentWeather()
	if !ok {
		t.Fatalf(
			"expected current weather payload, got kind %q",
			value.Kind(),
		)
	}

	if weatherPayload.Snapshot.Provider !=
		domainweather.ProviderOpenMeteo {
		t.Fatalf(
			"unexpected weather snapshot provider: %q",
			weatherPayload.Snapshot.Provider,
		)
	}
}

func TestBuildCurrentWeatherTaskPropagatesSourceError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"weather source failed",
	)

	task, err := BuildCurrentWeatherTask(
		CurrentWeatherTaskConfig{
			WeatherSource: &recordingWeatherSource{
				err: expectedErr,
			},
			Provider: providerpolicy.ProviderOpenMeteo,
		},
	)
	if err != nil {
		t.Fatalf(
			"build current weather task: %v",
			err,
		)
	}

	_, err = task.Function(
		context.Background(),
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected source error propagation, got %v",
			err,
		)
	}
}

func TestBuildTasksRequiresEachDependency(
	t *testing.T,
) {
	_, err := BuildTasks(
		TaskConfig{
			TrafficProvider: providerpolicy.ProviderAirplanesLive,
			WeatherSource:   &recordingWeatherSource{},
			WeatherProvider: providerpolicy.ProviderOpenMeteo,
		},
	)
	if !errors.Is(
		err,
		ErrRegionalTrafficSourceRequired,
	) {
		t.Fatalf(
			"expected missing traffic source error, got %v",
			err,
		)
	}

	_, err = BuildTasks(
		TaskConfig{
			TrafficSource:   &recordingTrafficSource{},
			WeatherSource:   &recordingWeatherSource{},
			WeatherProvider: providerpolicy.ProviderOpenMeteo,
		},
	)
	if !errors.Is(
		err,
		ErrRegionalTrafficProviderRequired,
	) {
		t.Fatalf(
			"expected missing traffic provider error, got %v",
			err,
		)
	}

	_, err = BuildTasks(
		TaskConfig{
			TrafficSource:   &recordingTrafficSource{},
			TrafficProvider: providerpolicy.ProviderAirplanesLive,
			WeatherProvider: providerpolicy.ProviderOpenMeteo,
		},
	)
	if !errors.Is(
		err,
		ErrCurrentWeatherSourceRequired,
	) {
		t.Fatalf(
			"expected missing weather source error, got %v",
			err,
		)
	}

	_, err = BuildTasks(
		TaskConfig{
			TrafficSource:   &recordingTrafficSource{},
			TrafficProvider: providerpolicy.ProviderAirplanesLive,
			WeatherSource:   &recordingWeatherSource{},
		},
	)
	if !errors.Is(
		err,
		ErrCurrentWeatherProviderRequired,
	) {
		t.Fatalf(
			"expected missing weather provider error, got %v",
			err,
		)
	}
}

func TestBuildTasksPreservesCompositeTaskOrderAndExplicitIdentity(
	t *testing.T,
) {
	const (
		latitude  = 40.4093
		longitude = 49.8671
		radius    = 250
	)

	tasks, err := BuildTasks(
		TaskConfig{
			TrafficSource:   &recordingTrafficSource{},
			WeatherSource:   &recordingWeatherSource{},
			TrafficProvider: providerpolicy.ProviderOpenSky,
			WeatherProvider: providerpolicy.ProviderOpenMeteo,
			Latitude:        latitude,
			Longitude:       longitude,
			Radius:          radius,
		},
	)
	if err != nil {
		t.Fatalf(
			"build shared snapshot tasks: %v",
			err,
		)
	}

	if len(tasks) != 2 {
		t.Fatalf(
			"unexpected task count: got %d, want 2",
			len(tasks),
		)
	}

	if tasks[0].ID != TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected first task identifier: %q",
			tasks[0].ID,
		)
	}

	if tasks[1].ID != TaskIDCurrentWeather {
		t.Fatalf(
			"unexpected second task identifier: %q",
			tasks[1].ID,
		)
	}

	if tasks[0].Provider != providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"unexpected first task provider: %q",
			tasks[0].Provider,
		)
	}

	if tasks[1].Provider != providerpolicy.ProviderOpenMeteo {
		t.Fatalf(
			"unexpected second task provider: %q",
			tasks[1].Provider,
		)
	}
}
