package config

import (
	"os"
	"strings"
)

const (
	DefaultServerAddress    = ":8080"
	DefaultStravaAPIBaseURL = "https://www.strava.com"

	EnvServerAddress            = "RUNTHREAD_SERVER_ADDR"
	EnvDatabaseURL              = "DATABASE_URL"
	EnvStravaClientID           = "STRAVA_CLIENT_ID"
	EnvStravaClientSecret       = "STRAVA_CLIENT_SECRET"
	EnvStravaOAuthRedirectURI   = "STRAVA_OAUTH_REDIRECT_URI"
	EnvStravaWebhookVerifyToken = "STRAVA_WEBHOOK_VERIFY_TOKEN"
	EnvStravaAPIBaseURL         = "STRAVA_API_BASE_URL"
)

type Config struct {
	ServerAddress            string
	DatabaseURL              string
	StravaClientID           string
	StravaClientSecret       string
	StravaOAuthRedirectURI   string
	StravaWebhookVerifyToken string
	StravaAPIBaseURL         string
}

func Load() Config {
	return Config{
		ServerAddress:            stringFromEnv(EnvServerAddress, DefaultServerAddress),
		DatabaseURL:              strings.TrimSpace(os.Getenv(EnvDatabaseURL)),
		StravaClientID:           strings.TrimSpace(os.Getenv(EnvStravaClientID)),
		StravaClientSecret:       strings.TrimSpace(os.Getenv(EnvStravaClientSecret)),
		StravaOAuthRedirectURI:   strings.TrimSpace(os.Getenv(EnvStravaOAuthRedirectURI)),
		StravaWebhookVerifyToken: strings.TrimSpace(os.Getenv(EnvStravaWebhookVerifyToken)),
		StravaAPIBaseURL:         stringFromEnv(EnvStravaAPIBaseURL, DefaultStravaAPIBaseURL),
	}
}

func (c Config) DatabaseConfigured() bool {
	return c.DatabaseURL != ""
}

func (c Config) StravaOAuthConfigured() bool {
	return c.StravaClientID != "" && c.StravaClientSecret != "" && c.StravaOAuthRedirectURI != ""
}

func stringFromEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
