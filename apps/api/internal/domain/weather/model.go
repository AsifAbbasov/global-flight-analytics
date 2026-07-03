package weather

import "time"

const ProviderOpenMeteo = "open_meteo"

type CurrentSnapshot struct {
	Provider                 string
	Latitude                 float64
	Longitude                float64
	ObservedAt               time.Time
	TemperatureCelsius       float64
	RelativeHumidityPercent  int
	PrecipitationMillimeters float64
	RainMillimeters          float64
	WeatherCode              int
	CloudCoverPercent        int
	SurfacePressureHPA       float64
	WindSpeedMetersPerSecond float64
	WindDirectionDegrees     int
	WindGustsMetersPerSecond float64
	RetrievedAt              time.Time
}
