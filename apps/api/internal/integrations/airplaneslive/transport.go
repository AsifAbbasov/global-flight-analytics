package airplaneslive

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) do(
	request *http.Request,
	target any,
) error {
	response, err := c.httpClient.Do(
		request,
	)
	if err != nil {
		return fmt.Errorf(
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
			return fmt.Errorf(
				"observe airplanes live response: %w",
				err,
			)
		}
	}

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf(
			"request failed with status %d",
			response.StatusCode,
		)
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(
		target,
	); err != nil {
		return fmt.Errorf(
			"decode response: %w",
			err,
		)
	}

	return nil
}
