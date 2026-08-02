package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
)

const (
	metricsKeySHA256EnvironmentVariable     = "METRICS_KEY_SHA256"
	ingestMetricsAddressEnvironmentVariable = "INGEST_METRICS_ADDRESS"
)

type ObservabilityConfig struct {
	MetricsKeyDigest     internalapikey.Digest
	MetricsKeyConfigured bool
	IngestMetricsAddress string
}

func LoadObservabilityConfig() (
	ObservabilityConfig,
	error,
) {
	keyDigest, keyConfigured, err := optionalMetricsKeyDigest()
	if err != nil {
		return ObservabilityConfig{}, err
	}

	ingestAddress := optionalTrimmedStringEnvironmentVariable(
		ingestMetricsAddressEnvironmentVariable,
	)
	if ingestAddress != "" {
		if !keyConfigured {
			return ObservabilityConfig{}, fmt.Errorf(
				"%s requires %s",
				ingestMetricsAddressEnvironmentVariable,
				metricsKeySHA256EnvironmentVariable,
			)
		}
		if err := validateMetricsAddress(ingestAddress); err != nil {
			return ObservabilityConfig{}, fmt.Errorf(
				"validate %s: %w",
				ingestMetricsAddressEnvironmentVariable,
				err,
			)
		}
	}

	return ObservabilityConfig{
		MetricsKeyDigest:     keyDigest,
		MetricsKeyConfigured: keyConfigured,
		IngestMetricsAddress: ingestAddress,
	}, nil
}

func optionalMetricsKeyDigest() (
	internalapikey.Digest,
	bool,
	error,
) {
	value, exists := os.LookupEnv(metricsKeySHA256EnvironmentVariable)
	if !exists || value == "" {
		return internalapikey.Digest{}, false, nil
	}

	digest, err := internalapikey.ParseDigestHex(value)
	if err != nil {
		return internalapikey.Digest{},
			false,
			fmt.Errorf(
				"load %s: %w",
				metricsKeySHA256EnvironmentVariable,
				err,
			)
	}
	return digest, true, nil
}

func validateMetricsAddress(
	value string,
) error {
	host, portValue, err := net.SplitHostPort(strings.TrimSpace(value))
	if err != nil {
		return fmt.Errorf("address must use host:port form: %w", err)
	}
	_ = host

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return fmt.Errorf("metrics port must be numeric: %w", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("metrics port must be between 1 and 65535")
	}
	return nil
}
