package domain

import (
	"strings"
	"testing"
	"time"
)

func TestAthleteProfileValidate(t *testing.T) {
	profile := AthleteProfile{
		ID:                    "athlete-1",
		ExperienceLevel:       ExperienceLevelBeginner,
		CurrentWeeklyDistance: Distance{Meters: 24000},
	}

	if err := profile.Validate(); err != nil {
		t.Fatalf("expected valid profile: %v", err)
	}

	profile.CurrentWeeklyDistance = Distance{Meters: -1}
	assertValidationError(t, profile.Validate(), "current weekly distance")
}

func TestTrainingGoalValidate(t *testing.T) {
	goal := TrainingGoal{
		ID:             "goal-1",
		AthleteID:      "athlete-1",
		Type:           GoalTypeRace,
		TargetDate:     date(2026, time.October, 18),
		TargetDistance: Distance{Meters: 21097.5},
		TargetDuration: 90 * time.Minute,
	}

	if err := goal.Validate(); err != nil {
		t.Fatalf("expected valid goal: %v", err)
	}

	goal.Type = GoalType("unsupported")
	assertValidationError(t, goal.Validate(), "invalid goal type")
}

func TestTrainingPlanValidateIncludesWeeksAndWorkouts(t *testing.T) {
	plan := TrainingPlan{
		ID:        "plan-1",
		AthleteID: "athlete-1",
		GoalID:    "goal-1",
		Status:    PlanStatusActive,
		StartsOn:  date(2026, time.June, 1),
		EndsOn:    date(2026, time.July, 12),
		Weeks: []PlanWeek{
			{
				ID:        "week-1",
				AthleteID: "athlete-1",
				GoalID:    "goal-1",
				PlanID:    "plan-1",
				WeekIndex: 1,
				StartsOn:  date(2026, time.June, 1),
				Focus:     WeekFocusBase,
				Workouts: []PlannedWorkout{
					{
						ID:             "workout-1",
						PlanID:         "plan-1",
						PlanWeekID:     "week-1",
						ScheduledFor:   date(2026, time.June, 2),
						Type:           WorkoutTypeEasy,
						Status:         PlannedWorkoutStatusScheduled,
						TargetDuration: 45 * time.Minute,
						Intensity:      IntensityTarget{Kind: IntensityKindEasy},
					},
				},
			},
		},
	}

	if err := plan.Validate(); err != nil {
		t.Fatalf("expected valid plan: %v", err)
	}

	plan.Weeks[0].Workouts[0].TargetDuration = 0
	assertValidationError(t, plan.Validate(), "target distance or duration")
}

func TestPlannedWorkoutValidateRequiresTargetDistanceOrDuration(t *testing.T) {
	workout := PlannedWorkout{
		ID:           "workout-1",
		PlanID:       "plan-1",
		PlanWeekID:   "week-1",
		ScheduledFor: date(2026, time.June, 2),
		Type:         WorkoutTypeEasy,
		Status:       PlannedWorkoutStatusScheduled,
	}

	assertValidationError(t, workout.Validate(), "target distance or duration")

	workout.TargetDistance = Distance{Meters: 5000}
	if err := workout.Validate(); err != nil {
		t.Fatalf("expected valid workout: %v", err)
	}
}

func TestImportedActivityValidateIsProviderNeutral(t *testing.T) {
	activity := ImportedActivity{
		ID:              "activity-1",
		AthleteID:       "athlete-1",
		Type:            ActivityTypeRun,
		StartedAt:       time.Date(2026, time.June, 2, 7, 30, 0, 0, time.UTC),
		Duration:        42 * time.Minute,
		Distance:        Distance{Meters: 8100},
		AveragePace:     Pace{SecondsPerKilometer: 311},
		AverageHeartBPM: 148,
	}

	if err := activity.Validate(); err != nil {
		t.Fatalf("expected valid imported activity: %v", err)
	}

	activity.Type = ActivityTypeRide
	if err := activity.Validate(); err != nil {
		t.Fatalf("expected ride imported activity to be valid: %v", err)
	}

	activity.Type = ActivityType("garmin_running_activity")
	assertValidationError(t, activity.Validate(), "invalid activity type")
}

func TestWorkoutMatchValidate(t *testing.T) {
	match := WorkoutMatch{
		ID:                 "match-1",
		PlannedWorkoutID:   "workout-1",
		ImportedActivityID: "activity-1",
		Status:             WorkoutMatchStatusMatched,
		Confidence:         MatchConfidenceHigh,
		MatchedBy:          MatchSourceAutomatic,
		MatchedAt:          time.Date(2026, time.June, 2, 9, 0, 0, 0, time.UTC),
	}

	if err := match.Validate(); err != nil {
		t.Fatalf("expected valid workout match: %v", err)
	}

	match.MatchedBy = MatchSource("watch")
	assertValidationError(t, match.Validate(), "invalid match source")
}

func TestWorkoutResultValidateRequiresActivityForCompletedOutcomes(t *testing.T) {
	result := WorkoutResult{
		ID:               "result-1",
		PlannedWorkoutID: "workout-1",
		Outcome:          WorkoutOutcomeCompletedAsPlanned,
		CompletedAt:      time.Date(2026, time.June, 2, 8, 30, 0, 0, time.UTC),
		Distance:         Distance{Meters: 8000},
		Duration:         42 * time.Minute,
	}

	assertValidationError(t, result.Validate(), "imported activity id")

	result.ImportedActivityID = "activity-1"
	if err := result.Validate(); err != nil {
		t.Fatalf("expected valid workout result: %v", err)
	}
}

func TestWorkoutResultValidateAllowsMissedWithoutActivity(t *testing.T) {
	result := WorkoutResult{
		ID:               "result-1",
		PlannedWorkoutID: "workout-1",
		Outcome:          WorkoutOutcomeMissed,
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("expected missed workout without activity to be valid: %v", err)
	}
}

func TestAdaptationEventValidate(t *testing.T) {
	event := AdaptationEvent{
		ID:        "adaptation-1",
		PlanID:    "plan-1",
		AthleteID: "athlete-1",
		Type:      AdaptationTypeMissedWorkout,
		Reason:    "Workout was missed.",
		Summary:   "The next easy run was shortened.",
		CreatedAt: time.Date(2026, time.June, 4, 12, 0, 0, 0, time.UTC),
		Changes: []PlanChange{
			{
				PlannedWorkoutID: "workout-2",
				Type:             PlanChangeTypeWorkoutAdjusted,
				Description:      "Reduced target duration from 50 minutes to 40 minutes.",
			},
		},
	}

	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid adaptation event: %v", err)
	}

	event.Changes[0].Type = PlanChangeType("rewrite_plan")
	assertValidationError(t, event.Validate(), "invalid plan change type")
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func assertValidationError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected validation error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected validation error containing %q, got %q", contains, err.Error())
	}
}
