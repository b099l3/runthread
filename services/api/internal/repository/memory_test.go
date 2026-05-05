package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

func TestInMemoryStoreSavesAndGetsCoreLoopRecords(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	profile := athleteProfile()
	goal := trainingGoal(profile.ID)
	week := planWeek()
	workout := week.Workouts[0]
	activity := importedActivity(profile.ID)
	match := workoutMatch(workout.ID, activity.ID)
	result := workoutResult(workout.ID, activity.ID)
	event := adaptationEvent(profile.ID, week.PlanID, week.Workouts[1].ID)

	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("save athlete profile: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("save training goal: %v", err)
	}
	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("save plan week: %v", err)
	}
	if err := store.SavePlannedWorkout(ctx, workout); err != nil {
		t.Fatalf("save planned workout: %v", err)
	}
	if err := store.SaveImportedActivity(ctx, activity); err != nil {
		t.Fatalf("save imported activity: %v", err)
	}
	if err := store.SaveWorkoutMatch(ctx, match); err != nil {
		t.Fatalf("save workout match: %v", err)
	}
	if err := store.SaveWorkoutResult(ctx, result); err != nil {
		t.Fatalf("save workout result: %v", err)
	}
	if err := store.SaveAdaptationEvent(ctx, event); err != nil {
		t.Fatalf("save adaptation event: %v", err)
	}

	assertID(t, mustGetAthleteProfile(t, store, profile.ID).ID, profile.ID)
	assertID(t, mustGetTrainingGoal(t, store, goal.ID).ID, goal.ID)
	assertID(t, mustGetPlanWeek(t, store, week.ID).ID, week.ID)
	assertID(t, mustGetPlannedWorkout(t, store, workout.ID).ID, workout.ID)
	assertID(t, mustGetImportedActivity(t, store, activity.ID).ID, activity.ID)
	assertID(t, mustGetWorkoutMatch(t, store, match.ID).ID, match.ID)
	assertID(t, mustGetWorkoutResult(t, store, result.ID).ID, result.ID)
	assertID(t, mustGetAdaptationEvent(t, store, event.ID).ID, event.ID)
}

func TestInMemoryStoreReturnsNotFound(t *testing.T) {
	_, err := NewInMemoryStore().GetAthleteProfile(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInMemoryStoreRejectsInvalidRecords(t *testing.T) {
	err := NewInMemoryStore().SaveAthleteProfile(context.Background(), domain.AthleteProfile{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestInMemoryStoreCopiesMutableSlices(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	profile := athleteProfile()
	profile.PreferredRunDays = []time.Weekday{time.Tuesday}
	profile.Constraints = []string{"No Friday runs"}
	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("save athlete profile: %v", err)
	}

	profile.PreferredRunDays[0] = time.Sunday
	profile.Constraints[0] = "Changed"

	stored := mustGetAthleteProfile(t, store, profile.ID)
	if stored.PreferredRunDays[0] != time.Tuesday {
		t.Fatalf("expected stored run day to remain Tuesday, got %s", stored.PreferredRunDays[0])
	}
	if stored.Constraints[0] != "No Friday runs" {
		t.Fatalf("expected stored constraint to remain unchanged, got %q", stored.Constraints[0])
	}

	stored.Constraints[0] = "Mutated after get"
	again := mustGetAthleteProfile(t, store, profile.ID)
	if again.Constraints[0] != "No Friday runs" {
		t.Fatalf("expected get result to be isolated from stored value, got %q", again.Constraints[0])
	}
}

func athleteProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-1",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func trainingGoal(athleteID string) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-1",
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     date(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func planWeek() domain.PlanWeek {
	return domain.PlanWeek{
		ID:        "week-1",
		AthleteID: "athlete-1",
		GoalID:    "goal-1",
		PlanID:    "plan-1",
		WeekIndex: 1,
		StartsOn:  date(2026, time.June, 1),
		Focus:     domain.WeekFocusBase,
		Workouts: []domain.PlannedWorkout{
			plannedWorkout("workout-1", date(2026, time.June, 2), domain.WorkoutTypeEasy),
			plannedWorkout("workout-2", date(2026, time.June, 4), domain.WorkoutTypeWorkout),
		},
	}
}

func plannedWorkout(id string, scheduledFor time.Time, workoutType domain.WorkoutType) domain.PlannedWorkout {
	return domain.PlannedWorkout{
		ID:             id,
		PlanID:         "plan-1",
		PlanWeekID:     "week-1",
		ScheduledFor:   scheduledFor,
		Type:           workoutType,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 8000},
		TargetDuration: 45 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
	}
}

func importedActivity(athleteID string) domain.ImportedActivity {
	return domain.ImportedActivity{
		ID:              "activity-1",
		AthleteID:       athleteID,
		Type:            domain.ActivityTypeRun,
		StartedAt:       dateTime(2026, time.June, 2, 7, 30),
		Duration:        44 * time.Minute,
		Distance:        domain.Distance{Meters: 7900},
		AveragePace:     domain.Pace{SecondsPerKilometer: 334},
		AverageHeartBPM: 146,
	}
}

func workoutMatch(workoutID string, activityID string) domain.WorkoutMatch {
	return domain.WorkoutMatch{
		ID:                 "match-1",
		PlannedWorkoutID:   workoutID,
		ImportedActivityID: activityID,
		Status:             domain.WorkoutMatchStatusMatched,
		Confidence:         domain.MatchConfidenceHigh,
		MatchedBy:          domain.MatchSourceAutomatic,
		MatchedAt:          dateTime(2026, time.June, 2, 9, 0),
	}
}

func workoutResult(workoutID string, activityID string) domain.WorkoutResult {
	return domain.WorkoutResult{
		ID:                 "result-1",
		PlannedWorkoutID:   workoutID,
		ImportedActivityID: activityID,
		Outcome:            domain.WorkoutOutcomeCompletedAsPlanned,
		CompletedAt:        dateTime(2026, time.June, 2, 8, 30),
		Distance:           domain.Distance{Meters: 7900},
		Duration:           44 * time.Minute,
	}
}

func adaptationEvent(athleteID string, planID string, workoutID string) domain.AdaptationEvent {
	return domain.AdaptationEvent{
		ID:        "adaptation-1",
		PlanID:    planID,
		AthleteID: athleteID,
		Type:      domain.AdaptationTypeUnderperformance,
		Reason:    "Workout came in below planned load.",
		Summary:   "Reduced pressure on the next workout.",
		CreatedAt: dateTime(2026, time.June, 2, 12, 0),
		Changes: []domain.PlanChange{
			{
				PlannedWorkoutID: workoutID,
				Type:             domain.PlanChangeTypeWorkoutAdjusted,
				Description:      "Keep the next workout conservative.",
			},
		},
	}
}

func mustGetAthleteProfile(t *testing.T, store *InMemoryStore, id string) domain.AthleteProfile {
	t.Helper()
	value, err := store.GetAthleteProfile(context.Background(), id)
	if err != nil {
		t.Fatalf("get athlete profile: %v", err)
	}
	return value
}

func mustGetTrainingGoal(t *testing.T, store *InMemoryStore, id string) domain.TrainingGoal {
	t.Helper()
	value, err := store.GetTrainingGoal(context.Background(), id)
	if err != nil {
		t.Fatalf("get training goal: %v", err)
	}
	return value
}

func mustGetPlanWeek(t *testing.T, store *InMemoryStore, id string) domain.PlanWeek {
	t.Helper()
	value, err := store.GetPlanWeek(context.Background(), id)
	if err != nil {
		t.Fatalf("get plan week: %v", err)
	}
	return value
}

func mustGetPlannedWorkout(t *testing.T, store *InMemoryStore, id string) domain.PlannedWorkout {
	t.Helper()
	value, err := store.GetPlannedWorkout(context.Background(), id)
	if err != nil {
		t.Fatalf("get planned workout: %v", err)
	}
	return value
}

func mustGetImportedActivity(t *testing.T, store *InMemoryStore, id string) domain.ImportedActivity {
	t.Helper()
	value, err := store.GetImportedActivity(context.Background(), id)
	if err != nil {
		t.Fatalf("get imported activity: %v", err)
	}
	return value
}

func mustGetWorkoutMatch(t *testing.T, store *InMemoryStore, id string) domain.WorkoutMatch {
	t.Helper()
	value, err := store.GetWorkoutMatch(context.Background(), id)
	if err != nil {
		t.Fatalf("get workout match: %v", err)
	}
	return value
}

func mustGetWorkoutResult(t *testing.T, store *InMemoryStore, id string) domain.WorkoutResult {
	t.Helper()
	value, err := store.GetWorkoutResult(context.Background(), id)
	if err != nil {
		t.Fatalf("get workout result: %v", err)
	}
	return value
}

func mustGetAdaptationEvent(t *testing.T, store *InMemoryStore, id string) domain.AdaptationEvent {
	t.Helper()
	value, err := store.GetAdaptationEvent(context.Background(), id)
	if err != nil {
		t.Fatalf("get adaptation event: %v", err)
	}
	return value
}

func assertID(t *testing.T, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("expected id %q, got %q", want, got)
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func dateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
