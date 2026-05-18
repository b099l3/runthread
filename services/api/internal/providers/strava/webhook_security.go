package strava

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SignatureVerifier struct {
	SigningSecret string
	Tolerance     time.Duration
	Now           func() time.Time
}

var _ WebhookVerifier = SignatureVerifier{}

func (v SignatureVerifier) VerifyWebhook(ctx context.Context, req VerifyWebhookRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if v.SigningSecret == "" {
		return errors.New("strava webhook signing secret is required")
	}
	timestamp, signature, err := parseStravaSignature(req.Signature)
	if err != nil {
		return err
	}
	tolerance := v.Tolerance
	if tolerance <= 0 {
		tolerance = 5 * time.Minute
	}
	now := time.Now
	if v.Now != nil {
		now = v.Now
	}
	signedAt := time.Unix(timestamp, 0)
	if delta := now().Sub(signedAt); delta > tolerance || delta < -tolerance {
		return errors.New("strava webhook signature timestamp outside tolerance")
	}

	mac := hmac.New(sha256.New, []byte(v.SigningSecret))
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(req.Body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return errors.New("invalid Strava webhook signature")
	}
	return nil
}

func parseStravaSignature(header string) (int64, string, error) {
	if strings.TrimSpace(header) == "" {
		return 0, "", errors.New("missing Strava webhook signature")
	}
	parts := strings.Split(header, ",")
	values := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			return 0, "", fmt.Errorf("invalid Strava webhook signature part %q", part)
		}
		values[key] = value
	}
	timestamp, err := strconv.ParseInt(values["t"], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid Strava webhook signature timestamp: %w", err)
	}
	signature := values["v1"]
	if signature == "" {
		return 0, "", errors.New("missing Strava webhook signature digest")
	}
	return timestamp, signature, nil
}

type InMemoryWebhookDeduper struct {
	mu   sync.Mutex
	seen map[string]bool
}

var _ WebhookDeduper = (*InMemoryWebhookDeduper)(nil)

func NewInMemoryWebhookDeduper() *InMemoryWebhookDeduper {
	return &InMemoryWebhookDeduper{seen: make(map[string]bool)}
}

func (d *InMemoryWebhookDeduper) SeenWebhookEvent(ctx context.Context, eventID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.seen[eventID], nil
}

func (d *InMemoryWebhookDeduper) MarkWebhookEventSeen(ctx context.Context, eventID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if eventID == "" {
		return errors.New("strava webhook event id is required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen[eventID] = true
	return nil
}
