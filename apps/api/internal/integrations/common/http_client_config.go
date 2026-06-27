package integrations

import "time"

type HTTPClientConfig struct {
	BaseURL   string
	Timeout   time.Duration
	UserAgent string
}

func DefaultHTTPClientConfig(baseURL string) HTTPClientConfig {
	return HTTPClientConfig{
		BaseURL:   baseURL,
		Timeout:   15 * time.Second,
		UserAgent: DefaultUserAgent,
	}
}
