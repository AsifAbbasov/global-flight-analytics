package opensky

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

const maxStatesResponseBytes int64 = 16 << 20

var (
	ErrPollingTooSoon          = errors.New("OpenSky state request is earlier than the configured polling interval")
	ErrRequestTimeInFuture     = errors.New("OpenSky request time must not be in the future")
	ErrAuthenticatedTimeWindow = errors.New("authenticated OpenSky state requests may look back at most one hour")
	ErrAnonymousHistoricalTime = errors.New("anonymous OpenSky state requests cannot request historical time")
)

type RateLimit struct {
	Remaining         *int64 `json:"remaining"`
	RetryAfterSeconds *int64 `json:"retry_after_seconds"`
}

type StatesRequest struct {
	Time        time.Time
	ICAO24      []string
	BoundingBox *BoundingBox
	Extended    bool
}

type StatesResult struct {
	ProviderTime  time.Time     `json:"provider_time"`
	States        []StateVector `json:"states"`
	RateLimit     RateLimit     `json:"rate_limit"`
	Authenticated bool          `json:"authenticated"`
}

type Client struct {
	config           Config
	tokenManager     *TokenManager
	responseObserver integrationcommon.ProviderResponseObserver

	pollMu            sync.Mutex
	lastStatesRequest time.Time
}

func NewClient(config Config) (*Client, error) {
	return NewClientWithResponseObserver(config, nil)
}

func NewClientWithResponseObserver(
	config Config,
	responseObserver integrationcommon.ProviderResponseObserver,
) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	client := &Client{
		config:           config,
		responseObserver: responseObserver,
	}
	if config.Authenticated() {
		manager, err := NewTokenManager(
			config.HTTPClient,
			config.TokenURL,
			config.ClientID,
			config.ClientSecret,
		)
		if err != nil {
			return nil, err
		}
		client.tokenManager = manager
	}
	return client, nil
}

func (client *Client) GetStates(
	ctx context.Context,
	input StatesRequest,
) (StatesResult, error) {
	if err := client.validateStatesRequest(input); err != nil {
		return StatesResult{}, err
	}
	if err := client.acquirePollSlot(); err != nil {
		return StatesResult{}, err
	}

	requestURL, err := client.statesURL(input)
	if err != nil {
		return StatesResult{}, err
	}
	response, err := client.do(ctx, http.MethodGet, requestURL)
	if err != nil {
		return StatesResult{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized && client.tokenManager != nil {
		client.tokenManager.Invalidate()
		response, err = client.do(ctx, http.MethodGet, requestURL)
		if err != nil {
			return StatesResult{}, err
		}
		defer response.Body.Close()
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		statusErr := integrationcommon.ProviderStatusError(response.StatusCode)
		if statusErr == nil {
			statusErr = fmt.Errorf("unexpected provider status %d", response.StatusCode)
		}
		return StatesResult{}, fmt.Errorf(
			"OpenSky states request failed: %w",
			statusErr,
		)
	}

	var payload StateResponse
	if err := integrationcommon.DecodeJSONHTTPResponse(
		response,
		sourceName,
		maxStatesResponseBytes,
		&payload,
	); err != nil {
		return StatesResult{}, fmt.Errorf("decode OpenSky states response: %w", err)
	}
	states, err := payload.ParseStates()
	if err != nil {
		return StatesResult{}, err
	}
	return StatesResult{
		ProviderTime:  time.Unix(payload.Time, 0).UTC(),
		States:        states,
		RateLimit:     parseRateLimit(response.Header),
		Authenticated: client.tokenManager != nil,
	}, nil
}

func (client *Client) validateStatesRequest(input StatesRequest) error {
	now := time.Now().UTC()
	if !input.Time.IsZero() {
		requested := input.Time.UTC()
		if requested.After(now) {
			return ErrRequestTimeInFuture
		}
		if client.tokenManager == nil {
			return ErrAnonymousHistoricalTime
		}
		if requested.Before(now.Add(-time.Hour)) {
			return ErrAuthenticatedTimeWindow
		}
	}
	if input.BoundingBox != nil {
		if err := input.BoundingBox.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) acquirePollSlot() error {
	client.pollMu.Lock()
	defer client.pollMu.Unlock()
	now := time.Now().UTC()
	if !client.lastStatesRequest.IsZero() &&
		now.Sub(client.lastStatesRequest) < client.config.PollingInterval {
		return ErrPollingTooSoon
	}
	client.lastStatesRequest = now
	return nil
}

func (client *Client) statesURL(input StatesRequest) (string, error) {
	base, err := url.Parse(strings.TrimRight(client.config.BaseURL, "/") + "/states/all")
	if err != nil {
		return "", fmt.Errorf("parse OpenSky states URL: %w", err)
	}
	query := base.Query()
	if input.Extended {
		query.Set("extended", "1")
	}
	if !input.Time.IsZero() {
		query.Set("time", strconv.FormatInt(input.Time.UTC().Unix(), 10))
	}
	for _, raw := range input.ICAO24 {
		icao24 := strings.ToLower(strings.TrimSpace(raw))
		if icao24 != "" {
			query.Add("icao24", icao24)
		}
	}
	if input.BoundingBox != nil {
		if err := input.BoundingBox.AddTo(query); err != nil {
			return "", err
		}
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func (client *Client) do(
	ctx context.Context,
	method string,
	requestURL string,
) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build OpenSky request: %w", err)
	}
	if strings.TrimSpace(client.config.UserAgent) != "" {
		request.Header.Set("User-Agent", client.config.UserAgent)
	}
	if client.tokenManager != nil {
		token, err := client.tokenManager.Token(ctx)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Authorization", "Bearer "+token)
	}

	startedAt := time.Now()
	response, err := client.config.HTTPClient.Do(request)
	latency := time.Since(startedAt)
	if err != nil {
		observer, supported := client.responseObserver.(integrationcommon.ProviderTransportFailureObserver)
		if supported {
			_ = observer.ObserveProviderTransportFailure(
				sourceName,
				err,
				latency,
			)
		}
		return nil, fmt.Errorf("execute OpenSky request: %w", err)
	}

	if client.responseObserver != nil {
		_ = client.responseObserver.ObserveProviderResponse(
			sourceName,
			response.StatusCode,
			response.Header.Clone(),
			latency,
		)
	}

	return response, nil
}

func parseRateLimit(headers http.Header) RateLimit {
	return RateLimit{
		Remaining:         parseOptionalHeaderInt64(headers.Get("X-Rate-Limit-Remaining")),
		RetryAfterSeconds: parseOptionalHeaderInt64(headers.Get("X-Rate-Limit-Retry-After-Seconds")),
	}
}

func parseOptionalHeaderInt64(value string) *int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}
