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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/joho/godotenv"
)

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

	databaseImportStartedAt := time.Now()

	reconciledCount, err := airportRepository.UpsertImported(
		ctx,
		result.Airports,
	)
	if err != nil {
		log.Fatalf(
			"reconcile OurAirports airports: %v",
			err,
		)
	}

	databaseImportDuration := time.Since(
		databaseImportStartedAt,
	)

	if err := persistHTTPValidator(
		ctx,
		validatorRepository,
		result,
	); err != nil {
		log.Fatal(err)
	}

	totalDuration := time.Since(
		totalStartedAt,
	)

	fmt.Printf(
		"source=%s countries=%s retrieval_status=downloaded received=%d reconciled=%d retrieved_at=%s source_load_duration=%s database_reconciliation_duration=%s total_duration=%s\n",
		ourairports.SourceName,
		strings.Join(
			countryCodes,
			",",
		),
		len(result.Airports),
		reconciledCount,
		result.RetrievedAt.Format(
			time.RFC3339,
		),
		sourceLoadDuration,
		databaseImportDuration,
		totalDuration,
	)
}
