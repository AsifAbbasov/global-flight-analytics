package observability

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerdecision"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type ProviderDecisionSnapshotSource interface {
	Snapshot(
		provider providerpolicy.Provider,
	) (providerdecision.Snapshot, error)
}

type ProviderDecisionCollector struct {
	source    ProviderDecisionSnapshotSource
	providers []providerpolicy.Provider
}

func NewProviderDecisionCollector(
	source ProviderDecisionSnapshotSource,
	providers []providerpolicy.Provider,
) (*ProviderDecisionCollector, error) {
	if source == nil {
		return nil, fmt.Errorf("provider decision metrics source is required")
	}

	seen := make(map[providerpolicy.Provider]struct{}, len(providers))
	normalized := make([]providerpolicy.Provider, 0, len(providers))
	for _, provider := range providers {
		if strings.TrimSpace(string(provider)) == "" {
			continue
		}
		if _, exists := seen[provider]; exists {
			continue
		}
		seen[provider] = struct{}{}
		normalized = append(normalized, provider)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("at least one provider decision metrics provider is required")
	}
	sort.Slice(normalized, func(left int, right int) bool {
		return normalized[left] < normalized[right]
	})

	return &ProviderDecisionCollector{
		source:    source,
		providers: normalized,
	}, nil
}

func (
	collector *ProviderDecisionCollector,
) Name() string {
	return "provider_decisions"
}

func (
	collector *ProviderDecisionCollector,
) WritePrometheus(
	_ context.Context,
	builder *strings.Builder,
) error {
	if collector == nil || collector.source == nil {
		return fmt.Errorf("provider decision metrics collector is unavailable")
	}

	decisionMetric := metricNamespace + "_provider_budget_decisions_total"
	writeHelpAndType(
		builder,
		decisionMetric,
		"Provider budget decisions partitioned by bounded provider and decision values.",
		"counter",
	)
	fallbackMetric := metricNamespace + "_provider_fallback_decisions_total"
	writeHelpAndType(
		builder,
		fallbackMetric,
		"Provider fallback decisions partitioned by bounded primary provider and outcome values.",
		"counter",
	)

	for _, provider := range collector.providers {
		snapshot, err := collector.source.Snapshot(provider)
		if err != nil {
			if errors.Is(err, providerdecision.ErrNoDecisionEvidence) {
				writeProviderDecisionZeros(builder, decisionMetric, fallbackMetric, provider)
				continue
			}
			return err
		}

		providerLabel := normalizeProvider(string(provider))
		for _, current := range []struct {
			decision string
			value    int64
		}{
			{decision: "allowed", value: snapshot.AllowedTotal},
			{decision: "denied", value: snapshot.DeniedTotal},
		} {
			fmt.Fprintf(
				builder,
				"%s%s %d\n",
				decisionMetric,
				formatLabels([]label{
					{name: "provider", value: providerLabel},
					{name: "decision", value: current.decision},
				}),
				current.value,
			)
		}

		for _, current := range []struct {
			outcome string
			value   int64
		}{
			{outcome: "primary_selected", value: snapshot.PrimarySelectedTotal},
			{outcome: "fallback_selected", value: snapshot.FallbackSelectedTotal},
			{outcome: "no_provider_available", value: snapshot.NoProviderAvailableTotal},
			{outcome: "terminal_failure", value: snapshot.TerminalFailureTotal},
		} {
			fmt.Fprintf(
				builder,
				"%s%s %d\n",
				fallbackMetric,
				formatLabels([]label{
					{name: "provider", value: providerLabel},
					{name: "outcome", value: current.outcome},
				}),
				current.value,
			)
		}
	}

	return nil
}

func writeProviderDecisionZeros(
	builder *strings.Builder,
	decisionMetric string,
	fallbackMetric string,
	provider providerpolicy.Provider,
) {
	providerLabel := normalizeProvider(string(provider))
	for _, decision := range []string{"allowed", "denied"} {
		fmt.Fprintf(
			builder,
			"%s%s 0\n",
			decisionMetric,
			formatLabels([]label{
				{name: "provider", value: providerLabel},
				{name: "decision", value: decision},
			}),
		)
	}
	for _, outcome := range []string{
		"fallback_selected",
		"no_provider_available",
		"primary_selected",
		"terminal_failure",
	} {
		fmt.Fprintf(
			builder,
			"%s%s 0\n",
			fallbackMetric,
			formatLabels([]label{
				{name: "provider", value: providerLabel},
				{name: "outcome", value: outcome},
			}),
		)
	}
}
