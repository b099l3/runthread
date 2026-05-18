package config

import (
	"os"
	"strconv"
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
	EnvProviderTokenKey         = "PROVIDER_TOKEN_ENCRYPTION_KEY"
	EnvStravaWebhookRetrySecs   = "STRAVA_WEBHOOK_RETRY_INTERVAL_SECONDS"
)

type Config struct {
	ServerAddress                     string
	DatabaseURL                       string
	StravaClientID                    string
	StravaClientSecret                string
	StravaOAuthRedirectURI            string
	StravaWebhookVerifyToken          string
	StravaAPIBaseURL                  string
	ProviderTokenKey                  string
	StravaWebhookRetryIntervalSeconds int
}

func Load() Config {
	return Config{
		ServerAddress:                     stringFromEnv(EnvServerAddress, DefaultServerAddress),
		DatabaseURL:                       strings.TrimSpace(os.Getenv(EnvDatabaseURL)),
		StravaClientID:                    strings.TrimSpace(os.Getenv(EnvStravaClientID)),
		StravaClientSecret:                strings.TrimSpace(os.Getenv(EnvStravaClientSecret)),
		StravaOAuthRedirectURI:            strings.TrimSpace(os.Getenv(EnvStravaOAuthRedirectURI)),
		StravaWebhookVerifyToken:          strings.TrimSpace(os.Getenv(EnvStravaWebhookVerifyToken)),
		StravaAPIBaseURL:                  stringFromEnv(EnvStravaAPIBaseURL, DefaultStravaAPIBaseURL),
		ProviderTokenKey:                  strings.TrimSpace(os.Getenv(EnvProviderTokenKey)),
		StravaWebhookRetryIntervalSeconds: positiveIntFromEnv(EnvStravaWebhookRetrySecs),
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

func positiveIntFromEnv(key string) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}
