package airplaneslive

import (
	"encoding/json"
	"fmt"
	"net/http"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func (c *Client) doStateRequest(
	request *http.Request,
) (*StateResponse, error) {
	response, err := c.httpClient.Do(
		request,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"execute request: %w",
			err,
		)
	}
	defer response.Body.Close()

	if c.responseObserver != nil {
		err := c.responseObserver.ObserveProviderResponse(
			sourceName,
			response.StatusCode,
			response.Header.Clone(),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"observe airplanes live response: %w",
				err,
			)
		}
	}

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"airplanes live request failed: %w",
			integrationcommon.ProviderStatusError(
				response.StatusCode,
			),
		)
	}

	var result StateResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(
		&result,
	); err != nil {
		return nil, fmt.Errorf(
			"decode response: %w",
			err,
		)
	}

	return &result, nil
}
