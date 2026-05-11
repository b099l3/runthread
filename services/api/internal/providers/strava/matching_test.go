package strava

import (
	"context"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestImportedStravaBackfillActivityMatchesPlannedWorkout(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	connection := connectedStravaConnection("connection-1", "athlete-strava-match")
	week := stravaMatchWeek()
	workout := week.Workouts[0]
	if err := store.SaveProviderConnection(ctx, connection); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("SavePlanWeek returned error: %v", err)
	}
	if err := store.SavePlannedWorkout(ctx, workout); err != nil {
		t.Fatalf("SavePlannedWorkout returned error: %v", err)
	}

	backfill := testBackfillService(t, store, &fakeActivityFetcher{
		summaries: []MockActivitySummary{{ActivityID: "strava-match-1"}},
		details: map[string]MockActivityPayload{
			"strava-match-1": stravaMatchPayload(workout),
		},
	})
	backfillResult, err := backfill.RunInitialBackfill(ctx, RunBackfillRequest{ProviderConnectionID: connection.ID})
	if err != nil {
		t.Fatalf("RunInitialBackfill returned error: %v", err)
	}
	if backfillResult.Imported != 1 {
		t.Fatalf("imported = %d, want 1", backfillResult.Imported)
	}

	matchService := app.ProviderActivityMatchService{
		Store:   store,
		Matcher: matching.Matcher{Now: stravaMatchNow},
		Now:     stravaMatchNow,
	}
	response, err := matchService.MatchProviderActivity(ctx, app.MatchProviderActivityRequest{
		ImportedActivityID: backfillResult.Imports[0].ImportedActivity.ID,
		PlanWeekID:         week.ID,
		PlannedWorkoutID:   workout.ID,
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("match status = %q, want matched", response.WorkoutMatch.Status)
	}
	if response.WorkoutMatch.MatchedBy != domain.MatchSourceAutomatic {
		t.Fatalf("matched by = %q, want automatic", response.WorkoutMatch.MatchedBy)
	}
	if _, err := store.GetWorkoutMatch(ctx, response.WorkoutMatch.ID); err != nil {
		t.Fatalf("expected persisted workout match: %v", err)
	}
}

func stravaMatchWeek() domain.PlanWeek {
	workout := domain.PlannedWorkout{
		ID:             "workout-strava-match",
		PlanID:         "plan-strava-match",
		PlanWeekID:     "week-strava-match",
		ScheduledFor:   time.Date(2026, time.June, 12, 0, 0, 0, 0, time.UTC),
		Type:           domain.WorkoutTypeEasy,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 8800},
		TargetDuration: 44 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
	}
	return domain.PlanWeek{
		ID:        "week-strava-match",
		AthleteID: "athlete-strava-match",
		PlanID:    "plan-strava-match",
		WeekIndex: 1,
		StartsOn:  time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC),
		Focus:     domain.WeekFocusBase,
		Workouts:  []domain.PlannedWorkout{workout},
	}
}

func stravaMatchPayload(workout domain.PlannedWorkout) MockActivityPayload {
	return MockActivityPayload{
		ActivityID:       "strava-match-1",
		AthleteID:        "athlete-strava-match",
		StravaSportType:  "Run",
		Name:             "Morning Run",
		StartDate:        workout.ScheduledFor.Add(7 * time.Hour),
		ElapsedTime:      2700,
		MovingTime:       2640,
		DistanceMeters:   8800,
		AverageHeartRate: 151,
	}
}

func stravaMatchNow() time.Time {
	return time.Date(2026, time.June, 12, 12, 0, 0, 0, time.UTC)
}
