package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIPort            = 8080
	defaultHealthcheckPath    = "/api/v1/ready"
	defaultHealthcheckTimeout = 2 * time.Second
)

var (
	errHealthcheckURLInvalid = errors.New(
		"healthcheck URL is invalid",
	)
	errHealthcheckTimeoutInvalid = errors.New(
		"healthcheck timeout must be greater than zero",
	)
	errAPIPortInvalid = errors.New(
		"API port must be between 1 and 65535",
	)
)

type healthcheckConfig struct {
	URL     string
	Timeout time.Duration
}

type HTTPClient interface {
	Do(
		request *http.Request,
	) (*http.Response, error)
}

func main() {
	config, err := loadHealthcheckConfig()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"load healthcheck configuration: %v\n",
			err,
		)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: config.Timeout,
		CheckRedirect: func(
			*http.Request,
			[]*http.Request,
		) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.Timeout,
	)
	defer cancel()

	if err := checkHealth(
		ctx,
		client,
		config.URL,
	); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"healthcheck failed: %v\n",
			err,
		)
		os.Exit(1)
	}
}

func loadHealthcheckConfig() (
	healthcheckConfig,
	error,
) {
	endpoint := strings.TrimSpace(
		os.Getenv(
			"HEALTHCHECK_URL",
		),
	)

	if endpoint == "" {
		port, err := loadAPIPort()
		if err != nil {
			return healthcheckConfig{}, err
		}

		endpoint = fmt.Sprintf(
			"http://127.0.0.1:%d%s",
			port,
			defaultHealthcheckPath,
		)
	}

	if err := validateHealthcheckURL(
		endpoint,
	); err != nil {
		return healthcheckConfig{}, err
	}

	timeout := defaultHealthcheckTimeout
	timeoutValue := strings.TrimSpace(
		os.Getenv(
			"HEALTHCHECK_TIMEOUT",
		),
	)

	if timeoutValue != "" {
		parsedTimeout, err := time.ParseDuration(
			timeoutValue,
		)
		if err != nil {
			return healthcheckConfig{},
				fmt.Errorf(
					"parse HEALTHCHECK_TIMEOUT: %w",
					err,
				)
		}

		if parsedTimeout <= 0 {
			return healthcheckConfig{},
				errHealthcheckTimeoutInvalid
		}

		timeout = parsedTimeout
	}

	return healthcheckConfig{
		URL:     endpoint,
		Timeout: timeout,
	}, nil
}

func loadAPIPort() (
	int,
	error,
) {
	portValue := strings.TrimSpace(
		os.Getenv(
			"API_PORT",
		),
	)

	if portValue == "" {
		return defaultAPIPort, nil
	}

	port, err := strconv.Atoi(
		portValue,
	)
	if err != nil {
		return 0,
			fmt.Errorf(
				"parse API_PORT: %w",
				err,
			)
	}

	if port < 1 || port > 65535 {
		return 0,
			errAPIPortInvalid
	}

	return port, nil
}

func validateHealthcheckURL(
	endpoint string,
) error {
	parsedURL, err := url.ParseRequestURI(
		endpoint,
	)
	if err != nil {
		return fmt.Errorf(
			"%w: %v",
			errHealthcheckURLInvalid,
			err,
		)
	}

	if parsedURL.Scheme != "http" &&
		parsedURL.Scheme != "https" {
		return errHealthcheckURLInvalid
	}

	if parsedURL.Host == "" {
		return errHealthcheckURLInvalid
	}

	return nil
}

func checkHealth(
	ctx context.Context,
	client HTTPClient,
	endpoint string,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if client == nil {
		return errors.New(
			"HTTP client is required",
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"create request: %w",
			err,
		)
	}

	request.Header.Set(
		"User-Agent",
		"global-flight-analytics-healthcheck",
	)

	response, err := client.Do(
		request,
	)
	if err != nil {
		return fmt.Errorf(
			"execute request: %w",
			err,
		)
	}
	defer response.Body.Close()

	_, readErr := io.Copy(
		io.Discard,
		io.LimitReader(
			response.Body,
			4096,
		),
	)
	if readErr != nil {
		return fmt.Errorf(
			"read response body: %w",
			readErr,
		)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected status: %d",
			response.StatusCode,
		)
	}

	return nil
}
