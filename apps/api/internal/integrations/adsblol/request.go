package adsblol

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

const adsbLOLContactURL = "https://github.com/AsifAbbasov/global-flight-analytics"

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	requestURL string,
) (*http.Request, error) {

	request, err := http.NewRequestWithContext(
		ctx,
		method,
		requestURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	request.Header.Set(
		integrationcommon.HeaderAccept,
		integrationcommon.ContentTypeJSON,
	)

	request.Header.Set(
		integrationcommon.HeaderUserAgent,
		identifiedUserAgent(c.userAgent),
	)

	return request, nil
}

func identifiedUserAgent(userAgent string) string {
	trimmed := strings.TrimSpace(userAgent)
	if strings.Contains(trimmed, adsbLOLContactURL) {
		return trimmed
	}

	return fmt.Sprintf("%s (+%s)", trimmed, adsbLOLContactURL)
}
