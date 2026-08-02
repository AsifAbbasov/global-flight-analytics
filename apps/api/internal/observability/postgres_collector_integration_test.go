package observability_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"
)

func TestPostgresCollectorExposesOperationalState(
	t *testing.T,
) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	pool, err := database.NewPostgresPool(databaseURL, 10*time.Second)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer pool.Close()

	collector, err := observability.NewPostgresCollector(pool)
	if err != nil {
		t.Fatalf("create collector: %v", err)
	}
	registry := observability.NewRegistry(observability.BuildInfo{})
	if err := registry.RegisterCollector(collector); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	output := registry.Render(context.Background())
	for _, expected := range []string{
		"global_flight_analytics_postgres_pool_total_connections",
		`global_flight_analytics_ingestion_runs{status="running"}`,
		`global_flight_analytics_reconciliation_tasks{status="pending"}`,
		"global_flight_analytics_reconciliation_oldest_pending_age_seconds",
		`global_flight_analytics_collector_last_scrape_success{collector="postgres"} 1`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q\n%s", expected, output)
		}
	}
}
