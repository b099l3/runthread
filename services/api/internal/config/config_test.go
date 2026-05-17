package config

import "testing"

func TestLoadUsesLocalDefaults(t *testing.T) {
	t.Setenv(EnvServerAddress, "")
	t.Setenv(EnvDatabaseURL, "")
	t.Setenv(EnvStravaAPIBaseURL, "")

	cfg := Load()

	if cfg.ServerAddress != DefaultServerAddress {
		t.Fatalf("ServerAddress = %q, want %q", cfg.ServerAddress, DefaultServerAddress)
	}
	if cfg.StravaAPIBaseURL != DefaultStravaAPIBaseURL {
		t.Fatalf("StravaAPIBaseURL = %q, want %q", cfg.StravaAPIBaseURL, DefaultStravaAPIBaseURL)
	}
	if cfg.DatabaseConfigured() {
		t.Fatal("DatabaseConfigured = true, want false")
	}
	if cfg.StravaOAuthConfigured() {
		t.Fatal("StravaOAuthConfigured = true, want false")
	}
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv(EnvServerAddress, "127.0.0.1:9090")
	t.Setenv(EnvDatabaseURL, "postgres://runthread:runthread@localhost:5432/runthread?sslmode=disable")
	t.Setenv(EnvStravaClientID, "strava-client-id")
	t.Setenv(EnvStravaClientSecret, "strava-client-secret")
	t.Setenv(EnvStravaOAuthRedirectURI, "runthread://provider/strava/callback")
	t.Setenv(EnvStravaWebhookVerifyToken, "strava-webhook-token")
	t.Setenv(EnvStravaAPIBaseURL, "https://strava.test")

	cfg := Load()

	if cfg.ServerAddress != "127.0.0.1:9090" {
		t.Fatalf("ServerAddress = %q, want configured address", cfg.ServerAddress)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL is empty, want configured URL")
	}
	if cfg.StravaClientID != "strava-client-id" {
		t.Fatalf("StravaClientID = %q, want configured value", cfg.StravaClientID)
	}
	if cfg.StravaClientSecret != "strava-client-secret" {
		t.Fatalf("StravaClientSecret = %q, want configured value", cfg.StravaClientSecret)
	}
	if cfg.StravaOAuthRedirectURI != "runthread://provider/strava/callback" {
		t.Fatalf("StravaOAuthRedirectURI = %q, want configured value", cfg.StravaOAuthRedirectURI)
	}
	if cfg.StravaWebhookVerifyToken != "strava-webhook-token" {
		t.Fatalf("StravaWebhookVerifyToken = %q, want configured value", cfg.StravaWebhookVerifyToken)
	}
	if cfg.StravaAPIBaseURL != "https://strava.test" {
		t.Fatalf("StravaAPIBaseURL = %q, want configured value", cfg.StravaAPIBaseURL)
	}
	if !cfg.DatabaseConfigured() {
		t.Fatal("DatabaseConfigured = false, want true")
	}
	if !cfg.StravaOAuthConfigured() {
		t.Fatal("StravaOAuthConfigured = false, want true")
	}
}

func TestLoadTrimsBlankEnvironmentValues(t *testing.T) {
	t.Setenv(EnvServerAddress, "  ")
	t.Setenv(EnvDatabaseURL, "  ")
	t.Setenv(EnvStravaClientID, "  ")
	t.Setenv(EnvStravaClientSecret, "  ")
	t.Setenv(EnvStravaOAuthRedirectURI, "  ")
	t.Setenv(EnvStravaWebhookVerifyToken, "  ")
	t.Setenv(EnvStravaAPIBaseURL, "  ")

	cfg := Load()

	if cfg.ServerAddress != DefaultServerAddress {
		t.Fatalf("ServerAddress = %q, want default for blank env", cfg.ServerAddress)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q, want empty for blank env", cfg.DatabaseURL)
	}
	if cfg.StravaClientID != "" {
		t.Fatalf("StravaClientID = %q, want empty for blank env", cfg.StravaClientID)
	}
	if cfg.StravaClientSecret != "" {
		t.Fatalf("StravaClientSecret = %q, want empty for blank env", cfg.StravaClientSecret)
	}
	if cfg.StravaOAuthRedirectURI != "" {
		t.Fatalf("StravaOAuthRedirectURI = %q, want empty for blank env", cfg.StravaOAuthRedirectURI)
	}
	if cfg.StravaWebhookVerifyToken != "" {
		t.Fatalf("StravaWebhookVerifyToken = %q, want empty for blank env", cfg.StravaWebhookVerifyToken)
	}
	if cfg.StravaAPIBaseURL != DefaultStravaAPIBaseURL {
		t.Fatalf("StravaAPIBaseURL = %q, want default for blank env", cfg.StravaAPIBaseURL)
	}
	if cfg.StravaOAuthConfigured() {
		t.Fatal("StravaOAuthConfigured = true, want false")
	}
}
