package weathercontext

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPostgresPoolRequired = errors.New(
		"Weather Context PostgreSQL pool is required",
	)
	ErrPostgresSnapshotPolicyInvalid = errors.New(
		"Weather Context PostgreSQL snapshot policy is invalid",
	)
	ErrWeatherSnapshotInvalid = errors.New(
		"Weather Context PostgreSQL weather snapshot is invalid",
	)
)

type PostgresSnapshotPolicy struct {
	Provider                      string
	MaximumCoordinateDeltaDegrees float64
}

func DefaultPostgresSnapshotPolicy() PostgresSnapshotPolicy {
	return PostgresSnapshotPolicy{
		Provider:                      domainweather.ProviderOpenMeteo,
		MaximumCoordinateDeltaDegrees: 1,
	}
}

func (policy PostgresSnapshotPolicy) Validate() error {
	if strings.TrimSpace(policy.Provider) == "" ||
		math.IsNaN(policy.MaximumCoordinateDeltaDegrees) ||
		math.IsInf(policy.MaximumCoordinateDeltaDegrees, 0) ||
		policy.MaximumCoordinateDeltaDegrees <= 0 ||
		policy.MaximumCoordinateDeltaDegrees > 10 {
		return ErrPostgresSnapshotPolicyInvalid
	}
	return nil
}

type weatherSnapshotQueryer interface {
	QueryRow(
		context.Context,
		string,
		...any,
	) pgx.Row
}

type PostgresSnapshotReader struct {
	queryer weatherSnapshotQueryer
	policy  PostgresSnapshotPolicy
}

func NewPostgresSnapshotReader(
	pool *pgxpool.Pool,
	policy PostgresSnapshotPolicy,
) (*PostgresSnapshotReader, error) {
	if pool == nil {
		return nil, ErrPostgresPoolRequired
	}
	return newPostgresSnapshotReader(
		pool,
		policy,
	)
}

func newPostgresSnapshotReader(
	queryer weatherSnapshotQueryer,
	policy PostgresSnapshotPolicy,
) (*PostgresSnapshotReader, error) {
	if queryer == nil {
		return nil, ErrPostgresPoolRequired
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	policy.Provider = strings.TrimSpace(
		policy.Provider,
	)
	return &PostgresSnapshotReader{
		queryer: queryer,
		policy:  policy,
	}, nil
}

const latestWeatherSnapshotQuery = `
SELECT
	provider,
	latitude::double precision,
	longitude::double precision,
	observed_at,
	retrieved_at,
	temperature_celsius::double precision,
	relative_humidity_percent,
	precipitation_mm::double precision,
	rain_mm::double precision,
	weather_code,
	cloud_cover_percent,
	surface_pressure_hpa::double precision,
	wind_speed_mps::double precision,
	wind_direction_degrees,
	wind_gusts_mps::double precision
FROM weather_snapshots
WHERE provider = $1
  AND observed_at <= $2
  AND retrieved_at <= $2
  AND latitude::double precision BETWEEN
      ($3::double precision - $5::double precision)
      AND
      ($3::double precision + $5::double precision)
  AND longitude::double precision BETWEEN
      ($4::double precision - $5::double precision)
      AND
      ($4::double precision + $5::double precision)
  AND temperature_celsius IS NOT NULL
  AND relative_humidity_percent IS NOT NULL
  AND precipitation_mm IS NOT NULL
  AND rain_mm IS NOT NULL
  AND weather_code IS NOT NULL
  AND cloud_cover_percent IS NOT NULL
  AND surface_pressure_hpa IS NOT NULL
  AND wind_speed_mps IS NOT NULL
  AND wind_direction_degrees IS NOT NULL
  AND wind_gusts_mps IS NOT NULL
ORDER BY
	POWER(
		latitude::double precision - $3::double precision,
		2
	) +
	POWER(
		longitude::double precision - $4::double precision,
		2
	) ASC,
	observed_at DESC,
	retrieved_at DESC,
	id ASC
LIMIT 1;
`

func (
	reader *PostgresSnapshotReader,
) GetLatestSnapshot(
	ctx context.Context,
	request WeatherSnapshotRequest,
) (domainweather.CurrentSnapshot, error) {
	if reader == nil || reader.queryer == nil {
		return domainweather.CurrentSnapshot{},
			ErrServiceUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return domainweather.CurrentSnapshot{}, err
	}
	if !aviationconstraints.IsLatitude(request.Latitude) ||
		!aviationconstraints.IsLongitude(request.Longitude) ||
		request.AsOfTime.IsZero() {
		return domainweather.CurrentSnapshot{},
			ErrInvalidRequest
	}

	asOfTime := request.AsOfTime.UTC()
	var snapshot domainweather.CurrentSnapshot
	err := reader.queryer.QueryRow(
		ctx,
		latestWeatherSnapshotQuery,
		reader.policy.Provider,
		asOfTime,
		request.Latitude,
		request.Longitude,
		reader.policy.MaximumCoordinateDeltaDegrees,
	).Scan(
		&snapshot.Provider,
		&snapshot.Latitude,
		&snapshot.Longitude,
		&snapshot.ObservedAt,
		&snapshot.RetrievedAt,
		&snapshot.TemperatureCelsius,
		&snapshot.RelativeHumidityPercent,
		&snapshot.PrecipitationMillimeters,
		&snapshot.RainMillimeters,
		&snapshot.WeatherCode,
		&snapshot.CloudCoverPercent,
		&snapshot.SurfacePressureHPA,
		&snapshot.WindSpeedMetersPerSecond,
		&snapshot.WindDirectionDegrees,
		&snapshot.WindGustsMetersPerSecond,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainweather.CurrentSnapshot{},
			ErrWeatherNotFound
	}
	if err != nil {
		return domainweather.CurrentSnapshot{},
			fmt.Errorf(
				"query Weather Context weather snapshot: %w",
				err,
			)
	}

	snapshot.Provider = strings.TrimSpace(
		snapshot.Provider,
	)
	snapshot.ObservedAt = snapshot.ObservedAt.UTC()
	snapshot.RetrievedAt = snapshot.RetrievedAt.UTC()
	if snapshot.Provider != reader.policy.Provider ||
		!aviationconstraints.IsLatitude(snapshot.Latitude) ||
		!aviationconstraints.IsLongitude(snapshot.Longitude) ||
		snapshot.ObservedAt.IsZero() ||
		snapshot.RetrievedAt.IsZero() ||
		snapshot.ObservedAt.After(asOfTime) ||
		snapshot.RetrievedAt.After(asOfTime) ||
		snapshot.RetrievedAt.Before(snapshot.ObservedAt) {
		return domainweather.CurrentSnapshot{},
			ErrWeatherSnapshotInvalid
	}

	return snapshot, nil
}
