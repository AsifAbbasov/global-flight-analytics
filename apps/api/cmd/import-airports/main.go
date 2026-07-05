package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	timeout := mustDurationEnv("OURAIRPORTS_TIMEOUT")
	countryCodes := mustCountryCodesEnv("OURAIRPORTS_COUNTRY_CODES")

	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer dbPool.Close()

	validatorRepository := postgres.NewSourceHTTPValidatorRepository(dbPool)

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
		log.Fatalf("create OurAirports client: %v", err)
	}

	sourceLoadStartedAt := time.Now()

	result, err := client.LoadAirportsConditional(
		ctx,
		conditionalRequest,
	)
	if err != nil {
		log.Fatalf("load OurAirports airports: %v", err)
	}

	sourceLoadDuration := time.Since(sourceLoadStartedAt)

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

		totalDuration := time.Since(totalStartedAt)

		fmt.Printf(
			"source=%s countries=%s retrieval_status=not_modified validator_write=%s checked_at=%s source_load_duration=%s total_duration=%s\n",
			ourairports.SourceName,
			strings.Join(countryCodes, ","),
			validatorWriteStatus,
			result.CheckedAt.Format(time.RFC3339),
			sourceLoadDuration,
			totalDuration,
		)

		return
	}

	airportRepository := postgres.NewAirportRepository(dbPool)

	databaseImportStartedAt := time.Now()

	importedCount, err := airportRepository.UpsertImported(
		ctx,
		result.Airports,
	)
	if err != nil {
		log.Fatalf("import OurAirports airports: %v", err)
	}

	databaseImportDuration := time.Since(databaseImportStartedAt)

	if err := persistHTTPValidator(
		ctx,
		validatorRepository,
		result,
	); err != nil {
		log.Fatal(err)
	}

	totalDuration := time.Since(totalStartedAt)

	fmt.Printf(
		"source=%s countries=%s retrieval_status=downloaded received=%d imported=%d retrieved_at=%s source_load_duration=%s database_import_duration=%s total_duration=%s\n",
		ourairports.SourceName,
		strings.Join(countryCodes, ","),
		len(result.Airports),
		importedCount,
		result.RetrievedAt.Format(time.RFC3339),
		sourceLoadDuration,
		databaseImportDuration,
		totalDuration,
	)
}

func mustDurationEnv(name string) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("parse %s: %v", name, err)
	}

	if parsed <= 0 {
		log.Fatalf("%s must be greater than zero", name)
	}

	return parsed
}

func mustCountryCodesEnv(name string) []string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		log.Fatalf("%s is required", name)
	}

	rawCountryCodes := strings.Split(value, ",")

	countryCodes := make([]string, 0, len(rawCountryCodes))
	seenCountryCodes := make(map[string]struct{}, len(rawCountryCodes))

	for _, rawCountryCode := range rawCountryCodes {
		countryCode := strings.ToUpper(
			strings.TrimSpace(rawCountryCode),
		)

		if countryCode == "" {
			continue
		}

		if _, exists := seenCountryCodes[countryCode]; exists {
			continue
		}

		seenCountryCodes[countryCode] = struct{}{}
		countryCodes = append(countryCodes, countryCode)
	}

	if len(countryCodes) == 0 {
		log.Fatalf("%s must contain at least one country code", name)
	}

	return countryCodes
}
