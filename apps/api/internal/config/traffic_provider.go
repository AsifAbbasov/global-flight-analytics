package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/opensky"
)

type TrafficProvider string

const (
	TrafficProviderAirplanesLive TrafficProvider = "airplanes.live"
	TrafficProviderOpenSky       TrafficProvider = "opensky"
	TrafficProviderAuto          TrafficProvider = "auto"
)

const (
	trafficProviderEnvironmentVariable        = "TRAFFIC_PROVIDER"
	openSkyBaseURLEnvironmentVariable         = "OPENSKY_BASE_URL"
	openSkyTokenURLEnvironmentVariable        = "OPENSKY_TOKEN_URL"
	openSkyClientIDEnvironmentVariable        = "OPENSKY_CLIENT_ID"
	openSkyClientSecretEnvironmentVariable    = "OPENSKY_CLIENT_SECRET"
	openSkyTimeoutEnvironmentVariable         = "OPENSKY_TIMEOUT"
	openSkyPollingIntervalEnvironmentVariable = "OPENSKY_POLLING_INTERVAL"
)

var (
	ErrTrafficProviderInvalid = errors.New(
		"traffic provider must be airplanes.live, opensky, or auto",
	)
	ErrOpenSkyCredentialPairRequired = errors.New(
		"OpenSky client id and client secret must be configured together",
	)
)

type TrafficProviderConfig struct {
	Provider TrafficProvider

	OpenSkyBaseURL         string
	OpenSkyTokenURL        string
	OpenSkyClientID        string
	OpenSkyClientSecret    string
	OpenSkyTimeout         time.Duration
	OpenSkyPollingInterval time.Duration
}

func LoadTrafficProviderConfig() (
	TrafficProviderConfig,
	error,
) {
	provider := TrafficProvider(
		strings.TrimSpace(
			os.Getenv(trafficProviderEnvironmentVariable),
		),
	)
	if provider == "" {
		provider = TrafficProviderAirplanesLive
	}
	if provider != TrafficProviderAirplanesLive &&
		provider != TrafficProviderOpenSky &&
		provider != TrafficProviderAuto {
		return TrafficProviderConfig{}, fmt.Errorf(
			"%w: %q",
			ErrTrafficProviderInvalid,
			provider,
		)
	}

	clientID := strings.TrimSpace(
		os.Getenv(openSkyClientIDEnvironmentVariable),
	)
	clientSecret := strings.TrimSpace(
		os.Getenv(openSkyClientSecretEnvironmentVariable),
	)
	if (clientID == "") != (clientSecret == "") {
		return TrafficProviderConfig{}, ErrOpenSkyCredentialPairRequired
	}

	timeout, err := trafficProviderOptionalPositiveDuration(
		openSkyTimeoutEnvironmentVariable,
		15*time.Second,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
	}

	defaultPollingInterval := 10 * time.Second
	if clientID != "" {
		defaultPollingInterval = 5 * time.Second
	}
	pollingInterval, err := trafficProviderOptionalPositiveDuration(
		openSkyPollingIntervalEnvironmentVariable,
		defaultPollingInterval,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
	}

	baseURL := trafficProviderOptionalTrimmedString(
		openSkyBaseURLEnvironmentVariable,
		opensky.DefaultBaseURL,
	)
	tokenURL := trafficProviderOptionalTrimmedString(
		openSkyTokenURLEnvironmentVariable,
		opensky.DefaultTokenURL,
	)

	result := TrafficProviderConfig{
		Provider:               provider,
		OpenSkyBaseURL:         baseURL,
		OpenSkyTokenURL:        tokenURL,
		OpenSkyClientID:        clientID,
		OpenSkyClientSecret:    clientSecret,
		OpenSkyTimeout:         timeout,
		OpenSkyPollingInterval: pollingInterval,
	}

	openSkyConfig := opensky.DefaultConfig()
	openSkyConfig.BaseURL = result.OpenSkyBaseURL
	openSkyConfig.TokenURL = result.OpenSkyTokenURL
	openSkyConfig.ClientID = result.OpenSkyClientID
	openSkyConfig.ClientSecret = result.OpenSkyClientSecret
	openSkyConfig.PollingInterval = result.OpenSkyPollingInterval
	if err := openSkyConfig.Validate(); err != nil {
		return TrafficProviderConfig{}, fmt.Errorf(
			"validate OpenSky traffic provider configuration: %w",
			err,
		)
	}

	return result, nil
}

func trafficProviderOptionalTrimmedString(
	name string,
	fallback string,
) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func trafficProviderOptionalPositiveDuration(
	name string,
	fallback time.Duration,
) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return parsed, nil
}
