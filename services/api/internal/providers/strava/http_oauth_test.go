package strava

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPCodeExchangerExchangesAuthorizationCode(t *testing.T) {
	var formBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
			t.Fatalf("content type = %q, want form", contentType)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		formBody = r.Form.Encode()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-token",
			"refresh_token": "refresh-token",
			"expires_at":    testOAuthNow().Add(time.Hour).Unix(),
			"scope":         "read,activity:read_all",
			"athlete": map[string]any{
				"id": 12345,
			},
		})
	}))
	defer server.Close()

	token, err := HTTPCodeExchanger{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		TokenURL:     server.URL,
	}.ExchangeCode(context.Background(), OAuthCodeExchangeRequest{Code: "auth-code-1"})
	if err != nil {
		t.Fatalf("ExchangeCode returned error: %v", err)
	}

	if !strings.Contains(formBody, "grant_type=authorization_code") {
		t.Fatalf("form body = %q, want grant type", formBody)
	}
	if token.ProviderUserID != "12345" {
		t.Fatalf("ProviderUserID = %q, want 12345", token.ProviderUserID)
	}
	if token.AccessToken != "access-token" {
		t.Fatalf("AccessToken = %q, want access-token", token.AccessToken)
	}
	if len(token.Scopes) != 2 || token.Scopes[1] != "activity:read_all" {
		t.Fatalf("Scopes = %#v, want read/activity", token.Scopes)
	}
}

func TestHTTPCodeExchangerReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad code", http.StatusBadRequest)
	}))
	defer server.Close()

	_, err := HTTPCodeExchanger{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		TokenURL:     server.URL,
	}.ExchangeCode(context.Background(), OAuthCodeExchangeRequest{Code: "auth-code-1"})
	if err == nil {
		t.Fatal("expected status error")
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("error = %q, want status", err.Error())
	}
}

func TestHTTPCodeExchangerRefreshesToken(t *testing.T) {
	var formBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		formBody = r.Form.Encode()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"expires_at":    testOAuthNow().Add(time.Hour).Unix(),
			"scope":         "activity:read_all",
		})
	}))
	defer server.Close()

	token, err := HTTPCodeExchanger{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		TokenURL:     server.URL,
	}.RefreshToken(context.Background(), "refresh-token")
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}

	if !strings.Contains(formBody, "grant_type=refresh_token") {
		t.Fatalf("form body = %q, want refresh grant type", formBody)
	}
	if token.AccessToken != "new-access-token" {
		t.Fatalf("access token = %q, want new-access-token", token.AccessToken)
	}
	if token.ProviderUserID != "" {
		t.Fatalf("provider user id = %q, want empty for refresh response", token.ProviderUserID)
	}
}
