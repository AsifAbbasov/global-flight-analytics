package airplaneslive

import (
	"net/http"

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
