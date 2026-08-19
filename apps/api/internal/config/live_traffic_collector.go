package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/adsblol"
)

const (
	liveTrafficCollectorEnabledEnvironmentVariable = "LIVE_TRAFFIC_COLLECTOR_ENABLED"
	liveTrafficProviderEnvironmentVariable         = "LIVE_TRAFFIC_PROVIDER"
	liveTrafficTargetNameEnvironmentVariable       = "LIVE_TRAFFIC_TARGET_NAME"
	liveTrafficLatitudeEnvironmentVariable         = "LIVE_TRAFFIC_LATITUDE"
	liveTrafficLongitudeEnvironmentVariable        = "LIVE_TRAFFIC_LONGITUDE"
	liveTrafficRadiusEnvironmentVariable           = "LIVE_TRAFFIC_RADIUS_NM"
	liveTrafficPollIntervalEnvironmentVariable     = "LIVE_TRAFFIC_POLL_INTERVAL"
	liveTrafficRequestTimeoutEnvironmentVariable   = "LIVE_TRAFFIC_REQUEST_TIMEOUT"
	liveTrafficMaxBackoffEnvironmentVariable       = "LIVE_TRAFFIC_MAX_BACKOFF"
	liveTrafficJitterRatioEnvironmentVariable      = "LIVE_TRAFFIC_JITTER_RATIO"
	liveTrafficBudgetTimeoutEnvironmentVariable    = "LIVE_TRAFFIC_BUDGET_TIMEOUT"

	adsbLOLBaseURLEnvironmentVariable                    = "ADSB_LOL_BASE_URL"
	adsbLOLTimeoutEnvironmentVariable                    = "ADSB_LOL_TIMEOUT"
	adsbLOLProductionContactConfirmedEnvironmentVariable = "ADSB_LOL_PRODUCTION_CONTACT_CONFIRMED"
)

const (
	defaultLiveTrafficPollInterval   = 10 * time.Second
	minimumLiveTrafficPollInterval   = 10 * time.Second
	defaultLiveTrafficRequestTimeout = 8 * time.Second
	defaultLiveTrafficMaxBackoff     = 5 * time.Minute
	defaultLiveTrafficJitterRatio    = 0.10
	defaultLiveTrafficBudgetTimeout  = 5 * time.Second
)

type LiveTrafficProvider string

const (
	LiveTrafficProviderADSBLOL LiveTrafficProvider = "adsb.lol"
)

var (
	ErrLiveTrafficProviderInvalid = errors.New(
		"live traffic provider must be adsb.lol",
	)
	ErrADSBLOLProductionContactRequired = errors.New(
		"adsb.lol production contact confirmation is required before enabling the live collector",
	)
)

type LiveTrafficCollectorConfig struct {
	Enabled  bool
	Provider LiveTrafficProvider

	TargetName string
	Latitude   float64
	Longitude  float64
	RadiusNM   int

	PollInterval   time.Duration
	RequestTimeout time.Duration
	MaxBackoff     time.Duration
	JitterRatio    float64
	BudgetTimeout  time.Duration

	ADSBLOLBaseURL                    string
	ADSBLOLTimeout                    time.Duration
	ADSBLOLProductionContactConfirmed bool
}

func LoadLiveTrafficCollectorConfig() (LiveTrafficCollectorConfig, error) {
	enabled, err := liveTrafficOptionalBoolean(
		liveTrafficCollectorEnabledEnvironmentVariable,
		false,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}
	if !enabled {
		return LiveTrafficCollectorConfig{
			Enabled:  false,
			Provider: LiveTrafficProviderADSBLOL,
		}, nil
	}

	provider := LiveTrafficProvider(
		strings.TrimSpace(os.Getenv(liveTrafficProviderEnvironmentVariable)),
	)
	if provider == "" {
		provider = LiveTrafficProviderADSBLOL
	}
	if provider != LiveTrafficProviderADSBLOL {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%w: %q",
			ErrLiveTrafficProviderInvalid,
			provider,
		)
	}

	contactConfirmed, err := liveTrafficOptionalBoolean(
		adsbLOLProductionContactConfirmedEnvironmentVariable,
		false,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}
	if !contactConfirmed {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%w: set %s=true only after the provider-contact dependency is actually resolved",
			ErrADSBLOLProductionContactRequired,
			adsbLOLProductionContactConfirmedEnvironmentVariable,
		)
	}

	targetName := strings.TrimSpace(os.Getenv(liveTrafficTargetNameEnvironmentVariable))
	if targetName == "" {
		targetName = "baku"
	}

	latitude, err := requiredFiniteFloat64EnvironmentVariable(
		liveTrafficLatitudeEnvironmentVariable,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, fmt.Errorf("load live traffic latitude: %w", err)
	}
	if latitude < -90 || latitude > 90 {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must be between -90 and 90",
			liveTrafficLatitudeEnvironmentVariable,
		)
	}

	longitude, err := requiredFiniteFloat64EnvironmentVariable(
		liveTrafficLongitudeEnvironmentVariable,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, fmt.Errorf("load live traffic longitude: %w", err)
	}
	if longitude < -180 || longitude > 180 {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must be between -180 and 180",
			liveTrafficLongitudeEnvironmentVariable,
		)
	}

	radius, err := requiredIntegerEnvironmentVariable(
		liveTrafficRadiusEnvironmentVariable,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, fmt.Errorf("load live traffic radius: %w", err)
	}
	if radius <= 0 || radius > 250 {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must be between 1 and 250 nautical miles",
			liveTrafficRadiusEnvironmentVariable,
		)
	}

	pollInterval, err := liveTrafficOptionalPositiveDuration(
		liveTrafficPollIntervalEnvironmentVariable,
		defaultLiveTrafficPollInterval,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}
	if pollInterval < minimumLiveTrafficPollInterval {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must be at least %s",
			liveTrafficPollIntervalEnvironmentVariable,
			minimumLiveTrafficPollInterval,
		)
	}

	requestTimeout, err := liveTrafficOptionalPositiveDuration(
		liveTrafficRequestTimeoutEnvironmentVariable,
		defaultLiveTrafficRequestTimeout,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}
	if requestTimeout > pollInterval {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must not exceed %s",
			liveTrafficRequestTimeoutEnvironmentVariable,
			liveTrafficPollIntervalEnvironmentVariable,
		)
	}

	maxBackoff, err := liveTrafficOptionalPositiveDuration(
		liveTrafficMaxBackoffEnvironmentVariable,
		defaultLiveTrafficMaxBackoff,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}
	if maxBackoff < pollInterval {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must be at least %s",
			liveTrafficMaxBackoffEnvironmentVariable,
			liveTrafficPollIntervalEnvironmentVariable,
		)
	}

	jitterRatio, err := liveTrafficOptionalRatio(
		liveTrafficJitterRatioEnvironmentVariable,
		defaultLiveTrafficJitterRatio,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}

	budgetTimeout, err := liveTrafficOptionalPositiveDuration(
		liveTrafficBudgetTimeoutEnvironmentVariable,
		defaultLiveTrafficBudgetTimeout,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}

	adsbTimeout, err := liveTrafficOptionalPositiveDuration(
		adsbLOLTimeoutEnvironmentVariable,
		requestTimeout,
	)
	if err != nil {
		return LiveTrafficCollectorConfig{}, err
	}
	if adsbTimeout > requestTimeout {
		return LiveTrafficCollectorConfig{}, fmt.Errorf(
			"%s must not exceed %s",
			adsbLOLTimeoutEnvironmentVariable,
			liveTrafficRequestTimeoutEnvironmentVariable,
		)
	}

	baseURL := strings.TrimSpace(os.Getenv(adsbLOLBaseURLEnvironmentVariable))
	if baseURL == "" {
		baseURL = adsblol.BaseURL
	}

	return LiveTrafficCollectorConfig{
		Enabled:  true,
		Provider: provider,

		TargetName: targetName,
		Latitude:   latitude,
		Longitude:  longitude,
		RadiusNM:   radius,

		PollInterval:   pollInterval,
		RequestTimeout: requestTimeout,
		MaxBackoff:     maxBackoff,
		JitterRatio:    jitterRatio,
		BudgetTimeout:  budgetTimeout,

		ADSBLOLBaseURL:                    baseURL,
		ADSBLOLTimeout:                    adsbTimeout,
		ADSBLOLProductionContactConfirmed: contactConfirmed,
	}, nil
}

func liveTrafficOptionalBoolean(
	name string,
	fallback bool,
) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	switch strings.ToLower(raw) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be true or false", name)
	}
}

func liveTrafficOptionalPositiveDuration(
	name string,
	fallback time.Duration,
) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return value, nil
}

func liveTrafficOptionalRatio(
	name string,
	fallback float64,
) (float64, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 0.5 {
		return 0, fmt.Errorf("%s must be finite and between 0 and 0.5", name)
	}
	return value, nil
}
