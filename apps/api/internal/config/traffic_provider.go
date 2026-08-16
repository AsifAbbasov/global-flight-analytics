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
	trafficProviderEnvironmentVariable                      = "TRAFFIC_PROVIDER"
	openSkyBaseURLEnvironmentVariable                       = "OPENSKY_BASE_URL"
	openSkyTokenURLEnvironmentVariable                      = "OPENSKY_TOKEN_URL"
	openSkyClientIDEnvironmentVariable                      = "OPENSKY_CLIENT_ID"
	openSkyClientSecretEnvironmentVariable                  = "OPENSKY_CLIENT_SECRET"
	openSkyTimeoutEnvironmentVariable                       = "OPENSKY_TIMEOUT"
	openSkyPollingIntervalEnvironmentVariable               = "OPENSKY_POLLING_INTERVAL"
	openSkyOperationalAgreementConfirmedEnvironmentVariable = "OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED"
)

var (
	ErrTrafficProviderInvalid = errors.New(
		"traffic provider must be airplanes.live, opensky, or auto",
	)
	ErrOpenSkyCredentialPairRequired = errors.New(
		"OpenSky client id and client secret must be configured together",
	)
	ErrOpenSkyOperationalAgreementRequired = errors.New(
		"OpenSky operational REST use requires a prior written agreement",
	)
	ErrOpenSkyOperationalAgreementInvalid = errors.New(
		"OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED must be true or false",
	)
)

type TrafficProviderConfig struct {
	Provider TrafficProvider

	OpenSkyBaseURL                       string
	OpenSkyTokenURL                      string
	OpenSkyClientID                      string
	OpenSkyClientSecret                  string
	OpenSkyTimeout                       time.Duration
	OpenSkyPollingInterval               time.Duration
	OpenSkyOperationalAgreementConfirmed bool
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

	openSkyOperationalAgreementConfirmed, err := trafficProviderOptionalBoolean(
		openSkyOperationalAgreementConfirmedEnvironmentVariable,
		false,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
	}
	if (provider == TrafficProviderOpenSky || provider == TrafficProviderAuto) &&
		!openSkyOperationalAgreementConfirmed {
		return TrafficProviderConfig{}, fmt.Errorf(
			"%w: set %s=true only after the required written agreement is confirmed",
			ErrOpenSkyOperationalAgreementRequired,
			openSkyOperationalAgreementConfirmedEnvironmentVariable,
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
		Provider:                             provider,
		OpenSkyBaseURL:                       baseURL,
		OpenSkyTokenURL:                      tokenURL,
		OpenSkyClientID:                      clientID,
		OpenSkyClientSecret:                  clientSecret,
		OpenSkyTimeout:                       timeout,
		OpenSkyPollingInterval:               pollingInterval,
		OpenSkyOperationalAgreementConfirmed: openSkyOperationalAgreementConfirmed,
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

func trafficProviderOptionalBoolean(
	name string,
	fallback bool,
) (bool, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}

	switch strings.ToLower(value) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf(
			"%w: %q",
			ErrOpenSkyOperationalAgreementInvalid,
			value,
		)
	}
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
