package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultTokenURL = "https://www.strava.com/oauth/token"

type HTTPCodeExchanger struct {
	Client       *http.Client
	ClientID     string
	ClientSecret string
	TokenURL     string
}

var _ CodeExchanger = HTTPCodeExchanger{}

func (e HTTPCodeExchanger) ExchangeCode(ctx context.Context, req OAuthCodeExchangeRequest) (OAuthToken, error) {
	if strings.TrimSpace(e.ClientID) == "" {
		return OAuthToken{}, fmt.Errorf("strava client id is required")
	}
	if strings.TrimSpace(e.ClientSecret) == "" {
		return OAuthToken{}, fmt.Errorf("strava client secret is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return OAuthToken{}, fmt.Errorf("authorization code is required")
	}

	tokenURL := e.TokenURL
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	form := url.Values{}
	form.Set("client_id", e.ClientID)
	form.Set("client_secret", e.ClientSecret)
	form.Set("code", req.Code)
	form.Set("grant_type", "authorization_code")

	return e.exchangeToken(ctx, tokenURL, form)
}

func (e HTTPCodeExchanger) RefreshToken(ctx context.Context, refreshToken string) (OAuthToken, error) {
	if strings.TrimSpace(e.ClientID) == "" {
		return OAuthToken{}, fmt.Errorf("strava client id is required")
	}
	if strings.TrimSpace(e.ClientSecret) == "" {
		return OAuthToken{}, fmt.Errorf("strava client secret is required")
	}
	if strings.TrimSpace(refreshToken) == "" {
		return OAuthToken{}, fmt.Errorf("refresh token is required")
	}

	tokenURL := e.TokenURL
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	form := url.Values{}
	form.Set("client_id", e.ClientID)
	form.Set("client_secret", e.ClientSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", "refresh_token")

	return e.exchangeToken(ctx, tokenURL, form)
}

func (e HTTPCodeExchanger) exchangeToken(ctx context.Context, tokenURL string, form url.Values) (OAuthToken, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthToken{}, fmt.Errorf("build Strava token exchange request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := e.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(httpReq)
	if err != nil {
		return OAuthToken{}, fmt.Errorf("call Strava token exchange: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return OAuthToken{}, fmt.Errorf("read Strava token exchange response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return OAuthToken{}, fmt.Errorf("Strava token exchange returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload tokenExchangeResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return OAuthToken{}, fmt.Errorf("decode Strava token exchange response: %w", err)
	}
	return payload.oauthToken()
}

type tokenExchangeResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresAt    int64           `json:"expires_at"`
	Scope        string          `json:"scope"`
	Athlete      tokenAthleteRef `json:"athlete"`
}

type tokenAthleteRef struct {
	ID json.Number `json:"id"`
}

func (r tokenExchangeResponse) oauthToken() (OAuthToken, error) {
	providerUserID := r.Athlete.ID.String()
	if providerUserID != "" {
		if _, err := strconv.ParseInt(providerUserID, 10, 64); err != nil {
			return OAuthToken{}, fmt.Errorf("invalid strava athlete id: %w", err)
		}
	}
	return OAuthToken{
		ProviderUserID: providerUserID,
		AccessToken:    r.AccessToken,
		RefreshToken:   r.RefreshToken,
		Scopes:         normaliseScopes(strings.Split(r.Scope, ",")),
		ExpiresAt:      time.Unix(r.ExpiresAt, 0).UTC(),
	}, nil
}
