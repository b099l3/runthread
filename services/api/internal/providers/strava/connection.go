package strava

import (
	"context"
	"fmt"

	"github.com/runthread/runthread/services/api/internal/app"
)

type ConnectionStarter struct {
	OAuth       *OAuthService
	RedirectURI string
	Scopes      []string
}

var _ app.ProviderConnectionStarter = ConnectionStarter{}

func (s ConnectionStarter) StartProviderConnection(ctx context.Context, req app.StartProviderConnectionRequest) (app.StartProviderConnectionResponse, error) {
	if s.OAuth == nil {
		return app.StartProviderConnectionResponse{}, fmt.Errorf("strava oauth service is required")
	}

	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = s.RedirectURI
	}

	response, err := s.OAuth.StartOAuth(ctx, StartOAuthRequest{
		AthleteID:   req.AthleteID,
		RedirectURI: redirectURI,
		Scopes:      s.Scopes,
	})
	if err != nil {
		return app.StartProviderConnectionResponse{}, err
	}

	return app.StartProviderConnectionResponse{
		Connection:       response.Connection,
		AuthorizationURL: response.AuthorizationURL,
		State:            response.State,
		OAuthReady:       true,
	}, nil
}
