package common

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrProviderRateLimited = errors.New(
		"provider rate limited",
	)

	ErrProviderUnauthorized = errors.New(
		"provider request unauthorized",
	)

	ErrProviderClient = errors.New(
		"provider returned client error",
	)

	ErrProviderServer = errors.New(
		"provider returned server error",
	)
)

func ClassifyProviderStatus(
	status int,
) error {
	switch {
	case status == http.StatusTooManyRequests:
		return ErrProviderRateLimited
	case status == http.StatusUnauthorized ||
		status == http.StatusForbidden:
		return ErrProviderUnauthorized
	case status >= http.StatusInternalServerError:
		return ErrProviderServer
	case status >= http.StatusBadRequest:
		return ErrProviderClient
	default:
		return nil
	}
}

func ProviderStatusError(
	status int,
) error {
	classified := ClassifyProviderStatus(
		status,
	)
	if classified == nil {
		return nil
	}

	return fmt.Errorf(
		"%w: status %d",
		classified,
		status,
	)
}
