package providerresponse

import (
	"errors"
	"net/http"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var ErrControllerRequired = errors.New(
	"provider response controller is required",
)

type IntegrationObserver struct {
	controller *Controller
}

func NewIntegrationObserver(
	controller *Controller,
) (*IntegrationObserver, error) {
	if controller == nil {
		return nil, ErrControllerRequired
	}

	return &IntegrationObserver{
		controller: controller,
	}, nil
}

func (observer *IntegrationObserver) ObserveProviderResponse(
	providerName string,
	statusCode int,
	headers http.Header,
) error {
	_, err := observer.controller.ObserveHTTPResponse(
		providerpolicy.Provider(
			strings.TrimSpace(providerName),
		),
		statusCode,
		headers,
	)

	return err
}
