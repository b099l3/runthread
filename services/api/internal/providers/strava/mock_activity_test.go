package strava

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

func TestNormaliseMockActivityMapsRun(t *testing.T) {
	activity := mustNormalise(t, validPayload("Run"))

	if activity.ID != "mock-imported-strava-strava-activity-1" {
		t.Fatalf("expected normalised id, got %q", activity.ID)
	}
	if activity.AthleteID != "athlete-1" {
		t.Fatalf("expected athlete-1, got %q", activity.AthleteID)
	}
	if activity.Type != domain.ActivityTypeRun {
		t.Fatalf("expected run activity type, got %q", activity.Type)
	}
	if activity.StartedAt != validStartTime() {
		t.Fatalf("expected start time %s, got %s", validStartTime(), activity.StartedAt)
	}
	if activity.Duration != 44*time.Minute {
		t.Fatalf("expected moving time duration, got %s", activity.Duration)
	}
	if activity.Distance.Meters != 8800 {
		t.Fatalf("expected 8800m, got %.0fm", activity.Distance.Meters)
	}
	if activity.AveragePace.SecondsPerKilometer != 300 {
		t.Fatalf("expected 300 sec/km pace, got %d", activity.AveragePace.SecondsPerKilometer)
	}
	if activity.AverageHeartBPM != 151 {
		t.Fatalf("expected average heart rate 151, got %d", activity.AverageHeartBPM)
	}
}

func TestNormaliseMockActivityMapsTrailAndVirtualRuns(t *testing.T) {
	trail := mustNormalise(t, validPayload("TrailRun"))
	if trail.Type != domain.ActivityTypeTrailRun {
		t.Fatalf("expected trail run, got %q", trail.Type)
	}

	treadmill := mustNormalise(t, validPayload("VirtualRun"))
	if treadmill.Type != domain.ActivityTypeTreadmill {
		t.Fatalf("expected treadmill, got %q", treadmill.Type)
	}
}

func TestNormaliseMockActivityMapsRideVariants(t *testing.T) {
	for _, sportType := range []string{"Ride", "VirtualRide", "virtual_ride", "GravelRide", "MountainBikeRide", "mountain_bike_ride"} {
		activity := mustNormalise(t, validPayload(sportType))
		if activity.Type != domain.ActivityTypeRide {
			t.Fatalf("sport type %q mapped to %q, want ride", sportType, activity.Type)
		}
	}
}

func TestNormaliseMockActivityUsesElapsedTimeWhenMovingTimeMissing(t *testing.T) {
	payload := validPayload("Run")
	payload.MovingTime = 0

	activity := mustNormalise(t, payload)

	if activity.Duration != 45*time.Minute {
		t.Fatalf("expected elapsed time duration, got %s", activity.Duration)
	}
}

func TestNormaliseMockActivityRejectsUnsupportedActivity(t *testing.T) {
	_, err := NormaliseMockActivity(validPayload("Swim"))

	if !errors.Is(err, ErrUnsupportedActivityType) {
		t.Fatalf("expected unsupported activity type error, got %v", err)
	}
}

func TestNormaliseMockActivityRejectsInvalidPayload(t *testing.T) {
	payload := validPayload("Run")
	payload.ElapsedTime = 0

	_, err := NormaliseMockActivity(payload)

	assertNormaliseError(t, err, "elapsed time")
}

func TestNormaliseMockActivityRejectsMissingProviderID(t *testing.T) {
	payload := validPayload("Run")
	payload.ActivityID = ""

	_, err := NormaliseMockActivity(payload)

	assertNormaliseError(t, err, "strava activity id")
}

func TestMockProviderNormalisesJSONPayload(t *testing.T) {
	payload := validPayload("Run")
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	activity, err := MockProvider{}.NormaliseActivity(context.Background(), body)
	if err != nil {
		t.Fatalf("expected normalised activity: %v", err)
	}

	if activity.Type != domain.ActivityTypeRun {
		t.Fatalf("expected run activity type, got %q", activity.Type)
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

func mustNormalise(t *testing.T, payload MockActivityPayload) domain.ImportedActivity {
	t.Helper()

	activity, err := NormaliseMockActivity(payload)
	if err != nil {
		t.Fatalf("expected normalised activity: %v", err)
	}
	return activity
}

func validPayload(sportType string) MockActivityPayload {
	return MockActivityPayload{
		ActivityID:       "strava-activity-1",
		AthleteID:        "athlete-1",
		StravaSportType:  sportType,
		Name:             "Morning Run",
		StartDate:        validStartTime(),
		ElapsedTime:      2700,
		MovingTime:       2640,
		DistanceMeters:   8800,
		AverageHeartRate: 151,
	}
}

func validStartTime() time.Time {
	return time.Date(2026, time.June, 3, 7, 15, 0, 0, time.UTC)
}

func assertNormaliseError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
