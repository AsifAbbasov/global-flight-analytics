package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/adsblol"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerdecision"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	providerhealthservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/providerhealth"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/livecollector"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errLiveTrafficDatabaseRequired = errors.New(
	"live traffic collector requires PostgreSQL for durable provider budget state",
)

type liveTrafficComposition struct {
	Store     *livetraffic.Store
	Collector *livecollector.Collector
}

func composeLiveTrafficCollector(
	dbPool *pgxpool.Pool,
	registry *observability.Registry,
	cfg config.LiveTrafficCollectorConfig,
	log *slog.Logger,
) (liveTrafficComposition, error) {
	if !cfg.Enabled {
		return liveTrafficComposition{}, nil
	}
	if dbPool == nil {
		return liveTrafficComposition{}, errLiveTrafficDatabaseRequired
	}
	if registry == nil {
		return liveTrafficComposition{}, errors.New(
			"live traffic collector requires observability registry",
		)
	}
	if cfg.Provider != config.LiveTrafficProviderADSBLOL {
		return liveTrafficComposition{}, fmt.Errorf(
			"unsupported live traffic provider: %s",
			cfg.Provider,
		)
	}

	store, err := livetraffic.NewStore(livetraffic.DefaultConfig())
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create live traffic store: %w",
			err,
		)
	}

	budgetStore, err := postgres.NewProviderBudgetStore(
		dbPool,
		cfg.BudgetTimeout,
	)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create durable live provider budget store: %w",
			err,
		)
	}
	budgetManager, err := providerbudget.NewDurable(budgetStore, nil)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create durable live provider budget manager: %w",
			err,
		)
	}
	responseController, err := providerresponse.New(
		providerresponse.Config{BudgetManager: budgetManager},
	)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create live provider response controller: %w",
			err,
		)
	}

	healthCollector := providerhealthservice.New(nil)
	decisionCollector := providerdecision.New(nil)
	responseObserver, err := providerresponse.NewIntegrationObserverWithRecorder(
		responseController,
		observability.NewProviderRecorder(registry, healthCollector),
	)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create live provider response observer: %w",
			err,
		)
	}

	orchestrator, err := ingestionorchestrator.
		NewDefaultWithDecisionRecorder[regionalprovider.ExecutionValue](
		responseController,
		decisionCollector,
	)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create live traffic ingestion orchestrator: %w",
			err,
		)
	}

	client, err := adsblol.NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   cfg.ADSBLOLBaseURL,
			Timeout:   cfg.ADSBLOLTimeout,
			UserAgent: "global-flight-analytics-live",
		},
		responseObserver,
	)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create adsb.lol client: %w",
			err,
		)
	}

	regional, err := regionalprovider.New(regionalprovider.Config{
		Provider:   adsblol.NewProvider(client),
		ProviderID: providerpolicy.ProviderADSBLOL,
		Executor:   orchestrator,
	})
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create orchestrated adsb.lol regional provider: %w",
			err,
		)
	}

	decisionMetrics, err := observability.NewProviderDecisionCollector(
		decisionCollector,
		[]providerpolicy.Provider{providerpolicy.ProviderADSBLOL},
	)
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create live provider decision metrics: %w",
			err,
		)
	}
	if err := registry.RegisterCollector(decisionMetrics); err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"register live provider decision metrics: %w",
			err,
		)
	}

	collector, err := livecollector.New(livecollector.Config{
		Loader: regional,
		Store:  store,
		Targets: []livecollector.Target{{
			Name:      cfg.TargetName,
			Latitude:  cfg.Latitude,
			Longitude: cfg.Longitude,
			RadiusNM:  cfg.RadiusNM,
		}},
		PollInterval:   cfg.PollInterval,
		RequestTimeout: cfg.RequestTimeout,
		MaxBackoff:     cfg.MaxBackoff,
		JitterRatio:    cfg.JitterRatio,
		Logger:         log,
	})
	if err != nil {
		return liveTrafficComposition{}, fmt.Errorf(
			"create central live traffic collector: %w",
			err,
		)
	}

	return liveTrafficComposition{
		Store:     store,
		Collector: collector,
	}, nil
}
