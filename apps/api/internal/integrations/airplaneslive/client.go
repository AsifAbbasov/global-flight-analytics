package airplaneslive

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

func NewClient(config integrationcommon.HTTPClientConfig) *Client {
	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		userAgent: config.UserAgent,
	}
}

func (c *Client) GetByCallsign(
	ctx context.Context,
	callsign string,
) (*StateResponse, error) {
	if callsign == "" {
		return nil, fmt.Errorf("callsign is required")
	}

	endpointPath := fmt.Sprintf(
		EndpointByCallsign,
		url.PathEscape(callsign),
	)

	requestURL, err := url.JoinPath(c.baseURL, endpointPath)
	if err != nil {
		return nil, fmt.Errorf(
			"build airplanes live callsign url: %w",
			err,
		)
	}

	request, err := c.newRequest(
		ctx,
		http.MethodGet,
		requestURL,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build airplanes live callsign request: %w",
			err,
		)
	}

	var result StateResponse

	if err := c.do(request, &result); err != nil {
		return nil, fmt.Errorf(
			"execute airplanes live callsign request: %w",
			err,
		)
	}

	return &result, nil
}

func (c *Client) GetByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) (*StateResponse, error) {
	if !aviationconstraints.IsLatitude(latitude) {
		return nil, fmt.Errorf(
			"latitude must be finite and between -90 and 90",
		)
	}

	if !aviationconstraints.IsLongitude(longitude) {
		return nil, fmt.Errorf(
			"longitude must be finite and between -180 and 180",
		)
	}

	if radius <= 0 {
		return nil, fmt.Errorf(
			"radius must be greater than zero",
		)
	}

	endpointPath := fmt.Sprintf(
		EndpointByPoint,
		latitude,
		longitude,
		radius,
	)

	requestURL, err := url.JoinPath(c.baseURL, endpointPath)
	if err != nil {
		return nil, fmt.Errorf(
			"build airplanes live point url: %w",
			err,
		)
	}

	request, err := c.newRequest(
		ctx,
		http.MethodGet,
		requestURL,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build airplanes live point request: %w",
			err,
		)
	}

	var result StateResponse

	if err := c.do(request, &result); err != nil {
		return nil, fmt.Errorf(
			"execute airplanes live point request: %w",
			err,
		)
	}

	return &result, nil
}
