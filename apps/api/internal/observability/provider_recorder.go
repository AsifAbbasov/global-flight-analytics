package observability

import (
	"net/http"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
)

type providerHTTPRecorder interface {
	RecordHTTPResponse(
		observation providerresponse.Observation,
		latency time.Duration,
	) error
}

type providerTransportRecorder interface {
	RecordTransportFailure(
		provider providerpolicy.Provider,
		outcome providerhealthdomain.RequestOutcome,
		latency time.Duration,
	) error
}

type providerResponseFailureRecorder interface {
	RecordResponseFailure(
		provider providerpolicy.Provider,
		latency time.Duration,
	) error
}

type ProviderRecorder struct {
	registry *Registry
	next     providerHTTPRecorder
}

func NewProviderRecorder(
	registry *Registry,
	next providerHTTPRecorder,
) *ProviderRecorder {
	return &ProviderRecorder{
		registry: registry,
		next:     next,
	}
}

func (
	recorder *ProviderRecorder,
) RecordHTTPResponse(
	observation providerresponse.Observation,
	latency time.Duration,
) error {
	if recorder != nil && recorder.registry != nil {
		recorder.registry.ObserveProviderRequest(
			string(observation.Provider),
			providerHTTPOutcome(observation.StatusCode),
			latency,
		)
	}

	if recorder == nil || recorder.next == nil {
		return nil
	}
	return recorder.next.RecordHTTPResponse(observation, latency)
}

func (
	recorder *ProviderRecorder,
) RecordTransportFailure(
	provider providerpolicy.Provider,
	outcome providerhealthdomain.RequestOutcome,
	latency time.Duration,
) error {
	if recorder != nil && recorder.registry != nil {
		recorder.registry.ObserveProviderRequest(
			string(provider),
			string(outcome),
			latency,
		)
	}

	if recorder == nil || recorder.next == nil {
		return nil
	}
	next, supported := recorder.next.(providerTransportRecorder)
	if !supported {
		return nil
	}
	return next.RecordTransportFailure(provider, outcome, latency)
}

func (
	recorder *ProviderRecorder,
) RecordResponseFailure(
	provider providerpolicy.Provider,
	latency time.Duration,
) error {
	if recorder != nil && recorder.registry != nil {
		recorder.registry.ObserveProviderRequest(
			string(provider),
			"invalid_response",
			latency,
		)
	}

	if recorder == nil || recorder.next == nil {
		return nil
	}
	next, supported := recorder.next.(providerResponseFailureRecorder)
	if !supported {
		return nil
	}
	return next.RecordResponseFailure(provider, latency)
}

func providerHTTPOutcome(
	statusCode int,
) string {
	switch {
	case statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices:
		return "success"
	case statusCode == http.StatusTooManyRequests:
		return "rate_limited"
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return "unauthorized"
	case statusCode >= http.StatusInternalServerError:
		return "server_error"
	default:
		return "client_error"
	}
}
