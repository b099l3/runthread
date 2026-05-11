package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestCompleteMatchedProviderActivityCreatesResultWithoutAdaptationWhenCompletedAsPlanned(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week, workout, activity, match := seedProviderActivityCompletionRecords(t, ctx, store, domain.WorkoutMatchStatusMatched)

	response, err := providerActivityCompletionService(store).CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: match.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-completed",
	})
	if err != nil {
		t.Fatalf("CompleteMatchedProviderActivity returned error: %v", err)
	}

	if response.WorkoutResult.Outcome != domain.WorkoutOutcomeCompletedAsPlanned {
		t.Fatalf("outcome = %q, want completed_as_planned", response.WorkoutResult.Outcome)
	}
	if response.WorkoutResult.ImportedActivityID != activity.ID {
		t.Fatalf("imported activity id = %q, want %q", response.WorkoutResult.ImportedActivityID, activity.ID)
	}
	if response.UpdatedWorkout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("workout status = %q, want completed", response.UpdatedWorkout.Status)
	}
	if response.UpdatedWorkout.ID != workout.ID {
		t.Fatalf("updated workout id = %q, want %q", response.UpdatedWorkout.ID, workout.ID)
	}
	if response.AdaptationEvent != nil {
		t.Fatalf("expected no adaptation event, got %#v", response.AdaptationEvent)
	}
	assertProviderActivityCompletionPersisted(t, ctx, store, response, false)
}

func TestCompleteMatchedProviderActivityCreatesUnderperformanceAdaptation(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week, _, _, match := seedProviderActivityCompletionRecords(t, ctx, store, domain.WorkoutMatchStatusMatched)

	response, err := providerActivityCompletionService(store).CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: match.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-underperformed",
		Outcome:        domain.WorkoutOutcomeUnderperformed,
	})
	if err != nil {
		t.Fatalf("CompleteMatchedProviderActivity returned error: %v", err)
	}

	if response.WorkoutResult.Outcome != domain.WorkoutOutcomeUnderperformed {
		t.Fatalf("outcome = %q, want underperformed", response.WorkoutResult.Outcome)
	}
	if response.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}
	if response.AdaptationEvent.Type != domain.AdaptationTypeUnderperformance {
		t.Fatalf("adaptation type = %q, want underperformance", response.AdaptationEvent.Type)
	}
	assertProviderActivityCompletionPersisted(t, ctx, store, response, true)
}

func TestCompleteMatchedProviderActivityCreatesOverperformanceAdaptation(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week, _, _, match := seedProviderActivityCompletionRecords(t, ctx, store, domain.WorkoutMatchStatusMatched)

	response, err := providerActivityCompletionService(store).CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: match.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-overperformed",
		Outcome:        domain.WorkoutOutcomeOverperformed,
	})
	if err != nil {
		t.Fatalf("CompleteMatchedProviderActivity returned error: %v", err)
	}

	if response.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}
	if response.AdaptationEvent.Type != domain.AdaptationTypeOverperformance {
		t.Fatalf("adaptation type = %q, want overperformance", response.AdaptationEvent.Type)
	}
	assertProviderActivityCompletionPersisted(t, ctx, store, response, true)
}

func TestCompleteMatchedProviderActivityCreatesPartialCompletionAdaptation(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week, _, _, match := seedProviderActivityCompletionRecords(t, ctx, store, domain.WorkoutMatchStatusMatched)

	response, err := providerActivityCompletionService(store).CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: match.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-partial",
		Outcome:        domain.WorkoutOutcomePartiallyCompleted,
	})
	if err != nil {
		t.Fatalf("CompleteMatchedProviderActivity returned error: %v", err)
	}

	if response.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}
	if response.AdaptationEvent.Type != domain.AdaptationTypePartialCompletion {
		t.Fatalf("adaptation type = %q, want partial_completion", response.AdaptationEvent.Type)
	}
	assertProviderActivityCompletionPersisted(t, ctx, store, response, true)
}

func TestCompleteMatchedProviderActivityRejectsUncertainMatchWithoutResult(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week, _, _, match := seedProviderActivityCompletionRecords(t, ctx, store, domain.WorkoutMatchStatusUncertain)

	_, err := providerActivityCompletionService(store).CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: match.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-uncertain",
	})

	assertProviderActivityCompletionError(t, err, "must be matched")
	if _, getErr := store.GetWorkoutResult(ctx, "result-uncertain"); getErr == nil {
		t.Fatal("expected no workout result for uncertain match")
	}
}

func TestCompleteMatchedProviderActivityRejectsRejectedMatchWithoutResult(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	week, _, _, match := seedProviderActivityCompletionRecords(t, ctx, store, domain.WorkoutMatchStatusRejected)

	_, err := providerActivityCompletionService(store).CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: match.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-rejected",
	})

	assertProviderActivityCompletionError(t, err, "must be matched")
	if _, getErr := store.GetWorkoutResult(ctx, "result-rejected"); getErr == nil {
		t.Fatal("expected no workout result for rejected match")
	}
}

func providerActivityCompletionService(store repository.Store) ProviderActivityCompletionService {
	return ProviderActivityCompletionService{
		Store:            store,
		AdaptationEngine: adaptation.Engine{Now: providerActivityCompletionNow},
		Now:              providerActivityCompletionNow,
	}
}

func seedProviderActivityCompletionRecords(t *testing.T, ctx context.Context, store *repository.InMemoryStore, status domain.WorkoutMatchStatus) (domain.PlanWeek, domain.PlannedWorkout, domain.ImportedActivity, domain.WorkoutMatch) {
	t.Helper()

	week := providerActivityCompletionWeek()
	workout := week.Workouts[0]
	activity := providerActivityCompletionActivity(workout)
	match := domain.WorkoutMatch{
		ID:                 "match-provider-completion-" + string(status),
		PlannedWorkoutID:   workout.ID,
		ImportedActivityID: activity.ID,
		Status:             status,
		Confidence:         domain.MatchConfidenceHigh,
		MatchedBy:          domain.MatchSourceAutomatic,
		MatchedAt:          providerActivityCompletionNow(),
		Notes:              "Test match.",
	}
	if status != domain.WorkoutMatchStatusMatched {
		match.Confidence = domain.MatchConfidenceMedium
	}
	if err := store.SavePlanWeek(ctx, week); err != nil {
		t.Fatalf("SavePlanWeek returned error: %v", err)
	}
	for _, plannedWorkout := range week.Workouts {
		if err := store.SavePlannedWorkout(ctx, plannedWorkout); err != nil {
			t.Fatalf("SavePlannedWorkout returned error: %v", err)
		}
	}
	if err := store.SaveImportedActivity(ctx, activity); err != nil {
		t.Fatalf("SaveImportedActivity returned error: %v", err)
	}
	if err := store.SaveWorkoutMatch(ctx, match); err != nil {
		t.Fatalf("SaveWorkoutMatch returned error: %v", err)
	}
	return week, workout, activity, match
}

func assertProviderActivityCompletionPersisted(t *testing.T, ctx context.Context, store *repository.InMemoryStore, response CompleteMatchedProviderActivityResponse, wantAdaptation bool) {
	t.Helper()

	if _, err := store.GetWorkoutResult(ctx, response.WorkoutResult.ID); err != nil {
		t.Fatalf("expected saved workout result: %v", err)
	}
	savedWorkout, err := store.GetPlannedWorkout(ctx, response.UpdatedWorkout.ID)
	if err != nil {
		t.Fatalf("expected saved planned workout: %v", err)
	}
	if savedWorkout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("saved workout status = %q, want completed", savedWorkout.Status)
	}
	savedWeek, err := store.GetPlanWeek(ctx, response.PlanWeek.ID)
	if err != nil {
		t.Fatalf("expected saved plan week: %v", err)
	}
	if savedWorkoutStatus(savedWeek, response.UpdatedWorkout.ID) != domain.PlannedWorkoutStatusCompleted {
		t.Fatal("expected saved plan week to include completed workout")
	}
	if wantAdaptation {
		if response.AdaptationEvent == nil {
			t.Fatal("expected adaptation event")
		}
		if _, err := store.GetAdaptationEvent(ctx, response.AdaptationEvent.ID); err != nil {
			t.Fatalf("expected saved adaptation event: %v", err)
		}
	}
}

func providerActivityCompletionWeek() domain.PlanWeek {
	first := domain.PlannedWorkout{
		ID:             "workout-provider-completion-1",
		PlanID:         "plan-provider-completion",
		PlanWeekID:     "week-provider-completion",
		ScheduledFor:   providerActivityCompletionDateTime(2026, time.June, 15, 0, 0),
		Type:           domain.WorkoutTypeWorkout,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 8000},
		TargetDuration: 45 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindTempo},
	}
	second := domain.PlannedWorkout{
		ID:             "workout-provider-completion-2",
		PlanID:         "plan-provider-completion",
		PlanWeekID:     "week-provider-completion",
		ScheduledFor:   providerActivityCompletionDateTime(2026, time.June, 17, 0, 0),
		Type:           domain.WorkoutTypeEasy,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 6000},
		TargetDuration: 35 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindEasy},
	}
	return domain.PlanWeek{
		ID:        "week-provider-completion",
		AthleteID: "athlete-provider-completion",
		PlanID:    "plan-provider-completion",
		WeekIndex: 1,
		StartsOn:  providerActivityCompletionDateTime(2026, time.June, 15, 0, 0),
		Focus:     domain.WeekFocusBuild,
		Workouts:  []domain.PlannedWorkout{first, second},
	}
}

func providerActivityCompletionActivity(workout domain.PlannedWorkout) domain.ImportedActivity {
	return domain.ImportedActivity{
		ID:              "imported-provider-completion",
		AthleteID:       "athlete-provider-completion",
		Type:            domain.ActivityTypeRun,
		StartedAt:       workout.ScheduledFor.Add(7 * time.Hour),
		Duration:        44 * time.Minute,
		Distance:        domain.Distance{Meters: 7900},
		AveragePace:     domain.Pace{SecondsPerKilometer: 334},
		AverageHeartBPM: 146,
	}
}

func providerActivityCompletionNow() time.Time {
	return providerActivityCompletionDateTime(2026, time.June, 15, 12, 0)
}

func providerActivityCompletionDateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func assertProviderActivityCompletionError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
}
