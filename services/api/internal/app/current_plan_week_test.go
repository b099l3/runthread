package app

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestGetCurrentPlanWeekReturnsSavedWeekByID(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)
	week := savedPlanWeek(profile.ID, goal.ID, date(2026, time.June, 1))
	activity := importedActivityForWorkout(profile.ID, week.Workouts[0])

	saveProfileGoalWeek(t, ctx, store, profile, goal, week)
	if err := store.SaveImportedActivity(ctx, activity); err != nil {
		t.Fatalf("SaveImportedActivity returned error: %v", err)
	}

	response, err := currentPlanWeekServiceWithStore(store).GetCurrentPlanWeek(ctx, GetCurrentPlanWeekRequest{
		PlanWeekID: week.ID,
	})
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}

	if response.PlanWeek.ID != week.ID {
		t.Fatalf("PlanWeek.ID = %q, want %q", response.PlanWeek.ID, week.ID)
	}
	if len(response.ImportedActivities) != 1 {
		t.Fatalf("ImportedActivities = %d, want 1", len(response.ImportedActivities))
	}
}

func TestGetCurrentPlanWeekGeneratesAndSavesWeekWhenMissing(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)

	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("SaveAthleteProfile returned error: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("SaveTrainingGoal returned error: %v", err)
	}

	response, err := currentPlanWeekServiceWithStore(store).GetCurrentPlanWeek(ctx, GetCurrentPlanWeekRequest{
		AthleteID:      profile.ID,
		GoalID:         goal.ID,
		TargetWeekDate: date(2026, time.June, 3),
	})
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}

	if len(response.PlanWeek.Workouts) != 7 {
		t.Fatalf("generated workouts = %d, want 7", len(response.PlanWeek.Workouts))
	}
	if _, err := store.GetPlanWeek(ctx, response.PlanWeek.ID); err != nil {
		t.Fatalf("expected generated week to be saved: %v", err)
	}
	for _, workout := range response.PlanWeek.Workouts {
		if _, err := store.GetPlannedWorkout(ctx, workout.ID); err != nil {
			t.Fatalf("expected generated workout %q to be saved: %v", workout.ID, err)
		}
	}
}

func TestGetCurrentPlanWeekDefaultsMissingTargetDateToCurrentWeek(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)

	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("SaveAthleteProfile returned error: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("SaveTrainingGoal returned error: %v", err)
	}

	response, err := currentPlanWeekServiceWithStore(store).GetCurrentPlanWeek(ctx, GetCurrentPlanWeekRequest{
		AthleteID: profile.ID,
		GoalID:    goal.ID,
	})
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}

	expectedStartsOn := startOfTestWeek(time.Now())
	if !sameTestDate(response.PlanWeek.StartsOn, expectedStartsOn) {
		t.Fatalf("PlanWeek.StartsOn = %s, want current week %s", response.PlanWeek.StartsOn, expectedStartsOn)
	}
}

func TestGetCurrentPlanWeekIncludesInMemoryCompletionState(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)
	targetWeekDate := date(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	completeResponse, err := coreLoopServiceWithStore(store).CompleteImportedActivity(ctx, CompleteImportedActivityRequest{
		AthleteProfile: profile,
		TrainingGoal:   goal,
		TargetWeekDate: targetWeekDate,
		ImportActivity: func(context.Context) (domain.ImportedActivity, error) {
			return importedActivityForWorkout(profile.ID, expectedWorkout), nil
		},
		ResultID: "result-1",
	})
	if err != nil {
		t.Fatalf("CompleteImportedActivity returned error: %v", err)
	}

	response, err := currentPlanWeekServiceWithStore(store).GetCurrentPlanWeek(ctx, GetCurrentPlanWeekRequest{
		AthleteID:      profile.ID,
		GoalID:         goal.ID,
		TargetWeekDate: targetWeekDate,
	})
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}

	if response.PlanWeek.ID != completeResponse.PlanWeek.ID {
		t.Fatalf("PlanWeek.ID = %q, want %q", response.PlanWeek.ID, completeResponse.PlanWeek.ID)
	}
	if len(response.ImportedActivities) != 1 {
		t.Fatalf("ImportedActivities = %d, want 1", len(response.ImportedActivities))
	}
	if len(response.WorkoutMatches) != 1 {
		t.Fatalf("WorkoutMatches = %d, want 1", len(response.WorkoutMatches))
	}
	if len(response.WorkoutResults) != 1 {
		t.Fatalf("WorkoutResults = %d, want 1", len(response.WorkoutResults))
	}
}

func currentPlanWeekServiceWithStore(store repository.Store) CurrentPlanWeekService {
	return CurrentPlanWeekService{Store: store}
}

func saveProfileGoalWeek(t *testing.T, ctx context.Context, store repository.Store, profile domain.AthleteProfile, goal domain.TrainingGoal, week domain.PlanWeek) {
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
}

func savedPlanWeek(athleteID string, goalID string, startsOn time.Time) domain.PlanWeek {
	return domain.PlanWeek{
		ID:        "week-1",
		AthleteID: athleteID,
		GoalID:    goalID,
		PlanID:    "plan-1",
		WeekIndex: 1,
		StartsOn:  startsOn,
		Focus:     domain.WeekFocusBase,
		Workouts: []domain.PlannedWorkout{
			{
				ID:             "workout-1",
				PlanID:         "plan-1",
				PlanWeekID:     "week-1",
				ScheduledFor:   startsOn.AddDate(0, 0, 1),
				Type:           domain.WorkoutTypeEasy,
				Status:         domain.PlannedWorkoutStatusScheduled,
				TargetDistance: domain.Distance{Meters: 5000},
				TargetDuration: 30 * time.Minute,
				Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
			},
		},
	}
}

func startOfTestWeek(date time.Time) time.Time {
	year, month, day := date.Date()
	location := date.Location()
	normalized := time.Date(year, month, day, 0, 0, 0, 0, location)
	offset := (int(normalized.Weekday()) + 6) % 7
	return normalized.AddDate(0, 0, -offset)
}

func sameTestDate(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
