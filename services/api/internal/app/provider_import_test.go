package app

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestCompleteProviderImportUsesInMemoryRepositories(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	workout := providerImportFirstRunWorkout(t, week)
	connection := providerImportConnection(profile.ID)

	seedProviderImportRecords(t, ctx, store, profile, goal, week, connection)

	response, err := providerImportService(t, store).CompleteProviderImport(ctx, CompleteProviderImportRequest{
		Import: providerimport.ImportRequest{
			AthleteID:            profile.ID,
			ProviderConnectionID: connection.ID,
			ProviderActivity: providerimport.ProviderActivityInput{
				ProviderActivityID:   "garmin-provider-import-1",
				ProviderActivityType: "running",
				StartedAt:            workout.ScheduledFor.Add(7 * time.Hour),
			},
			ImportedActivity: providerImportActivityForWorkout(profile.ID, workout),
			RawPayload:       []byte(`{"activityId":"garmin-provider-import-1"}`),
			PayloadKind:      "activity",
			Delivery: providerimport.DeliveryMetadata{
				EventType:  "mock_activity_import",
				DeliveryID: "delivery-provider-import-1",
				ReceivedAt: providerImportDateTime(2026, time.June, 2, 8, 0),
			},
		},
		PlanWeekID:       week.ID,
		PlannedWorkoutID: workout.ID,
		ResultID:         "result-provider-import-1",
		Outcome:          domain.WorkoutOutcomeUnderperformed,
	})
	if err != nil {
		t.Fatalf("CompleteProviderImport returned error: %v", err)
	}

	if response.ProviderImport.ProviderActivity.ImportedActivityID != response.ProviderImport.ImportedActivity.ID {
		t.Fatalf("provider activity imported id = %q, want %q", response.ProviderImport.ProviderActivity.ImportedActivityID, response.ProviderImport.ImportedActivity.ID)
	}
	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("workout match status = %q, want matched", response.WorkoutMatch.Status)
	}
	if response.UpdatedWorkout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("updated workout status = %q, want completed", response.UpdatedWorkout.Status)
	}
	if response.WorkoutResult.Outcome != domain.WorkoutOutcomeUnderperformed {
		t.Fatalf("workout result outcome = %q, want underperformed", response.WorkoutResult.Outcome)
	}
	if response.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}

	assertProviderImportCompletionPersisted(t, ctx, store, response)
}

func TestCompleteProviderImportFindsBestWorkoutWhenWorkoutIDIsOmitted(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	workout := providerImportFirstRunWorkout(t, week)
	connection := providerImportConnection(profile.ID)
	seedProviderImportRecords(t, ctx, store, profile, goal, week, connection)

	response, err := providerImportService(t, store).CompleteProviderImport(ctx, CompleteProviderImportRequest{
		Import: providerimport.ImportRequest{
			AthleteID:            profile.ID,
			ProviderConnectionID: connection.ID,
			ProviderActivity: providerimport.ProviderActivityInput{
				ProviderActivityID:   "garmin-provider-import-2",
				ProviderActivityType: "running",
				StartedAt:            workout.ScheduledFor.Add(7 * time.Hour),
			},
			ImportedActivity: providerImportActivityForWorkout(profile.ID, workout),
		},
		PlanWeek: week,
		ResultID: "result-provider-import-2",
	})
	if err != nil {
		t.Fatalf("CompleteProviderImport returned error: %v", err)
	}
	if response.UpdatedWorkout.ID != workout.ID {
		t.Fatalf("updated workout id = %q, want %q", response.UpdatedWorkout.ID, workout.ID)
	}
	if response.AdaptationEvent != nil {
		t.Fatalf("expected no adaptation event, got %#v", response.AdaptationEvent)
	}
}

func TestCompleteProviderImportPersistsRideCrossTrainingCompletion(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	workout := providerImportFirstRunWorkout(t, week)
	connection := providerImportConnection(profile.ID)
	seedProviderImportRecords(t, ctx, store, profile, goal, week, connection)

	ride := providerImportActivityForWorkout(profile.ID, workout)
	ride.ID = "imported-ride-cross-training"
	ride.Type = domain.ActivityTypeRide
	ride.Distance = domain.Distance{Meters: workout.TargetDistance.Meters * 3}

	response, err := providerImportService(t, store).CompleteProviderImport(ctx, CompleteProviderImportRequest{
		Import: providerimport.ImportRequest{
			AthleteID:            profile.ID,
			ProviderConnectionID: connection.ID,
			ProviderActivity: providerimport.ProviderActivityInput{
				ProviderActivityID:   "garmin-provider-ride-1",
				ProviderActivityType: "cycling",
				StartedAt:            workout.ScheduledFor.Add(7 * time.Hour),
			},
			ImportedActivity: ride,
		},
		PlanWeek: week,
		ResultID: "result-provider-ride-1",
		Outcome:  domain.WorkoutOutcomeCompletedAsPlanned,
	})
	if err != nil {
		t.Fatalf("CompleteProviderImport returned error: %v", err)
	}

	if response.ProviderImport.ImportedActivity.Type != domain.ActivityTypeRide {
		t.Fatalf("imported activity type = %q, want ride", response.ProviderImport.ImportedActivity.Type)
	}
	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("workout match status = %q, want matched", response.WorkoutMatch.Status)
	}
	if response.WorkoutMatch.Confidence != domain.MatchConfidenceMedium {
		t.Fatalf("workout match confidence = %q, want medium", response.WorkoutMatch.Confidence)
	}
	if response.WorkoutResult.Notes != "Completed by ride cross-training activity." {
		t.Fatalf("workout result notes = %q", response.WorkoutResult.Notes)
	}
	assertProviderImportCompletionPersisted(t, ctx, store, response)
}

func TestCompleteProviderImportRejectsIgnoredImport(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	connection := providerImportConnection(profile.ID)
	seedProviderImportRecords(t, ctx, store, profile, goal, week, connection)

	_, err := providerImportService(t, store).CompleteProviderImport(ctx, CompleteProviderImportRequest{
		Import: providerimport.ImportRequest{
			AthleteID:            profile.ID,
			ProviderConnectionID: connection.ID,
			ProviderActivity: providerimport.ProviderActivityInput{
				ProviderActivityID:   "garmin-provider-import-ignored",
				ProviderActivityType: "walking",
				StartedAt:            providerImportDateTime(2026, time.June, 2, 8, 0),
			},
			IgnoreReason: "unsupported activity type",
		},
		PlanWeek: week,
		ResultID: "result-provider-import-ignored",
	})
	if err == nil {
		t.Fatal("expected ignored provider import to fail completion")
	}
}

func providerImportService(t *testing.T, store *repository.InMemoryStore) ProviderImportService {
	t.Helper()

	service, err := NewProviderImportService(store, store)
	if err != nil {
		t.Fatalf("NewProviderImportService returned error: %v", err)
	}
	service.Matcher = matching.Matcher{Now: providerImportNow}
	service.AdaptationEngine = adaptation.Engine{Now: providerImportNow}
	if concrete, ok := service.Importer.(providerimport.Service); ok {
		concrete.Now = providerImportNow
		service.Importer = concrete
	}
	return service
}

func seedProviderImportRecords(t *testing.T, ctx context.Context, store *repository.InMemoryStore, profile domain.AthleteProfile, goal domain.TrainingGoal, week domain.PlanWeek, connection repository.ProviderConnection) {
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

func assertProviderImportCompletionPersisted(t *testing.T, ctx context.Context, store *repository.InMemoryStore, response CompleteProviderImportResponse) {
	t.Helper()

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
	savedWorkout, err := store.GetPlannedWorkout(ctx, response.UpdatedWorkout.ID)
	if err != nil {
		t.Fatalf("expected saved planned workout: %v", err)
	}
	if savedWorkout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("saved planned workout status = %q, want completed", savedWorkout.Status)
	}
	savedWeek, err := store.GetPlanWeek(ctx, response.PlanWeek.ID)
	if err != nil {
		t.Fatalf("expected saved plan week: %v", err)
	}
	if savedWorkoutStatus(savedWeek, response.UpdatedWorkout.ID) != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("expected saved plan week to include completed workout")
	}
	if response.AdaptationEvent != nil {
		if _, err := store.GetAdaptationEvent(ctx, response.AdaptationEvent.ID); err != nil {
			t.Fatalf("expected saved adaptation event: %v", err)
		}
	}
}

func providerImportProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-provider-import",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func providerImportGoal(athleteID string) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-provider-import",
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     providerImportDate(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func providerImportConnection(athleteID string) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             "connection-provider-import",
		AthleteID:      athleteID,
		Provider:       "garmin",
		ProviderUserID: "garmin-user-provider-import",
		Status:         repository.ProviderConnectionStatusConnected,
		ConnectedAt:    providerImportDateTime(2026, time.June, 1, 8, 0),
	}
}

func providerImportWeek(t *testing.T, profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time) domain.PlanWeek {
	t.Helper()

	week, err := planning.GenerateWeek(profile, goal, targetWeekDate)
	if err != nil {
		t.Fatalf("generate plan week: %v", err)
	}
	return week
}

func providerImportFirstRunWorkout(t *testing.T, week domain.PlanWeek) domain.PlannedWorkout {
	t.Helper()

	for _, workout := range week.Workouts {
		if workout.Type == domain.WorkoutTypeEasy || workout.Type == domain.WorkoutTypeLongRun || workout.Type == domain.WorkoutTypeWorkout {
			return workout
		}
	}
	t.Fatal("expected at least one run workout")
	return domain.PlannedWorkout{}
}

func providerImportActivityForWorkout(athleteID string, workout domain.PlannedWorkout) *domain.ImportedActivity {
	return &domain.ImportedActivity{
		ID:              "imported-" + workout.ID,
		AthleteID:       athleteID,
		Type:            domain.ActivityTypeRun,
		StartedAt:       workout.ScheduledFor.Add(7 * time.Hour),
		Duration:        workout.TargetDuration,
		Distance:        workout.TargetDistance,
		AveragePace:     domain.Pace{SecondsPerKilometer: 330},
		AverageHeartBPM: 145,
	}
}

func providerImportNow() time.Time {
	return providerImportDateTime(2026, time.June, 2, 12, 0)
}

func providerImportDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func providerImportDateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
