package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/ourairports"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/joho/godotenv"
)

const ourAirportsPublicationLeaseDuration = 30 * time.Minute

func main() {
	totalStartedAt := time.Now()

	_ = godotenv.Load()

	cfg, err := config.LoadImportAirportsConfig()
	if err != nil {
		log.Fatalf(
			"load import-airports configuration: %v",
			err,
		)
	}

	timeout := cfg.OurAirportsTimeout
	countryCodes := cfg.OurAirportsCountryCodes

	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		log.Fatalf(
			"connect postgres: %v",
			err,
		)
	}
	defer dbPool.Close()

	validatorRepository := postgres.NewSourceHTTPValidatorRepository(
		dbPool,
	)

	conditionalRequest, err := loadConditionalRequest(
		ctx,
		validatorRepository,
	)
	if err != nil {
		log.Fatal(err)
	}

	client, err := ourairports.NewClient(
		ourairports.ClientConfig{
			Timeout:      timeout,
			CountryCodes: countryCodes,
		},
	)
	if err != nil {
		log.Fatalf(
			"create OurAirports client: %v",
			err,
		)
	}

	sourceLoadStartedAt := time.Now()

	result, err := client.LoadAirportsConditional(
		ctx,
		conditionalRequest,
	)
	if err != nil {
		log.Fatalf(
			"load OurAirports airports: %v",
			err,
		)
	}

	sourceLoadDuration := time.Since(
		sourceLoadStartedAt,
	)

	if result.NotModified {
		validatorWritten, err := persistHTTPValidatorIfChanged(
			ctx,
			validatorRepository,
			conditionalRequest,
			result,
		)
		if err != nil {
			log.Fatal(err)
		}

		validatorWriteStatus := "skipped"
		if validatorWritten {
			validatorWriteStatus = "updated"
		}

		totalDuration := time.Since(
			totalStartedAt,
		)

		fmt.Printf(
			"source=%s countries=%s retrieval_status=not_modified validator_write=%s checked_at=%s source_load_duration=%s total_duration=%s\n",
			ourairports.SourceName,
			strings.Join(
				countryCodes,
				",",
			),
			validatorWriteStatus,
			result.CheckedAt.Format(
				time.RFC3339,
			),
			sourceLoadDuration,
			totalDuration,
		)

		return
	}

	airportRepository := postgres.NewAirportRepository(
		dbPool,
	)
	publicationRepository := postgres.NewProviderPublicationRepository(
		dbPool,
		ourAirportsPublicationLeaseDuration,
		nil,
	)
	publicationOrchestrator, err := ingestionorchestrator.NewPublicationOnly[airportImportExecutionValue](
		publicationRepository,
		nil,
	)
	if err != nil {
		log.Fatalf(
			"create OurAirports publication orchestrator: %v",
			err,
		)
	}

	databaseImportStartedAt := time.Now()
	publicationOutcome, err := executeAirportPublication(
		ctx,
		publicationOrchestrator,
		airportRepository,
		result,
	)
	if err != nil {
		log.Fatalf(
			"execute OurAirports publication import: %v",
			err,
		)
	}
	databaseImportDuration := time.Since(databaseImportStartedAt)

	if err := persistHTTPValidator(
		ctx,
		validatorRepository,
		result,
	); err != nil {
		log.Fatal(err)
	}

	retrievalStatus := "downloaded"
	if publicationOutcome.AlreadyProcessed {
		retrievalStatus = "already_processed"
	}

	totalDuration := time.Since(
		totalStartedAt,
	)

	fmt.Printf(
		"source=%s countries=%s retrieval_status=%s publication_id=%s received=%d reconciled=%d shared=%t retrieved_at=%s source_load_duration=%s database_reconciliation_duration=%s total_duration=%s\n",
		ourairports.SourceName,
		strings.Join(
			countryCodes,
			",",
		),
		retrievalStatus,
		result.PublicationID,
		len(result.Airports),
		publicationOutcome.ReconciledCount,
		publicationOutcome.Shared,
		result.RetrievedAt.Format(
			time.RFC3339,
		),
		sourceLoadDuration,
		databaseImportDuration,
		totalDuration,
	)

}
