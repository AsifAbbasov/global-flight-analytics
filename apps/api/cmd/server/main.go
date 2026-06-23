package main

import (
	"log"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
)

func main() {
	cfg := config.Load()
	app := server.New()

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
