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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrator"
	applogger "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	status := flag.Bool(
		"status",
		false,
		"print migration status without applying SQL",
	)

	flag.Parse()

	log := applogger.New()

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		log.Error(
			"failed to load migration configuration",
			"error",
			err,
		)
		os.Exit(1)
	}

	migrationsDir, err := validateMigrationsDir(
		cfg.MigrationsDir,
	)
	if err != nil {
		log.Error(
			"failed to validate migrations directory",
			"error",
			err,
		)
		os.Exit(1)
	}

	pool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		log.Error(
			"failed to connect postgres",
			"error",
			err,
		)
		os.Exit(1)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.MigrationTimeout,
	)
	defer cancel()

	runner, err := migrator.NewRunner(
		pool,
		migrationsDir,
	)
	if err != nil {
		log.Error(
			"failed to create migration runner",
			"error",
			err,
		)
		os.Exit(1)
	}

	if *status {
		statuses, err := runner.Status(
			ctx,
		)
		if err != nil {
			log.Error(
				"failed to load migration status",
				"error",
				err,
			)
			os.Exit(1)
		}

		printStatuses(
			statuses,
		)

		return
	}

	applied, err := runner.ApplyPending(
		ctx,
	)
	if err != nil {
		log.Error(
			"failed to apply migrations",
			"error",
			err,
		)
		os.Exit(1)
	}

	if len(applied) == 0 {
		log.Info(
			"no pending migrations",
		)

		return
	}

	for _, migration := range applied {
		log.Info(
			"migration applied",
			"version",
			migration.Version,
			"name",
			migration.Name,
		)
	}
}

func validateMigrationsDir(
	path string,
) (string, error) {
	trimmedPath := strings.TrimSpace(
		path,
	)
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

	info, err := os.Stat(
		cleanPath,
	)
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

func printStatuses(
	statuses []migrator.MigrationStatus,
) {
	if len(statuses) == 0 {
		fmt.Println(
			"No migrations found",
		)

		return
	}

	for _, status := range statuses {
		state := "pending"

		if status.Applied {
			state = "applied"
		}

		fmt.Printf(
			"%s %s %s\n",
			status.Migration.Version,
			status.Migration.Name,
			state,
		)
	}
}
