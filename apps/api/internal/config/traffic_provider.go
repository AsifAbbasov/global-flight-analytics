package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/adsblol"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/opensky"
)

type TrafficProvider string

const (
	TrafficProviderADSBLOL       TrafficProvider = "adsb.lol"
	TrafficProviderAirplanesLive TrafficProvider = "airplanes.live"
	TrafficProviderOpenSky       TrafficProvider = "opensky"
	TrafficProviderAuto          TrafficProvider = "auto"
)

const (
	trafficProviderEnvironmentVariable = "TRAFFIC_PROVIDER"

	adsbLOLBaseURLEnvironmentVariable = "ADSBLOL_BASE_URL"
	adsbLOLTimeoutEnvironmentVariable = "ADSBLOL_TIMEOUT"

	airplanesLiveTimeoutEnvironmentVariable        = "AIRPLANES_LIVE_TIMEOUT"
	airplanesLiveAccessApprovedEnvironmentVariable = "AIRPLANES_LIVE_ACCESS_APPROVED"

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
		"traffic provider must be adsb.lol, airplanes.live, opensky, or auto",
	)
	ErrOpenSkyCredentialPairRequired = errors.New(
		"OpenSky client id and client secret must be configured together",
	)
	ErrAirplanesLiveAccessApprovalRequired = errors.New(
		"airplanes.live access must be explicitly approved before runtime selection",
	)
	ErrOpenSkyOperationalAgreementRequired = errors.New(
		"OpenSky operational agreement must be explicitly confirmed before runtime selection",
	)
)

type ADSBLOLProviderConfig struct {
	BaseURL string
	Timeout time.Duration
}

type AirplanesLiveProviderConfig struct {
	Timeout        time.Duration
	AccessApproved bool
}

type OpenSkyProviderConfig struct {
	BaseURL                       string
	TokenURL                      string
	ClientID                      string
	ClientSecret                  string
	Timeout                       time.Duration
	PollingInterval               time.Duration
	OperationalAgreementConfirmed bool
}

type TrafficProviderConfig struct {
	Provider TrafficProvider

	ADSBLOL       ADSBLOLProviderConfig
	AirplanesLive AirplanesLiveProviderConfig
	OpenSky       OpenSkyProviderConfig
}

func (config TrafficProviderConfig) RequireEligible(
	provider TrafficProvider,
) error {
	switch provider {
	case TrafficProviderADSBLOL:
		return nil
	case TrafficProviderAirplanesLive:
		if !config.AirplanesLive.AccessApproved {
			return ErrAirplanesLiveAccessApprovalRequired
		}
		return nil
	case TrafficProviderOpenSky:
		if !config.OpenSky.OperationalAgreementConfirmed {
			return ErrOpenSkyOperationalAgreementRequired
		}
		return nil
	default:
		return fmt.Errorf(
			"%w: %q",
			ErrTrafficProviderInvalid,
			provider,
		)
	}
}

func (config TrafficProviderConfig) AutomaticCandidates() []TrafficProvider {
	candidates := []TrafficProvider{
		TrafficProviderADSBLOL,
	}

	if config.AirplanesLive.AccessApproved {
		return append(
			candidates,
			TrafficProviderAirplanesLive,
		)
	}

	if config.OpenSky.OperationalAgreementConfirmed {
		return append(
			candidates,
			TrafficProviderOpenSky,
		)
	}

	return candidates
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
		provider = TrafficProviderADSBLOL
	}
	if provider != TrafficProviderADSBLOL &&
		provider != TrafficProviderAirplanesLive &&
		provider != TrafficProviderOpenSky &&
		provider != TrafficProviderAuto {
		return TrafficProviderConfig{}, fmt.Errorf(
			"%w: %q",
			ErrTrafficProviderInvalid,
			provider,
		)
	}

	adsbLOLTimeout, err := trafficProviderOptionalPositiveDuration(
		adsbLOLTimeoutEnvironmentVariable,
		15*time.Second,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
	}

	airplanesLiveTimeout, err := trafficProviderOptionalPositiveDuration(
		airplanesLiveTimeoutEnvironmentVariable,
		15*time.Second,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
	}
	airplanesLiveAccessApproved, err := trafficProviderOptionalBoolean(
		airplanesLiveAccessApprovedEnvironmentVariable,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
	}

	openSkyOperationalAgreementConfirmed, err := trafficProviderOptionalBoolean(
		openSkyOperationalAgreementConfirmedEnvironmentVariable,
	)
	if err != nil {
		return TrafficProviderConfig{}, err
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

	openSkyTimeout, err := trafficProviderOptionalPositiveDuration(
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

	result := TrafficProviderConfig{
		Provider: provider,
		ADSBLOL: ADSBLOLProviderConfig{
			BaseURL: trafficProviderOptionalTrimmedString(
				adsbLOLBaseURLEnvironmentVariable,
				adsblol.BaseURL,
			),
			Timeout: adsbLOLTimeout,
		},
		AirplanesLive: AirplanesLiveProviderConfig{
			Timeout:        airplanesLiveTimeout,
			AccessApproved: airplanesLiveAccessApproved,
		},
		OpenSky: OpenSkyProviderConfig{
			BaseURL: trafficProviderOptionalTrimmedString(
				openSkyBaseURLEnvironmentVariable,
				opensky.DefaultBaseURL,
			),
			TokenURL: trafficProviderOptionalTrimmedString(
				openSkyTokenURLEnvironmentVariable,
				opensky.DefaultTokenURL,
			),
			ClientID:                      clientID,
			ClientSecret:                  clientSecret,
			Timeout:                       openSkyTimeout,
			PollingInterval:               pollingInterval,
			OperationalAgreementConfirmed: openSkyOperationalAgreementConfirmed,
		},
	}

	openSkyConfig := opensky.DefaultConfig()
	openSkyConfig.BaseURL = result.OpenSky.BaseURL
	openSkyConfig.TokenURL = result.OpenSky.TokenURL
	openSkyConfig.ClientID = result.OpenSky.ClientID
	openSkyConfig.ClientSecret = result.OpenSky.ClientSecret
	openSkyConfig.PollingInterval = result.OpenSky.PollingInterval
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

func trafficProviderOptionalBoolean(
	name string,
) (bool, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return false, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s as boolean: %w", name, err)
	}
	return parsed, nil
}
