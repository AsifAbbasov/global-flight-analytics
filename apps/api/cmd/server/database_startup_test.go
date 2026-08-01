package main

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
)

func TestOpenServerDatabaseRequiresContext(
	t *testing.T,
) {
	pool, err := openServerDatabase(
		nil,
		config.ServerConfig{},
		slog.Default(),
	)

	if pool != nil {
		pool.Close()
		t.Fatal(
			"expected nil postgres pool when server context is missing",
		)
	}
	if !errors.Is(
		err,
		errServerContextRequired,
	) {
		t.Fatalf(
			"expected server context requirement, got %v",
			err,
		)
	}
}

func TestOpenServerDatabaseAllowsDatabaseOptionalMode(
	t *testing.T,
) {
	pool, err := openServerDatabase(
		context.Background(),
		config.ServerConfig{},
		slog.Default(),
	)

	if err != nil {
		t.Fatalf(
			"expected database-optional startup, got %v",
			err,
		)
	}
	if pool != nil {
		pool.Close()
		t.Fatal(
			"expected nil postgres pool in database-optional mode",
		)
	}
}

func TestOpenServerDatabasePreservesCanceledStartupContext(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	pool, err := openServerDatabase(
		ctx,
		config.ServerConfig{
			Database: &config.PostgresConfig{
				URL:            "postgres://postgres:postgres@127.0.0.1:5432/global_flight_analytics?sslmode=disable",
				ConnectTimeout: time.Second,
			},
		},
		slog.Default(),
	)

	if pool != nil {
		pool.Close()
		t.Fatal(
			"expected nil postgres pool after startup cancellation",
		)
	}
	if !errors.Is(
		err,
		errServerDatabaseConnection,
	) {
		t.Fatalf(
			"expected server database classification, got %v",
			err,
		)
	}
	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected preserved startup cancellation, got %v",
			err,
		)
	}
	if code := serverFailureCode(err); code != "SERVER_DATABASE_CONNECTION_FAILED" {
		t.Fatalf(
			"expected database failure code, got %q",
			code,
		)
	}
}
