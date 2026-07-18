package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurepipeline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")
	os.Exit(runCommand(os.Args[1:], os.Stdout, os.Stderr, time.Now))
}

func runCommand(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	now func() time.Time,
) int {
	if now == nil {
		fmt.Fprintln(stderr, "ERROR: command clock is required")
		return 1
	}

	options, err := parseCommandOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: parse command options: %v\n", err)
		return 1
	}

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load feature materialization configuration: %v\n",
			err,
		)
		return 1
	}

	pool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: connect PostgreSQL: %v\n", err)
		return 1
	}
	defer pool.Close()

	trajectoryRepository := postgres.NewTrajectoryRepository(pool)
	aircraftRepository := postgres.NewAircraftRepository(pool)
	composition, err := featurepipeline.NewPostgres(
		featurepipeline.PostgresConfig{
			Extractor: extractorcomposition.Config{
				AircraftLookup: aircraftRepository,
				IsAircraftNotFound: func(err error) bool {
					return errors.Is(err, aircraft.ErrNotFound)
				},
				Now: now,
			},
			Pool: pool,
			Now:  now,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose PostgreSQL feature pipeline: %v\n",
			err,
		)
		return 1
	}

	operation, err := newMaterializationOperation(
		trajectoryRepository,
		composition.Pipeline,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose feature materialization operation: %v\n",
			err,
		)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.MigrationTimeout,
	)
	defer cancel()

	report, err := operation.Execute(ctx, options)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: execute flight feature materialization: %v\n",
			err,
		)
		return 1
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: encode feature materialization report: %v\n",
			err,
		)
		return 1
	}

	return 0
}
