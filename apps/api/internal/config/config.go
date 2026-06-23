package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	databaseURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("API_PORT")

	if port == "" {
		port = "8080"
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}
}
