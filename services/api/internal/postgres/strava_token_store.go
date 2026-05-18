package postgres

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/providers/strava"
)

type StravaTokenStore struct {
	db     *sql.DB
	cipher cipher.AEAD
}

var _ strava.TokenStore = (*StravaTokenStore)(nil)
var _ strava.TokenRepository = (*StravaTokenStore)(nil)

func NewStravaTokenStore(db *sql.DB, key string) (*StravaTokenStore, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres db is required")
	}
	aead, err := providerTokenCipher(key)
	if err != nil {
		return nil, err
	}
	return &StravaTokenStore{db: db, cipher: aead}, nil
}

func (s *StravaTokenStore) StoreToken(ctx context.Context, token strava.OAuthToken) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	reference := "strava-token-" + uuid.NewString()
	encrypted, err := s.encryptToken(token)
	if err != nil {
		return "", err
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO provider_tokens (
    reference,
    provider,
    provider_connection_id,
    provider_user_id,
    encrypted_token
) VALUES ($1, $2, $3, $4, $5)
`, reference, strava.ProviderName, token.ProviderConnectionID, token.ProviderUserID, encrypted); err != nil {
		return "", fmt.Errorf("store Strava token: %w", err)
	}
	return reference, nil
}

func (s *StravaTokenStore) LoadToken(ctx context.Context, reference string) (strava.OAuthToken, error) {
	if err := ctx.Err(); err != nil {
		return strava.OAuthToken{}, err
	}
	var encrypted []byte
	err := s.db.QueryRowContext(ctx, `
SELECT encrypted_token
FROM provider_tokens
WHERE reference = $1 AND provider = $2
`, reference, strava.ProviderName).Scan(&encrypted)
	if err == sql.ErrNoRows {
		return strava.OAuthToken{}, fmt.Errorf("strava token reference not found")
	}
	if err != nil {
		return strava.OAuthToken{}, fmt.Errorf("load Strava token: %w", err)
	}
	return s.decryptToken(encrypted)
}

func (s *StravaTokenStore) UpdateToken(ctx context.Context, reference string, token strava.OAuthToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	encrypted, err := s.encryptToken(token)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE provider_tokens
SET
    provider_connection_id = $3,
    provider_user_id = $4,
    encrypted_token = $5,
    updated_at = now()
WHERE reference = $1 AND provider = $2
`, reference, strava.ProviderName, token.ProviderConnectionID, token.ProviderUserID, encrypted)
	if err != nil {
		return fmt.Errorf("update Strava token: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check Strava token update: %w", err)
	}
	if updated == 0 {
		return fmt.Errorf("strava token reference not found")
	}
	return nil
}

func (s *StravaTokenStore) encryptToken(token strava.OAuthToken) ([]byte, error) {
	payload, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshal Strava token: %w", err)
	}
	nonce := make([]byte, s.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate token nonce: %w", err)
	}
	sealed := s.cipher.Seal(nil, nonce, payload, []byte(strava.ProviderName))
	return append(nonce, sealed...), nil
}

func (s *StravaTokenStore) decryptToken(encrypted []byte) (strava.OAuthToken, error) {
	nonceSize := s.cipher.NonceSize()
	if len(encrypted) <= nonceSize {
		return strava.OAuthToken{}, fmt.Errorf("encrypted Strava token is invalid")
	}
	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]
	payload, err := s.cipher.Open(nil, nonce, ciphertext, []byte(strava.ProviderName))
	if err != nil {
		return strava.OAuthToken{}, fmt.Errorf("decrypt Strava token: %w", err)
	}
	var token strava.OAuthToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return strava.OAuthToken{}, fmt.Errorf("decode Strava token: %w", err)
	}
	return token, nil
}

func providerTokenCipher(key string) (cipher.AEAD, error) {
	keyBytes, err := providerTokenKeyBytes(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create token cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create token cipher mode: %w", err)
	}
	return aead, nil
}

func providerTokenKeyBytes(key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("provider token encryption key is required")
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	raw := []byte(key)
	if len(raw) == 32 {
		return raw, nil
	}
	return nil, fmt.Errorf("provider token encryption key must be 32 raw bytes or base64-encoded 32 bytes")
}
