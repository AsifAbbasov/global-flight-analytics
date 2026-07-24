package config

import (
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/clientidentity"
)

const (
	apiTrustedProxyRangesEnvironmentVariable = "API_TRUSTED_PROXY_RANGES"
	apiClientIPHeaderEnvironmentVariable     = "API_CLIENT_IP_HEADER"
)

type TrustedProxyConfig struct {
	ClientIPHeader     string
	TrustedProxyRanges []string
}

func LoadTrustedProxyConfig() (
	TrustedProxyConfig,
	error,
) {
	rawRanges := optionalTrimmedStringEnvironmentVariable(
		apiTrustedProxyRangesEnvironmentVariable,
	)
	rawHeader := optionalTrimmedStringEnvironmentVariable(
		apiClientIPHeaderEnvironmentVariable,
	)

	if rawRanges == "" && rawHeader == "" {
		return TrustedProxyConfig{}, nil
	}

	ranges := make(
		[]string,
		0,
	)
	for _, part := range strings.Split(
		rawRanges,
		",",
	) {
		normalized := strings.TrimSpace(
			part,
		)
		if normalized == "" {
			continue
		}
		ranges = append(
			ranges,
			normalized,
		)
	}

	policy, err := clientidentity.NewPolicy(
		clientidentity.Config{
			Header:             rawHeader,
			TrustedProxyRanges: ranges,
		},
	)
	if err != nil {
		return TrustedProxyConfig{},
			fmt.Errorf(
				"load trusted proxy client identity configuration: %w",
				err,
			)
	}

	return TrustedProxyConfig{
		ClientIPHeader:     policy.Header(),
		TrustedProxyRanges: policy.TrustedProxyRanges(),
	}, nil
}
