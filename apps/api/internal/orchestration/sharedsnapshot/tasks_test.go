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
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	source.called = true
	source.latitude = latitude
	source.longitude = longitude
	source.radius = radius

	return source.states, source.err
}

type recordingWeatherSource struct {
	called  bool
	request openmeteo.CurrentWeatherRequest
	err     error
}

func (
	source *recordingWeatherSource,
) GetCurrentWeather(
	ctx context.Context,
	request openmeteo.CurrentWeatherRequest,
) (domainweather.CurrentSnapshot, error) {
	source.called = true
	source.request = request

	return domainweather.CurrentSnapshot{}, source.err
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

	expectedStates := []flightstate.FlightState{
		{
			ID:     "state-1",
			ICAO24: "abc123",
		},
	}

	source := &recordingTrafficSource{
		states: expectedStates,
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

	if source.latitude != latitude {
		t.Fatalf(
			"unexpected traffic latitude: got %f, want %f",
			source.latitude,
			latitude,
		)
	}

	if source.longitude != longitude {
		t.Fatalf(
			"unexpected traffic longitude: got %f, want %f",
			source.longitude,
			longitude,
		)
	}

	if source.radius != radius {
		t.Fatalf(
			"unexpected traffic radius: got %d, want %d",
			source.radius,
			radius,
		)
	}

	states, ok := value.([]flightstate.FlightState)
	if !ok {
		t.Fatalf(
			"unexpected traffic task result type: %T",
			value,
		)
	}

	if len(states) != 1 {
		t.Fatalf(
			"unexpected traffic task state count: got %d, want 1",
			len(states),
		)
	}

	if states[0].ID != "state-1" {
		t.Fatalf(
			"unexpected traffic state identifier: got %q, want %q",
			states[0].ID,
			"state-1",
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

	request := openmeteo.CurrentWeatherRequest{
		Latitude:  latitude,
		Longitude: longitude,
	}

	expectedRequestKey := weatherprovider.CurrentWeatherRequestKey(
		request,
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

	source := &recordingWeatherSource{}

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

	if _, err := task.Function(
		context.Background(),
	); err != nil {
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

	if source.request.Latitude != latitude {
		t.Fatalf(
			"unexpected weather latitude: got %f, want %f",
			source.request.Latitude,
			latitude,
		)
	}

	if source.request.Longitude != longitude {
		t.Fatalf(
			"unexpected weather longitude: got %f, want %f",
			source.request.Longitude,
			longitude,
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

func TestBuildTasksRequiresTrafficSource(
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
			"expected ErrRegionalTrafficSourceRequired, got %v",
			err,
		)
	}
}

func TestBuildTasksRequiresTrafficProviderIdentity(
	t *testing.T,
) {
	_, err := BuildTasks(
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
			"expected ErrRegionalTrafficProviderRequired, got %v",
			err,
		)
	}
}

func TestBuildTasksRequiresWeatherSource(
	t *testing.T,
) {
	_, err := BuildTasks(
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
			"expected ErrCurrentWeatherSourceRequired, got %v",
			err,
		)
	}
}

func TestBuildTasksRequiresWeatherProviderIdentity(
	t *testing.T,
) {
	_, err := BuildTasks(
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
			"expected ErrCurrentWeatherProviderRequired, got %v",
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

	trafficTask := tasks[0]

	if trafficTask.ID != TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected first task identifier: got %q, want %q",
			trafficTask.ID,
			TaskIDRegionalTraffic,
		)
	}

	if trafficTask.Provider != providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"unexpected first task provider: got %q, want %q",
			trafficTask.Provider,
			providerpolicy.ProviderOpenSky,
		)
	}

	expectedTrafficRequestKey := regionalprovider.PointRequestKey(
		latitude,
		longitude,
		radius,
	)

	if trafficTask.RequestKey != expectedTrafficRequestKey {
		t.Fatalf(
			"unexpected first task request key: got %q, want %q",
			trafficTask.RequestKey,
			expectedTrafficRequestKey,
		)
	}

	weatherTask := tasks[1]

	if weatherTask.ID != TaskIDCurrentWeather {
		t.Fatalf(
			"unexpected second task identifier: got %q, want %q",
			weatherTask.ID,
			TaskIDCurrentWeather,
		)
	}

	if weatherTask.Provider != providerpolicy.ProviderOpenMeteo {
		t.Fatalf(
			"unexpected second task provider: got %q, want %q",
			weatherTask.Provider,
			providerpolicy.ProviderOpenMeteo,
		)
	}

	expectedWeatherRequestKey := weatherprovider.CurrentWeatherRequestKey(
		openmeteo.CurrentWeatherRequest{
			Latitude:  latitude,
			Longitude: longitude,
		},
	)

	if weatherTask.RequestKey != expectedWeatherRequestKey {
		t.Fatalf(
			"unexpected second task request key: got %q, want %q",
			weatherTask.RequestKey,
			expectedWeatherRequestKey,
		)
	}
}
