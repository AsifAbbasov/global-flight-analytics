package config

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestLoadIngestConfig(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	loadedConfig, err := LoadIngestConfig()
	if err != nil {
		t.Fatalf(
			"expected valid ingest configuration, got error: %v",
			err,
		)
	}

	if loadedConfig.Database.URL != "postgresql://user:password@host/database" {
		t.Fatalf(
			"expected database url %q, got %q",
			"postgresql://user:password@host/database",
			loadedConfig.Database.URL,
		)
	}

	if loadedConfig.Database.ConnectTimeout != 3*time.Second {
		t.Fatalf(
			"expected database connect timeout %s, got %s",
			3*time.Second,
			loadedConfig.Database.ConnectTimeout,
		)
	}

	if math.Abs(
		loadedConfig.TrafficIngestionLatitude-40.4093,
	) > 1e-12 {
		t.Fatalf(
			"expected latitude %f, got %f",
			40.4093,
			loadedConfig.TrafficIngestionLatitude,
		)
	}

	if math.Abs(
		loadedConfig.TrafficIngestionLongitude-49.8671,
	) > 1e-12 {
		t.Fatalf(
			"expected longitude %f, got %f",
			49.8671,
			loadedConfig.TrafficIngestionLongitude,
		)
	}

	if loadedConfig.TrafficIngestionRadius != 250 {
		t.Fatalf(
			"expected radius %d, got %d",
			250,
			loadedConfig.TrafficIngestionRadius,
		)
	}

	if loadedConfig.AirplanesLiveTimeout != 10*time.Second {
		t.Fatalf(
			"expected airplanes.live timeout %s, got %s",
			10*time.Second,
			loadedConfig.AirplanesLiveTimeout,
		)
	}

	if loadedConfig.TrajectoryMaxTimeGap != 90*time.Second {
		t.Fatalf(
			"expected trajectory maximum time gap %s, got %s",
			90*time.Second,
			loadedConfig.TrajectoryMaxTimeGap,
		)
	}

	if math.Abs(
		loadedConfig.TrajectoryMaxGroundSpeedMPS-420.5,
	) > 1e-12 {
		t.Fatalf(
			"expected trajectory maximum ground speed %f, got %f",
			420.5,
			loadedConfig.TrajectoryMaxGroundSpeedMPS,
		)
	}
}

func TestLoadIngestConfigDoesNotRequireOpenMeteoTimeout(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		"",
	)

	_, err := LoadIngestConfig()
	if err != nil {
		t.Fatalf(
			"expected ingest configuration not to require open-meteo timeout, got %v",
			err,
		)
	}
}

func TestLoadIngestConfigRejectsMissingDatabaseURL(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	t.Setenv(
		databaseURLEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadIngestConfig()

	if err == nil {
		t.Fatal(
			"expected ingest configuration error, got nil",
		)
	}

	if loadedConfig != (IngestConfig{}) {
		t.Fatalf(
			"expected zero ingest configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load database url: DATABASE_URL is required",
	) {
		t.Fatalf(
			"expected contextual database url error, got %q",
			err.Error(),
		)
	}
}

func TestLoadIngestConfigRejectsInvalidDatabaseConnectTimeout(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"0s",
	)

	loadedConfig, err := LoadIngestConfig()

	if err == nil {
		t.Fatal(
			"expected ingest configuration error, got nil",
		)
	}

	if loadedConfig != (IngestConfig{}) {
		t.Fatalf(
			"expected zero ingest configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load database connect timeout: DATABASE_CONNECT_TIMEOUT must be greater than zero",
	) {
		t.Fatalf(
			"expected contextual database timeout error, got %q",
			err.Error(),
		)
	}
}

func TestLoadIngestConfigRejectsNonFiniteCoordinates(
	t *testing.T,
) {
	tests := []struct {
		name                    string
		environmentVariableName string
		value                   string
		expectedError           string
	}{
		{
			name:                    "not a number latitude",
			environmentVariableName: trafficIngestionLatitudeEnvironmentVariable,
			value:                   "NaN",
			expectedError: "load traffic ingestion latitude: " +
				"TRAFFIC_INGESTION_LATITUDE must be a finite value",
		},
		{
			name:                    "positive infinity latitude",
			environmentVariableName: trafficIngestionLatitudeEnvironmentVariable,
			value:                   "+Inf",
			expectedError: "load traffic ingestion latitude: " +
				"TRAFFIC_INGESTION_LATITUDE must be a finite value",
		},
		{
			name:                    "negative infinity longitude",
			environmentVariableName: trafficIngestionLongitudeEnvironmentVariable,
			value:                   "-Inf",
			expectedError: "load traffic ingestion longitude: " +
				"TRAFFIC_INGESTION_LONGITUDE must be a finite value",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				setValidIngestEnvironment(
					t,
				)

				t.Setenv(
					test.environmentVariableName,
					test.value,
				)

				loadedConfig, err := LoadIngestConfig()

				if err == nil {
					t.Fatal(
						"expected ingest configuration error, got nil",
					)
				}

				if loadedConfig != (IngestConfig{}) {
					t.Fatalf(
						"expected zero ingest configuration, got %+v",
						loadedConfig,
					)
				}

				if !strings.Contains(
					err.Error(),
					test.expectedError,
				) {
					t.Fatalf(
						"expected error containing %q, got %q",
						test.expectedError,
						err.Error(),
					)
				}
			},
		)
	}
}

func TestLoadIngestConfigRejectsInvalidRadius(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	t.Setenv(
		trafficIngestionRadiusEnvironmentVariable,
		"250.5",
	)

	loadedConfig, err := LoadIngestConfig()

	if err == nil {
		t.Fatal(
			"expected ingest configuration error, got nil",
		)
	}

	if loadedConfig != (IngestConfig{}) {
		t.Fatalf(
			"expected zero ingest configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load traffic ingestion radius: parse TRAFFIC_INGESTION_RADIUS as integer",
	) {
		t.Fatalf(
			"expected contextual radius error, got %q",
			err.Error(),
		)
	}
}

func TestLoadIngestConfigRejectsNonPositiveProviderTimeouts(
	t *testing.T,
) {
	tests := []struct {
		name                    string
		environmentVariableName string
		value                   string
		expectedError           string
	}{
		{
			name:                    "zero airplanes live timeout",
			environmentVariableName: airplanesLiveTimeoutEnvironmentVariable,
			value:                   "0s",
			expectedError: "load airplanes.live timeout: " +
				"AIRPLANES_LIVE_TIMEOUT must be greater than zero",
		},
		{
			name:                    "negative airplanes live timeout",
			environmentVariableName: airplanesLiveTimeoutEnvironmentVariable,
			value:                   "-1s",
			expectedError: "load airplanes.live timeout: " +
				"AIRPLANES_LIVE_TIMEOUT must be greater than zero",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				setValidIngestEnvironment(
					t,
				)

				t.Setenv(
					test.environmentVariableName,
					test.value,
				)

				loadedConfig, err := LoadIngestConfig()

				if err == nil {
					t.Fatal(
						"expected ingest configuration error, got nil",
					)
				}

				if loadedConfig != (IngestConfig{}) {
					t.Fatalf(
						"expected zero ingest configuration, got %+v",
						loadedConfig,
					)
				}

				if !strings.Contains(
					err.Error(),
					test.expectedError,
				) {
					t.Fatalf(
						"expected error containing %q, got %q",
						test.expectedError,
						err.Error(),
					)
				}
			},
		)
	}
}

func TestLoadIngestConfigAcceptsDisabledTrajectoryThresholds(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	t.Setenv(
		trajectoryMaxTimeGapEnvironmentVariable,
		"0s",
	)

	t.Setenv(
		trajectoryMaxGroundSpeedMPSEnvironmentVariable,
		"0",
	)

	loadedConfig, err := LoadIngestConfig()
	if err != nil {
		t.Fatalf(
			"expected disabled trajectory thresholds to be valid, got error: %v",
			err,
		)
	}

	if loadedConfig.TrajectoryMaxTimeGap != 0 {
		t.Fatalf(
			"expected disabled trajectory time gap, got %s",
			loadedConfig.TrajectoryMaxTimeGap,
		)
	}

	if loadedConfig.TrajectoryMaxGroundSpeedMPS != 0 {
		t.Fatalf(
			"expected disabled trajectory ground speed threshold, got %f",
			loadedConfig.TrajectoryMaxGroundSpeedMPS,
		)
	}
}

func TestLoadIngestConfigRejectsNegativeTrajectoryTimeGap(
	t *testing.T,
) {
	setValidIngestEnvironment(
		t,
	)

	t.Setenv(
		trajectoryMaxTimeGapEnvironmentVariable,
		"-1s",
	)

	loadedConfig, err := LoadIngestConfig()

	if err == nil {
		t.Fatal(
			"expected ingest configuration error, got nil",
		)
	}

	if loadedConfig != (IngestConfig{}) {
		t.Fatalf(
			"expected zero ingest configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load trajectory maximum time gap: TRAJECTORY_MAX_TIME_GAP must be non-negative",
	) {
		t.Fatalf(
			"expected trajectory time gap validation error, got %q",
			err.Error(),
		)
	}
}

func TestLoadIngestConfigRejectsInvalidTrajectoryGroundSpeed(
	t *testing.T,
) {
	tests := []struct {
		name          string
		value         string
		expectedError string
	}{
		{
			name:  "negative finite value",
			value: "-1",
			expectedError: "load trajectory maximum ground speed: " +
				"TRAJECTORY_MAX_GROUND_SPEED_MPS must be non-negative",
		},
		{
			name:  "not a number",
			value: "NaN",
			expectedError: "load trajectory maximum ground speed: " +
				"TRAJECTORY_MAX_GROUND_SPEED_MPS must be a finite value",
		},
		{
			name:  "positive infinity",
			value: "+Inf",
			expectedError: "load trajectory maximum ground speed: " +
				"TRAJECTORY_MAX_GROUND_SPEED_MPS must be a finite value",
		},
		{
			name:  "negative infinity",
			value: "-Inf",
			expectedError: "load trajectory maximum ground speed: " +
				"TRAJECTORY_MAX_GROUND_SPEED_MPS must be a finite value",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				setValidIngestEnvironment(
					t,
				)

				t.Setenv(
					trajectoryMaxGroundSpeedMPSEnvironmentVariable,
					test.value,
				)

				loadedConfig, err := LoadIngestConfig()

				if err == nil {
					t.Fatal(
						"expected ingest configuration error, got nil",
					)
				}

				if loadedConfig != (IngestConfig{}) {
					t.Fatalf(
						"expected zero ingest configuration, got %+v",
						loadedConfig,
					)
				}

				if !strings.Contains(
					err.Error(),
					test.expectedError,
				) {
					t.Fatalf(
						"expected error containing %q, got %q",
						test.expectedError,
						err.Error(),
					)
				}
			},
		)
	}
}

func setValidIngestEnvironment(
	t *testing.T,
) {
	t.Helper()

	t.Setenv(
		databaseURLEnvironmentVariable,
		"  postgresql://user:password@host/database  ",
	)

	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"  3s  ",
	)

	t.Setenv(
		trafficIngestionLatitudeEnvironmentVariable,
		"  40.4093  ",
	)

	t.Setenv(
		trafficIngestionLongitudeEnvironmentVariable,
		"  49.8671  ",
	)

	t.Setenv(
		trafficIngestionRadiusEnvironmentVariable,
		"  250  ",
	)

	t.Setenv(
		airplanesLiveTimeoutEnvironmentVariable,
		"  10s  ",
	)

	t.Setenv(
		trajectoryMaxTimeGapEnvironmentVariable,
		"  90s  ",
	)

	t.Setenv(
		trajectoryMaxGroundSpeedMPSEnvironmentVariable,
		"  420.5  ",
	)
}
