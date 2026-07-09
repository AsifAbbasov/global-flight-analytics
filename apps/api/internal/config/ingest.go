package config

import "fmt"

const (
	trafficIngestionLatitudeEnvironmentVariable = "TRAFFIC_INGESTION_LATITUDE"

	trafficIngestionLongitudeEnvironmentVariable = "TRAFFIC_INGESTION_LONGITUDE"

	trafficIngestionRadiusEnvironmentVariable = "TRAFFIC_INGESTION_RADIUS"

	airplanesLiveTimeoutEnvironmentVariable = "AIRPLANES_LIVE_TIMEOUT"

	trajectoryMaxTimeGapEnvironmentVariable = "TRAJECTORY_MAX_TIME_GAP"

	trajectoryMaxGroundSpeedMPSEnvironmentVariable = "TRAJECTORY_MAX_GROUND_SPEED_MPS"
)

func LoadIngestConfig() (
	IngestConfig,
	error,
) {
	databaseURL, err := requiredTrimmedStringEnvironmentVariable(
		databaseURLEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load database url: %w",
			err,
		)
	}

	databaseConnectTimeout, err := requiredPositiveDurationEnvironmentVariable(
		databaseConnectTimeoutEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load database connect timeout: %w",
			err,
		)
	}

	trafficIngestionLatitude, err := requiredFiniteFloat64EnvironmentVariable(
		trafficIngestionLatitudeEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load traffic ingestion latitude: %w",
			err,
		)
	}

	trafficIngestionLongitude, err := requiredFiniteFloat64EnvironmentVariable(
		trafficIngestionLongitudeEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load traffic ingestion longitude: %w",
			err,
		)
	}

	trafficIngestionRadius, err := requiredIntegerEnvironmentVariable(
		trafficIngestionRadiusEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load traffic ingestion radius: %w",
			err,
		)
	}

	airplanesLiveTimeout, err := requiredPositiveDurationEnvironmentVariable(
		airplanesLiveTimeoutEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load airplanes.live timeout: %w",
			err,
		)
	}

	trajectoryMaxTimeGap, err := requiredNonNegativeDurationEnvironmentVariable(
		trajectoryMaxTimeGapEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load trajectory maximum time gap: %w",
			err,
		)
	}

	trajectoryMaxGroundSpeedMPS, err := requiredNonNegativeFiniteFloat64EnvironmentVariable(
		trajectoryMaxGroundSpeedMPSEnvironmentVariable,
	)
	if err != nil {
		return IngestConfig{}, fmt.Errorf(
			"load trajectory maximum ground speed: %w",
			err,
		)
	}

	return IngestConfig{
		Database: PostgresConfig{
			URL:            databaseURL,
			ConnectTimeout: databaseConnectTimeout,
		},
		TrafficIngestionLatitude:    trafficIngestionLatitude,
		TrafficIngestionLongitude:   trafficIngestionLongitude,
		TrafficIngestionRadius:      trafficIngestionRadius,
		AirplanesLiveTimeout:        airplanesLiveTimeout,
		TrajectoryMaxTimeGap:        trajectoryMaxTimeGap,
		TrajectoryMaxGroundSpeedMPS: trajectoryMaxGroundSpeedMPS,
	}, nil
}
