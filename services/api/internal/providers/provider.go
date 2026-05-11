package providers

import (
	"context"

	"github.com/runthread/runthread/services/api/internal/domain"
)

// ActivityProvider normalises provider-specific activity payloads into the
// provider-neutral domain model used by matching and adaptation.
type ActivityProvider interface {
	ProviderName() string
	NormaliseActivity(ctx context.Context, payload []byte) (domain.ImportedActivity, error)
}
