package main

import (
	"fmt"
	"net/http"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/adsblol"
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
	selection config.TrafficProviderConfig,
	executor regionalprovider.Executor,
	responseObserver integrationcommon.ProviderResponseObserver,
	fallbackRecorder providerfallback.DecisionRecorder,
	healthSources ...trafficProviderHealthSource,
) (trafficProviderSelection, error) {
	switch selection.Provider {
	case config.TrafficProviderADSBLOL,
		config.TrafficProviderAirplanesLive,
		config.TrafficProviderOpenSky:
		result, err := buildSingleTrafficProvider(
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
		candidates := selection.AutomaticCandidates()
		if len(candidates) == 0 {
			return trafficProviderSelection{}, fmt.Errorf(
				"automatic traffic provider selection returned no candidates",
			)
		}

		primary, err := buildSingleTrafficProvider(
			selection,
			candidates[0],
			executor,
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"build primary %s provider: %w",
				candidates[0],
				err,
			)
		}

		if len(candidates) == 1 {
			primary.Mode = config.TrafficProviderAuto
			return primary, nil
		}

		secondary, err := buildSingleTrafficProvider(
			selection,
			candidates[1],
			executor,
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"build secondary %s provider: %w",
				candidates[1],
				err,
			)
		}

		fallbackProvider, err := newTrafficFallbackProvider(
			primary,
			secondary,
			providerfallback.New(nil),
			fallbackRecorder,
			healthSources...,
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
	selection config.TrafficProviderConfig,
	providerName config.TrafficProvider,
	executor regionalprovider.Executor,
	responseObserver integrationcommon.ProviderResponseObserver,
) (trafficProviderSelection, error) {
	if err := selection.RequireEligible(providerName); err != nil {
		return trafficProviderSelection{}, err
	}

	switch providerName {
	case config.TrafficProviderADSBLOL:
		client, err := adsblol.NewClientWithResponseObserver(
			integrationcommon.HTTPClientConfig{
				BaseURL:   selection.ADSBLOL.BaseURL,
				Timeout:   selection.ADSBLOL.Timeout,
				UserAgent: "global-flight-analytics-ingest",
			},
			responseObserver,
		)
		if err != nil {
			return trafficProviderSelection{}, fmt.Errorf(
				"create ADSB.lol client: %w",
				err,
			)
		}

		return orchestrateTrafficProvider(
			adsblol.NewProvider(client),
			providerpolicy.ProviderADSBLOL,
			executor,
		)

	case config.TrafficProviderAirplanesLive:
		client, err := airplaneslive.NewClientWithResponseObserver(
			integrationcommon.HTTPClientConfig{
				BaseURL:   airplaneslive.BaseURL,
				Timeout:   selection.AirplanesLive.Timeout,
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
		clientConfig.BaseURL = selection.OpenSky.BaseURL
		clientConfig.TokenURL = selection.OpenSky.TokenURL
		clientConfig.ClientID = selection.OpenSky.ClientID
		clientConfig.ClientSecret = selection.OpenSky.ClientSecret
		clientConfig.HTTPClient = &http.Client{
			Timeout: selection.OpenSky.Timeout,
		}
		clientConfig.UserAgent = "global-flight-analytics-ingest"
		clientConfig.PollingInterval = selection.OpenSky.PollingInterval

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
