package garmin

import (
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

func TestNormalizeMockActivityMapsRoadRun(t *testing.T) {
	activity := mustNormalize(t, MockActivityPayload{
		ActivityID:         "garmin-activity-1",
		AthleteID:          "athlete-1",
		GarminActivityType: "running",
		StartTime:          time.Date(2026, time.June, 2, 7, 30, 0, 0, time.UTC),
		DurationSeconds:    2700,
		DistanceMeters:     9000,
		AverageHeartRate:   148,
	})

	if activity.ID != "mock-imported-garmin-activity-1" {
		t.Fatalf("expected normalised id, got %q", activity.ID)
	}
	if activity.Type != domain.ActivityTypeRun {
		t.Fatalf("expected run activity type, got %q", activity.Type)
	}
	if activity.Duration != 45*time.Minute {
		t.Fatalf("expected 45 minute duration, got %s", activity.Duration)
	}
	if activity.Distance.Meters != 9000 {
		t.Fatalf("expected 9000m, got %.0fm", activity.Distance.Meters)
	}
	if activity.AveragePace.SecondsPerKilometer != 300 {
		t.Fatalf("expected 300 sec/km pace, got %d", activity.AveragePace.SecondsPerKilometer)
	}
	if activity.AverageHeartBPM != 148 {
		t.Fatalf("expected average heart rate 148, got %d", activity.AverageHeartBPM)
	}
}

func TestNormalizeMockActivityMapsTrailAndTreadmillRuns(t *testing.T) {
	trail := mustNormalize(t, validPayload("trail_running"))
	if trail.Type != domain.ActivityTypeTrailRun {
		t.Fatalf("expected trail run, got %q", trail.Type)
	}

	treadmill := mustNormalize(t, validPayload("treadmill_running"))
	if treadmill.Type != domain.ActivityTypeTreadmill {
		t.Fatalf("expected treadmill, got %q", treadmill.Type)
	}
}

func TestNormalizeMockActivityMapsUnknownTypeToOther(t *testing.T) {
	activity := mustNormalize(t, validPayload("indoor_cardio"))

	if activity.Type != domain.ActivityTypeOther {
		t.Fatalf("expected other activity type, got %q", activity.Type)
	}
}

func TestNormalizeMockActivityRejectsInvalidPayload(t *testing.T) {
	_, err := NormalizeMockActivity(MockActivityPayload{
		ActivityID:         "garmin-activity-1",
		AthleteID:          "athlete-1",
		GarminActivityType: "running",
		StartTime:          time.Date(2026, time.June, 2, 7, 30, 0, 0, time.UTC),
		DurationSeconds:    0,
		DistanceMeters:     9000,
	})

	assertNormalizeError(t, err, "duration seconds")
}

func TestNormalizeMockActivityRejectsMissingProviderID(t *testing.T) {
	payload := validPayload("running")
	payload.ActivityID = ""

	_, err := NormalizeMockActivity(payload)

	assertNormalizeError(t, err, "garmin activity id")
}

func mustNormalize(t *testing.T, payload MockActivityPayload) domain.ImportedActivity {
	t.Helper()

	activity, err := NormalizeMockActivity(payload)
	if err != nil {
		t.Fatalf("expected normalised activity: %v", err)
	}
	return activity
}

func validPayload(activityType string) MockActivityPayload {
	return MockActivityPayload{
		ActivityID:         "garmin-activity-1",
		AthleteID:          "athlete-1",
		GarminActivityType: activityType,
		StartTime:          time.Date(2026, time.June, 2, 7, 30, 0, 0, time.UTC),
		DurationSeconds:    2700,
		DistanceMeters:     9000,
		AverageHeartRate:   148,
	}
}

func assertNormalizeError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
