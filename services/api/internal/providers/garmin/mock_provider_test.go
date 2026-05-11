package garmin

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	legacygarmin "github.com/runthread/runthread/services/api/internal/garmin"
)

func TestMockProviderNormalisesGarminPayloadThroughActivityProviderBoundary(t *testing.T) {
	body, err := json.Marshal(validMockPayload("running"))
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	activity, err := MockProvider{}.NormaliseActivity(context.Background(), body)
	if err != nil {
		t.Fatalf("NormaliseActivity returned error: %v", err)
	}

	if activity.Type != domain.ActivityTypeRun {
		t.Fatalf("activity type = %q, want run", activity.Type)
	}
	if activity.AthleteID != "athlete-1" {
		t.Fatalf("athlete id = %q, want athlete-1", activity.AthleteID)
	}
	if activity.Distance.Meters != 9000 {
		t.Fatalf("distance = %.0f, want 9000", activity.Distance.Meters)
	}
}

func TestMockProviderHonoursCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := MockProvider{}.NormaliseActivity(ctx, []byte(`{}`))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func validMockPayload(activityType string) legacygarmin.MockActivityPayload {
	return legacygarmin.MockActivityPayload{
		ActivityID:         "garmin-activity-1",
		AthleteID:          "athlete-1",
		GarminActivityType: activityType,
		StartTime:          time.Date(2026, time.June, 16, 7, 30, 0, 0, time.UTC),
		DurationSeconds:    2700,
		DistanceMeters:     9000,
		AverageHeartRate:   148,
	}
}
