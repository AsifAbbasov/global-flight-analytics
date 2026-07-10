package common

import "net/http"

type ProviderResponseObserver interface {
	ObserveProviderResponse(
		providerName string,
		statusCode int,
		headers http.Header,
	) error
}
