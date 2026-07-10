package sharedsnapshot

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/weatherprovider"
)

const (
	TaskIDRegionalTraffic = "regional-traffic"
	TaskIDCurrentWeather  = "current-weather"
)

var (
	ErrRegionalTrafficSourceRequired = errors.New(
		"shared snapshot regional traffic source is required",
	)

	ErrRegionalTrafficProviderRequired = errors.New(
		"shared snapshot regional traffic provider identity is required",
	)

	ErrCurrentWeatherSourceRequired = errors.New(
		"shared snapshot current weather source is required",
	)

	ErrCurrentWeatherProviderRequired = errors.New(
		"shared snapshot current weather provider identity is required",
	)
)

type RegionalTrafficSource interface {
	LoadByPoint(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) ([]flightstate.FlightState, error)
}

type CurrentWeatherSource interface {
	GetCurrentWeather(
		ctx context.Context,
		request openmeteo.CurrentWeatherRequest,
	) (domainweather.CurrentSnapshot, error)
}

type RegionalTrafficTaskConfig struct {
	TrafficSource RegionalTrafficSource
	Provider      providerpolicy.Provider

	Latitude  float64
	Longitude float64
	Radius    int
}

type CurrentWeatherTaskConfig struct {
	WeatherSource CurrentWeatherSource
	Provider      providerpolicy.Provider

	Latitude  float64
	Longitude float64
}

type TaskConfig struct {
	TrafficSource RegionalTrafficSource
	WeatherSource CurrentWeatherSource

	TrafficProvider providerpolicy.Provider
	WeatherProvider providerpolicy.Provider

	Latitude  float64
	Longitude float64
	Radius    int
}

func BuildRegionalTrafficTask(
	config RegionalTrafficTaskConfig,
) (providerfanout.Task[Payload], error) {
	if config.TrafficSource == nil {
		return providerfanout.Task[Payload]{},
			ErrRegionalTrafficSourceRequired
	}

	if config.Provider == "" {
		return providerfanout.Task[Payload]{},
			ErrRegionalTrafficProviderRequired
	}

	return providerfanout.Task[Payload]{
		ID:       TaskIDRegionalTraffic,
		Provider: config.Provider,
		RequestKey: regionalprovider.PointRequestKey(
			config.Latitude,
			config.Longitude,
			config.Radius,
		),
		Function: func(
			ctx context.Context,
		) (Payload, error) {
			states, err := config.TrafficSource.LoadByPoint(
				ctx,
				config.Latitude,
				config.Longitude,
				config.Radius,
			)
			if err != nil {
				return Payload{},
					err
			}

			return NewRegionalTrafficPayload(
				states,
			), nil
		},
	}, nil
}

func BuildCurrentWeatherTask(
	config CurrentWeatherTaskConfig,
) (providerfanout.Task[Payload], error) {
	if config.WeatherSource == nil {
		return providerfanout.Task[Payload]{},
			ErrCurrentWeatherSourceRequired
	}

	if config.Provider == "" {
		return providerfanout.Task[Payload]{},
			ErrCurrentWeatherProviderRequired
	}

	weatherRequest := openmeteo.CurrentWeatherRequest{
		Latitude:  config.Latitude,
		Longitude: config.Longitude,
	}

	return providerfanout.Task[Payload]{
		ID:       TaskIDCurrentWeather,
		Provider: config.Provider,
		RequestKey: weatherprovider.CurrentWeatherRequestKey(
			weatherRequest,
		),
		Function: func(
			ctx context.Context,
		) (Payload, error) {
			snapshot, err := config.WeatherSource.GetCurrentWeather(
				ctx,
				weatherRequest,
			)
			if err != nil {
				return Payload{},
					err
			}

			return NewCurrentWeatherPayload(
				snapshot,
			), nil
		},
	}, nil
}

func BuildTasks(
	config TaskConfig,
) ([]providerfanout.Task[Payload], error) {
	trafficTask, err := BuildRegionalTrafficTask(
		RegionalTrafficTaskConfig{
			TrafficSource: config.TrafficSource,
			Provider:      config.TrafficProvider,
			Latitude:      config.Latitude,
			Longitude:     config.Longitude,
			Radius:        config.Radius,
		},
	)
	if err != nil {
		return nil, err
	}

	weatherTask, err := BuildCurrentWeatherTask(
		CurrentWeatherTaskConfig{
			WeatherSource: config.WeatherSource,
			Provider:      config.WeatherProvider,
			Latitude:      config.Latitude,
			Longitude:     config.Longitude,
		},
	)
	if err != nil {
		return nil, err
	}

	return []providerfanout.Task[Payload]{
		trafficTask,
		weatherTask,
	}, nil
}
