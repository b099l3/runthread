package garmin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

type MockActivityPayload struct {
	ActivityID         string
	AthleteID          string
	GarminActivityType string
	StartTime          time.Time
	DurationSeconds    int
	DistanceMeters     float64
	AverageHeartRate   int
}

func NormalizeMockActivity(payload MockActivityPayload) (domain.ImportedActivity, error) {
	if payload.ActivityID == "" {
		return domain.ImportedActivity{}, errors.New("garmin activity id is required")
	}
	if payload.AthleteID == "" {
		return domain.ImportedActivity{}, errors.New("athlete id is required")
	}
	if payload.StartTime.IsZero() {
		return domain.ImportedActivity{}, errors.New("start time is required")
	}
	if payload.DurationSeconds <= 0 {
		return domain.ImportedActivity{}, errors.New("duration seconds must be positive")
	}
	if payload.DistanceMeters < 0 {
		return domain.ImportedActivity{}, errors.New("distance meters cannot be negative")
	}
	if payload.AverageHeartRate < 0 {
		return domain.ImportedActivity{}, errors.New("average heart rate cannot be negative")
	}

	activity := domain.ImportedActivity{
		ID:              fmt.Sprintf("mock-imported-%s", payload.ActivityID),
		AthleteID:       payload.AthleteID,
		Type:            normalizeActivityType(payload.GarminActivityType),
		StartedAt:       payload.StartTime,
		Duration:        time.Duration(payload.DurationSeconds) * time.Second,
		Distance:        domain.Distance{Meters: payload.DistanceMeters},
		AveragePace:     normalizeAveragePace(payload.DistanceMeters, payload.DurationSeconds),
		AverageHeartBPM: payload.AverageHeartRate,
	}
	if err := activity.Validate(); err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("normalised imported activity is invalid: %w", err)
	}
	return activity, nil
}

func normalizeActivityType(garminType string) domain.ActivityType {
	switch strings.ToLower(strings.TrimSpace(garminType)) {
	case "running", "run", "road_running":
		return domain.ActivityTypeRun
	case "trail_running", "trail_run":
		return domain.ActivityTypeTrailRun
	case "treadmill_running", "treadmill":
		return domain.ActivityTypeTreadmill
	case "walking", "walk":
		return domain.ActivityTypeWalk
	case "cycling", "cycle", "biking", "bike", "road_biking", "road_cycling", "indoor_cycling", "virtual_ride", "mountain_biking", "mountain_bike", "gravel_cycling", "gravel_biking":
		return domain.ActivityTypeRide
	default:
		return domain.ActivityTypeOther
	}
}

func normalizeAveragePace(distanceMeters float64, durationSeconds int) domain.Pace {
	if distanceMeters <= 0 || durationSeconds <= 0 {
		return domain.Pace{}
	}
	secondsPerKilometer := int((float64(durationSeconds) / distanceMeters) * 1000)
	return domain.Pace{SecondsPerKilometer: secondsPerKilometer}
}
