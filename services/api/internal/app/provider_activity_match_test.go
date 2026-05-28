package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestMatchProviderActivityPersistsConfidentMatch(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week := providerActivityMatchWeek()
	workout := week.Workouts[0]
	activity := providerActivityMatchActivity(workout)
	seedProviderActivityMatchRecords(t, ctx, store, week, activity)

	response, err := providerActivityMatchService(store).MatchProviderActivity(ctx, MatchProviderActivityRequest{
		ImportedActivityID: activity.ID,
		PlanWeekID:         week.ID,
		PlannedWorkoutID:   workout.ID,
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("status = %q, want matched", response.WorkoutMatch.Status)
	}
	if response.WorkoutMatch.Confidence != domain.MatchConfidenceHigh {
		t.Fatalf("confidence = %q, want high", response.WorkoutMatch.Confidence)
	}
	assertProviderActivityMatchPersisted(t, ctx, store, response.WorkoutMatch)
}

func TestMatchProviderActivityPersistsUncertainMatch(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week := providerActivityMatchWeek()
	workout := week.Workouts[0]
	activity := providerActivityMatchActivity(workout)
	activity.ID = "imported-uncertain"
	activity.StartedAt = workout.ScheduledFor.AddDate(0, 0, 1).Add(7 * time.Hour)
	activity.Duration = 75 * time.Minute
	seedProviderActivityMatchRecords(t, ctx, store, week, activity)

	response, err := providerActivityMatchService(store).MatchProviderActivity(ctx, MatchProviderActivityRequest{
		ImportedActivityID: activity.ID,
		PlanWeekID:         week.ID,
		PlannedWorkoutID:   workout.ID,
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusUncertain {
		t.Fatalf("status = %q, want uncertain", response.WorkoutMatch.Status)
	}
	if response.WorkoutMatch.Confidence != domain.MatchConfidenceMedium {
		t.Fatalf("confidence = %q, want medium", response.WorkoutMatch.Confidence)
	}
	assertProviderActivityMatchPersisted(t, ctx, store, response.WorkoutMatch)
}

func TestMatchProviderActivityPersistsRejectedMatch(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week := providerActivityMatchWeek()
	workout := week.Workouts[0]
	activity := providerActivityMatchActivity(workout)
	activity.ID = "imported-rejected"
	activity.Type = domain.ActivityTypeWalk
	seedProviderActivityMatchRecords(t, ctx, store, week, activity)

	response, err := providerActivityMatchService(store).MatchProviderActivity(ctx, MatchProviderActivityRequest{
		ImportedActivityID: activity.ID,
		PlanWeekID:         week.ID,
		PlannedWorkoutID:   workout.ID,
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("status = %q, want rejected", response.WorkoutMatch.Status)
	}
	assertProviderActivityMatchPersisted(t, ctx, store, response.WorkoutMatch)
}

func TestMatchProviderActivityDoesNotDefaultRejectedMatchToRestWorkout(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week := providerActivityMatchWeekWithRestFirst()
	workout := week.Workouts[1]
	activity := providerActivityMatchActivity(workout)
	activity.ID = "imported-rejected-long-run"
	activity.Distance = domain.Distance{Meters: 20000}
	activity.Duration = 2 * time.Hour
	seedProviderActivityMatchRecords(t, ctx, store, week, activity)
	if err := store.SaveWorkoutMatch(ctx, domain.WorkoutMatch{
		ID:                 "stale-rest-match",
		PlannedWorkoutID:   week.Workouts[0].ID,
		ImportedActivityID: activity.ID,
		Status:             domain.WorkoutMatchStatusRejected,
		Confidence:         domain.MatchConfidenceLow,
		MatchedBy:          domain.MatchSourceAutomatic,
		MatchedAt:          providerActivityMatchNow().Add(-time.Hour),
		Notes:              "Rejected because activity type does not match the planned workout.",
	}); err != nil {
		t.Fatalf("SaveWorkoutMatch returned error: %v", err)
	}

	response, err := providerActivityMatchService(store).MatchProviderActivity(ctx, MatchProviderActivityRequest{
		ImportedActivityID: activity.ID,
		PlanWeekID:         week.ID,
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	if response.PlannedWorkout.ID != workout.ID {
		t.Fatalf("planned workout id = %q, want %q", response.PlannedWorkout.ID, workout.ID)
	}
	if response.PlannedWorkout.Type == domain.WorkoutTypeRest {
		t.Fatal("expected rejected match not to default to rest workout")
	}
	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusRejected {
		t.Fatalf("status = %q, want rejected", response.WorkoutMatch.Status)
	}
	if _, err := store.GetWorkoutMatch(ctx, "stale-rest-match"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected stale rest match to be pruned, got err %v", err)
	}
	assertProviderActivityMatchPersisted(t, ctx, store, response.WorkoutMatch)
}

func TestMatchProviderActivitySupportsManualOverride(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week := providerActivityMatchWeek()
	workout := week.Workouts[0]
	activity := providerActivityMatchActivity(workout)
	activity.ID = "imported-manual"
	activity.Type = domain.ActivityTypeWalk
	seedProviderActivityMatchRecords(t, ctx, store, week, activity)

	response, err := providerActivityMatchService(store).MatchProviderActivity(ctx, MatchProviderActivityRequest{
		ImportedActivityID: activity.ID,
		PlanWeekID:         week.ID,
		PlannedWorkoutID:   workout.ID,
		Manual:             true,
		ManualNotes:        "Runner confirmed this imported activity belongs to the planned workout.",
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	if response.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("status = %q, want matched", response.WorkoutMatch.Status)
	}
	if response.WorkoutMatch.MatchedBy != domain.MatchSourceManual {
		t.Fatalf("matched by = %q, want manual", response.WorkoutMatch.MatchedBy)
	}
	assertProviderActivityMatchPersisted(t, ctx, store, response.WorkoutMatch)
}

func providerActivityMatchService(store repository.Store) ProviderActivityMatchService {
	return ProviderActivityMatchService{
		Store:   store,
		Matcher: matching.Matcher{Now: providerActivityMatchNow},
		Now:     providerActivityMatchNow,
	}
}

func seedProviderActivityMatchRecords(t *testing.T, ctx context.Context, store *repository.InMemoryStore, week domain.PlanWeek, activity domain.ImportedActivity) {
	t.Helper()

	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("SavePlanWeek returned error: %v", err)
	}
	for _, workout := range week.Workouts {
		if err := store.SavePlannedWorkout(ctx, workout); err != nil {
			t.Fatalf("SavePlannedWorkout returned error: %v", err)
		}
	}
	if err := store.SaveImportedActivity(ctx, activity); err != nil {
		t.Fatalf("SaveImportedActivity returned error: %v", err)
	}
}

func assertProviderActivityMatchPersisted(t *testing.T, ctx context.Context, store *repository.InMemoryStore, match domain.WorkoutMatch) {
	t.Helper()

	stored, err := store.GetWorkoutMatch(ctx, match.ID)
	if err != nil {
		t.Fatalf("GetWorkoutMatch returned error: %v", err)
	}
	if stored.Status != match.Status {
		t.Fatalf("stored status = %q, want %q", stored.Status, match.Status)
	}
}

func providerActivityMatchWeek() domain.PlanWeek {
	workout := domain.PlannedWorkout{
		ID:             "workout-provider-match-1",
		PlanID:         "plan-provider-match",
		PlanWeekID:     "week-provider-match",
		ScheduledFor:   providerActivityMatchDateTime(2026, time.June, 11, 0, 0),
		Type:           domain.WorkoutTypeEasy,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 8000},
		TargetDuration: 45 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
	}
	return domain.PlanWeek{
		ID:        "week-provider-match",
		AthleteID: "athlete-provider-match",
		PlanID:    "plan-provider-match",
		WeekIndex: 1,
		StartsOn:  providerActivityMatchDateTime(2026, time.June, 8, 0, 0),
		Focus:     domain.WeekFocusBase,
		Workouts:  []domain.PlannedWorkout{workout},
	}
}

func providerActivityMatchWeekWithRestFirst() domain.PlanWeek {
	week := providerActivityMatchWeek()
	rest := domain.PlannedWorkout{
		ID:           "workout-provider-match-rest",
		PlanID:       week.PlanID,
		PlanWeekID:   week.ID,
		ScheduledFor: week.Workouts[0].ScheduledFor,
		Type:         domain.WorkoutTypeRest,
		Status:       domain.PlannedWorkoutStatusScheduled,
		Notes:        "Rest day.",
	}
	week.Workouts = []domain.PlannedWorkout{rest, week.Workouts[0]}
	return week
}

func providerActivityMatchActivity(workout domain.PlannedWorkout) domain.ImportedActivity {
	return domain.ImportedActivity{
		ID:              "imported-provider-match",
		AthleteID:       "athlete-provider-match",
		Type:            domain.ActivityTypeRun,
		StartedAt:       workout.ScheduledFor.Add(7 * time.Hour),
		Duration:        44 * time.Minute,
		Distance:        domain.Distance{Meters: 7900},
		AveragePace:     domain.Pace{SecondsPerKilometer: 334},
		AverageHeartBPM: 146,
	}
}

func providerActivityMatchNow() time.Time {
	return providerActivityMatchDateTime(2026, time.June, 11, 12, 0)
}

func providerActivityMatchDateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
