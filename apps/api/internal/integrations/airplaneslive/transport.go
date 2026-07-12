package airplaneslive

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func (c *Client) doStateRequest(
	request *http.Request,
) (*StateResponse, error) {
	startedAt := time.Now()

	response, err := c.httpClient.Do(
		request,
	)
	latency := time.Since(startedAt)

	if err != nil {
		requestErr := fmt.Errorf(
			"execute request: %w",
			err,
		)

		if observeErr := c.observeTransportFailure(
			err,
			latency,
		); observeErr != nil {
			return nil, errors.Join(
				requestErr,
				fmt.Errorf(
					"observe airplanes live transport failure: %w",
					observeErr,
				),
			)
		}

		return nil, requestErr
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		if err := c.observeResponse(
			response,
			latency,
		); err != nil {
			return nil, fmt.Errorf(
				"observe airplanes live response: %w",
				err,
			)
		}

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
		decodeErr := fmt.Errorf(
			"decode response: %w",
			err,
		)

		if observeErr := c.observeResponseFailure(
			err,
			latency,
		); observeErr != nil {
			return nil, errors.Join(
				decodeErr,
				fmt.Errorf(
					"observe airplanes live response failure: %w",
					observeErr,
				),
			)
		}

		return nil, decodeErr
	}

	if err := c.observeResponse(
		response,
		latency,
	); err != nil {
		return nil, fmt.Errorf(
			"observe airplanes live response: %w",
			err,
		)
	}

	return &result, nil
}

func (c *Client) observeResponse(
	response *http.Response,
	latency time.Duration,
) error {
	if c.responseObserver == nil {
		return nil
	}

	return c.responseObserver.ObserveProviderResponse(
		sourceName,
		response.StatusCode,
		response.Header.Clone(),
		latency,
	)
}

func (c *Client) observeTransportFailure(
	requestErr error,
	latency time.Duration,
) error {
	if c.responseObserver == nil {
		return nil
	}

	observer, supported :=
		c.responseObserver.(integrationcommon.ProviderTransportFailureObserver)
	if !supported {
		return nil
	}

	return observer.ObserveProviderTransportFailure(
		sourceName,
		requestErr,
		latency,
	)
}

func (c *Client) observeResponseFailure(
	responseErr error,
	latency time.Duration,
) error {
	if c.responseObserver == nil {
		return nil
	}

	observer, supported :=
		c.responseObserver.(integrationcommon.ProviderResponseFailureObserver)
	if !supported {
		return nil
	}

	return observer.ObserveProviderResponseFailure(
		sourceName,
		responseErr,
		latency,
	)
}
