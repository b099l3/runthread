package app

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestRetryImportedActivitiesCompletesUnmatchedActivity(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	workout := providerImportFirstRunWorkout(t, week)
	activity := providerImportActivityForWorkout(profile.ID, workout)
	activity.ID = "imported-retry-1"

	seedRetryRecords(t, ctx, store, profile, goal, week, *activity)

	response, err := ImportedActivityRetryService{
		Store: store,
		Now:   providerActivityCompletionNow,
	}.RetryImportedActivities(ctx, RetryImportedActivitiesRequest{
		AthleteID: profile.ID,
		GoalID:    goal.ID,
	})
	if err != nil {
		t.Fatalf("RetryImportedActivities returned error: %v", err)
	}

	if response.Scanned != 1 {
		t.Fatalf("scanned = %d, want 1", response.Scanned)
	}
	if response.Completed != 1 {
		t.Fatalf("completed = %d, want 1", response.Completed)
	}
	if response.Failed != 0 {
		t.Fatalf("failed = %d, failures = %#v", response.Failed, response.Failures)
	}

	plan, err := CurrentPlanWeekService{Store: store}.GetCurrentPlanWeek(ctx, GetCurrentPlanWeekRequest{
		AthleteID:      profile.ID,
		GoalID:         goal.ID,
		TargetWeekDate: activity.StartedAt,
	})
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}
	if len(plan.WorkoutResults) != 1 {
		t.Fatalf("workout results = %d, want 1", len(plan.WorkoutResults))
	}
	if plan.WorkoutResults[0].ImportedActivityID != activity.ID {
		t.Fatalf("result imported activity = %q, want %q", plan.WorkoutResults[0].ImportedActivityID, activity.ID)
	}
}

func TestRetryImportedActivitiesSkipsExistingResult(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	workout := providerImportFirstRunWorkout(t, week)
	activity := providerImportActivityForWorkout(profile.ID, workout)
	activity.ID = "imported-retry-completed"

	seedRetryRecords(t, ctx, store, profile, goal, week, *activity)
	if err := store.SaveWorkoutResult(ctx, domain.WorkoutResult{
		ID:                 "result-retry-completed",
		PlannedWorkoutID:   workout.ID,
		ImportedActivityID: activity.ID,
		Outcome:            domain.WorkoutOutcomeCompletedAsPlanned,
		CompletedAt:        activity.StartedAt.Add(activity.Duration),
		Distance:           activity.Distance,
		Duration:           activity.Duration,
	}); err != nil {
		t.Fatalf("SaveWorkoutResult returned error: %v", err)
	}

	response, err := ImportedActivityRetryService{
		Store: store,
		Now:   providerActivityCompletionNow,
	}.RetryImportedActivities(ctx, RetryImportedActivitiesRequest{
		AthleteID: profile.ID,
		GoalID:    goal.ID,
	})
	if err != nil {
		t.Fatalf("RetryImportedActivities returned error: %v", err)
	}

	if response.AlreadyCompleted != 1 {
		t.Fatalf("already completed = %d, want 1", response.AlreadyCompleted)
	}
	if response.Completed != 0 {
		t.Fatalf("completed = %d, want 0", response.Completed)
	}
}

func TestRetryImportedActivitiesSkipsAlreadyCompletedWorkout(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := providerImportProfile()
	goal := providerImportGoal(profile.ID)
	week := providerImportWeek(t, profile, goal, providerImportDate(2026, time.June, 3))
	workout := providerImportFirstRunWorkout(t, week)
	completedWorkout := workout
	completedWorkout.Status = domain.PlannedWorkoutStatusCompleted
	for index := range week.Workouts {
		if week.Workouts[index].ID == workout.ID {
			week.Workouts[index] = completedWorkout
		}
	}
	activity := providerImportActivityForWorkout(profile.ID, workout)
	activity.ID = "imported-retry-completed-workout"

	seedRetryRecords(t, ctx, store, profile, goal, week, *activity)

	response, err := ImportedActivityRetryService{
		Store: store,
		Now:   providerActivityCompletionNow,
	}.RetryImportedActivities(ctx, RetryImportedActivitiesRequest{
		AthleteID: profile.ID,
		GoalID:    goal.ID,
	})
	if err != nil {
		t.Fatalf("RetryImportedActivities returned error: %v", err)
	}

	if response.SkippedCompletedWorkout != 1 {
		t.Fatalf("skipped completed workout = %d, want 1", response.SkippedCompletedWorkout)
	}
	if response.Failed != 0 {
		t.Fatalf("failed = %d, failures = %#v", response.Failed, response.Failures)
	}
}

func seedRetryRecords(t *testing.T, ctx context.Context, store *repository.InMemoryStore, profile domain.AthleteProfile, goal domain.TrainingGoal, week domain.PlanWeek, activity domain.ImportedActivity) {
	t.Helper()

	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("SaveAthleteProfile returned error: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("SaveTrainingGoal returned error: %v", err)
	}
	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("SavePlanWeek returned error: %v", err)
	}
	for _, workout := range week.Workouts {
		if err := store.SavePlannedWorkout(ctx, workout); err != nil {
			t.Fatalf("SavePlannedWorkout returned error: %v", err)
		}
	}
	if err := store.SaveImportedActivity(ctx, activity); err != nil {
		t.Fatalf("SaveImportedActivity returned error: %v", err)
	}
}
