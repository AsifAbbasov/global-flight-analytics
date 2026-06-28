package airplaneslive

import (
	"context"
	"fmt"
	"net/http"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

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
		c.userAgent,
	)

	return request, nil
}
