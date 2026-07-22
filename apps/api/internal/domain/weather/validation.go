package weather

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrWeatherProviderRequired     = errors.New("weather provider is required")
	ErrWeatherCoordinatesInvalid   = errors.New("weather coordinates are invalid")
	ErrWeatherTimestampsInvalid    = errors.New("weather timestamps are invalid")
	ErrWeatherHumidityInvalid      = errors.New("weather relative humidity is invalid")
	ErrWeatherCloudCoverInvalid    = errors.New("weather cloud cover is invalid")
	ErrWeatherPrecipitationInvalid = errors.New("weather precipitation is invalid")
	ErrWeatherPressureInvalid      = errors.New("weather surface pressure is invalid")
	ErrWeatherWindInvalid          = errors.New("weather wind evidence is invalid")
	ErrWeatherTemperatureInvalid   = errors.New("weather temperature is invalid")
)

func (value CurrentSnapshot) Validate() error {
	if strings.TrimSpace(value.Provider) == "" {
		return ErrWeatherProviderRequired
	}
	if !finiteWeather(value.Latitude) || value.Latitude < -90 || value.Latitude > 90 ||
		!finiteWeather(value.Longitude) || value.Longitude < -180 || value.Longitude > 180 {
		return ErrWeatherCoordinatesInvalid
	}
	if value.ObservedAt.IsZero() || value.RetrievedAt.IsZero() || value.RetrievedAt.Before(value.ObservedAt) {
		return ErrWeatherTimestampsInvalid
	}
	if !finiteWeather(value.TemperatureCelsius) {
		return ErrWeatherTemperatureInvalid
	}
	if value.RelativeHumidityPercent < 0 || value.RelativeHumidityPercent > 100 {
		return ErrWeatherHumidityInvalid
	}
	if value.CloudCoverPercent < 0 || value.CloudCoverPercent > 100 {
		return ErrWeatherCloudCoverInvalid
	}
	if !finiteWeather(value.PrecipitationMillimeters) || value.PrecipitationMillimeters < 0 ||
		!finiteWeather(value.RainMillimeters) || value.RainMillimeters < 0 {
		return ErrWeatherPrecipitationInvalid
	}
	if !finiteWeather(value.SurfacePressureHPA) || value.SurfacePressureHPA <= 0 {
		return ErrWeatherPressureInvalid
	}
	if !finiteWeather(value.WindSpeedMetersPerSecond) || value.WindSpeedMetersPerSecond < 0 ||
		!finiteWeather(value.WindGustsMetersPerSecond) || value.WindGustsMetersPerSecond < 0 ||
		value.WindDirectionDegrees < 0 || value.WindDirectionDegrees >= 360 {
		return ErrWeatherWindInvalid
	}
	return nil
}

func finiteWeather(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
