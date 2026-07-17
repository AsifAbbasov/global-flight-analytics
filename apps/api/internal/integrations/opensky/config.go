package opensky

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL  = "https://opensky-network.org/api"
	DefaultTokenURL = "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token"
)

var (
	ErrBaseURLRequired        = errors.New("OpenSky base URL is required")
	ErrTokenURLRequired       = errors.New("OpenSky token URL is required when OAuth2 credentials are configured")
	ErrCredentialPairRequired = errors.New("OpenSky client id and client secret must be configured together")
	ErrHTTPClientRequired     = errors.New("OpenSky HTTP client is required")
	ErrPollingIntervalInvalid = errors.New("OpenSky polling interval is below the supported minimum")
)

type Config struct {
	BaseURL         string
	TokenURL        string
	ClientID        string
	ClientSecret    string
	HTTPClient      *http.Client
	UserAgent       string
	PollingInterval time.Duration
}

func DefaultConfig() Config {
	return Config{
		BaseURL:         DefaultBaseURL,
		TokenURL:        DefaultTokenURL,
		HTTPClient:      &http.Client{Timeout: 15 * time.Second},
		UserAgent:       "global-flight-analytics",
		PollingInterval: 10 * time.Second,
	}
}

func (config Config) Authenticated() bool {
	return strings.TrimSpace(config.ClientID) != "" &&
		strings.TrimSpace(config.ClientSecret) != ""
}

func (config Config) Validate() error {
	if strings.TrimSpace(config.BaseURL) == "" {
		return ErrBaseURLRequired
	}
	if config.HTTPClient == nil {
		return ErrHTTPClientRequired
	}

	clientIDSet := strings.TrimSpace(config.ClientID) != ""
	clientSecretSet := strings.TrimSpace(config.ClientSecret) != ""
	if clientIDSet != clientSecretSet {
		return ErrCredentialPairRequired
	}
	if clientIDSet && strings.TrimSpace(config.TokenURL) == "" {
		return ErrTokenURLRequired
	}

	minimum := 10 * time.Second
	if clientIDSet {
		minimum = 5 * time.Second
	}
	if config.PollingInterval < minimum {
		return ErrPollingIntervalInvalid
	}
	return nil
}
