package plugin_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryData_VisualMode(t *testing.T) {
	keycloak := makeKeycloakMock(t)
	defer keycloak.Close()

	tstore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/read/historical-data" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body []string
		json.NewDecoder(r.Body).Decode(&body)
		if len(body) != 1 || body[0] != "plant-a|sensor_id=123" {
			t.Errorf("unexpected lookup body: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"plant-a|sensor_id=123": []map[string]interface{}{
				{"ts": "2024-01-01T00:00:00Z", "v": "42.5", "dv": "42.5"},
				{"ts": "2024-01-01T00:01:00Z", "v": "43.0", "dv": "43.0"},
			},
		})
	}))
	defer tstore.Close()

	ds, err := makeTestPlugin(tstore.URL, keycloak.URL)
	if err != nil {
		t.Fatal(err)
	}

	queryJSON, _ := json.Marshal(map[string]interface{}{
		"queryType": "visual",
		"lookups":   []string{"plant-a|sensor_id=123"},
		"aggType":   "avg",
		"aggInt":    "1m",
		"tz":        "UTC",
	})

	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{
				RefID:     "A",
				JSON:      queryJSON,
				TimeRange: backend.TimeRange{From: time.Now().Add(-1 * time.Hour), To: time.Now()},
			},
		},
	}

	resp, err := ds.QueryData(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	result, ok := resp.Responses["A"]
	if !ok {
		t.Fatal("missing response for RefID A")
	}
	if result.Error != nil {
		t.Fatalf("query error: %v", result.Error)
	}
	if len(result.Frames) != 1 {
		t.Errorf("expected 1 frame, got %d", len(result.Frames))
	}
	if result.Frames[0].Rows() != 2 {
		t.Errorf("expected 2 rows, got %d", result.Frames[0].Rows())
	}
}

func TestQueryData_RawMode(t *testing.T) {
	keycloak := makeKeycloakMock(t)
	defer keycloak.Close()

	called := false
	tstore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer tstore.Close()

	ds, err := makeTestPlugin(tstore.URL, keycloak.URL)
	if err != nil {
		t.Fatal(err)
	}

	rawBody := `{"lookups":["plant-a|sensor_id=123"]}`
	queryJSON, _ := json.Marshal(map[string]interface{}{
		"queryType": "raw",
		"rawJson":   rawBody,
	})

	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{
				RefID:     "A",
				JSON:      queryJSON,
				TimeRange: backend.TimeRange{From: time.Now().Add(-1 * time.Hour), To: time.Now()},
			},
		},
	}

	_, err = ds.QueryData(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected tstore to be called")
	}
}
