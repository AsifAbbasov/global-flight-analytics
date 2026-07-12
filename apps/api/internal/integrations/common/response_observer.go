package common

import (
	"net/http"
	"time"
)

type ProviderResponseObserver interface {
	ObserveProviderResponse(
		providerName string,
		statusCode int,
		headers http.Header,
		latency time.Duration,
	) error
}

type ProviderTransportFailureObserver interface {
	ObserveProviderTransportFailure(
		providerName string,
		requestErr error,
		latency time.Duration,
	) error
}

type ProviderResponseFailureObserver interface {
	ObserveProviderResponseFailure(
		providerName string,
		responseErr error,
		latency time.Duration,
	) error
}
