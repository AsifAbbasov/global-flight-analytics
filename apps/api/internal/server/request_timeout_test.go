package server

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeProtectionConfigDefaultsRequestTimeoutToWriteTimeout(
	t *testing.T,
) {
	normalized, err := normalizeProtectionConfig(
		ProtectionConfig{},
	)
	if err != nil {
		t.Fatalf(
			"normalize protection configuration: %v",
			err,
		)
	}
	if normalized.RequestTimeout != normalized.WriteTimeout {
		t.Fatalf(
			"expected request timeout to default to write timeout %s, got %s",
			normalized.WriteTimeout,
			normalized.RequestTimeout,
		)
	}
}

func TestNormalizeProtectionConfigRejectsRequestTimeoutAboveWriteTimeout(
	t *testing.T,
) {
	normalized, err := normalizeProtectionConfig(
		ProtectionConfig{
			RequestTimeout: 20 * time.Second,
			WriteTimeout:   15 * time.Second,
		},
	)
	if err == nil {
		t.Fatal(
			"expected request timeout relationship error",
		)
	}
	if normalized.RequestTimeout != 0 ||
		normalized.WriteTimeout != 0 ||
		normalized.ReadTimeout != 0 ||
		normalized.IdleTimeout != 0 ||
		normalized.BodyLimitBytes != 0 ||
		normalized.AllowedOrigins != "" ||
		len(normalized.TrustedProxyRanges) != 0 {
		t.Fatalf(
			"expected zero protection configuration, got %+v",
			normalized,
		)
	}
	if !strings.Contains(
		err.Error(),
		"request timeout must not exceed write timeout",
	) {
		t.Fatalf(
			"unexpected request timeout error: %v",
			err,
		)
	}
}
