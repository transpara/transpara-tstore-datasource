package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

var _ backend.QueryDataHandler = (*Datasource)(nil)
var _ backend.CheckHealthHandler = (*Datasource)(nil)
var _ backend.CallResourceHandler = (*Datasource)(nil)
var _ instancemgmt.Instance = (*Datasource)(nil)

// Settings holds the non-secret plugin configuration.
type Settings struct {
	URL      string `json:"url"`
	TokenURL string `json:"tokenUrl"`
	ClientID string `json:"clientId"`
}

// Datasource is the plugin instance.
type Datasource struct {
	settings     Settings
	clientSecret string
	tokenCache   *TokenCache
	httpClient   *http.Client
}

// NewDatasource creates a new plugin instance from Grafana's instance settings.
// Returns *Datasource (which satisfies instancemgmt.Instance).
func NewDatasource(_ context.Context, s backend.DataSourceInstanceSettings) (*Datasource, error) {
	var settings Settings
	if err := json.Unmarshal(s.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("parsing datasource settings: %w", err)
	}
	if settings.URL == "" {
		settings.URL = s.URL
	}
	return &Datasource{
		settings:     settings,
		clientSecret: s.DecryptedSecureJSONData["clientSecret"],
		tokenCache:   NewTokenCache(),
		httpClient:   &http.Client{},
	}, nil
}

// Dispose cleans up the instance (required by instancemgmt.Instance).
func (d *Datasource) Dispose() {}

// authSettings returns the AuthSettings for token fetching.
func (d *Datasource) authSettings() AuthSettings {
	return AuthSettings{TokenURL: d.settings.TokenURL, ClientID: d.settings.ClientID}
}

// doRequest executes an HTTP request against tstore-interface with a Bearer token.
// On 401, clears the token cache and retries once.
func (d *Datasource) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	token, err := d.tokenCache.GetToken(ctx, d.authSettings(), d.clientSecret)
	if err != nil {
		return nil, fmt.Errorf("fetching token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		d.tokenCache = NewTokenCache()
		token, err = d.tokenCache.GetToken(ctx, d.authSettings(), d.clientSecret)
		if err != nil {
			return nil, fmt.Errorf("refreshing token: %w", err)
		}
		req2 := req.Clone(ctx)
		req2.Header.Set("Authorization", "Bearer "+token)
		return d.httpClient.Do(req2)
	}

	return resp, nil
}

// CheckHealth verifies Keycloak auth and tstore-interface reachability.
func (d *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.settings.URL+"/api/v1/up", nil)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}

	resp, err := d.doRequest(ctx, req)
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("tstore-interface returned %d", resp.StatusCode),
		}, nil
	}

	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "OK"}, nil
}

// CallResource proxies dropdown requests to tstore-interface.
//
// Supported paths:
//   - "datasets"               -> GET /api/v1/dataset
//   - "lookups?dataset=<name>" -> GET /api/v1/lookups?filter=<name>&limit=1000
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	var targetURL string

	switch {
	case req.Path == "datasets":
		targetURL = d.settings.URL + "/api/v1/dataset"
	case strings.HasPrefix(req.Path, "lookups"):
		params := url.Values{}
		parsed, err := url.Parse("/?" + req.URL)
		if err == nil {
			if ds := parsed.Query().Get("dataset"); ds != "" {
				params.Set("filter", ds)
			}
		}
		params.Set("limit", "1000")
		targetURL = d.settings.URL + "/api/v1/lookups?" + params.Encode()
	default:
		return sender.Send(&backend.CallResourceResponse{Status: http.StatusNotFound})
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}

	resp, err := d.doRequest(ctx, httpReq)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadGateway,
			Body:   []byte(err.Error()),
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// For lookups, unwrap PaginatedResponse and return just the results array.
	if strings.HasPrefix(req.Path, "lookups") {
		var paginated struct {
			Results []string `json:"results"`
		}
		if jsonErr := json.Unmarshal(body, &paginated); jsonErr == nil {
			body, _ = json.Marshal(paginated.Results)
		}
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: resp.StatusCode,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: body,
	})
}

// QueryData is implemented in query.go.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// Placeholder -- will be replaced in Task 7.
	return backend.NewQueryDataResponse(), nil
}
