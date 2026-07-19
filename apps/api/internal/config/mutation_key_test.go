package config

import (
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
)

func TestLoadAPIProtectionConfigParsesMutationDigest(
	t *testing.T,
) {
	clearAPIProtectionEnvironment(t)

	digest := internalapikey.DigestCandidate(
		strings.Repeat(
			"configured-mutation-key-",
			2,
		),
	)
	t.Setenv(
		apiMutationKeySHA256EnvironmentVariable,
		digest.Hex(),
	)

	config, err :=
		loadAPIProtectionConfig()
	if err != nil {
		t.Fatalf(
			"load protection config: %v",
			err,
		)
	}
	if !config.MutationKeyConfigured {
		t.Fatal(
			"mutation key must be configured",
		)
	}
	if config.MutationKeyDigest != digest {
		t.Fatalf(
			"digest = %x, want %x",
			config.MutationKeyDigest,
			digest,
		)
	}
}

func TestLoadAPIProtectionConfigAllowsMissingMutationDigestWithoutDatabase(
	t *testing.T,
) {
	clearAPIProtectionEnvironment(t)

	config, err :=
		loadAPIProtectionConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.MutationKeyConfigured {
		t.Fatal(
			"mutation key must remain unconfigured",
		)
	}
	if !config.MutationKeyDigest.IsZero() {
		t.Fatal(
			"missing mutation key must use zero digest",
		)
	}
}

func TestLoadAPIProtectionConfigRejectsInvalidMutationDigest(
	t *testing.T,
) {
	clearAPIProtectionEnvironment(t)
	t.Setenv(
		apiMutationKeySHA256EnvironmentVariable,
		strings.Repeat("z", 64),
	)

	_, err := loadAPIProtectionConfig()
	if err == nil {
		t.Fatal(
			"expected invalid digest error",
		)
	}
	if !strings.Contains(
		err.Error(),
		apiMutationKeySHA256EnvironmentVariable,
	) {
		t.Fatalf(
			"error = %q",
			err,
		)
	}
}

func TestLoadServerConfigRequiresMutationDigestWithDatabase(
	t *testing.T,
) {
	t.Setenv(
		apiPortEnvironmentVariable,
		"8080",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		"postgresql://user:password@host/database",
	)
	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"3s",
	)
	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		"5s",
	)
	clearAPIProtectionEnvironment(t)

	_, err := LoadServerConfig()
	if err == nil {
		t.Fatal(
			"expected missing mutation digest error",
		)
	}
	if !strings.Contains(
		err.Error(),
		apiMutationKeySHA256EnvironmentVariable+
			" is required when DATABASE_URL is configured",
	) {
		t.Fatalf(
			"error = %q",
			err,
		)
	}
}
