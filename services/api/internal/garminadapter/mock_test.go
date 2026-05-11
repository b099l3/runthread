package garminadapter

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/garmin"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestMockGarminAdapterFeedsProviderImportAppService(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := adapterProfile()
	goal := adapterGoal(profile.ID)
	week := adapterWeek(t, profile, goal, adapterDate(2026, time.June, 3))
	workout := adapterFirstRunWorkout(t, week)
	connection := adapterConnection(profile.ID)

	seedAdapterRecords(t, ctx, store, profile, goal, week, connection)

	request, err := BuildMockCompleteProviderImportRequest(MockCompleteProviderImportInput{
		AthleteID:            profile.ID,
		ProviderConnectionID: connection.ID,
		Payload:              adapterPayloadForWorkout(profile.ID, workout),
		RawPayload:           []byte(`{"activityId":"garmin-adapter-activity"}`),
		ReceivedAt:           adapterDateTime(2026, time.June, 2, 8, 0),
		DeliveryID:           "delivery-adapter-1",
		PlanWeekID:           week.ID,
		PlannedWorkoutID:     workout.ID,
		ResultID:             "result-adapter-1",
		Outcome:              domain.WorkoutOutcomeUnderperformed,
	})
	if err != nil {
		t.Fatalf("BuildMockCompleteProviderImportRequest returned error: %v", err)
	}

	service := adapterProviderImportService(t, store)
	response, err := service.CompleteProviderImport(ctx, request)
	if err != nil {
		t.Fatalf("CompleteProviderImport returned error: %v", err)
	}

	if response.ProviderImport.ProviderActivity.ImportedActivityID != response.ProviderImport.ImportedActivity.ID {
		t.Fatalf("provider activity imported id = %q, want %q", response.ProviderImport.ProviderActivity.ImportedActivityID, response.ProviderImport.ImportedActivity.ID)
	}
	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("match status = %q, want matched", response.WorkoutMatch.Status)
	}
	if response.WorkoutResult.Outcome != domain.WorkoutOutcomeUnderperformed {
		t.Fatalf("workout result outcome = %q, want underperformed", response.WorkoutResult.Outcome)
	}
	if response.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}

	if _, err := store.GetProviderActivity(ctx, response.ProviderImport.ProviderActivity.ID); err != nil {
		t.Fatalf("expected saved provider activity: %v", err)
	}
	if _, err := store.GetImportedActivity(ctx, response.ProviderImport.ImportedActivity.ID); err != nil {
		t.Fatalf("expected saved imported activity: %v", err)
	}
	if _, err := store.GetWorkoutMatch(ctx, response.WorkoutMatch.ID); err != nil {
		t.Fatalf("expected saved workout match: %v", err)
	}
	if _, err := store.GetWorkoutResult(ctx, response.WorkoutResult.ID); err != nil {
		t.Fatalf("expected saved workout result: %v", err)
	}
	if _, err := store.GetAdaptationEvent(ctx, response.AdaptationEvent.ID); err != nil {
		t.Fatalf("expected saved adaptation event: %v", err)
	}
	payloads, err := store.ListProviderActivityPayloads(ctx, response.ProviderImport.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("list provider payloads: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("payload count = %d, want 1", len(payloads))
	}
}

func adapterProviderImportService(t *testing.T, store *repository.InMemoryStore) app.ProviderImportService {
	t.Helper()

	service, err := app.NewProviderImportService(store, store)
	if err != nil {
		t.Fatalf("NewProviderImportService returned error: %v", err)
	}
	service.Matcher = matching.Matcher{Now: adapterNow}
	service.AdaptationEngine = adaptation.Engine{Now: adapterNow}
	if importer, ok := service.Importer.(providerimport.Service); ok {
		importer.Now = adapterNow
		service.Importer = importer
	}
	return service
}

func seedAdapterRecords(t *testing.T, ctx context.Context, store *repository.InMemoryStore, profile domain.AthleteProfile, goal domain.TrainingGoal, week domain.PlanWeek, connection repository.ProviderConnection) {
	t.Helper()

	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("save athlete profile: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("save training goal: %v", err)
	}
	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("save plan week: %v", err)
	}
	for _, workout := range week.Workouts {
		if err := store.SavePlannedWorkout(ctx, workout); err != nil {
			t.Fatalf("save planned workout %q: %v", workout.ID, err)
		}
	}
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}
}

func adapterProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-adapter",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func adapterGoal(athleteID string) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-adapter",
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     adapterDate(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func adapterConnection(athleteID string) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             "connection-adapter",
		AthleteID:      athleteID,
		Provider:       garmin.ProviderNameGarmin,
		ProviderUserID: "garmin-user-adapter",
		Status:         repository.ProviderConnectionStatusConnected,
		ConnectedAt:    adapterDateTime(2026, time.June, 1, 8, 0),
	}
}

func adapterWeek(t *testing.T, profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time) domain.PlanWeek {
	t.Helper()

	week, err := planning.GenerateWeek(profile, goal, targetWeekDate)
	if err != nil {
		t.Fatalf("generate plan week: %v", err)
	}
	return week
}

func adapterFirstRunWorkout(t *testing.T, week domain.PlanWeek) domain.PlannedWorkout {
	t.Helper()

	for _, workout := range week.Workouts {
		if workout.Type == domain.WorkoutTypeEasy || workout.Type == domain.WorkoutTypeLongRun || workout.Type == domain.WorkoutTypeWorkout {
			return workout
		}
	}
	t.Fatal("expected at least one run workout")
	return domain.PlannedWorkout{}
}

func adapterPayloadForWorkout(athleteID string, workout domain.PlannedWorkout) garmin.MockActivityPayload {
	return garmin.MockActivityPayload{
		ActivityID:         "garmin-adapter-activity",
		AthleteID:          athleteID,
		GarminActivityType: "running",
		StartTime:          workout.ScheduledFor.Add(7 * time.Hour),
		DurationSeconds:    int(workout.TargetDuration.Seconds()),
		DistanceMeters:     workout.TargetDistance.Meters,
		AverageHeartRate:   145,
	}
}

func adapterNow() time.Time {
	return adapterDateTime(2026, time.June, 2, 12, 0)
}

func adapterDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func adapterDateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
