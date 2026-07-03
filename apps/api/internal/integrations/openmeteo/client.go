package openmeteo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
)

const (
	defaultBaseURL = "https://api.open-meteo.com"
	defaultTimeout = 10 * time.Second
)

var (
	ErrInvalidCoordinates = errors.New("invalid coordinates")
	ErrInvalidBaseURL     = errors.New("invalid open-meteo base url")
)

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type CurrentWeatherRequest struct {
	Latitude  float64
	Longitude float64
}

func New(config Config) (*Client, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, ErrInvalidBaseURL
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		timeout := config.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}

		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}, nil
}

func (client *Client) GetCurrentWeather(ctx context.Context, request CurrentWeatherRequest) (weather.CurrentSnapshot, error) {
	if !isValidLatitude(request.Latitude) || !isValidLongitude(request.Longitude) {
		return weather.CurrentSnapshot{}, ErrInvalidCoordinates
	}

	endpoint, err := url.Parse(client.baseURL + "/v1/forecast")
	if err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf("build open-meteo endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("latitude", formatCoordinate(request.Latitude))
	query.Set("longitude", formatCoordinate(request.Longitude))
	query.Set("current", strings.Join(currentWeatherVariables(), ","))
	query.Set("wind_speed_unit", "ms")
	query.Set("timezone", "UTC")
	query.Set("forecast_days", "1")
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf("create open-meteo request: %w", err)
	}

	response, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf("open-meteo request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return weather.CurrentSnapshot{}, fmt.Errorf("open-meteo unexpected status: %d", response.StatusCode)
	}

	var payload forecastResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf("decode open-meteo response: %w", err)
	}

	observedAt, err := parseOpenMeteoTime(payload.Current.Time)
	if err != nil {
		return weather.CurrentSnapshot{}, fmt.Errorf("parse open-meteo current time: %w", err)
	}

	return weather.CurrentSnapshot{
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
	}, nil
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

func isValidLatitude(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= -90 && value <= 90
}

func isValidLongitude(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= -180 && value <= 180
}

func formatCoordinate(value float64) string {
	return fmt.Sprintf("%.6f", value)
}

func parseOpenMeteoTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("empty time")
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err == nil {
		return parsed.UTC(), nil
	}

	parsed, err = time.ParseInLocation("2006-01-02T15:04", trimmed, time.UTC)
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
