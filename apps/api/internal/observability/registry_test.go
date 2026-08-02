package observability

import (
	"context"
	"strings"
	"testing"
	"time"
)

type staticCollector struct {
	name    string
	content string
	err     error
}

func (collector staticCollector) Name() string {
	return collector.name
}

func (collector staticCollector) WritePrometheus(
	context.Context,
	*strings.Builder,
) error {
	return collector.err
}

func TestRegistryRendersBoundedPrometheusMetrics(
	t *testing.T,
) {
	registry := NewRegistry(
		BuildInfo{
			Version:  "1.2.3",
			Revision: "abc123",
		},
	)

	registry.BeginHTTPRequest()
	registry.FinishHTTPRequest(
		"GET",
		"/api/v1/aircraft/:icao24",
		200,
		150*time.Millisecond,
	)
	registry.ObserveProviderRequest(
		"open_meteo",
		"success",
		250*time.Millisecond,
	)
	registry.ObserveIngestionCycle(
		"failed",
		2*time.Second,
		3,
		5*time.Second,
	)

	output := registry.Render(context.Background())
	for _, expected := range []string{
		`global_flight_analytics_build_info{version="1.2.3",revision="abc123"} 1`,
		`global_flight_analytics_http_requests_total{method="GET",route="/api/v1/aircraft/:icao24",status_class="2xx"} 1`,
		`global_flight_analytics_provider_requests_total{provider="open_meteo",outcome="success"} 1`,
		`global_flight_analytics_ingestion_cycles_total{result="failed"} 1`,
		`global_flight_analytics_ingestion_consecutive_failures 3`,
		`global_flight_analytics_ingestion_next_delay_seconds 5`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q\n%s", expected, output)
		}
	}

	for _, forbidden := range []string{
		"request_id=",
		"icao24=",
		"ip=",
		"error=",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("unexpected high-cardinality or sensitive label %q\n%s", forbidden, output)
		}
	}
}

func TestRegistryRejectsRawOrUnboundedRouteValues(
	t *testing.T,
) {
	registry := NewRegistry(BuildInfo{})
	registry.BeginHTTPRequest()
	registry.FinishHTTPRequest(
		"GET",
		"/api/v1/aircraft/abc123?secret=value",
		200,
		time.Millisecond,
	)

	output := registry.Render(context.Background())
	if !strings.Contains(output, `route="unmatched"`) {
		t.Fatalf("expected unsafe route to be normalized\n%s", output)
	}
	if strings.Contains(output, "abc123") || strings.Contains(output, "secret") {
		t.Fatalf("raw request data leaked into metrics\n%s", output)
	}
}

func TestRegistryTracksCollectorFailureWithoutExposingErrorText(
	t *testing.T,
) {
	registry := NewRegistry(BuildInfo{})
	if err := registry.RegisterCollector(
		staticCollector{
			name: "postgres",
			err:  context.DeadlineExceeded,
		},
	); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	output := registry.Render(context.Background())
	if !strings.Contains(
		output,
		`global_flight_analytics_collector_errors_total{collector="postgres"} 1`,
	) {
		t.Fatalf("expected collector error counter\n%s", output)
	}
	if strings.Contains(output, context.DeadlineExceeded.Error()) {
		t.Fatalf("collector error text leaked into metrics\n%s", output)
	}
}
