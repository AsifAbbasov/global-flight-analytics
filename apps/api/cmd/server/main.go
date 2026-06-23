package main

import (
	"log"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Println("DATABASE_URL is not set. Starting API without database connection.")
	} else {
		dbPool, err := database.NewPostgresPool(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}

		defer dbPool.Close()

		log.Println("PostgreSQL connection established.")
	}

	app := server.New()

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
