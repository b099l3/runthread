package strava

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/repository"
)

const defaultAuthorizeURL = "https://www.strava.com/oauth/authorize"

type TokenStore interface {
	StoreToken(ctx context.Context, token OAuthToken) (string, error)
}

type CodeExchanger interface {
	ExchangeCode(ctx context.Context, req OAuthCodeExchangeRequest) (OAuthToken, error)
}

type OAuthCodeExchangeRequest struct {
	Code        string
	RedirectURI string
}

type OAuthToken struct {
	ProviderConnectionID string
	ProviderUserID       string
	AccessToken          string
	RefreshToken         string
	Scopes               []string
	ExpiresAt            time.Time
}

type OAuthService struct {
	Store        repository.ProviderStore
	Exchanger    CodeExchanger
	Tokens       TokenStore
	ClientID     string
	AuthorizeURL string
	Now          func() time.Time
	NewState     func() (string, error)
}

type StartOAuthRequest struct {
	AthleteID   string
	RedirectURI string
	Scopes      []string
}

type StartOAuthResponse struct {
	Connection       repository.ProviderConnection
	AuthorizationURL string
	State            string
}

type CompleteOAuthCallbackRequest struct {
	State       string
	Code        string
	RedirectURI string
}

type CompleteOAuthCallbackResponse struct {
	Connection repository.ProviderConnection
}

func (s OAuthService) StartOAuth(ctx context.Context, req StartOAuthRequest) (StartOAuthResponse, error) {
	if err := s.validateStartRequest(req); err != nil {
		return StartOAuthResponse{}, err
	}

	connection, found, err := s.pendingConnection(ctx, req.AthleteID)
	if err != nil {
		return StartOAuthResponse{}, err
	}
	now := s.now()
	if !found {
		connection = repository.ProviderConnection{
			ID:        uuid.NewString(),
			AthleteID: req.AthleteID,
			Provider:  ProviderName,
			Status:    repository.ProviderConnectionStatusPending,
			CreatedAt: now,
		}
	}

	state, err := s.newState()
	if err != nil {
		return StartOAuthResponse{}, fmt.Errorf("generate OAuth state: %w", err)
	}
	connection.LastImportCursor = state
	connection.UpdatedAt = now
	if err := s.Store.SaveProviderConnection(ctx, connection); err != nil {
		return StartOAuthResponse{}, fmt.Errorf("save provider connection: %w", err)
	}

	return StartOAuthResponse{
		Connection:       connection,
		AuthorizationURL: s.authorizationURL(req.RedirectURI, state, req.Scopes),
		State:            state,
	}, nil
}

func (s OAuthService) CompleteOAuthCallback(ctx context.Context, req CompleteOAuthCallbackRequest) (CompleteOAuthCallbackResponse, error) {
	if err := s.validateCallbackRequest(req); err != nil {
		return CompleteOAuthCallbackResponse{}, err
	}

	connection, err := s.connectionForState(ctx, req.State)
	if err != nil {
		return CompleteOAuthCallbackResponse{}, err
	}

	token, err := s.Exchanger.ExchangeCode(ctx, OAuthCodeExchangeRequest{
		Code:        req.Code,
		RedirectURI: req.RedirectURI,
	})
	if err != nil {
		return CompleteOAuthCallbackResponse{}, s.markConnectionError(ctx, connection, fmt.Sprintf("exchange Strava authorization code: %v", err))
	}
	if err := validateOAuthToken(token); err != nil {
		return CompleteOAuthCallbackResponse{}, s.markConnectionError(ctx, connection, fmt.Sprintf("invalid Strava token response: %v", err))
	}
	token.ProviderConnectionID = connection.ID
	tokenReference, err := s.Tokens.StoreToken(ctx, token)
	if err != nil {
		return CompleteOAuthCallbackResponse{}, s.markConnectionError(ctx, connection, fmt.Sprintf("store Strava token: %v", err))
	}

	now := s.now()
	connection.ProviderUserID = token.ProviderUserID
	connection.Status = repository.ProviderConnectionStatusConnected
	connection.ConnectedAt = now
	connection.TokenReference = tokenReference
	connection.TokenExpiresAt = token.ExpiresAt
	connection.LastImportCursor = ""
	connection.LastError = ""
	connection.UpdatedAt = now
	if err := s.Store.SaveProviderConnection(ctx, connection); err != nil {
		return CompleteOAuthCallbackResponse{}, fmt.Errorf("save connected provider connection: %w", err)
	}

	return CompleteOAuthCallbackResponse{
		Connection: connection,
	}, nil
}

func (s OAuthService) validateStartRequest(req StartOAuthRequest) error {
	if s.Store == nil {
		return fmt.Errorf("provider store is required")
	}
	if s.Tokens == nil {
		return fmt.Errorf("token store is required")
	}
	if strings.TrimSpace(s.ClientID) == "" {
		return fmt.Errorf("strava client id is required")
	}
	if strings.TrimSpace(req.AthleteID) == "" {
		return fmt.Errorf("athlete id is required")
	}
	if strings.TrimSpace(req.RedirectURI) == "" {
		return fmt.Errorf("redirect uri is required")
	}
	return nil
}

func (s OAuthService) validateCallbackRequest(req CompleteOAuthCallbackRequest) error {
	if s.Store == nil {
		return fmt.Errorf("provider store is required")
	}
	if s.Exchanger == nil {
		return fmt.Errorf("code exchanger is required")
	}
	if s.Tokens == nil {
		return fmt.Errorf("token store is required")
	}
	if strings.TrimSpace(req.State) == "" {
		return fmt.Errorf("state is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return fmt.Errorf("authorization code is required")
	}
	return nil
}

func validateOAuthToken(token OAuthToken) error {
	if strings.TrimSpace(token.ProviderUserID) == "" {
		return fmt.Errorf("strava provider user id is required")
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return fmt.Errorf("access token is required")
	}
	if strings.TrimSpace(token.RefreshToken) == "" {
		return fmt.Errorf("refresh token is required")
	}
	return nil
}

func (s OAuthService) pendingConnection(ctx context.Context, athleteID string) (repository.ProviderConnection, bool, error) {
	connections, err := s.Store.ListProviderConnectionsByAthlete(ctx, athleteID)
	if err != nil {
		return repository.ProviderConnection{}, false, fmt.Errorf("list provider connections: %w", err)
	}
	for _, connection := range connections {
		if connection.Provider == ProviderName && connection.Status == repository.ProviderConnectionStatusPending {
			return connection, true, nil
		}
	}
	return repository.ProviderConnection{}, false, nil
}

func (s OAuthService) connectionForState(ctx context.Context, state string) (repository.ProviderConnection, error) {
	connections, err := s.Store.ListProviderConnectionsByStatus(ctx, repository.ProviderConnectionStatusPending)
	if err != nil {
		return repository.ProviderConnection{}, fmt.Errorf("list pending provider connections: %w", err)
	}
	for _, connection := range connections {
		if connection.Provider == ProviderName && connection.LastImportCursor == state {
			return connection, nil
		}
	}
	return repository.ProviderConnection{}, fmt.Errorf("pending Strava connection for state not found")
}

func (s OAuthService) markConnectionError(ctx context.Context, connection repository.ProviderConnection, message string) error {
	connection.Status = repository.ProviderConnectionStatusError
	connection.LastError = message
	connection.UpdatedAt = s.now()
	if err := s.Store.SaveProviderConnection(ctx, connection); err != nil {
		return fmt.Errorf("save errored provider connection: %w", err)
	}
	return errors.New(message)
}

func (s OAuthService) authorizationURL(redirectURI string, state string, scopes []string) string {
	base := s.AuthorizeURL
	if base == "" {
		base = defaultAuthorizeURL
	}
	values := url.Values{}
	values.Set("client_id", s.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	values.Set("approval_prompt", "auto")
	values.Set("state", state)
	values.Set("scope", strings.Join(normaliseScopes(scopes), ","))
	return base + "?" + values.Encode()
}

func normaliseScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return []string{"activity:read_all"}
	}
	seen := make(map[string]struct{}, len(scopes))
	normalised := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		normalised = append(normalised, scope)
	}
	if len(normalised) == 0 {
		return []string{"activity:read_all"}
	}
	return normalised
}

func (s OAuthService) newState() (string, error) {
	if s.NewState != nil {
		return s.NewState()
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func (s OAuthService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
