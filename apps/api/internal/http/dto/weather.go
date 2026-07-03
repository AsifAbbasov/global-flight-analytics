package dto

import (
	"time"

	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
)

type CurrentWeatherResponse struct {
	SnapshotID               string    `json:"snapshot_id"`
	Provider                 string    `json:"provider"`
	Latitude                 float64   `json:"latitude"`
	Longitude                float64   `json:"longitude"`
	ObservedAt               time.Time `json:"observed_at"`
	RetrievedAt              time.Time `json:"retrieved_at"`
	StoredAt                 time.Time `json:"stored_at"`
	TemperatureCelsius       float64   `json:"temperature_celsius"`
	RelativeHumidityPercent  int       `json:"relative_humidity_percent"`
	PrecipitationMillimeters float64   `json:"precipitation_mm"`
	RainMillimeters          float64   `json:"rain_mm"`
	WeatherCode              int       `json:"weather_code"`
	CloudCoverPercent        int       `json:"cloud_cover_percent"`
	SurfacePressureHPA       float64   `json:"surface_pressure_hpa"`
	WindSpeedMetersPerSecond float64   `json:"wind_speed_mps"`
	WindDirectionDegrees     int       `json:"wind_direction_degrees"`
	WindGustsMetersPerSecond float64   `json:"wind_gusts_mps"`
}

func ToCurrentWeather(result weatherservice.CurrentWeatherResult) CurrentWeatherResponse {
	snapshot := result.Snapshot

	return CurrentWeatherResponse{
		SnapshotID:               result.SnapshotID,
		Provider:                 snapshot.Provider,
		Latitude:                 snapshot.Latitude,
		Longitude:                snapshot.Longitude,
		ObservedAt:               snapshot.ObservedAt,
		RetrievedAt:              snapshot.RetrievedAt,
		StoredAt:                 result.StoredAt,
		TemperatureCelsius:       snapshot.TemperatureCelsius,
		RelativeHumidityPercent:  snapshot.RelativeHumidityPercent,
		PrecipitationMillimeters: snapshot.PrecipitationMillimeters,
		RainMillimeters:          snapshot.RainMillimeters,
		WeatherCode:              snapshot.WeatherCode,
		CloudCoverPercent:        snapshot.CloudCoverPercent,
		SurfacePressureHPA:       snapshot.SurfacePressureHPA,
		WindSpeedMetersPerSecond: snapshot.WindSpeedMetersPerSecond,
		WindDirectionDegrees:     snapshot.WindDirectionDegrees,
		WindGustsMetersPerSecond: snapshot.WindGustsMetersPerSecond,
	}
}
