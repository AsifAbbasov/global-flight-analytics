package openmeteo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

const (
	defaultBaseURL          = "https://api.open-meteo.com"
	maxWeatherResponseBytes = 1 << 20
)

var (
	ErrInvalidCoordinates = errors.New(
		"invalid coordinates",
	)

	ErrInvalidBaseURL = errors.New(
		"invalid open-meteo base url",
	)

	ErrInvalidTimeout = errors.New(
		"open-meteo timeout must be greater than zero",
	)
)

type Config struct {
	BaseURL          string
	HTTPClient       *http.Client
	Timeout          time.Duration
	ResponseObserver integrationcommon.ProviderResponseObserver
}

type Client struct {
	baseURL          string
	httpClient       *http.Client
	responseObserver integrationcommon.ProviderResponseObserver
}

type CurrentWeatherRequest struct {
	Latitude  float64
	Longitude float64
}

func New(
	config Config,
) (*Client, error) {
	baseURL := strings.TrimSpace(
		config.BaseURL,
	)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	parsedURL, err := url.Parse(
		baseURL,
	)
	if err != nil ||
		parsedURL.Scheme == "" ||
		parsedURL.Host == "" {
		return nil, ErrInvalidBaseURL
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		if config.Timeout <= 0 {
			return nil, ErrInvalidTimeout
		}

		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &Client{
		baseURL: strings.TrimRight(
			baseURL,
			"/",
		),
		httpClient:       httpClient,
		responseObserver: config.ResponseObserver,
	}, nil
}

func (client *Client) GetCurrentWeather(
	ctx context.Context,
	request CurrentWeatherRequest,
) (weather.CurrentSnapshot, error) {
	if !aviationconstraints.IsLatitude(
		request.Latitude,
	) ||
		!aviationconstraints.IsLongitude(
			request.Longitude,
		) {
		return weather.CurrentSnapshot{}, ErrInvalidCoordinates
	}

	endpoint, err := url.Parse(
		client.baseURL + "/v1/forecast",
	)
	if err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf(
			"build open-meteo endpoint: %w",
			err,
		)
	}

	query := endpoint.Query()

	query.Set(
		"latitude",
		formatCoordinate(
			request.Latitude,
		),
	)

	query.Set(
		"longitude",
		formatCoordinate(
			request.Longitude,
		),
	)

	query.Set(
		"current",
		strings.Join(
			currentWeatherVariables(),
			",",
		),
	)

	query.Set(
		"wind_speed_unit",
		"ms",
	)

	query.Set(
		"timezone",
		"UTC",
	)

	query.Set(
		"forecast_days",
		"1",
	)

	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf(
			"create open-meteo request: %w",
			err,
		)
	}

	requestStartedAt := time.Now()

	response, err := client.httpClient.Do(
		httpRequest,
	)
	latency := time.Since(requestStartedAt)

	if err != nil {
		requestErr := fmt.Errorf(
			"open-meteo request failed: %w",
			err,
		)

		if observeErr := client.observeProviderTransportFailure(
			err,
			latency,
		); observeErr != nil {
			return weather.CurrentSnapshot{}, errors.Join(
				requestErr,
				fmt.Errorf(
					"observe open-meteo transport failure: %w",
					observeErr,
				),
			)
		}

		return weather.CurrentSnapshot{}, requestErr
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		statusErr := integrationcommon.ProviderStatusError(
			response.StatusCode,
		)
		if statusErr == nil {
			statusErr = fmt.Errorf(
				"unexpected provider status %d",
				response.StatusCode,
			)
		}
		requestErr := fmt.Errorf(
			"open-meteo request failed: %w",
			statusErr,
		)

		if observeErr := client.observeProviderResponse(
			response,
			latency,
		); observeErr != nil {
			return weather.CurrentSnapshot{}, errors.Join(
				requestErr,
				fmt.Errorf(
					"observe open-meteo response: %w",
					observeErr,
				),
			)
		}

		return weather.CurrentSnapshot{}, requestErr
	}

	var payload forecastResponse

	if err := integrationcommon.DecodeJSONHTTPResponse(
		response,
		weather.ProviderOpenMeteo,
		maxWeatherResponseBytes,
		&payload,
	); err != nil {
		decodeErr := fmt.Errorf(
			"decode open-meteo response: %w",
			err,
		)

		if observeErr := client.observeProviderResponseFailure(
			err,
			latency,
		); observeErr != nil {
			return weather.CurrentSnapshot{}, errors.Join(
				decodeErr,
				fmt.Errorf(
					"observe open-meteo response failure: %w",
					observeErr,
				),
			)
		}

		return weather.CurrentSnapshot{}, decodeErr
	}

	observedAt, err := parseOpenMeteoTime(
		payload.Current.Time,
	)
	if err != nil {
		parseErr := fmt.Errorf(
			"parse open-meteo current time: %w",
			err,
		)

		if observeErr := client.observeProviderResponseFailure(
			err,
			latency,
		); observeErr != nil {
			return weather.CurrentSnapshot{}, errors.Join(
				parseErr,
				fmt.Errorf(
					"observe open-meteo response failure: %w",
					observeErr,
				),
			)
		}

		return weather.CurrentSnapshot{}, parseErr
	}

	snapshot := weather.CurrentSnapshot{
		Provider:                 weather.ProviderOpenMeteo,
		Latitude:                 payload.Latitude,
		Longitude:                payload.Longitude,
		ObservedAt:               observedAt,
		TemperatureCelsius:       payload.Current.Temperature2M,
		RelativeHumidityPercent:  payload.Current.RelativeHumidity2M,
		PrecipitationMillimeters: payload.Current.Precipitation,
		RainMillimeters:          payload.Current.Rain,
		WeatherCode:              payload.Current.WeatherCode,
		CloudCoverPercent:        payload.Current.CloudCover,
		SurfacePressureHPA:       payload.Current.SurfacePressure,
		WindSpeedMetersPerSecond: payload.Current.WindSpeed10M,
		WindDirectionDegrees:     payload.Current.WindDirection10M,
		WindGustsMetersPerSecond: payload.Current.WindGusts10M,
		RetrievedAt:              time.Now().UTC(),
	}

	_ = client.observeProviderResponse(
		response,
		latency,
	)

	return snapshot, nil
}

func (client *Client) observeProviderResponse(
	response *http.Response,
	latency time.Duration,
) error {
	if client.responseObserver == nil {
		return nil
	}

	return client.responseObserver.ObserveProviderResponse(
		weather.ProviderOpenMeteo,
		response.StatusCode,
		response.Header.Clone(),
		latency,
	)
}

func (client *Client) observeProviderTransportFailure(
	requestErr error,
	latency time.Duration,
) error {
	if client.responseObserver == nil {
		return nil
	}

	observer, supported :=
		client.responseObserver.(integrationcommon.ProviderTransportFailureObserver)
	if !supported {
		return nil
	}

	return observer.ObserveProviderTransportFailure(
		weather.ProviderOpenMeteo,
		requestErr,
		latency,
	)
}

func (client *Client) observeProviderResponseFailure(
	responseErr error,
	latency time.Duration,
) error {
	if client.responseObserver == nil {
		return nil
	}

	observer, supported :=
		client.responseObserver.(integrationcommon.ProviderResponseFailureObserver)
	if !supported {
		return nil
	}

	return observer.ObserveProviderResponseFailure(
		weather.ProviderOpenMeteo,
		responseErr,
		latency,
	)
}

func currentWeatherVariables() []string {
	return []string{
		"temperature_2m",
		"relative_humidity_2m",
		"precipitation",
		"rain",
		"weather_code",
		"cloud_cover",
		"surface_pressure",
		"wind_speed_10m",
		"wind_direction_10m",
		"wind_gusts_10m",
	}
}

func formatCoordinate(
	value float64,
) string {
	return fmt.Sprintf(
		"%.6f",
		value,
	)
}

func parseOpenMeteoTime(
	value string,
) (time.Time, error) {
	trimmed := strings.TrimSpace(
		value,
	)

	if trimmed == "" {
		return time.Time{}, errors.New(
			"empty time",
		)
	}

	parsed, err := time.Parse(
		time.RFC3339,
		trimmed,
	)
	if err == nil {
		return parsed.UTC(), nil
	}

	parsed, err = time.ParseInLocation(
		"2006-01-02T15:04",
		trimmed,
		time.UTC,
	)
	if err == nil {
		return parsed.UTC(), nil
	}

	return time.Time{}, err
}

type forecastResponse struct {
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Current   currentWeather `json:"current"`
}

type currentWeather struct {
	Time               string  `json:"time"`
	Temperature2M      float64 `json:"temperature_2m"`
	RelativeHumidity2M int     `json:"relative_humidity_2m"`
	Precipitation      float64 `json:"precipitation"`
	Rain               float64 `json:"rain"`
	WeatherCode        int     `json:"weather_code"`
	CloudCover         int     `json:"cloud_cover"`
	SurfacePressure    float64 `json:"surface_pressure"`
	WindSpeed10M       float64 `json:"wind_speed_10m"`
	WindDirection10M   int     `json:"wind_direction_10m"`
	WindGusts10M       float64 `json:"wind_gusts_10m"`
}
