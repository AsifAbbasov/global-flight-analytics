package config

import (
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
)

const observabilityTestKey = "0123456789abcdef0123456789abcdef"

func TestLoadObservabilityConfigAllowsDisabledMetrics(
	t *testing.T,
) {
	clearObservabilityEnvironment(t)

	config, err := LoadObservabilityConfig()
	if err != nil {
		t.Fatalf("load observability config: %v", err)
	}
	if config.MetricsKeyConfigured ||
		!config.MetricsKeyDigest.IsZero() ||
		config.IngestMetricsAddress != "" {
		t.Fatalf("unexpected disabled config: %+v", config)
	}
}

func TestLoadObservabilityConfigUsesProtectedIngestAddress(
	t *testing.T,
) {
	clearObservabilityEnvironment(t)
	digest := internalapikey.DigestCandidate(observabilityTestKey)
	t.Setenv(metricsKeySHA256EnvironmentVariable, digest.Hex())
	t.Setenv(ingestMetricsAddressEnvironmentVariable, ":9091")

	config, err := LoadObservabilityConfig()
	if err != nil {
		t.Fatalf("load observability config: %v", err)
	}
	if !config.MetricsKeyConfigured ||
		config.MetricsKeyDigest != digest ||
		config.IngestMetricsAddress != ":9091" {
		t.Fatalf("unexpected observability config: %+v", config)
	}
}

func TestLoadObservabilityConfigRejectsUnprotectedIngestAddress(
	t *testing.T,
) {
	clearObservabilityEnvironment(t)
	t.Setenv(ingestMetricsAddressEnvironmentVariable, ":9091")

	config, err := LoadObservabilityConfig()
	if err == nil {
		t.Fatal("expected unprotected metrics address error")
	}
	if config != (ObservabilityConfig{}) {
		t.Fatalf("expected zero config, got %+v", config)
	}
	if !strings.Contains(err.Error(), metricsKeySHA256EnvironmentVariable) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadObservabilityConfigRejectsInvalidAddress(
	t *testing.T,
) {
	clearObservabilityEnvironment(t)
	digest := internalapikey.DigestCandidate(observabilityTestKey)
	t.Setenv(metricsKeySHA256EnvironmentVariable, digest.Hex())
	t.Setenv(ingestMetricsAddressEnvironmentVariable, "invalid")

	_, err := LoadObservabilityConfig()
	if err == nil {
		t.Fatal("expected invalid metrics address error")
	}
}

func clearObservabilityEnvironment(
	t *testing.T,
) {
	t.Helper()
	t.Setenv(metricsKeySHA256EnvironmentVariable, "")
	t.Setenv(ingestMetricsAddressEnvironmentVariable, "")
}
