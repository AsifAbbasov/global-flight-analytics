package integrations

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var (
	ErrHTTPClientBaseURLRequired = errors.New(
		"http client base url is required",
	)

	ErrHTTPClientBaseURLInvalid = errors.New(
		"http client base url must be an absolute http or https url",
	)

	ErrHTTPClientTimeoutInvalid = errors.New(
		"http client timeout must be greater than zero",
	)

	ErrHTTPClientUserAgentRequired = errors.New(
		"http client user agent is required",
	)
)

type HTTPClientConfig struct {
	BaseURL   string
	Timeout   time.Duration
	UserAgent string
}

func (config HTTPClientConfig) Validate() error {
	baseURL := strings.TrimSpace(
		config.BaseURL,
	)
	if baseURL == "" {
		return ErrHTTPClientBaseURLRequired
	}

	parsedURL, err := url.Parse(
		baseURL,
	)
	if err != nil ||
		parsedURL.Host == "" ||
		(parsedURL.Scheme != "http" &&
			parsedURL.Scheme != "https") {
		return fmt.Errorf(
			"%w: %q",
			ErrHTTPClientBaseURLInvalid,
			config.BaseURL,
		)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrHTTPClientTimeoutInvalid,
			config.Timeout,
		)
	}

	if strings.TrimSpace(
		config.UserAgent,
	) == "" {
		return ErrHTTPClientUserAgentRequired
	}

	return nil
}
