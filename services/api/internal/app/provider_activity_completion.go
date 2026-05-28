package app

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type ProviderActivityCompletionService struct {
	Store            repository.Store
	AdaptationEngine adaptation.Engine
	Now              func() time.Time
}

type CompleteMatchedProviderActivityRequest struct {
	WorkoutMatchID   string
	WorkoutMatch     domain.WorkoutMatch
	ImportedActivity domain.ImportedActivity
	PlanWeekID       string
	PlanWeek         domain.PlanWeek
	PlannedWorkout   domain.PlannedWorkout
	ResultID         string
	Outcome          domain.WorkoutOutcome
	Notes            string
}

type CompleteMatchedProviderActivityResponse struct {
	WorkoutMatch     domain.WorkoutMatch
	ImportedActivity domain.ImportedActivity
	PlanWeek         domain.PlanWeek
	UpdatedWorkout   domain.PlannedWorkout
	WorkoutResult    domain.WorkoutResult
	AdaptationEvent  *domain.AdaptationEvent
}

func (s ProviderActivityCompletionService) CompleteMatchedProviderActivity(ctx context.Context, request CompleteMatchedProviderActivityRequest) (CompleteMatchedProviderActivityResponse, error) {
	if s.Store == nil {
		return CompleteMatchedProviderActivityResponse{}, fmt.Errorf("repository store is required")
	}

	match, err := s.workoutMatch(ctx, request)
	if err != nil {
		return CompleteMatchedProviderActivityResponse{}, err
	}
	if match.Status != domain.WorkoutMatchStatusMatched {
		return CompleteMatchedProviderActivityResponse{WorkoutMatch: match}, fmt.Errorf("workout match must be matched before completion: %s", match.Status)
	}

	activity, err := s.importedActivity(ctx, request, match)
	if err != nil {
		return CompleteMatchedProviderActivityResponse{WorkoutMatch: match}, err
	}
	week, err := s.planWeek(ctx, request)
	if err != nil {
		return CompleteMatchedProviderActivityResponse{WorkoutMatch: match, ImportedActivity: activity}, err
	}
	workout, err := s.plannedWorkout(ctx, request, week, match)
	if err != nil {
		return CompleteMatchedProviderActivityResponse{WorkoutMatch: match, ImportedActivity: activity, PlanWeek: week}, err
	}

	outcome := request.Outcome
	if outcome == "" {
		outcome = domain.WorkoutOutcomeCompletedAsPlanned
	}
	resultID := request.ResultID
	if resultID == "" {
		resultID = deterministicUUID("runthread:workout-result", match.ID)
	}
	notes := request.Notes
	if notes == "" {
		notes = "Created from matched provider activity."
		if matching.IsRideCrossTrainingMatch(workout.Type, activity.Type) {
			notes = "Completed by ride cross-training activity."
		}
	}

	updatedWorkout, result, err := domain.MarkWorkoutCompleted(workout, domain.WorkoutCompletion{
		ResultID:           resultID,
		ImportedActivityID: activity.ID,
		CompletedAt:        activity.StartedAt.Add(activity.Duration),
		Distance:           activity.Distance,
		Duration:           activity.Duration,
		Outcome:            outcome,
		Notes:              notes,
	})
	if err != nil {
		return CompleteMatchedProviderActivityResponse{WorkoutMatch: match, ImportedActivity: activity, PlanWeek: week}, fmt.Errorf("mark workout completed: %w", err)
	}

	updatedWeek := replaceWorkout(week, updatedWorkout)
	adaptationEvent, err := s.adaptationEngine().AdaptWorkoutResult(adaptation.WorkoutResultInput{
		AthleteID: updatedWeek.AthleteID,
		PlanWeek:  updatedWeek,
		Result:    result,
	})
	if err != nil {
		return CompleteMatchedProviderActivityResponse{}, fmt.Errorf("adapt workout result: %w", err)
	}
	if err := s.persistCompletion(ctx, updatedWeek, updatedWorkout, result, adaptationEvent); err != nil {
		return CompleteMatchedProviderActivityResponse{}, err
	}

	return CompleteMatchedProviderActivityResponse{
		WorkoutMatch:     match,
		ImportedActivity: activity,
		PlanWeek:         updatedWeek,
		UpdatedWorkout:   updatedWorkout,
		WorkoutResult:    result,
		AdaptationEvent:  adaptationEvent,
	}, nil
}

func (s ProviderActivityCompletionService) workoutMatch(ctx context.Context, request CompleteMatchedProviderActivityRequest) (domain.WorkoutMatch, error) {
	if request.WorkoutMatch.ID != "" {
		if err := request.WorkoutMatch.Validate(); err != nil {
			return domain.WorkoutMatch{}, fmt.Errorf("invalid workout match: %w", err)
		}
		return request.WorkoutMatch, nil
	}
	if request.WorkoutMatchID == "" {
		return domain.WorkoutMatch{}, fmt.Errorf("workout match or workout match id is required")
	}
	match, err := s.Store.GetWorkoutMatch(ctx, request.WorkoutMatchID)
	if err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("get workout match: %w", err)
	}
	return match, nil
}

func (s ProviderActivityCompletionService) importedActivity(ctx context.Context, request CompleteMatchedProviderActivityRequest, match domain.WorkoutMatch) (domain.ImportedActivity, error) {
	if request.ImportedActivity.ID != "" {
		if err := request.ImportedActivity.Validate(); err != nil {
			return domain.ImportedActivity{}, fmt.Errorf("invalid imported activity: %w", err)
		}
		if request.ImportedActivity.ID != match.ImportedActivityID {
			return domain.ImportedActivity{}, fmt.Errorf("imported activity %q does not match workout match imported activity %q", request.ImportedActivity.ID, match.ImportedActivityID)
		}
		return request.ImportedActivity, nil
	}
	activity, err := s.Store.GetImportedActivity(ctx, match.ImportedActivityID)
	if err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("get imported activity: %w", err)
	}
	return activity, nil
}

func (s ProviderActivityCompletionService) planWeek(ctx context.Context, request CompleteMatchedProviderActivityRequest) (domain.PlanWeek, error) {
	if request.PlanWeek.ID != "" {
		if err := request.PlanWeek.Validate(); err != nil {
			return domain.PlanWeek{}, fmt.Errorf("invalid plan week: %w", err)
		}
		return request.PlanWeek, nil
	}
	if request.PlanWeekID == "" {
		return domain.PlanWeek{}, fmt.Errorf("plan week or plan week id is required")
	}
	week, err := s.Store.GetPlanWeek(ctx, request.PlanWeekID)
	if err != nil {
		return domain.PlanWeek{}, fmt.Errorf("get plan week: %w", err)
	}
	return week, nil
}

func (s ProviderActivityCompletionService) plannedWorkout(ctx context.Context, request CompleteMatchedProviderActivityRequest, week domain.PlanWeek, match domain.WorkoutMatch) (domain.PlannedWorkout, error) {
	if request.PlannedWorkout.ID != "" {
		if err := request.PlannedWorkout.Validate(); err != nil {
			return domain.PlannedWorkout{}, fmt.Errorf("invalid planned workout: %w", err)
		}
		if request.PlannedWorkout.ID != match.PlannedWorkoutID {
			return domain.PlannedWorkout{}, fmt.Errorf("planned workout %q does not match workout match planned workout %q", request.PlannedWorkout.ID, match.PlannedWorkoutID)
		}
		return request.PlannedWorkout, nil
	}
	return plannedWorkoutFromWeekOrStore(ctx, s.Store, week, match.PlannedWorkoutID)
}

func (s ProviderActivityCompletionService) persistCompletion(ctx context.Context, week domain.PlanWeek, workout domain.PlannedWorkout, result domain.WorkoutResult, adaptationEvent *domain.AdaptationEvent) error {
	if err := s.Store.SaveWorkoutResult(ctx, result); err != nil {
		return fmt.Errorf("save workout result: %w", err)
	}
	if err := s.Store.SavePlannedWorkout(ctx, workout); err != nil {
		return fmt.Errorf("save planned workout: %w", err)
	}
	if err := s.Store.SavePlanWeek(ctx, week); err != nil {
		return fmt.Errorf("save plan week: %w", err)
	}
	if adaptationEvent != nil {
		if err := s.Store.SaveAdaptationEvent(ctx, *adaptationEvent); err != nil {
			return fmt.Errorf("save adaptation event: %w", err)
		}
	}
	return nil
}

func (s ProviderActivityCompletionService) adaptationEngine() adaptation.Engine {
	if s.AdaptationEngine.Now != nil {
		return s.AdaptationEngine
	}
	return adaptation.Engine{Now: s.now}
}

func (s ProviderActivityCompletionService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
