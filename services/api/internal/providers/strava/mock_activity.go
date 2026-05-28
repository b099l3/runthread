package strava

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/providers"
)

var ErrUnsupportedActivityType = errors.New("unsupported Strava activity type")

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

	var activity MockActivityPayload
	if err := json.Unmarshal(payload, &activity); err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("decode mock Strava activity payload: %w", err)
	}
	return NormaliseMockActivity(activity)
}

type MockActivityPayload struct {
	ActivityID       string    `json:"id"`
	AthleteID        string    `json:"athlete_id"`
	StravaSportType  string    `json:"sport_type"`
	Name             string    `json:"name"`
	StartDate        time.Time `json:"start_date"`
	ElapsedTime      int       `json:"elapsed_time"`
	MovingTime       int       `json:"moving_time"`
	DistanceMeters   float64   `json:"distance"`
	AverageHeartRate int       `json:"average_heartrate"`
}

func NormaliseMockActivity(payload MockActivityPayload) (domain.ImportedActivity, error) {
	if payload.ActivityID == "" {
		return domain.ImportedActivity{}, errors.New("strava activity id is required")
	}
	if payload.AthleteID == "" {
		return domain.ImportedActivity{}, errors.New("athlete id is required")
	}
	if payload.StartDate.IsZero() {
		return domain.ImportedActivity{}, errors.New("start date is required")
	}
	if payload.ElapsedTime <= 0 {
		return domain.ImportedActivity{}, errors.New("elapsed time must be positive")
	}
	if payload.MovingTime < 0 {
		return domain.ImportedActivity{}, errors.New("moving time cannot be negative")
	}
	if payload.DistanceMeters < 0 {
		return domain.ImportedActivity{}, errors.New("distance meters cannot be negative")
	}
	if payload.AverageHeartRate < 0 {
		return domain.ImportedActivity{}, errors.New("average heart rate cannot be negative")
	}

	activityType, err := normaliseSportType(payload.StravaSportType)
	if err != nil {
		return domain.ImportedActivity{}, err
	}

	durationSeconds := payload.MovingTime
	if durationSeconds == 0 {
		durationSeconds = payload.ElapsedTime
	}

	activity := domain.ImportedActivity{
		ID:              fmt.Sprintf("mock-imported-strava-%s", payload.ActivityID),
		AthleteID:       payload.AthleteID,
		Type:            activityType,
		StartedAt:       payload.StartDate,
		Duration:        time.Duration(durationSeconds) * time.Second,
		Distance:        domain.Distance{Meters: payload.DistanceMeters},
		AveragePace:     normaliseAveragePace(payload.DistanceMeters, durationSeconds),
		AverageHeartBPM: payload.AverageHeartRate,
	}
	if err := activity.Validate(); err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("normalised imported activity is invalid: %w", err)
	}
	return activity, nil
}

func normaliseSportType(sportType string) (domain.ActivityType, error) {
	switch strings.ToLower(strings.TrimSpace(sportType)) {
	case "run":
		return domain.ActivityTypeRun, nil
	case "trailrun", "trail_run":
		return domain.ActivityTypeTrailRun, nil
	case "virtualrun", "virtual_run":
		return domain.ActivityTypeTreadmill, nil
	case "ride", "virtualride", "virtual_ride", "gravelride", "gravel_ride", "mountainbikeride", "mountain_bike_ride", "mtb":
		return domain.ActivityTypeRide, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedActivityType, sportType)
	}
}

func normaliseAveragePace(distanceMeters float64, durationSeconds int) domain.Pace {
	if distanceMeters <= 0 || durationSeconds <= 0 {
		return domain.Pace{}
	}
	secondsPerKilometer := int((float64(durationSeconds) / distanceMeters) * 1000)
	return domain.Pace{SecondsPerKilometer: secondsPerKilometer}
}
