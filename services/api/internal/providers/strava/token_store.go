package strava

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type InMemoryTokenStore struct {
	mu     sync.Mutex
	tokens map[string]OAuthToken
}

var _ TokenStore = (*InMemoryTokenStore)(nil)

func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		tokens: make(map[string]OAuthToken),
	}
}

func (s *InMemoryTokenStore) StoreToken(ctx context.Context, token OAuthToken) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if s == nil {
		return "", fmt.Errorf("strava token store is required")
	}

	reference := "strava-token-" + uuid.NewString()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens == nil {
		s.tokens = make(map[string]OAuthToken)
	}
	s.tokens[reference] = token
	return reference, nil
}

func (s *InMemoryTokenStore) LoadToken(ctx context.Context, reference string) (OAuthToken, error) {
	if err := ctx.Err(); err != nil {
		return OAuthToken{}, err
	}
	if s == nil {
		return OAuthToken{}, fmt.Errorf("strava token store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.tokens[reference]
	if !ok {
		return OAuthToken{}, fmt.Errorf("strava token reference not found")
	}
	return token, nil
}

func (s *InMemoryTokenStore) UpdateToken(ctx context.Context, reference string, token OAuthToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("strava token store is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens == nil {
		s.tokens = make(map[string]OAuthToken)
	}
	if _, ok := s.tokens[reference]; !ok {
		return fmt.Errorf("strava token reference not found")
	}
	s.tokens[reference] = token
	return nil
}
