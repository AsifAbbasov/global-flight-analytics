package config

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/clientidentity"
)

func TestLoadTrustedProxyConfigUsesDirectConnectionByDefault(
	t *testing.T,
) {
	clearTrustedProxyEnvironment(
		t,
	)

	config, err := LoadTrustedProxyConfig()
	if err != nil {
		t.Fatalf(
			"load default trusted proxy configuration: %v",
			err,
		)
	}
	if config.ClientIPHeader != "" ||
		len(config.TrustedProxyRanges) != 0 {
		t.Fatalf(
			"expected disabled trusted proxy configuration, got %+v",
			config,
		)
	}
}

func TestLoadTrustedProxyConfigNormalizesConfiguredPolicy(
	t *testing.T,
) {
	clearTrustedProxyEnvironment(
		t,
	)
	t.Setenv(
		apiTrustedProxyRangesEnvironmentVariable,
		" 192.0.2.10,192.0.2.10/32,10.0.0.0/8 ",
	)
	t.Setenv(
		apiClientIPHeaderEnvironmentVariable,
		"x-real-ip",
	)

	config, err := LoadTrustedProxyConfig()
	if err != nil {
		t.Fatalf(
			"load trusted proxy configuration: %v",
			err,
		)
	}
	if config.ClientIPHeader !=
		clientidentity.HeaderXRealIP {
		t.Fatalf(
			"unexpected header: %q",
			config.ClientIPHeader,
		)
	}
	if len(config.TrustedProxyRanges) != 2 ||
		config.TrustedProxyRanges[0] !=
			"192.0.2.10/32" ||
		config.TrustedProxyRanges[1] !=
			"10.0.0.0/8" {
		t.Fatalf(
			"unexpected ranges: %#v",
			config.TrustedProxyRanges,
		)
	}
}

func TestLoadTrustedProxyConfigDefaultsToForwardedFor(
	t *testing.T,
) {
	clearTrustedProxyEnvironment(
		t,
	)
	t.Setenv(
		apiTrustedProxyRangesEnvironmentVariable,
		"192.0.2.0/24",
	)

	config, err := LoadTrustedProxyConfig()
	if err != nil {
		t.Fatalf(
			"load trusted proxy configuration: %v",
			err,
		)
	}
	if config.ClientIPHeader !=
		clientidentity.HeaderXForwardedFor {
		t.Fatalf(
			"unexpected default header: %q",
			config.ClientIPHeader,
		)
	}
}

func TestLoadTrustedProxyConfigRejectsHeaderWithoutRanges(
	t *testing.T,
) {
	clearTrustedProxyEnvironment(
		t,
	)
	t.Setenv(
		apiClientIPHeaderEnvironmentVariable,
		clientidentity.HeaderXForwardedFor,
	)

	_, err := LoadTrustedProxyConfig()
	if !errors.Is(
		err,
		clientidentity.ErrTrustedProxyRangesRequired,
	) {
		t.Fatalf(
			"expected trusted proxy range error, got %v",
			err,
		)
	}
}

func clearTrustedProxyEnvironment(
	t *testing.T,
) {
	t.Helper()
	t.Setenv(
		apiTrustedProxyRangesEnvironmentVariable,
		"",
	)
	t.Setenv(
		apiClientIPHeaderEnvironmentVariable,
		"",
	)
}
