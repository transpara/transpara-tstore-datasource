package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AuthSettings holds non-secret auth configuration.
type AuthSettings struct {
	TokenURL string
	ClientID string
}

// tokenResponse is the Keycloak token endpoint response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// FetchToken performs a client_credentials grant against the Keycloak token endpoint.
// Returns the access token and its expiry time.
func FetchToken(ctx context.Context, httpClient *http.Client, settings AuthSettings, clientSecret string) (string, time.Time, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {settings.ClientID},
		"client_secret": {clientSecret},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, settings.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", time.Time{}, fmt.Errorf("decoding token response: %w", err)
	}

	// Subtract 30s buffer so we refresh before actual expiry.
	expiry := time.Now().Add(time.Duration(tr.ExpiresIn-30) * time.Second)
	return tr.AccessToken, expiry, nil
}

// TokenCache caches a single bearer token and refreshes it on expiry.
type TokenCache struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// NewTokenCache returns an empty TokenCache.
func NewTokenCache() *TokenCache {
	return &TokenCache{}
}

// Reset clears the cached token, forcing the next GetToken call to fetch a new one.
func (c *TokenCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
	c.expiresAt = time.Time{}
}

// GetToken returns the cached token, fetching a new one if expired or absent.
func (c *TokenCache) GetToken(ctx context.Context, httpClient *http.Client, settings AuthSettings, clientSecret string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}

	token, expiry, err := FetchToken(ctx, httpClient, settings, clientSecret)
	if err != nil {
		return "", err
	}
	c.token = token
	c.expiresAt = expiry
	return c.token, nil
}
