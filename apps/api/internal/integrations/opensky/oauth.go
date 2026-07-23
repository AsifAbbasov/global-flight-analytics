package opensky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrTokenResponseInvalid = errors.New("OpenSky OAuth2 token response is invalid")
)

const (
	tokenRefreshMargin          = 30 * time.Second
	defaultTokenLifetimeSeconds = int64(1800)
	maxTokenLifetimeSeconds     = int64(^uint64(0)>>1) / int64(time.Second)
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type TokenManager struct {
	client       *http.Client
	tokenURL     string
	clientID     string
	clientSecret string

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewTokenManager(
	client *http.Client,
	tokenURL string,
	clientID string,
	clientSecret string,
) (*TokenManager, error) {
	if client == nil {
		return nil, ErrHTTPClientRequired
	}
	if strings.TrimSpace(tokenURL) == "" {
		return nil, ErrTokenURLRequired
	}
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, ErrCredentialPairRequired
	}
	return &TokenManager{
		client:       client,
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

func (manager *TokenManager) Token(ctx context.Context) (string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.token != "" && time.Now().UTC().Before(manager.expiresAt) {
		return manager.token, nil
	}
	return manager.refresh(ctx)
}

func (manager *TokenManager) Invalidate() {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.token = ""
	manager.expiresAt = time.Time{}
}

func (manager *TokenManager) refresh(ctx context.Context) (string, error) {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", manager.clientID)
	values.Set("client_secret", manager.clientSecret)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		manager.tokenURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("build OpenSky OAuth2 request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := manager.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("execute OpenSky OAuth2 request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("OpenSky OAuth2 request failed with status %d", response.StatusCode)
	}

	var payload tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode OpenSky OAuth2 response: %w", err)
	}
	payload.AccessToken = strings.TrimSpace(payload.AccessToken)
	if payload.AccessToken == "" {
		return "", ErrTokenResponseInvalid
	}
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = defaultTokenLifetimeSeconds
	}
	if payload.ExpiresIn > maxTokenLifetimeSeconds {
		return "", ErrTokenResponseInvalid
	}

	lifetime := time.Duration(payload.ExpiresIn) * time.Second
	if lifetime > tokenRefreshMargin {
		lifetime -= tokenRefreshMargin
	}
	manager.token = payload.AccessToken
	manager.expiresAt = time.Now().UTC().Add(lifetime)
	return manager.token, nil
}
