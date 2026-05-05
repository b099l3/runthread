package adaptation

import (
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

func TestAdaptWorkoutResultCreatesMissedWorkoutEvent(t *testing.T) {
	event := mustAdapt(t, workoutResult(domain.WorkoutOutcomeMissed))

	if event.Type != domain.AdaptationTypeMissedWorkout {
		t.Fatalf("expected missed workout adaptation, got %q", event.Type)
	}
	if len(event.Changes) != 1 {
		t.Fatalf("expected one plan change, got %d", len(event.Changes))
	}
	if event.Changes[0].Type != domain.PlanChangeTypeWorkoutAdjusted {
		t.Fatalf("expected workout adjustment, got %q", event.Changes[0].Type)
	}
	if event.Changes[0].PlannedWorkoutID != "workout-threshold" {
		t.Fatalf("expected threshold workout to be adjusted, got %q", event.Changes[0].PlannedWorkoutID)
	}
}

func TestAdaptWorkoutResultCreatesPartialCompletionEvent(t *testing.T) {
	event := mustAdapt(t, workoutResult(domain.WorkoutOutcomePartiallyCompleted))

	if event.Type != domain.AdaptationTypePartialCompletion {
		t.Fatalf("expected partial completion adaptation, got %q", event.Type)
	}
	if !strings.Contains(event.Summary, "conservative") {
		t.Fatalf("expected conservative summary, got %q", event.Summary)
	}
}

func TestAdaptWorkoutResultCreatesOverperformanceEvent(t *testing.T) {
	event := mustAdapt(t, workoutResult(domain.WorkoutOutcomeOverperformed))

	if event.Type != domain.AdaptationTypeOverperformance {
		t.Fatalf("expected overperformance adaptation, got %q", event.Type)
	}
	if !strings.Contains(event.Changes[0].Description, "Avoid adding extra intensity") {
		t.Fatalf("expected conservative overperformance change, got %q", event.Changes[0].Description)
	}
}

func TestAdaptWorkoutResultCreatesUnderperformanceEvent(t *testing.T) {
	event := mustAdapt(t, workoutResult(domain.WorkoutOutcomeUnderperformed))

	if event.Type != domain.AdaptationTypeUnderperformance {
		t.Fatalf("expected underperformance adaptation, got %q", event.Type)
	}
	if !strings.Contains(event.Changes[0].Description, "Reduce pressure") {
		t.Fatalf("expected reduced pressure change, got %q", event.Changes[0].Description)
	}
}

func TestAdaptWorkoutResultNoopsForCompletedAsPlanned(t *testing.T) {
	engine := Engine{Now: fixedNow}

	event, err := engine.AdaptWorkoutResult(WorkoutResultInput{
		AthleteID: "athlete-1",
		PlanWeek:  planWeek(),
		Result:    workoutResult(domain.WorkoutOutcomeCompletedAsPlanned),
	})
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if event != nil {
		t.Fatalf("expected no adaptation event, got %#v", event)
	}
}

func TestAdaptWorkoutResultFallsBackToPlanNoteWhenNoFutureWorkout(t *testing.T) {
	week := planWeek()
	result := workoutResult(domain.WorkoutOutcomeMissed)
	result.PlannedWorkoutID = "workout-long"

	engine := Engine{Now: fixedNow}
	event, err := engine.AdaptWorkoutResult(WorkoutResultInput{
		AthleteID: "athlete-1",
		PlanWeek:  week,
		Result:    result,
	})
	if err != nil {
		t.Fatalf("expected adaptation event: %v", err)
	}
	if event.Changes[0].Type != domain.PlanChangeTypePlanNoteAdded {
		t.Fatalf("expected plan note change, got %q", event.Changes[0].Type)
	}
}

func mustAdapt(t *testing.T, result domain.WorkoutResult) domain.AdaptationEvent {
	t.Helper()

	engine := Engine{Now: fixedNow}
	event, err := engine.AdaptWorkoutResult(WorkoutResultInput{
		AthleteID: "athlete-1",
		PlanWeek:  planWeek(),
		Result:    result,
	})
	if err != nil {
		t.Fatalf("expected adaptation event: %v", err)
	}
	if event == nil {
		t.Fatal("expected adaptation event, got nil")
	}
	return *event
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
			plannedWorkout("workout-easy", date(2026, time.June, 2), domain.WorkoutTypeEasy, 8000, 45*time.Minute),
			plannedWorkout("workout-threshold", date(2026, time.June, 4), domain.WorkoutTypeWorkout, 7000, 35*time.Minute),
			plannedWorkout("workout-long", date(2026, time.June, 7), domain.WorkoutTypeLongRun, 13000, 80*time.Minute),
		},
	}
}

func plannedWorkout(id string, scheduledFor time.Time, workoutType domain.WorkoutType, distance float64, duration time.Duration) domain.PlannedWorkout {
	return domain.PlannedWorkout{
		ID:             id,
		PlanID:         "plan-1",
		PlanWeekID:     "week-1",
		ScheduledFor:   scheduledFor,
		Type:           workoutType,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: distance},
		TargetDuration: duration,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
	}
}

func workoutResult(outcome domain.WorkoutOutcome) domain.WorkoutResult {
	result := domain.WorkoutResult{
		ID:               "result-1",
		PlannedWorkoutID: "workout-easy",
		Outcome:          outcome,
		Distance:         domain.Distance{Meters: 7000},
		Duration:         38 * time.Minute,
		Notes:            "Test result.",
	}
	if outcome != domain.WorkoutOutcomeMissed && outcome != domain.WorkoutOutcomeSkipped {
		result.ImportedActivityID = "activity-1"
		result.CompletedAt = dateTime(2026, time.June, 2, 8, 0)
	}
	return result
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
