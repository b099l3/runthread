package garmin

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestMockGarminImportCanDriveCoreLoopInMemory(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := e2eAthleteProfile()
	goal := e2eTrainingGoal(profile.ID)
	targetWeekDate := e2eDate(2026, time.June, 3)
	week := e2ePlanWeek(t, profile, goal, targetWeekDate)
	workout := e2eFirstRunWorkout(t, week)
	connection := e2eProviderConnection(profile.ID)

	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("save athlete profile: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("save training goal: %v", err)
	}
	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("save plan week: %v", err)
	}
	for _, planned := range week.Workouts {
		if err := store.SavePlannedWorkout(ctx, planned); err != nil {
			t.Fatalf("save planned workout %q: %v", planned.ID, err)
		}
	}
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("save provider connection: %v", err)
	}

	importer := newTestMockImportService(t, store)
	importResult, err := importer.ImportActivity(ctx, MockImportRequest{
		Connection: connection,
		Payload:    e2ePayloadForWorkout(profile.ID, workout),
		RawPayload: []byte(`{"activityId":"garmin-e2e-activity"}`),
		ReceivedAt: e2eDateTime(2026, time.June, 2, 8, 0),
		DeliveryID: "delivery-e2e-1",
	})
	if err != nil {
		t.Fatalf("import mock Garmin activity: %v", err)
	}

	match, err := matching.Matcher{Now: e2eNow}.MatchActivity(workout, importResult.ImportedActivity)
	if err != nil {
		t.Fatalf("match imported activity: %v", err)
	}
	if match.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("match status = %q, want matched", match.Status)
	}
	if err := store.SaveWorkoutMatch(ctx, match); err != nil {
		t.Fatalf("save workout match: %v", err)
	}

	updatedWorkout, result, err := domain.MarkWorkoutCompleted(workout, domain.WorkoutCompletion{
		ResultID:           "result-garmin-e2e-1",
		ImportedActivityID: importResult.ImportedActivity.ID,
		CompletedAt:        importResult.ImportedActivity.StartedAt.Add(importResult.ImportedActivity.Duration),
		Distance:           importResult.ImportedActivity.Distance,
		Duration:           importResult.ImportedActivity.Duration,
		Outcome:            domain.WorkoutOutcomeUnderperformed,
		Notes:              "Created by mock Garmin import e2e test.",
	})
	if err != nil {
		t.Fatalf("mark workout completed: %v", err)
	}
	if err := store.SaveWorkoutResult(ctx, result); err != nil {
		t.Fatalf("save workout result: %v", err)
	}
	if err := store.SavePlannedWorkout(ctx, updatedWorkout); err != nil {
		t.Fatalf("save updated workout: %v", err)
	}

	updatedWeek := e2eReplaceWorkout(week, updatedWorkout)
	adaptationEvent, err := (adaptation.Engine{Now: e2eNow}).AdaptWorkoutResult(adaptation.WorkoutResultInput{
		AthleteID: profile.ID,
		PlanWeek:  updatedWeek,
		Result:    result,
	})
	if err != nil {
		t.Fatalf("run adaptation logic: %v", err)
	}
	if adaptationEvent == nil {
		t.Fatal("expected underperformed result to produce adaptation event")
	}
	if err := store.SaveAdaptationEvent(ctx, *adaptationEvent); err != nil {
		t.Fatalf("save adaptation event: %v", err)
	}

	assertE2EProviderImportPersisted(t, ctx, store, importResult)
	if _, err := store.GetWorkoutMatch(ctx, match.ID); err != nil {
		t.Fatalf("expected saved workout match: %v", err)
	}
	if _, err := store.GetWorkoutResult(ctx, result.ID); err != nil {
		t.Fatalf("expected saved workout result: %v", err)
	}
	if _, err := store.GetAdaptationEvent(ctx, adaptationEvent.ID); err != nil {
		t.Fatalf("expected saved adaptation event: %v", err)
	}
}

func assertE2EProviderImportPersisted(t *testing.T, ctx context.Context, store *repository.InMemoryStore, result MockImportResult) {
	t.Helper()

	providerActivity, err := store.GetProviderActivity(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("expected saved provider activity: %v", err)
	}
	if providerActivity.ImportedActivityID != result.ImportedActivity.ID {
		t.Fatalf("provider activity imported id = %q, want %q", providerActivity.ImportedActivityID, result.ImportedActivity.ID)
	}
	if _, err := store.GetImportedActivity(ctx, result.ImportedActivity.ID); err != nil {
		t.Fatalf("expected saved imported activity: %v", err)
	}
	events, err := store.ListProviderImportEventsByConnection(ctx, result.ProviderActivity.ProviderConnectionID)
	if err != nil {
		t.Fatalf("list provider import events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("provider import event count = %d, want 1", len(events))
	}
	if events[0].Status != repository.ProviderImportEventStatusProcessed {
		t.Fatalf("provider import event status = %q, want processed", events[0].Status)
	}
	payloads, err := store.ListProviderActivityPayloads(ctx, result.ProviderActivity.ID)
	if err != nil {
		t.Fatalf("list provider activity payloads: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("provider payload count = %d, want 1", len(payloads))
	}
}

func e2eAthleteProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-garmin-e2e",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func e2eTrainingGoal(athleteID string) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-garmin-e2e",
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     e2eDate(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func e2eProviderConnection(athleteID string) repository.ProviderConnection {
	return repository.ProviderConnection{
		ID:             "connection-garmin-e2e",
		AthleteID:      athleteID,
		Provider:       ProviderNameGarmin,
		ProviderUserID: "garmin-user-e2e",
		Status:         repository.ProviderConnectionStatusConnected,
		ConnectedAt:    e2eDateTime(2026, time.June, 1, 8, 0),
	}
}

func e2ePlanWeek(t *testing.T, profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time) domain.PlanWeek {
	t.Helper()

	week, err := planning.GenerateWeek(profile, goal, targetWeekDate)
	if err != nil {
		t.Fatalf("generate plan week: %v", err)
	}
	return week
}

func e2eFirstRunWorkout(t *testing.T, week domain.PlanWeek) domain.PlannedWorkout {
	t.Helper()

	for _, workout := range week.Workouts {
		if workout.Type == domain.WorkoutTypeEasy || workout.Type == domain.WorkoutTypeLongRun || workout.Type == domain.WorkoutTypeWorkout {
			return workout
		}
	}
	t.Fatal("expected at least one run workout")
	return domain.PlannedWorkout{}
}

func e2ePayloadForWorkout(athleteID string, workout domain.PlannedWorkout) MockActivityPayload {
	return MockActivityPayload{
		ActivityID:         "garmin-e2e-activity",
		AthleteID:          athleteID,
		GarminActivityType: "running",
		StartTime:          workout.ScheduledFor.Add(7 * time.Hour),
		DurationSeconds:    int(workout.TargetDuration.Seconds()),
		DistanceMeters:     workout.TargetDistance.Meters,
		AverageHeartRate:   145,
	}
}

func e2eReplaceWorkout(week domain.PlanWeek, updated domain.PlannedWorkout) domain.PlanWeek {
	for i, workout := range week.Workouts {
		if workout.ID == updated.ID {
			week.Workouts[i] = updated
			return week
		}
	}
	return week
}

func e2eNow() time.Time {
	return e2eDateTime(2026, time.June, 2, 12, 0)
}

func e2eDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func e2eDateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
