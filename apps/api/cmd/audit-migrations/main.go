package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrationaudit"
	"github.com/joho/godotenv"
)

const blockerExitCode = 2

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
		"audit-migrations",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	outputFormat := flags.String(
		"format",
		migrationaudit.OutputFormatText,
		"report format: text or json",
	)
	strict := flags.Bool(
		"strict",
		false,
		"exit with code 2 when audit blockers are found",
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

	migrationsDir, err := validateMigrationsDir(
		cfg.MigrationsDir,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: validate migrations directory: %v\n",
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

	stateLoader, err :=
		migrationaudit.NewPostgresStateLoader(pool)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: create migration state loader: %v\n",
			err,
		)
		return 1
	}

	auditor, err := migrationaudit.New(
		migrationaudit.Config{
			MigrationsDir: migrationsDir,
			StateLoader:   stateLoader,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: create migration auditor: %v\n",
			err,
		)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.MigrationTimeout,
	)
	defer cancel()

	report, err := auditor.Audit(ctx)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: audit migration history: %v\n",
			err,
		)
		return 1
	}

	if err := migrationaudit.WriteReport(
		stdout,
		report,
		*outputFormat,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: write migration audit report: %v\n",
			err,
		)
		return 1
	}

	if *strict && report.BlockerCount > 0 {
		return blockerExitCode
	}

	return 0
}

func validateMigrationsDir(
	path string,
) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", fmt.Errorf(
			"migrations directory path is required",
		)
	}

	absolutePath, err := filepath.Abs(
		trimmedPath,
	)
	if err != nil {
		return "", fmt.Errorf(
			"resolve migrations directory path %q: %w",
			trimmedPath,
			err,
		)
	}

	cleanPath := filepath.Clean(
		absolutePath,
	)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf(
			"stat migrations directory %q: %w",
			cleanPath,
			err,
		)
	}
	if !info.IsDir() {
		return "", fmt.Errorf(
			"%q is not a directory",
			cleanPath,
		)
	}

	return cleanPath, nil
}
