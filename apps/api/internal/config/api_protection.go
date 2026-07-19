package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
)

const (
	apiAllowedOriginsEnvironmentVariable    = "API_ALLOWED_ORIGINS"
	apiBodyLimitBytesEnvironmentVariable    = "API_BODY_LIMIT_BYTES"
	apiReadTimeoutEnvironmentVariable       = "API_READ_TIMEOUT"
	apiWriteTimeoutEnvironmentVariable      = "API_WRITE_TIMEOUT"
	apiIdleTimeoutEnvironmentVariable       = "API_IDLE_TIMEOUT"
	apiRateLimitMaxEnvironmentVariable      = "API_RATE_LIMIT_MAX"
	apiRateLimitWindowEnvironmentVariable   = "API_RATE_LIMIT_WINDOW"
	apiMutationKeySHA256EnvironmentVariable = "API_MUTATION_KEY_SHA256"

	defaultAPIAllowedOrigins  = "http://localhost:3000,http://localhost:3001"
	defaultAPIBodyLimitBytes  = 1024 * 1024
	defaultAPIReadTimeout     = 10 * time.Second
	defaultAPIWriteTimeout    = 15 * time.Second
	defaultAPIIdleTimeout     = 60 * time.Second
	defaultAPIRateLimitMax    = 120
	defaultAPIRateLimitWindow = time.Minute
)

func loadAPIProtectionConfig() (
	APIProtectionConfig,
	error,
) {
	allowedOrigins, err := optionalAllowedOriginsEnvironmentVariable(
		apiAllowedOriginsEnvironmentVariable,
		defaultAPIAllowedOrigins,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	bodyLimitBytes, err := optionalPositiveIntegerEnvironmentVariable(
		apiBodyLimitBytesEnvironmentVariable,
		defaultAPIBodyLimitBytes,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	readTimeout, err := optionalPositiveDurationEnvironmentVariable(
		apiReadTimeoutEnvironmentVariable,
		defaultAPIReadTimeout,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	writeTimeout, err := optionalPositiveDurationEnvironmentVariable(
		apiWriteTimeoutEnvironmentVariable,
		defaultAPIWriteTimeout,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	idleTimeout, err := optionalPositiveDurationEnvironmentVariable(
		apiIdleTimeoutEnvironmentVariable,
		defaultAPIIdleTimeout,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	rateLimitMax, err := optionalPositiveIntegerEnvironmentVariable(
		apiRateLimitMaxEnvironmentVariable,
		defaultAPIRateLimitMax,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	rateLimitWindow, err := optionalPositiveDurationEnvironmentVariable(
		apiRateLimitWindowEnvironmentVariable,
		defaultAPIRateLimitWindow,
	)
	if err != nil {
		return APIProtectionConfig{}, err
	}

	mutationKeyDigest, mutationKeyConfigured, err :=
		optionalMutationKeyDigestEnvironmentVariable()
	if err != nil {
		return APIProtectionConfig{}, err
	}

	return APIProtectionConfig{
		AllowedOrigins:        allowedOrigins,
		BodyLimitBytes:        bodyLimitBytes,
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		IdleTimeout:           idleTimeout,
		RateLimitMax:          rateLimitMax,
		RateLimitWindow:       rateLimitWindow,
		MutationKeyDigest:     mutationKeyDigest,
		MutationKeyConfigured: mutationKeyConfigured,
	}, nil
}

func optionalMutationKeyDigestEnvironmentVariable() (
	internalapikey.Digest,
	bool,
	error,
) {
	value, exists := os.LookupEnv(
		apiMutationKeySHA256EnvironmentVariable,
	)
	if !exists || value == "" {
		return internalapikey.Digest{},
			false,
			nil
	}

	digest, err := internalapikey.ParseDigestHex(
		value,
	)
	if err != nil {
		return internalapikey.Digest{},
			false,
			fmt.Errorf(
				"load %s: %w",
				apiMutationKeySHA256EnvironmentVariable,
				err,
			)
	}

	return digest, true, nil
}

func optionalAllowedOriginsEnvironmentVariable(
	name string,
	defaultValue string,
) (string, error) {
	value := optionalTrimmedStringEnvironmentVariable(
		name,
	)
	if value == "" {
		value = defaultValue
	}

	parts := strings.Split(
		value,
		",",
	)

	seen := make(
		map[string]struct{},
		len(parts),
	)

	normalized := make(
		[]string,
		0,
		len(parts),
	)

	for _, part := range parts {
		origin := strings.TrimSpace(
			part,
		)
		if origin == "" {
			continue
		}

		if origin == "*" {
			return "", fmt.Errorf(
				"%s must not contain wildcard origins",
				name,
			)
		}

		parsed, err := url.Parse(
			origin,
		)
		if err != nil ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" ||
			parsed.User != nil ||
			parsed.Path != "" ||
			parsed.RawQuery != "" ||
			parsed.Fragment != "" {
			return "", fmt.Errorf(
				"%s contains invalid origin %q",
				name,
				origin,
			)
		}

		if _, exists := seen[origin]; exists {
			continue
		}

		seen[origin] = struct{}{}
		normalized = append(
			normalized,
			origin,
		)
	}

	if len(normalized) == 0 {
		return "", fmt.Errorf(
			"%s must contain at least one origin",
			name,
		)
	}

	return strings.Join(
		normalized,
		",",
	), nil
}

func optionalPositiveIntegerEnvironmentVariable(
	name string,
	defaultValue int,
) (int, error) {
	value := optionalTrimmedStringEnvironmentVariable(
		name,
	)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(
		value,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %s as integer: %w",
			name,
			err,
		)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf(
			"%s must be greater than zero",
			name,
		)
	}

	return parsed, nil
}

func optionalPositiveDurationEnvironmentVariable(
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	value := optionalTrimmedStringEnvironmentVariable(
		name,
	)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := time.ParseDuration(
		value,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %s as duration: %w",
			name,
			err,
		)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf(
			"%s must be greater than zero",
			name,
		)
	}

	return parsed, nil
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
