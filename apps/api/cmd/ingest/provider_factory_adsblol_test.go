package main

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
)

type providerFactoryExecutorStub struct{}

func (providerFactoryExecutorStub) Execute(
	context.Context,
	providerpolicy.Provider,
	string,
	ingestionorchestrator.Function[regionalprovider.ExecutionValue],
) (
	ingestionorchestrator.ExecuteResult[regionalprovider.ExecutionValue],
	error,
) {
	return ingestionorchestrator.ExecuteResult[regionalprovider.ExecutionValue]{}, nil
}

func loadProviderFactoryConfig(
	t *testing.T,
	provider config.TrafficProvider,
) config.TrafficProviderConfig {
	t.Helper()

	t.Setenv("TRAFFIC_PROVIDER", string(provider))
	t.Setenv("ADSBLOL_BASE_URL", "")
	t.Setenv("ADSBLOL_TIMEOUT", "")
	t.Setenv("AIRPLANES_LIVE_TIMEOUT", "")
	t.Setenv("AIRPLANES_LIVE_ACCESS_APPROVED", "")
	t.Setenv("OPENSKY_BASE_URL", "")
	t.Setenv("OPENSKY_TOKEN_URL", "")
	t.Setenv("OPENSKY_CLIENT_ID", "")
	t.Setenv("OPENSKY_CLIENT_SECRET", "")
	t.Setenv("OPENSKY_TIMEOUT", "")
	t.Setenv("OPENSKY_POLLING_INTERVAL", "")
	t.Setenv("OPENSKY_OPERATIONAL_AGREEMENT_CONFIRMED", "")

	loaded, err := config.LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load provider config: %v", err)
	}

	return loaded
}

func TestBuildTrafficProviderAutoDefaultsToADSBLOLOnly(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderAuto,
	)

	result, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if err != nil {
		t.Fatalf("build auto provider: %v", err)
	}

	if result.ProviderID != providerpolicy.ProviderADSBLOL {
		t.Fatalf("provider ID = %s, want adsb.lol", result.ProviderID)
	}
	if len(result.ProviderIDs) != 1 ||
		result.ProviderIDs[0] != providerpolicy.ProviderADSBLOL {
		t.Fatalf("provider IDs = %+v, want adsb.lol only", result.ProviderIDs)
	}
	if result.Mode != config.TrafficProviderAuto {
		t.Fatalf("mode = %s, want auto", result.Mode)
	}
	if result.Provider.SourceName() != "adsb.lol" {
		t.Fatalf("source = %q, want adsb.lol", result.Provider.SourceName())
	}
}

func TestBuildTrafficProviderAutoAddsApprovedAirplanesLiveSecondary(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderAuto,
	)
	selection.AirplanesLive.AccessApproved = true

	result, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if err != nil {
		t.Fatalf("build approved auto provider: %v", err)
	}

	if len(result.ProviderIDs) != 2 {
		t.Fatalf("provider IDs = %+v, want 2 providers", result.ProviderIDs)
	}
	if result.ProviderIDs[0] != providerpolicy.ProviderADSBLOL ||
		result.ProviderIDs[1] != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"provider order = %+v, want adsb.lol -> airplanes.live",
			result.ProviderIDs,
		)
	}
}

func TestBuildTrafficProviderAutoAddsApprovedOpenSkyFallback(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderAuto,
	)
	selection.OpenSky.OperationalAgreementConfirmed = true

	result, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if err != nil {
		t.Fatalf("build approved auto provider: %v", err)
	}

	if len(result.ProviderIDs) != 2 {
		t.Fatalf("provider IDs = %+v, want 2 providers", result.ProviderIDs)
	}
	if result.ProviderIDs[0] != providerpolicy.ProviderADSBLOL ||
		result.ProviderIDs[1] != providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"provider order = %+v, want adsb.lol -> opensky",
			result.ProviderIDs,
		)
	}
}

func TestBuildTrafficProviderPrefersApprovedAirplanesLiveOverOpenSky(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderAuto,
	)
	selection.AirplanesLive.AccessApproved = true
	selection.OpenSky.OperationalAgreementConfirmed = true

	result, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if err != nil {
		t.Fatalf("build approved auto provider: %v", err)
	}

	if len(result.ProviderIDs) != 2 ||
		result.ProviderIDs[1] != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"provider IDs = %+v, want adsb.lol -> airplanes.live",
			result.ProviderIDs,
		)
	}
}

func TestBuildTrafficProviderRejectsUnapprovedAirplanesLive(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderAirplanesLive,
	)

	_, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if !errors.Is(
		err,
		config.ErrAirplanesLiveAccessApprovalRequired,
	) {
		t.Fatalf(
			"expected airplanes.live approval error, got %v",
			err,
		)
	}
}

func TestBuildTrafficProviderRejectsUnapprovedOpenSky(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderOpenSky,
	)

	_, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if !errors.Is(
		err,
		config.ErrOpenSkyOperationalAgreementRequired,
	) {
		t.Fatalf(
			"expected OpenSky agreement error, got %v",
			err,
		)
	}
}

func TestBuildTrafficProviderAcceptsDirectADSBLOL(
	t *testing.T,
) {
	selection := loadProviderFactoryConfig(
		t,
		config.TrafficProviderADSBLOL,
	)

	result, err := buildTrafficProvider(
		selection,
		providerFactoryExecutorStub{},
		nil,
		&fallbackDecisionRecorderStub{},
	)
	if err != nil {
		t.Fatalf("build direct ADSB.lol provider: %v", err)
	}

	if result.ProviderID != providerpolicy.ProviderADSBLOL ||
		result.Provider.SourceName() != "adsb.lol" {
		t.Fatalf("unexpected direct ADSB.lol selection: %+v", result)
	}
}
