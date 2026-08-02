package observability

import (
	"context"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerdecision"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestProviderDecisionCollectorExportsBoundedFallbackMetrics(
	t *testing.T,
) {
	source := providerdecision.New(nil)
	source.RecordBudgetDecision(
		providerpolicy.ProviderOpenSky,
		"sensitive-request-key",
		"sensitive-publication-id",
		providerbudget.Decision{
			Provider: providerpolicy.ProviderOpenSky,
			Allowed:  true,
		},
	)
	source.RecordFallbackDecision(
		providerfallback.Decision{
			PrimaryProvider: providerpolicy.ProviderOpenSky,
			Outcome:         providerfallback.OutcomeFallbackSelected,
		},
	)

	collector, err := NewProviderDecisionCollector(
		source,
		[]providerpolicy.Provider{providerpolicy.ProviderOpenSky},
	)
	if err != nil {
		t.Fatalf("create provider decision collector: %v", err)
	}
	registry := NewRegistry(BuildInfo{})
	if err := registry.RegisterCollector(collector); err != nil {
		t.Fatalf("register provider decision collector: %v", err)
	}

	output := registry.Render(context.Background())
	for _, expected := range []string{
		`global_flight_analytics_provider_budget_decisions_total{provider="opensky",decision="allowed"} 1`,
		`global_flight_analytics_provider_fallback_decisions_total{provider="opensky",outcome="fallback_selected"} 1`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q\n%s", expected, output)
		}
	}
	if strings.Contains(output, "sensitive-request-key") ||
		strings.Contains(output, "sensitive-publication-id") {
		t.Fatalf("high-cardinality provider decision evidence leaked\n%s", output)
	}
}
