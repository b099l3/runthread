package matching

import (
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

func TestMatchActivityHighConfidenceOnDateDistanceAndDuration(t *testing.T) {
	match := mustMatch(t, plannedWorkout(), importedActivity())

	if match.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("expected matched status, got %q", match.Status)
	}
	if match.Confidence != domain.MatchConfidenceHigh {
		t.Fatalf("expected high confidence, got %q", match.Confidence)
	}
	if match.MatchedBy != domain.MatchSourceAutomatic {
		t.Fatalf("expected automatic match, got %q", match.MatchedBy)
	}
}

func TestMatchActivityRejectsTypeMismatch(t *testing.T) {
	activity := importedActivity()
	activity.Type = domain.ActivityTypeWalk

	match := mustMatch(t, plannedWorkout(), activity)

	if match.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("expected rejected status, got %q", match.Status)
	}
	if match.Confidence != domain.MatchConfidenceLow {
		t.Fatalf("expected low confidence, got %q", match.Confidence)
	}
}

func TestMatchActivityAcceptsRideAsRunCrossTrainingByDuration(t *testing.T) {
	activity := importedActivity()
	activity.Type = domain.ActivityTypeRide
	activity.Distance = domain.Distance{Meters: 26000}

	match := mustMatch(t, plannedWorkout(), activity)

	if match.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("expected matched status, got %q", match.Status)
	}
	if match.Confidence != domain.MatchConfidenceMedium {
		t.Fatalf("expected medium confidence, got %q", match.Confidence)
	}
	if match.Notes != "Completed by ride cross-training activity." {
		t.Fatalf("match notes = %q", match.Notes)
	}
}

func TestMatchActivityRejectsWeakRideCrossTrainingMatch(t *testing.T) {
	activity := importedActivity()
	activity.Type = domain.ActivityTypeRide
	activity.Duration = 20 * time.Minute

	match := mustMatch(t, plannedWorkout(), activity)

	if match.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("expected rejected status, got %q", match.Status)
	}
}

func TestMatchActivityKeepsRaceRunOnly(t *testing.T) {
	workout := plannedWorkout()
	workout.Type = domain.WorkoutTypeRace
	activity := importedActivity()
	activity.Type = domain.ActivityTypeRide

	match := mustMatch(t, workout, activity)

	if match.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("expected rejected race ride match, got %q", match.Status)
	}
}

func TestMatchActivityRejectsLargeDateGap(t *testing.T) {
	activity := importedActivity()
	activity.StartedAt = dateTime(2026, time.June, 8, 7, 30)

	match := mustMatch(t, plannedWorkout(), activity)

	if match.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("expected rejected status, got %q", match.Status)
	}
}

func TestMatchActivityUncertainOnNearDateAndDistanceWithPoorDuration(t *testing.T) {
	activity := importedActivity()
	activity.StartedAt = dateTime(2026, time.June, 3, 7, 30)
	activity.Duration = 75 * time.Minute

	match := mustMatch(t, plannedWorkout(), activity)

	if match.Status != domain.WorkoutMatchStatusUncertain {
		t.Fatalf("expected uncertain status, got %q", match.Status)
	}
	if match.Confidence != domain.MatchConfidenceMedium {
		t.Fatalf("expected medium confidence, got %q", match.Confidence)
	}
}

func TestMatchActivityRejectsDistanceAndDurationMismatch(t *testing.T) {
	activity := importedActivity()
	activity.Distance = domain.Distance{Meters: 2500}
	activity.Duration = 15 * time.Minute

	match := mustMatch(t, plannedWorkout(), activity)

	if match.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("expected rejected status, got %q", match.Status)
	}
}

func TestManualMatchCreatesHighConfidenceManualMatch(t *testing.T) {
	matchedAt := dateTime(2026, time.June, 2, 12, 0)
	match, err := ManualMatch(plannedWorkout(), importedActivity(), matchedAt, "Runner confirmed this was the planned run.")
	if err != nil {
		t.Fatalf("expected manual match: %v", err)
	}

	if match.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("expected matched status, got %q", match.Status)
	}
	if match.Confidence != domain.MatchConfidenceHigh {
		t.Fatalf("expected high confidence, got %q", match.Confidence)
	}
	if match.MatchedBy != domain.MatchSourceManual {
		t.Fatalf("expected manual source, got %q", match.MatchedBy)
	}
}

func mustMatch(t *testing.T, workout domain.PlannedWorkout, activity domain.ImportedActivity) domain.WorkoutMatch {
	t.Helper()

	matcher := Matcher{Now: func() time.Time {
		return dateTime(2026, time.June, 2, 12, 0)
	}}
	match, err := matcher.MatchActivity(workout, activity)
	if err != nil {
		t.Fatalf("expected match result: %v", err)
	}
	return match
}

func plannedWorkout() domain.PlannedWorkout {
	return domain.PlannedWorkout{
		ID:             "workout-1",
		PlanID:         "plan-1",
		PlanWeekID:     "week-1",
		ScheduledFor:   dateTime(2026, time.June, 2, 0, 0),
		Type:           domain.WorkoutTypeEasy,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 8000},
		TargetDuration: 45 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
	}
}

func importedActivity() domain.ImportedActivity {
	return domain.ImportedActivity{
		ID:              "activity-1",
		AthleteID:       "athlete-1",
		Type:            domain.ActivityTypeRun,
		StartedAt:       dateTime(2026, time.June, 2, 7, 30),
		Duration:        44 * time.Minute,
		Distance:        domain.Distance{Meters: 7900},
		AveragePace:     domain.Pace{SecondsPerKilometer: 334},
		AverageHeartBPM: 146,
	}
}

func dateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
