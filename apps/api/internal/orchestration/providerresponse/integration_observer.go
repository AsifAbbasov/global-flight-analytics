package providerresponse

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrControllerRequired = errors.New(
		"provider response controller is required",
	)
	ErrTransportFailureRequired = errors.New(
		"provider transport failure is required",
	)
	ErrResponseFailureRequired = errors.New(
		"provider response failure is required",
	)
)

type ObservationRecorder interface {
	RecordHTTPResponse(
		observation Observation,
		latency time.Duration,
	) error
}

type TransportFailureRecorder interface {
	RecordTransportFailure(
		provider providerpolicy.Provider,
		outcome providerhealthdomain.RequestOutcome,
		latency time.Duration,
	) error
}

type ResponseFailureRecorder interface {
	RecordResponseFailure(
		provider providerpolicy.Provider,
		latency time.Duration,
	) error
}

type IntegrationObserver struct {
	controller *Controller
	recorder   ObservationRecorder
}

func NewIntegrationObserver(
	controller *Controller,
) (*IntegrationObserver, error) {
	return NewIntegrationObserverWithRecorder(
		controller,
		nil,
	)
}

func NewIntegrationObserverWithRecorder(
	controller *Controller,
	recorder ObservationRecorder,
) (*IntegrationObserver, error) {
	if controller == nil {
		return nil, ErrControllerRequired
	}

	return &IntegrationObserver{
		controller: controller,
		recorder:   recorder,
	}, nil
}

func (observer *IntegrationObserver) ObserveProviderResponse(
	providerName string,
	statusCode int,
	headers http.Header,
	latency time.Duration,
) error {
	observation, err := observer.controller.ObserveHTTPResponse(
		providerpolicy.Provider(
			strings.TrimSpace(providerName),
		),
		statusCode,
		headers,
	)
	if err != nil {
		return err
	}

	if observer.recorder == nil {
		return nil
	}

	return observer.recorder.RecordHTTPResponse(
		observation,
		latency,
	)
}

func (observer *IntegrationObserver) ObserveProviderTransportFailure(
	providerName string,
	requestErr error,
	latency time.Duration,
) error {
	if requestErr == nil {
		return ErrTransportFailureRequired
	}
	if errors.Is(
		requestErr,
		context.Canceled,
	) {
		return nil
	}

	provider := providerpolicy.Provider(
		strings.TrimSpace(providerName),
	)
	if _, err := providerpolicy.Get(provider); err != nil {
		return err
	}

	failureRecorder, supported :=
		observer.recorder.(TransportFailureRecorder)
	if !supported {
		return nil
	}

	return failureRecorder.RecordTransportFailure(
		provider,
		classifyTransportFailure(requestErr),
		latency,
	)
}

func (observer *IntegrationObserver) ObserveProviderResponseFailure(
	providerName string,
	responseErr error,
	latency time.Duration,
) error {
	if responseErr == nil {
		return ErrResponseFailureRequired
	}
	if errors.Is(
		responseErr,
		context.Canceled,
	) {
		return nil
	}

	provider := providerpolicy.Provider(
		strings.TrimSpace(providerName),
	)
	if _, err := providerpolicy.Get(provider); err != nil {
		return err
	}

	failureRecorder, supported :=
		observer.recorder.(ResponseFailureRecorder)
	if !supported {
		return nil
	}

	return failureRecorder.RecordResponseFailure(
		provider,
		latency,
	)
}

func classifyTransportFailure(
	requestErr error,
) providerhealthdomain.RequestOutcome {
	if errors.Is(
		requestErr,
		context.DeadlineExceeded,
	) {
		return providerhealthdomain.RequestOutcomeTimeout
	}

	var networkError net.Error
	if errors.As(
		requestErr,
		&networkError,
	) && networkError.Timeout() {
		return providerhealthdomain.RequestOutcomeTimeout
	}

	return providerhealthdomain.RequestOutcomeNetworkError
}
