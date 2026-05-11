package garmin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/runthread/runthread/services/api/internal/domain"
	legacygarmin "github.com/runthread/runthread/services/api/internal/garmin"
	"github.com/runthread/runthread/services/api/internal/providers"
)

type MockProvider struct{}

var _ providers.ActivityProvider = MockProvider{}

func (MockProvider) ProviderName() string {
	return ProviderName
}

func (MockProvider) NormaliseActivity(ctx context.Context, payload []byte) (domain.ImportedActivity, error) {
	select {
	case <-ctx.Done():
		return domain.ImportedActivity{}, ctx.Err()
	default:
	}

	var activity legacygarmin.MockActivityPayload
	if err := json.Unmarshal(payload, &activity); err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("decode mock Garmin activity payload: %w", err)
	}
	return legacygarmin.NormalizeMockActivity(activity)
}
