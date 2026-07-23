package weather

import "time"

const ProviderOpenMeteo = "open_meteo"

type CurrentMetricAvailability struct {
	TemperatureCelsius       bool
	RelativeHumidityPercent  bool
	PrecipitationMillimeters bool
	RainMillimeters          bool
	WeatherCode              bool
	CloudCoverPercent        bool
	SurfacePressureHPA       bool
	WindSpeedMetersPerSecond bool
	WindDirectionDegrees     bool
	WindGustsMetersPerSecond bool
}

func AllCurrentMetricsAvailable() CurrentMetricAvailability {
	return CurrentMetricAvailability{
		TemperatureCelsius:       true,
		RelativeHumidityPercent:  true,
		PrecipitationMillimeters: true,
		RainMillimeters:          true,
		WeatherCode:              true,
		CloudCoverPercent:        true,
		SurfacePressureHPA:       true,
		WindSpeedMetersPerSecond: true,
		WindDirectionDegrees:     true,
		WindGustsMetersPerSecond: true,
	}
}

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
	MetricAvailabilityKnown  bool
	MetricAvailability       CurrentMetricAvailability
	RetrievedAt              time.Time
}

func (snapshot CurrentSnapshot) ResolvedMetricAvailability() CurrentMetricAvailability {
	if !snapshot.MetricAvailabilityKnown {
		return AllCurrentMetricsAvailable()
	}

	return snapshot.MetricAvailability
}
