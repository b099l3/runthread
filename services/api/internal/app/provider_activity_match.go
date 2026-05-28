package app

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type ProviderActivityMatchService struct {
	Store   repository.Store
	Matcher matching.Matcher
	Now     func() time.Time
}

type automaticWorkoutMatchPruner interface {
	DeleteAutomaticWorkoutMatchesByImportedActivity(ctx context.Context, importedActivityID string, keepMatchID string) error
}

type MatchProviderActivityRequest struct {
	ImportedActivityID string
	ImportedActivity   domain.ImportedActivity
	PlanWeekID         string
	PlanWeek           domain.PlanWeek
	PlannedWorkoutID   string
	PlannedWorkout     domain.PlannedWorkout
	Manual             bool
	ManualNotes        string
}

type MatchProviderActivityResponse struct {
	ImportedActivity domain.ImportedActivity
	PlanWeek         domain.PlanWeek
	PlannedWorkout   domain.PlannedWorkout
	WorkoutMatch     domain.WorkoutMatch
}

func (s ProviderActivityMatchService) MatchProviderActivity(ctx context.Context, request MatchProviderActivityRequest) (MatchProviderActivityResponse, error) {
	if s.Store == nil {
		return MatchProviderActivityResponse{}, fmt.Errorf("repository store is required")
	}

	activity, err := s.importedActivity(ctx, request)
	if err != nil {
		return MatchProviderActivityResponse{}, err
	}
	week, err := s.planWeek(ctx, request)
	if err != nil {
		return MatchProviderActivityResponse{ImportedActivity: activity}, err
	}
	workout, match, err := s.workoutAndMatch(ctx, request, week, activity)
	if err != nil {
		return MatchProviderActivityResponse{ImportedActivity: activity, PlanWeek: week}, err
	}
	if err := s.Store.SaveWorkoutMatch(ctx, match); err != nil {
		return MatchProviderActivityResponse{ImportedActivity: activity, PlanWeek: week, PlannedWorkout: workout, WorkoutMatch: match}, fmt.Errorf("save workout match: %w", err)
	}
	if pruner, ok := s.Store.(automaticWorkoutMatchPruner); ok {
		if err := pruner.DeleteAutomaticWorkoutMatchesByImportedActivity(ctx, activity.ID, match.ID); err != nil {
			return MatchProviderActivityResponse{ImportedActivity: activity, PlanWeek: week, PlannedWorkout: workout, WorkoutMatch: match}, fmt.Errorf("prune stale workout matches: %w", err)
		}
	}
	return MatchProviderActivityResponse{
		ImportedActivity: activity,
		PlanWeek:         week,
		PlannedWorkout:   workout,
		WorkoutMatch:     match,
	}, nil
}

func (s ProviderActivityMatchService) importedActivity(ctx context.Context, request MatchProviderActivityRequest) (domain.ImportedActivity, error) {
	if request.ImportedActivity.ID != "" {
		if err := request.ImportedActivity.Validate(); err != nil {
			return domain.ImportedActivity{}, fmt.Errorf("invalid imported activity: %w", err)
		}
		return request.ImportedActivity, nil
	}
	if request.ImportedActivityID == "" {
		return domain.ImportedActivity{}, fmt.Errorf("imported activity or imported activity id is required")
	}
	activity, err := s.Store.GetImportedActivity(ctx, request.ImportedActivityID)
	if err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("get imported activity: %w", err)
	}
	return activity, nil
}

func (s ProviderActivityMatchService) planWeek(ctx context.Context, request MatchProviderActivityRequest) (domain.PlanWeek, error) {
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

func (s ProviderActivityMatchService) workoutAndMatch(ctx context.Context, request MatchProviderActivityRequest, week domain.PlanWeek, activity domain.ImportedActivity) (domain.PlannedWorkout, domain.WorkoutMatch, error) {
	workout, err := s.plannedWorkout(ctx, request, week, activity)
	if err != nil {
		return domain.PlannedWorkout{}, domain.WorkoutMatch{}, err
	}
	if request.Manual {
		match, err := matching.ManualMatch(workout, activity, s.now(), request.ManualNotes)
		return workout, match, err
	}
	match, err := s.matcher().MatchActivity(workout, activity)
	return workout, match, err
}

func (s ProviderActivityMatchService) plannedWorkout(ctx context.Context, request MatchProviderActivityRequest, week domain.PlanWeek, activity domain.ImportedActivity) (domain.PlannedWorkout, error) {
	if request.PlannedWorkout.ID != "" {
		if err := request.PlannedWorkout.Validate(); err != nil {
			return domain.PlannedWorkout{}, fmt.Errorf("invalid planned workout: %w", err)
		}
		return request.PlannedWorkout, nil
	}
	if request.PlannedWorkoutID != "" {
		return plannedWorkoutFromWeekOrStore(ctx, s.Store, week, request.PlannedWorkoutID)
	}
	workout, _, err := bestProviderImportMatch(s.matcher(), week, activity)
	return workout, err
}

func (s ProviderActivityMatchService) matcher() matching.Matcher {
	if s.Matcher.Now != nil {
		return s.Matcher
	}
	return matching.Matcher{Now: s.now}
}

func (s ProviderActivityMatchService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
