package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/airplaneslive"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/opensky"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
)

type trafficProviderSelection struct {
	Provider    trafficingestion.RegionalProvider
	ProviderID  providerpolicy.Provider
	ProviderIDs []providerpolicy.Provider
	Mode        config.TrafficProvider
}

func buildTrafficProvider(
	airplanesLiveTimeout time.Duration,
	selection config.TrafficProviderConfig,
	executor regionalprovider.Executor,
	responseObserver integrationcommon.ProviderResponseObserver,
	fallbackRecorder providerfallback.DecisionRecorder,
) (trafficProviderSelection, error) {
	switch selection.Provider {
	case config.TrafficProviderAirplanesLive,
		config.TrafficProviderOpenSky:
		result, err := buildSingleTrafficProvider(
			airplanesLiveTimeout,
			selection,
			selection.Provider,
			executor,
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, err
		}
		result.Mode = selection.Provider
		return result, nil

	case config.TrafficProviderAuto:
		primary, err := buildSingleTrafficProvider(
			airplanesLiveTimeout,
			selection,
			config.TrafficProviderAirplanesLive,
			executor,
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"build primary airplanes.live provider: %w",
				err,
			)
		}

		secondary, err := buildSingleTrafficProvider(
			airplanesLiveTimeout,
			selection,
			config.TrafficProviderOpenSky,
			executor,
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"build secondary OpenSky provider: %w",
				err,
			)
		}

		fallbackProvider, err := newTrafficFallbackProvider(
			primary,
			secondary,
			providerfallback.New(nil),
			fallbackRecorder,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"create automatic traffic fallback provider: %w",
				err,
			)
		}

		return trafficProviderSelection{
			Provider:   fallbackProvider,
			ProviderID: primary.ProviderID,
			ProviderIDs: []providerpolicy.Provider{
				primary.ProviderID,
				secondary.ProviderID,
			},
			Mode: config.TrafficProviderAuto,
		}, nil

	default:
		return trafficProviderSelection{}, fmt.Errorf(
			"unsupported traffic provider: %s",
			selection.Provider,
		)
	}
}

func buildSingleTrafficProvider(
	airplanesLiveTimeout time.Duration,
	selection config.TrafficProviderConfig,
	providerName config.TrafficProvider,
	executor regionalprovider.Executor,
	responseObserver integrationcommon.ProviderResponseObserver,
) (trafficProviderSelection, error) {
	switch providerName {
	case config.TrafficProviderAirplanesLive:
		client, err := airplaneslive.NewClientWithResponseObserver(
			integrationcommon.HTTPClientConfig{
				BaseURL:   airplaneslive.BaseURL,
				Timeout:   airplanesLiveTimeout,
				UserAgent: "global-flight-analytics-ingest",
			},
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"create airplanes.live client: %w",
				err,
			)
		}

		return orchestrateTrafficProvider(
			airplaneslive.NewProvider(client),
			providerpolicy.ProviderAirplanesLive,
			executor,
		)

	case config.TrafficProviderOpenSky:
		clientConfig := opensky.DefaultConfig()
		clientConfig.BaseURL = selection.OpenSkyBaseURL
		clientConfig.TokenURL = selection.OpenSkyTokenURL
		clientConfig.ClientID = selection.OpenSkyClientID
		clientConfig.ClientSecret = selection.OpenSkyClientSecret
		clientConfig.HTTPClient = &http.Client{
			Timeout: selection.OpenSkyTimeout,
		}
		clientConfig.UserAgent = "global-flight-analytics-ingest"
		clientConfig.PollingInterval = selection.OpenSkyPollingInterval

		client, err := opensky.NewClientWithResponseObserver(
			clientConfig,
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"create OpenSky client: %w",
				err,
			)
		}
		provider, err := opensky.NewProvider(client)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"create OpenSky regional provider: %w",
				err,
			)
		}

		return orchestrateTrafficProvider(
			provider,
			providerpolicy.ProviderOpenSky,
			executor,
		)

	default:
		return trafficProviderSelection{}, fmt.Errorf(
			"unsupported single traffic provider: %s",
			providerName,
		)
	}
}

func orchestrateTrafficProvider(
	delegate regionalprovider.Delegate,
	providerID providerpolicy.Provider,
	executor regionalprovider.Executor,
) (trafficProviderSelection, error) {
	provider, err := regionalprovider.New(
		regionalprovider.Config{
			Provider:   delegate,
			ProviderID: providerID,
			Executor:   executor,
		},
	)
	if err != nil {
		return trafficProviderSelection{}, fmt.Errorf(
			"create orchestrated regional traffic provider: %w",
			err,
		)
	}

	return trafficProviderSelection{
		Provider:   provider,
		ProviderID: providerID,
		ProviderIDs: []providerpolicy.Provider{
			providerID,
		},
	}, nil
}
