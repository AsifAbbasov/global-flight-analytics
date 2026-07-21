package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrationrepair"
	"github.com/joho/godotenv"
)

const blockedExitCode = 2

func main() {
	os.Exit(run(
		os.Args[1:],
		os.Stdout,
		os.Stderr,
	))
}

func run(
	args []string,
	stdout *os.File,
	stderr *os.File,
) int {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	flags := flag.NewFlagSet(
		"verify-migration-sequence-repair",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	outputFormat := flags.String(
		"format",
		migrationrepair.OutputFormatText,
		"report format: text or json",
	)
	strict := flags.Bool(
		"strict",
		false,
		"exit with code 2 when repair is blocked",
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load migration configuration: %v\n",
			err,
		)
		return 1
	}

	pool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: connect postgres: %v\n",
			err,
		)
		return 1
	}
	defer pool.Close()

	inspector, err :=
		migrationrepair.NewPostgresInspector(pool)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: create migration repair inspector: %v\n",
			err,
		)
		return 1
	}

	verifier, err := migrationrepair.New(
		migrationrepair.Config{
			Inspector:     inspector,
			MigrationsDir: cfg.MigrationsDir,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: create migration repair verifier: %v\n",
			err,
		)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.MigrationTimeout,
	)
	defer cancel()

	report, err := verifier.Verify(ctx)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify migration sequence repair: %v\n",
			err,
		)
		return 1
	}

	if err := migrationrepair.WriteReport(
		stdout,
		report,
		*outputFormat,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: write migration repair report: %v\n",
			err,
		)
		return 1
	}

	if *strict && !report.Ready {
		return blockedExitCode
	}

	return 0
}
