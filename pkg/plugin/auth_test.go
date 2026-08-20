package plugin_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/transpara/t-store-datasource/pkg/plugin"
)

func TestFetchToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("grant_type") != "client_credentials" {
			t.Errorf("expected grant_type=client_credentials, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("client_id") != "test-client" {
			t.Errorf("expected client_id=test-client, got %s", r.FormValue("client_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token-123",
			"expires_in":   3600,
		})
	}))
	defer srv.Close()

	settings := plugin.AuthSettings{
		TokenURL: srv.URL,
		ClientID: "test-client",
	}

	token, _, err := plugin.FetchToken(context.Background(), http.DefaultClient, settings, "test-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "test-token-123" {
		t.Errorf("expected token=test-token-123, got %s", token)
	}
}

func TestTokenCache_RefreshesOnExpiry(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token-" + string(rune('0'+n)),
			"expires_in":   1, // 1 second — expires immediately for test
		})
	}))
	defer srv.Close()

	cache := plugin.NewTokenCache()
	settings := plugin.AuthSettings{TokenURL: srv.URL, ClientID: "c"}

	tok1, err := cache.GetToken(context.Background(), http.DefaultClient, settings, "secret")
	if err != nil {
		t.Fatal(err)
	}

	// Force expiry
	time.Sleep(2 * time.Second)

	tok2, err := cache.GetToken(context.Background(), http.DefaultClient, settings, "secret")
	if err != nil {
		t.Fatal(err)
	}

	if tok1 == tok2 {
		t.Error("expected different tokens after cache expiry")
	}
	if callCount.Load() != 2 {
		t.Errorf("expected 2 token fetches, got %d", callCount.Load())
	}
}
