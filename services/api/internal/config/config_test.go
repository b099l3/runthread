package config

import "testing"

func TestLoadUsesLocalDefaults(t *testing.T) {
	t.Setenv(EnvServerAddress, "")
	t.Setenv(EnvDatabaseURL, "")

	cfg := Load()

	if cfg.ServerAddress != DefaultServerAddress {
		t.Fatalf("ServerAddress = %q, want %q", cfg.ServerAddress, DefaultServerAddress)
	}
	if cfg.DatabaseConfigured() {
		t.Fatal("DatabaseConfigured = true, want false")
	}
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv(EnvServerAddress, "127.0.0.1:9090")
	t.Setenv(EnvDatabaseURL, "postgres://runthread:runthread@localhost:5432/runthread?sslmode=disable")

	cfg := Load()

	if cfg.ServerAddress != "127.0.0.1:9090" {
		t.Fatalf("ServerAddress = %q, want configured address", cfg.ServerAddress)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL is empty, want configured URL")
	}
	if !cfg.DatabaseConfigured() {
		t.Fatal("DatabaseConfigured = false, want true")
	}
}

func TestLoadTrimsBlankEnvironmentValues(t *testing.T) {
	t.Setenv(EnvServerAddress, "  ")
	t.Setenv(EnvDatabaseURL, "  ")

	cfg := Load()

	if cfg.ServerAddress != DefaultServerAddress {
		t.Fatalf("ServerAddress = %q, want default for blank env", cfg.ServerAddress)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q, want empty for blank env", cfg.DatabaseURL)
	}
}
