package plugin_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/transpara/t-store-datasource/pkg/plugin"
)

func makeTestPlugin(tstoreURL, keycloakURL string) (*plugin.Datasource, error) {
	settings := backend.DataSourceInstanceSettings{
		URL: tstoreURL,
		JSONData: json.RawMessage(`{"url":"` + tstoreURL + `","tokenUrl":"` + keycloakURL + `","clientId":"test-client"}`),
		DecryptedSecureJSONData: map[string]string{
			"clientSecret": "test-secret",
		},
	}
	return plugin.NewDatasource(context.Background(), settings)
}

func makeKeycloakMock(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-token",
			"expires_in":   3600,
		})
	}))
}

func TestCheckHealth_Success(t *testing.T) {
	keycloak := makeKeycloakMock(t)
	defer keycloak.Close()

	tstore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/up" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer tstore.Close()

	ds, err := makeTestPlugin(tstore.URL, keycloak.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Errorf("expected OK, got %v: %s", result.Status, result.Message)
	}
}

func TestCheckHealth_BadKeycloak(t *testing.T) {
	keycloak := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer keycloak.Close()

	tstore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer tstore.Close()

	ds, err := makeTestPlugin(tstore.URL, keycloak.URL)
	if err != nil {
		t.Fatal(err)
	}

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != backend.HealthStatusError {
		t.Errorf("expected Error status, got %v", result.Status)
	}
}

func TestCallResource_Datasets(t *testing.T) {
	keycloak := makeKeycloakMock(t)
	defer keycloak.Close()

	tstore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/dataset" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]string{"plant-a", "plant-b"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer tstore.Close()

	ds, err := makeTestPlugin(tstore.URL, keycloak.URL)
	if err != nil {
		t.Fatal(err)
	}

	var captured *backend.CallResourceResponse
	sender := &mockSender{capture: func(r *backend.CallResourceResponse) { captured = r }}

	err = ds.CallResource(context.Background(), &backend.CallResourceRequest{Path: "datasets"}, sender)
	if err != nil {
		t.Fatal(err)
	}
	if captured.Status != http.StatusOK {
		t.Errorf("expected 200, got %d", captured.Status)
	}
	var result []string
	json.Unmarshal(captured.Body, &result)
	if len(result) != 2 || result[0] != "plant-a" {
		t.Errorf("unexpected datasets: %v", result)
	}
}

type mockSender struct {
	capture func(*backend.CallResourceResponse)
}

func (m *mockSender) Send(r *backend.CallResourceResponse) error {
	m.capture(r)
	return nil
}
