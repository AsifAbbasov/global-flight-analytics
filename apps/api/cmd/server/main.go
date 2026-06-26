package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	applogger "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/logger"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	log := applogger.New()

	var dbPool *pgxpool.Pool

	if cfg.DatabaseURL == "" {
		log.Warn("database url is not set; starting api without database connection")
	} else {
		var err error

		dbPool, err = database.NewPostgresPool(cfg.DatabaseURL)
		if err != nil {
			log.Error("failed to connect postgres", "error", err)
			os.Exit(1)
		}

		log.Info("postgres connection established")
	}

	app := server.New(dbPool)

	go func() {
		log.Info("api server starting", "port", cfg.Port)

		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdownSignal
	log.Info("shutdown signal received", "signal", sig.String())

	if err := app.Shutdown(); err != nil {
		log.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}

	if dbPool != nil {
		dbPool.Close()
		log.Info("postgres connection closed")
	}

	log.Info("api server stopped")
}
