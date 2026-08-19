package adsblol

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

const (
	BaseURL                     = "https://api.adsb.lol"
	EndpointByPoint             = "/v2/point/%f/%f/%d"
	maxStateResponseBytes int64 = 8 << 20
	sourceName                  = "adsb.lol"
)

type Client struct {
	baseURL          string
	httpClient       *http.Client
	userAgent        string
	responseObserver integrationcommon.ProviderResponseObserver
}

func NewClient(config integrationcommon.HTTPClientConfig) (*Client, error) {
	return NewClientWithResponseObserver(config, nil)
}

func NewClientWithResponseObserver(
	config integrationcommon.HTTPClientConfig,
	responseObserver integrationcommon.ProviderResponseObserver,
) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate adsb.lol http client config: %w", err)
	}
	return &Client{
		baseURL:          strings.TrimRight(strings.TrimSpace(config.BaseURL), "/"),
		httpClient:       &http.Client{Timeout: config.Timeout},
		userAgent:        strings.TrimSpace(config.UserAgent),
		responseObserver: responseObserver,
	}, nil
}

func (client *Client) GetByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) (*StateResponse, error) {
	if !aviationconstraints.IsLatitude(latitude) {
		return nil, fmt.Errorf("latitude must be finite and between -90 and 90")
	}
	if !aviationconstraints.IsLongitude(longitude) {
		return nil, fmt.Errorf("longitude must be finite and between -180 and 180")
	}
	if radius <= 0 || radius > 250 {
		return nil, fmt.Errorf("radius must be between 1 and 250 nautical miles")
	}

	endpoint := fmt.Sprintf(EndpointByPoint, latitude, longitude, radius)
	requestURL, err := url.JoinPath(client.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("build adsb.lol point url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build adsb.lol point request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", client.userAgent)

	return client.doStateRequest(request)
}

func (client *Client) doStateRequest(request *http.Request) (*StateResponse, error) {
	startedAt := time.Now()
	response, err := client.httpClient.Do(request)
	latency := time.Since(startedAt)
	if err != nil {
		requestErr := fmt.Errorf("execute adsb.lol request: %w", err)
		if observer, ok := client.responseObserver.(integrationcommon.ProviderTransportFailureObserver); ok {
			if observeErr := observer.ObserveProviderTransportFailure(
				sourceName,
				err,
				latency,
			); observeErr != nil {
				return nil, errors.Join(requestErr, observeErr)
			}
		}
		return nil, requestErr
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		statusErr := integrationcommon.ProviderStatusError(response.StatusCode)
		if statusErr == nil {
			statusErr = fmt.Errorf("unexpected provider status %d", response.StatusCode)
		}
		if client.responseObserver != nil {
			if observeErr := client.responseObserver.ObserveProviderResponse(
				sourceName,
				response.StatusCode,
				response.Header.Clone(),
				latency,
			); observeErr != nil {
				return nil, errors.Join(statusErr, observeErr)
			}
		}
		return nil, fmt.Errorf("adsb.lol request failed: %w", statusErr)
	}

	var result StateResponse
	if err := integrationcommon.DecodeJSONHTTPResponse(
		response,
		sourceName,
		maxStateResponseBytes,
		&result,
	); err != nil {
		if observer, ok := client.responseObserver.(integrationcommon.ProviderResponseFailureObserver); ok {
			_ = observer.ObserveProviderResponseFailure(
				sourceName,
				err,
				latency,
			)
		}
		return nil, fmt.Errorf("decode adsb.lol response: %w", err)
	}

	if client.responseObserver != nil {
		if err := client.responseObserver.ObserveProviderResponse(
			sourceName,
			response.StatusCode,
			response.Header.Clone(),
			latency,
		); err != nil {
			return nil, fmt.Errorf("observe adsb.lol response: %w", err)
		}
	}
	return &result, nil
}
