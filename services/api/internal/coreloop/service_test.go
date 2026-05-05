package coreloop

import (
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/garmin"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/planning"
)

func TestCompleteImportedActivityCoreLoopCompletedAsPlannedNoAdaptation(t *testing.T) {
	profile := beginnerProfile()
	goal := halfMarathonGoal(profile.ID)
	targetWeekDate := date(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	result, err := testService().CompleteImportedActivity(CompleteImportedActivityInput{
		AthleteProfile: profile,
		TrainingGoal:   goal,
		TargetWeekDate: targetWeekDate,
		ImportActivity: func() (domain.ImportedActivity, error) {
			return garmin.NormalizeMockActivity(mockPayloadForWorkout(profile.ID, expectedWorkout))
		},
		ResultID: "result-1",
	})
	if err != nil {
		t.Fatalf("expected completed core loop: %v", err)
	}

	if len(result.PlanWeek.Workouts) != 7 {
		t.Fatalf("expected generated seven-day week, got %d workouts", len(result.PlanWeek.Workouts))
	}
	if result.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("expected confident match, got %q", result.WorkoutMatch.Status)
	}
	if result.UpdatedWorkout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("expected completed workout status, got %q", result.UpdatedWorkout.Status)
	}
	if result.WorkoutResult.Outcome != domain.WorkoutOutcomeCompletedAsPlanned {
		t.Fatalf("expected completed-as-planned result, got %q", result.WorkoutResult.Outcome)
	}
	if result.AdaptationEvent != nil {
		t.Fatalf("expected no adaptation event, got %#v", result.AdaptationEvent)
	}
}

func TestCompleteImportedActivityCoreLoopUnderperformedProducesAdaptation(t *testing.T) {
	profile := intermediateProfile()
	goal := halfMarathonGoal(profile.ID)
	targetWeekDate := date(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	result, err := testService().CompleteImportedActivity(CompleteImportedActivityInput{
		AthleteProfile: profile,
		TrainingGoal:   goal,
		TargetWeekDate: targetWeekDate,
		ImportActivity: func() (domain.ImportedActivity, error) {
			return garmin.NormalizeMockActivity(mockPayloadForWorkout(profile.ID, expectedWorkout))
		},
		ResultID: "result-1",
		Outcome:  domain.WorkoutOutcomeUnderperformed,
	})
	if err != nil {
		t.Fatalf("expected completed core loop with adaptation: %v", err)
	}

	if result.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("expected confident match, got %q", result.WorkoutMatch.Status)
	}
	if result.WorkoutResult.Outcome != domain.WorkoutOutcomeUnderperformed {
		t.Fatalf("expected underperformed result, got %q", result.WorkoutResult.Outcome)
	}
	if result.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}
	if result.AdaptationEvent.Type != domain.AdaptationTypeUnderperformance {
		t.Fatalf("expected underperformance adaptation, got %q", result.AdaptationEvent.Type)
	}
	if len(result.AdaptationEvent.Changes) != 1 {
		t.Fatalf("expected one adaptation change, got %d", len(result.AdaptationEvent.Changes))
	}
	if result.AdaptationEvent.Changes[0].Type != domain.PlanChangeTypeWorkoutAdjusted {
		t.Fatalf("expected workout adjustment, got %q", result.AdaptationEvent.Changes[0].Type)
	}
}

func testService() Service {
	return Service{
		Planner:          planning.NewWeeklyPlanner(),
		Matcher:          fixedMatcher(),
		AdaptationEngine: fixedAdaptationEngine(),
	}
}

func fixedMatcher() matching.Matcher {
	return matching.Matcher{Now: fixedNow}
}

func fixedAdaptationEngine() adaptation.Engine {
	return adaptation.Engine{Now: fixedNow}
}

func beginnerProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-beginner",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
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

func mockPayloadForWorkout(athleteID string, workout domain.PlannedWorkout) garmin.MockActivityPayload {
	return garmin.MockActivityPayload{
		ActivityID:         "activity-" + workout.ID,
		AthleteID:          athleteID,
		GarminActivityType: "running",
		StartTime:          workout.ScheduledFor.Add(7 * time.Hour),
		DurationSeconds:    int(workout.TargetDuration.Seconds()),
		DistanceMeters:     workout.TargetDistance.Meters,
		AverageHeartRate:   145,
	}
}

func fixedNow() time.Time {
	return dateTime(2026, time.June, 2, 12, 0)
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func dateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
