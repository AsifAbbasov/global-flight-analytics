package airplaneslive

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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

func (c *Client) GetByCallsign(ctx context.Context, callsign string) (*StateResponse, error) {
	if callsign == "" {
		return nil, fmt.Errorf("callsign is required")
	}
	endpointPath := fmt.Sprintf(EndpointByCallsign, url.PathEscape(callsign))

	requestURL, err := url.JoinPath(c.baseURL, endpointPath)
	if err != nil {
		return nil, fmt.Errorf("build airplanes live callsign url: %w", err)
	}

	request, err := c.newRequest(ctx, http.MethodGet, requestURL)
	if err != nil {
		return nil, fmt.Errorf("build airplanes live callsign request: %w", err)
	}

	var result StateResponse

	if err := c.do(request, &result); err != nil {
		return nil, fmt.Errorf("execute airplanes live callsign request: %w", err)
	}

	return &result, nil
}
