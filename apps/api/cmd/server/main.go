package main

import (
	"log"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	var dbPool *pgxpool.Pool

	if cfg.DatabaseURL == "" {
		log.Println("DATABASE_URL is not set. Starting API without database connection.")
	} else {
		var err error

		dbPool, err = database.NewPostgresPool(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}

		defer dbPool.Close()

		log.Println("PostgreSQL connection established.")
	}

	app := server.New(dbPool)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
