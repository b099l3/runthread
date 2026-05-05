package domain

import (
	"strings"
	"testing"
	"time"
)

func TestMarkWorkoutCompletedCreatesResultAndUpdatesStatus(t *testing.T) {
	workout := scheduledWorkout()

	updated, result, err := MarkWorkoutCompleted(workout, WorkoutCompletion{
		ResultID:           "result-1",
		ImportedActivityID: "activity-1",
		CompletedAt:        time.Date(2026, time.June, 2, 8, 30, 0, 0, time.UTC),
		Distance:           Distance{Meters: 8200},
		Duration:           42 * time.Minute,
	})
	if err != nil {
		t.Fatalf("expected completed workout: %v", err)
	}

	if updated.Status != PlannedWorkoutStatusCompleted {
		t.Fatalf("expected workout status completed, got %q", updated.Status)
	}
	if result.PlannedWorkoutID != workout.ID {
		t.Fatalf("expected result planned workout id %q, got %q", workout.ID, result.PlannedWorkoutID)
	}
	if result.Outcome != WorkoutOutcomeCompletedAsPlanned {
		t.Fatalf("expected completed as planned outcome, got %q", result.Outcome)
	}
	if result.ImportedActivityID != "activity-1" {
		t.Fatalf("expected imported activity id to be set")
	}
}

func TestMarkWorkoutCompletedSupportsPartialOutcome(t *testing.T) {
	workout := scheduledWorkout()

	_, result, err := MarkWorkoutCompleted(workout, WorkoutCompletion{
		ResultID:           "result-1",
		ImportedActivityID: "activity-1",
		Outcome:            WorkoutOutcomePartiallyCompleted,
		CompletedAt:        time.Date(2026, time.June, 2, 8, 30, 0, 0, time.UTC),
		Distance:           Distance{Meters: 4000},
		Duration:           24 * time.Minute,
	})
	if err != nil {
		t.Fatalf("expected partial completion: %v", err)
	}
	if result.Outcome != WorkoutOutcomePartiallyCompleted {
		t.Fatalf("expected partial outcome, got %q", result.Outcome)
	}
}

func TestMarkWorkoutCompletedRequiresActivityForCompletion(t *testing.T) {
	workout := scheduledWorkout()

	_, _, err := MarkWorkoutCompleted(workout, WorkoutCompletion{
		ResultID:    "result-1",
		CompletedAt: time.Date(2026, time.June, 2, 8, 30, 0, 0, time.UTC),
		Distance:    Distance{Meters: 8200},
		Duration:    42 * time.Minute,
	})

	assertWorkoutFlowError(t, err, "imported activity id")
}

func TestMarkWorkoutMissedCreatesMissedResult(t *testing.T) {
	workout := scheduledWorkout()

	updated, result, err := MarkWorkoutMissed(workout, "result-1", "Runner did not complete the planned run.")
	if err != nil {
		t.Fatalf("expected missed workout: %v", err)
	}

	if updated.Status != PlannedWorkoutStatusMissed {
		t.Fatalf("expected missed status, got %q", updated.Status)
	}
	if result.Outcome != WorkoutOutcomeMissed {
		t.Fatalf("expected missed outcome, got %q", result.Outcome)
	}
	if result.ImportedActivityID != "" {
		t.Fatal("expected missed result without imported activity")
	}
}

func TestMarkWorkoutSkippedCreatesSkippedResult(t *testing.T) {
	workout := scheduledWorkout()

	updated, result, err := MarkWorkoutSkipped(workout, "result-1", "Runner chose to skip this workout.")
	if err != nil {
		t.Fatalf("expected skipped workout: %v", err)
	}

	if updated.Status != PlannedWorkoutStatusSkipped {
		t.Fatalf("expected skipped status, got %q", updated.Status)
	}
	if result.Outcome != WorkoutOutcomeSkipped {
		t.Fatalf("expected skipped outcome, got %q", result.Outcome)
	}
}

func TestMoveWorkoutCreatesMovedResultAndUpdatesDate(t *testing.T) {
	workout := scheduledWorkout()
	newDate := time.Date(2026, time.June, 4, 0, 0, 0, 0, time.UTC)

	updated, result, err := MoveWorkout(workout, WorkoutMove{
		ResultID:        "result-1",
		NewScheduledFor: newDate,
		MovedAt:         time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC),
		Notes:           "Moved to fit the runner's week.",
	})
	if err != nil {
		t.Fatalf("expected moved workout: %v", err)
	}

	if updated.Status != PlannedWorkoutStatusMoved {
		t.Fatalf("expected moved status, got %q", updated.Status)
	}
	if !updated.ScheduledFor.Equal(newDate) {
		t.Fatalf("expected scheduled date %s, got %s", newDate, updated.ScheduledFor)
	}
	if result.Outcome != WorkoutOutcomeMoved {
		t.Fatalf("expected moved outcome, got %q", result.Outcome)
	}
}

func TestTerminalWorkoutCannotBeMarkedAgain(t *testing.T) {
	workout := scheduledWorkout()
	workout.Status = PlannedWorkoutStatusCompleted

	_, _, err := MarkWorkoutMissed(workout, "result-1", "Too late.")

	assertWorkoutFlowError(t, err, "already completed")
}

func TestMovedWorkoutCanStillBeCompleted(t *testing.T) {
	workout := scheduledWorkout()
	workout.Status = PlannedWorkoutStatusMoved

	updated, result, err := MarkWorkoutCompleted(workout, WorkoutCompletion{
		ResultID:           "result-1",
		ImportedActivityID: "activity-1",
		CompletedAt:        time.Date(2026, time.June, 4, 8, 30, 0, 0, time.UTC),
		Distance:           Distance{Meters: 8200},
		Duration:           42 * time.Minute,
	})
	if err != nil {
		t.Fatalf("expected moved workout to be completable: %v", err)
	}
	if updated.Status != PlannedWorkoutStatusCompleted {
		t.Fatalf("expected completed status, got %q", updated.Status)
	}
	if result.Outcome != WorkoutOutcomeCompletedAsPlanned {
		t.Fatalf("expected completed outcome, got %q", result.Outcome)
	}
}

func scheduledWorkout() PlannedWorkout {
	return PlannedWorkout{
		ID:             "workout-1",
		PlanID:         "plan-1",
		PlanWeekID:     "week-1",
		ScheduledFor:   time.Date(2026, time.June, 2, 0, 0, 0, 0, time.UTC),
		Type:           WorkoutTypeEasy,
		Status:         PlannedWorkoutStatusScheduled,
		TargetDistance: Distance{Meters: 8000},
		TargetDuration: 45 * time.Minute,
		Intensity:      IntensityTarget{Kind: IntensityKindEasy},
	}
}

func assertWorkoutFlowError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
