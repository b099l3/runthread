package app

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestCompleteImportedActivityUsesInMemoryCoreLoop(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)
	targetWeekDate := date(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	response, err := coreLoopServiceWithStore(store).CompleteImportedActivity(ctx, CompleteImportedActivityRequest{
		AthleteProfile: profile,
		TrainingGoal:   goal,
		TargetWeekDate: targetWeekDate,
		ImportActivity: func(context.Context) (domain.ImportedActivity, error) {
			return importedActivityForWorkout(profile.ID, expectedWorkout), nil
		},
		ResultID: "result-1",
	})
	if err != nil {
		t.Fatalf("expected completed imported activity: %v", err)
	}

	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("expected matched workout, got %q", response.WorkoutMatch.Status)
	}
	if response.UpdatedWorkout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("expected completed workout, got %q", response.UpdatedWorkout.Status)
	}
	if response.WorkoutResult.Outcome != domain.WorkoutOutcomeCompletedAsPlanned {
		t.Fatalf("expected completed-as-planned result, got %q", response.WorkoutResult.Outcome)
	}
	if response.AdaptationEvent != nil {
		t.Fatalf("expected no adaptation event, got %#v", response.AdaptationEvent)
	}

	assertSavedCoreLoopRecords(t, ctx, store, profile.ID, goal.ID, response)
}

func TestCompleteImportedActivityRequiresRunner(t *testing.T) {
	_, err := CoreLoopService{}.CompleteImportedActivity(context.Background(), CompleteImportedActivityRequest{
		ImportActivity: func(context.Context) (domain.ImportedActivity, error) {
			return domain.ImportedActivity{}, nil
		},
	})
	if err == nil {
		t.Fatal("expected missing runner error")
	}
}

func TestCompleteImportedActivityPersistsAdaptationEventWhenPresent(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)
	targetWeekDate := date(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	response, err := coreLoopServiceWithStore(store).CompleteImportedActivity(ctx, CompleteImportedActivityRequest{
		AthleteProfile: profile,
		TrainingGoal:   goal,
		TargetWeekDate: targetWeekDate,
		ImportActivity: func(context.Context) (domain.ImportedActivity, error) {
			return importedActivityForWorkout(profile.ID, expectedWorkout), nil
		},
		ResultID: "result-1",
		Outcome:  domain.WorkoutOutcomeUnderperformed,
	})
	if err != nil {
		t.Fatalf("expected completed imported activity: %v", err)
	}
	if response.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}

	saved, err := store.GetAdaptationEvent(ctx, response.AdaptationEvent.ID)
	if err != nil {
		t.Fatalf("expected saved adaptation event: %v", err)
	}
	if saved.Type != domain.AdaptationTypeUnderperformance {
		t.Fatalf("expected underperformance event, got %q", saved.Type)
	}
}

func coreLoopServiceWithStore(store repository.Store) CoreLoopService {
	service := NewInMemoryCoreLoopService()
	service.Store = store
	return service
}

func assertSavedCoreLoopRecords(t *testing.T, ctx context.Context, store repository.Store, athleteID string, goalID string, response CompleteImportedActivityResponse) {
	t.Helper()

	if _, err := store.GetAthleteProfile(ctx, athleteID); err != nil {
		t.Fatalf("expected saved athlete profile: %v", err)
	}
	if _, err := store.GetTrainingGoal(ctx, goalID); err != nil {
		t.Fatalf("expected saved training goal: %v", err)
	}
	savedWeek, err := store.GetPlanWeek(ctx, response.PlanWeek.ID)
	if err != nil {
		t.Fatalf("expected saved plan week: %v", err)
	}
	if _, err := store.GetImportedActivity(ctx, response.ImportedActivity.ID); err != nil {
		t.Fatalf("expected saved imported activity: %v", err)
	}
	if _, err := store.GetWorkoutMatch(ctx, response.WorkoutMatch.ID); err != nil {
		t.Fatalf("expected saved workout match: %v", err)
	}
	if _, err := store.GetWorkoutResult(ctx, response.WorkoutResult.ID); err != nil {
		t.Fatalf("expected saved workout result: %v", err)
	}
	if savedWorkoutStatus(savedWeek, response.UpdatedWorkout.ID) != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("expected saved plan week to include completed workout")
	}
	assertSavedPlannedWorkouts(t, ctx, store, response.PlanWeek.Workouts, response.UpdatedWorkout.ID)
}

func savedWorkoutStatus(week domain.PlanWeek, workoutID string) domain.PlannedWorkoutStatus {
	for _, workout := range week.Workouts {
		if workout.ID == workoutID {
			return workout.Status
		}
	}
	return ""
}

func assertSavedPlannedWorkouts(t *testing.T, ctx context.Context, store repository.Store, workouts []domain.PlannedWorkout, completedWorkoutID string) {
	t.Helper()

	if len(workouts) == 0 {
		t.Fatal("expected planned workouts to be present")
	}

	for _, workout := range workouts {
		saved, err := store.GetPlannedWorkout(ctx, workout.ID)
		if err != nil {
			t.Fatalf("expected saved planned workout %q: %v", workout.ID, err)
		}
		if saved.ID != workout.ID {
			t.Fatalf("saved workout ID = %q, want %q", saved.ID, workout.ID)
		}
		if saved.ID == completedWorkoutID && saved.Status != domain.PlannedWorkoutStatusCompleted {
			t.Fatalf("saved completed workout status = %q, want %q", saved.Status, domain.PlannedWorkoutStatusCompleted)
		}
	}
}

func beginnerProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-beginner",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func halfMarathonGoal(athleteID string) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-1",
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     date(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func firstWorkoutOfType(t *testing.T, profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time, workoutType domain.WorkoutType) domain.PlannedWorkout {
	t.Helper()

	week, err := planning.GenerateWeek(profile, goal, targetWeekDate)
	if err != nil {
		t.Fatalf("expected generated week: %v", err)
	}
	for _, workout := range week.Workouts {
		if workout.Type == workoutType {
			return workout
		}
	}
	t.Fatalf("expected generated workout type %q", workoutType)
	return domain.PlannedWorkout{}
}

func importedActivityForWorkout(athleteID string, workout domain.PlannedWorkout) domain.ImportedActivity {
	return domain.ImportedActivity{
		ID:              "activity-" + workout.ID,
		AthleteID:       athleteID,
		Type:            domain.ActivityTypeRun,
		StartedAt:       workout.ScheduledFor.Add(7 * time.Hour),
		Duration:        workout.TargetDuration,
		Distance:        workout.TargetDistance,
		AveragePace:     domain.Pace{SecondsPerKilometer: 330},
		AverageHeartBPM: 145,
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
