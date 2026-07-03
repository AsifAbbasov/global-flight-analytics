package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/gofiber/fiber/v2"
)

type CurrentWeatherService interface {
	GetAndStoreCurrentWeather(ctx context.Context, request weatherservice.CurrentWeatherRequest) (weatherservice.CurrentWeatherResult, error)
}

type WeatherHandler struct {
	service CurrentWeatherService
}

func NewWeatherHandler(service CurrentWeatherService) *WeatherHandler {
	return &WeatherHandler{
		service: service,
	}
}

func (handler *WeatherHandler) GetCurrent(c *fiber.Ctx) error {
	if handler == nil || handler.service == nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "WEATHER_SERVICE_UNAVAILABLE", "Weather service is unavailable")
	}

	latitude, longitude, ok := parseWeatherCoordinates(c)
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_WEATHER_COORDINATES", "Invalid weather coordinates")
	}

	result, err := handler.service.GetAndStoreCurrentWeather(c.Context(), weatherservice.CurrentWeatherRequest{
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		if errors.Is(err, weatherservice.ErrInvalidWeatherCoordinates) {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_WEATHER_COORDINATES", "Invalid weather coordinates")
		}

		if errors.Is(err, weatherservice.ErrWeatherClientRequired) ||
			errors.Is(err, weatherservice.ErrWeatherRepositoryRequired) {
			return response.Error(c, fiber.StatusServiceUnavailable, "WEATHER_SERVICE_UNAVAILABLE", "Weather service is unavailable")
		}

		return response.Error(c, fiber.StatusInternalServerError, "WEATHER_LOAD_FAILED", "Failed to load current weather")
	}

	return response.OK(c, dto.ToCurrentWeather(result))
}

func parseWeatherCoordinates(c *fiber.Ctx) (float64, float64, bool) {
	latitude, latitudeOK := parseWeatherCoordinate(c.Query("lat"))
	if !latitudeOK {
		return 0, 0, false
	}

	longitude, longitudeOK := parseWeatherCoordinate(c.Query("lon"))
	if !longitudeOK {
		return 0, 0, false
	}

	return latitude, longitude, true
}

func parseWeatherCoordinate(value string) (float64, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, false
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}
