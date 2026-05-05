package planning

import (
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

func TestGenerateWeekHasSevenDays(t *testing.T) {
	week := mustGenerateWeek(t, beginnerProfile(), halfMarathonGoal(), date(2026, time.June, 3))

	if len(week.Workouts) != 7 {
		t.Fatalf("expected 7 planned workouts, got %d", len(week.Workouts))
	}

	for i, workout := range week.Workouts {
		expectedDate := week.StartsOn.AddDate(0, 0, i)
		if !sameDate(workout.ScheduledFor, expectedDate) {
			t.Fatalf("workout %d scheduled for %s, want %s", i, workout.ScheduledFor, expectedDate)
		}
	}
}

func TestGenerateWeekAvoidsBackToBackHardDays(t *testing.T) {
	week := mustGenerateWeek(t, intermediateProfile(), halfMarathonGoal(), date(2026, time.June, 3))

	for i := 0; i < len(week.Workouts)-1; i++ {
		if isHardWorkout(week.Workouts[i].Type) && isHardWorkout(week.Workouts[i+1].Type) {
			t.Fatalf("found back-to-back hard days at indexes %d and %d", i, i+1)
		}
	}
}

func TestGenerateWeekIncludesLongRunOnSundayByDefault(t *testing.T) {
	week := mustGenerateWeek(t, beginnerProfile(), halfMarathonGoal(), date(2026, time.June, 3))

	longRun := findWorkout(t, week, domain.WorkoutTypeLongRun)
	if longRun.ScheduledFor.Weekday() != time.Sunday {
		t.Fatalf("expected long run on Sunday, got %s", longRun.ScheduledFor.Weekday())
	}
}

func TestGenerateWeekUsesPreferredSaturdayLongRun(t *testing.T) {
	profile := beginnerProfile()
	profile.PreferredRunDays = []time.Weekday{time.Tuesday, time.Thursday, time.Saturday}

	week := mustGenerateWeek(t, profile, halfMarathonGoal(), date(2026, time.June, 3))

	longRun := findWorkout(t, week, domain.WorkoutTypeLongRun)
	if longRun.ScheduledFor.Weekday() != time.Saturday {
		t.Fatalf("expected long run on preferred Saturday, got %s", longRun.ScheduledFor.Weekday())
	}
}

func TestGenerateWeekVolumeIsSensible(t *testing.T) {
	profile := beginnerProfile()
	profile.CurrentWeeklyDistance = domain.Distance{Meters: 20000}

	week := mustGenerateWeek(t, profile, halfMarathonGoal(), date(2026, time.June, 3))

	total := runningVolume(week.Workouts)
	if total < profile.CurrentWeeklyDistance.Meters {
		t.Fatalf("expected weekly volume not below current volume, got %.0fm", total)
	}
	maxIncrease := profile.CurrentWeeklyDistance.Meters * 1.05
	if total > maxIncrease+200 {
		t.Fatalf("expected weekly volume to stay near 5%% cap %.0fm, got %.0fm", maxIncrease, total)
	}
}

func TestGenerateWeekDiffersForBeginnerAndIntermediate(t *testing.T) {
	beginnerWeek := mustGenerateWeek(t, beginnerProfile(), halfMarathonGoal(), date(2026, time.June, 3))
	intermediateWeek := mustGenerateWeek(t, intermediateProfile(), halfMarathonGoal(), date(2026, time.June, 3))

	if countWorkoutType(beginnerWeek, domain.WorkoutTypeWorkout) != 0 {
		t.Fatal("expected beginner week to avoid threshold workout")
	}
	if countWorkoutType(intermediateWeek, domain.WorkoutTypeWorkout) != 1 {
		t.Fatal("expected intermediate week to include one threshold workout")
	}
	if countWorkoutType(intermediateWeek, domain.WorkoutTypeStrength) != 1 {
		t.Fatal("expected intermediate week to include one optional strength day")
	}
}

func mustGenerateWeek(t *testing.T, profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time) domain.PlanWeek {
	t.Helper()

	week, err := GenerateWeek(profile, goal, targetWeekDate)
	if err != nil {
		t.Fatalf("expected generated week: %v", err)
	}
	return week
}

func beginnerProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-beginner",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 16000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func intermediateProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-intermediate",
		ExperienceLevel:       domain.ExperienceLevelIntermediate,
		CurrentWeeklyDistance: domain.Distance{Meters: 32000},
		PreferredRunDays:      []time.Weekday{time.Monday, time.Tuesday, time.Thursday, time.Friday, time.Sunday},
	}
}

func halfMarathonGoal() domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-1",
		AthleteID:      "athlete-1",
		Type:           domain.GoalTypeRace,
		TargetDate:     date(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func findWorkout(t *testing.T, week domain.PlanWeek, workoutType domain.WorkoutType) domain.PlannedWorkout {
	t.Helper()

	for _, workout := range week.Workouts {
		if workout.Type == workoutType {
			return workout
		}
	}
	t.Fatalf("expected workout type %q", workoutType)
	return domain.PlannedWorkout{}
}

func countWorkoutType(week domain.PlanWeek, workoutType domain.WorkoutType) int {
	count := 0
	for _, workout := range week.Workouts {
		if workout.Type == workoutType {
			count++
		}
	}
	return count
}

func sameDate(a, b time.Time) bool {
	aYear, aMonth, aDay := a.Date()
	bYear, bMonth, bDay := b.Date()
	return aYear == bYear && aMonth == bMonth && aDay == bDay
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
