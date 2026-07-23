package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWeatherRepositoryPoolRequired = errors.New("weather repository pool is required")
	ErrInvalidWeatherProvider        = errors.New("invalid weather provider")
	ErrInvalidWeatherCoordinates     = errors.New("invalid weather coordinates")
	ErrInvalidWeatherObservedAt      = errors.New("invalid weather observed time")
	ErrInvalidWeatherHumidity        = errors.New("invalid weather relative humidity")
	ErrInvalidWeatherCloudCover      = errors.New("invalid weather cloud cover")
	ErrInvalidWeatherPrecipitation   = errors.New("invalid weather precipitation")
	ErrInvalidWeatherPressure        = errors.New("invalid weather pressure")
	ErrInvalidWeatherWind            = errors.New("invalid weather wind")
)

type WeatherRepository struct {
	pool *pgxpool.Pool
}

func NewWeatherRepository(pool *pgxpool.Pool) *WeatherRepository {
	return &WeatherRepository{
		pool: pool,
	}
}

func (repository *WeatherRepository) SaveCurrentSnapshot(ctx context.Context, snapshot weather.CurrentSnapshot) (string, error) {
	if repository == nil || repository.pool == nil {
		return "", ErrWeatherRepositoryPoolRequired
	}

	normalizedSnapshot, err := normalizeCurrentWeatherSnapshot(snapshot)
	if err != nil {
		return "", err
	}
	metricAvailability := normalizedSnapshot.ResolvedMetricAvailability()

	var snapshotID string

	err = repository.pool.QueryRow(ctx, `
		INSERT INTO weather_snapshots (
			provider,
			latitude,
			longitude,
			observed_at,
			retrieved_at,
			temperature_celsius,
			relative_humidity_percent,
			precipitation_mm,
			rain_mm,
			weather_code,
			cloud_cover_percent,
			surface_pressure_hpa,
			wind_speed_mps,
			wind_direction_degrees,
			wind_gusts_mps,
			metadata_json
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			NULL
		)
		ON CONFLICT (provider, latitude, longitude, observed_at)
		DO UPDATE SET
			retrieved_at = EXCLUDED.retrieved_at,
			temperature_celsius = EXCLUDED.temperature_celsius,
			relative_humidity_percent = EXCLUDED.relative_humidity_percent,
			precipitation_mm = EXCLUDED.precipitation_mm,
			rain_mm = EXCLUDED.rain_mm,
			weather_code = EXCLUDED.weather_code,
			cloud_cover_percent = EXCLUDED.cloud_cover_percent,
			surface_pressure_hpa = EXCLUDED.surface_pressure_hpa,
			wind_speed_mps = EXCLUDED.wind_speed_mps,
			wind_direction_degrees = EXCLUDED.wind_direction_degrees,
			wind_gusts_mps = EXCLUDED.wind_gusts_mps
		RETURNING id::text;
	`,
		normalizedSnapshot.Provider,
		normalizedSnapshot.Latitude,
		normalizedSnapshot.Longitude,
		normalizedSnapshot.ObservedAt,
		normalizedSnapshot.RetrievedAt,
		nullableWeatherFloat64(normalizedSnapshot.TemperatureCelsius, metricAvailability.TemperatureCelsius),
		nullableWeatherInt(normalizedSnapshot.RelativeHumidityPercent, metricAvailability.RelativeHumidityPercent),
		nullableWeatherFloat64(normalizedSnapshot.PrecipitationMillimeters, metricAvailability.PrecipitationMillimeters),
		nullableWeatherFloat64(normalizedSnapshot.RainMillimeters, metricAvailability.RainMillimeters),
		nullableWeatherInt(normalizedSnapshot.WeatherCode, metricAvailability.WeatherCode),
		nullableWeatherInt(normalizedSnapshot.CloudCoverPercent, metricAvailability.CloudCoverPercent),
		nullableWeatherFloat64(normalizedSnapshot.SurfacePressureHPA, metricAvailability.SurfacePressureHPA),
		nullableWeatherFloat64(normalizedSnapshot.WindSpeedMetersPerSecond, metricAvailability.WindSpeedMetersPerSecond),
		nullableWeatherInt(normalizedSnapshot.WindDirectionDegrees, metricAvailability.WindDirectionDegrees),
		nullableWeatherFloat64(normalizedSnapshot.WindGustsMetersPerSecond, metricAvailability.WindGustsMetersPerSecond),
	).Scan(&snapshotID)
	if err != nil {
		return "", fmt.Errorf("save weather snapshot: %w", err)
	}

	return snapshotID, nil
}

func normalizeCurrentWeatherSnapshot(snapshot weather.CurrentSnapshot) (weather.CurrentSnapshot, error) {
	normalizedSnapshot := snapshot

	normalizedSnapshot.Provider = strings.TrimSpace(normalizedSnapshot.Provider)
	if normalizedSnapshot.Provider == "" {
		normalizedSnapshot.Provider = weather.ProviderOpenMeteo
	}

	if normalizedSnapshot.Provider == "" {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherProvider
	}

	if !aviationconstraints.IsLatitude(normalizedSnapshot.Latitude) ||
		!aviationconstraints.IsLongitude(normalizedSnapshot.Longitude) {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherCoordinates
	}

	if normalizedSnapshot.ObservedAt.IsZero() {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherObservedAt
	}

	normalizedSnapshot.ObservedAt = normalizedSnapshot.ObservedAt.UTC()

	if normalizedSnapshot.RetrievedAt.IsZero() {
		normalizedSnapshot.RetrievedAt = time.Now().UTC()
	} else {
		normalizedSnapshot.RetrievedAt = normalizedSnapshot.RetrievedAt.UTC()
	}

	metricAvailability := normalizedSnapshot.ResolvedMetricAvailability()

	if metricAvailability.RelativeHumidityPercent &&
		!aviationconstraints.IsPercentInt(normalizedSnapshot.RelativeHumidityPercent) {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherHumidity
	}

	if metricAvailability.CloudCoverPercent &&
		!aviationconstraints.IsPercentInt(normalizedSnapshot.CloudCoverPercent) {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherCloudCover
	}

	if (metricAvailability.PrecipitationMillimeters &&
		!aviationconstraints.IsNonNegativeFloat64(normalizedSnapshot.PrecipitationMillimeters)) ||
		(metricAvailability.RainMillimeters &&
			!aviationconstraints.IsNonNegativeFloat64(normalizedSnapshot.RainMillimeters)) {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherPrecipitation
	}

	if metricAvailability.SurfacePressureHPA &&
		!aviationconstraints.IsPositiveFloat64(normalizedSnapshot.SurfacePressureHPA) {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherPressure
	}

	if (metricAvailability.WindSpeedMetersPerSecond &&
		!aviationconstraints.IsNonNegativeFloat64(normalizedSnapshot.WindSpeedMetersPerSecond)) ||
		(metricAvailability.WindGustsMetersPerSecond &&
			!aviationconstraints.IsNonNegativeFloat64(normalizedSnapshot.WindGustsMetersPerSecond)) ||
		(metricAvailability.WindDirectionDegrees &&
			!aviationconstraints.IsHeadingDegreesInclusive(normalizedSnapshot.WindDirectionDegrees)) {
		return weather.CurrentSnapshot{}, ErrInvalidWeatherWind
	}

	return normalizedSnapshot, nil
}

func nullableWeatherFloat64(value float64, available bool) any {
	if !available {
		return nil
	}
	return value
}

func nullableWeatherInt(value int, available bool) any {
	if !available {
		return nil
	}
	return value
}
