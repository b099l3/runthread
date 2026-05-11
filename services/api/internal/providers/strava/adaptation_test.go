package strava

import (
	"context"
	"testing"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestImportedStravaActivityCanProduceWorkoutResultAndAdaptation(t *testing.T) {
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

	matchService := app.ProviderActivityMatchService{
		Store:   store,
		Matcher: matching.Matcher{Now: stravaMatchNow},
		Now:     stravaMatchNow,
	}
	matchResponse, err := matchService.MatchProviderActivity(ctx, app.MatchProviderActivityRequest{
		ImportedActivityID: backfillResult.Imports[0].ImportedActivity.ID,
		PlanWeekID:         week.ID,
		PlannedWorkoutID:   workout.ID,
	})
	if err != nil {
		t.Fatalf("MatchProviderActivity returned error: %v", err)
	}

	completionService := app.ProviderActivityCompletionService{
		Store:            store,
		AdaptationEngine: adaptation.Engine{Now: stravaMatchNow},
		Now:              stravaMatchNow,
	}
	completionResponse, err := completionService.CompleteMatchedProviderActivity(ctx, app.CompleteMatchedProviderActivityRequest{
		WorkoutMatchID: matchResponse.WorkoutMatch.ID,
		PlanWeekID:     week.ID,
		ResultID:       "result-strava-underperformed",
		Outcome:        domain.WorkoutOutcomeUnderperformed,
	})
	if err != nil {
		t.Fatalf("CompleteMatchedProviderActivity returned error: %v", err)
	}

	if completionResponse.WorkoutResult.ImportedActivityID != backfillResult.Imports[0].ImportedActivity.ID {
		t.Fatalf("result imported activity = %q, want %q", completionResponse.WorkoutResult.ImportedActivityID, backfillResult.Imports[0].ImportedActivity.ID)
	}
	if completionResponse.AdaptationEvent == nil {
		t.Fatal("expected adaptation event")
	}
	if completionResponse.AdaptationEvent.Type != domain.AdaptationTypeUnderperformance {
		t.Fatalf("adaptation type = %q, want underperformance", completionResponse.AdaptationEvent.Type)
	}
	if _, err := store.GetWorkoutResult(ctx, completionResponse.WorkoutResult.ID); err != nil {
		t.Fatalf("expected saved workout result: %v", err)
	}
	if _, err := store.GetAdaptationEvent(ctx, completionResponse.AdaptationEvent.ID); err != nil {
		t.Fatalf("expected saved adaptation event: %v", err)
	}
}
