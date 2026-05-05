package app

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/coreloop"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type ActivityImporter func(context.Context) (domain.ImportedActivity, error)

type CoreLoopRunner interface {
	CompleteImportedActivity(coreloop.CompleteImportedActivityInput) (coreloop.CompleteImportedActivityResult, error)
}

type CoreLoopService struct {
	Runner CoreLoopRunner
	Store  repository.Store
}

type CompleteImportedActivityRequest struct {
	AthleteProfile domain.AthleteProfile
	TrainingGoal   domain.TrainingGoal
	TargetWeekDate time.Time
	ImportActivity ActivityImporter
	ResultID       string
	Outcome        domain.WorkoutOutcome
}

type CompleteImportedActivityResponse struct {
	PlanWeek         domain.PlanWeek
	ImportedActivity domain.ImportedActivity
	WorkoutMatch     domain.WorkoutMatch
	UpdatedWorkout   domain.PlannedWorkout
	WorkoutResult    domain.WorkoutResult
	AdaptationEvent  *domain.AdaptationEvent
}

func NewInMemoryCoreLoopService() CoreLoopService {
	return CoreLoopService{
		Runner: coreloop.NewService(),
		Store:  repository.NewInMemoryStore(),
	}
}

func (s CoreLoopService) CompleteImportedActivity(ctx context.Context, request CompleteImportedActivityRequest) (CompleteImportedActivityResponse, error) {
	if s.Runner == nil {
		return CompleteImportedActivityResponse{}, fmt.Errorf("core loop runner is required")
	}
	if s.Store == nil {
		return CompleteImportedActivityResponse{}, fmt.Errorf("repository store is required")
	}
	if request.ImportActivity == nil {
		return CompleteImportedActivityResponse{}, fmt.Errorf("activity importer is required")
	}

	result, err := s.Runner.CompleteImportedActivity(coreloop.CompleteImportedActivityInput{
		AthleteProfile: request.AthleteProfile,
		TrainingGoal:   request.TrainingGoal,
		TargetWeekDate: request.TargetWeekDate,
		ImportActivity: func() (domain.ImportedActivity, error) {
			return request.ImportActivity(ctx)
		},
		ResultID: request.ResultID,
		Outcome:  request.Outcome,
	})
	if err != nil {
		return CompleteImportedActivityResponse{}, err
	}

	persistedWeek := replaceWorkout(result.PlanWeek, result.UpdatedWorkout)
	result.PlanWeek = persistedWeek

	if err := s.persistCoreLoopResult(ctx, request, result); err != nil {
		return CompleteImportedActivityResponse{}, err
	}

	return CompleteImportedActivityResponse{
		PlanWeek:         result.PlanWeek,
		ImportedActivity: result.ImportedActivity,
		WorkoutMatch:     result.WorkoutMatch,
		UpdatedWorkout:   result.UpdatedWorkout,
		WorkoutResult:    result.WorkoutResult,
		AdaptationEvent:  result.AdaptationEvent,
	}, nil
}

func (s CoreLoopService) persistCoreLoopResult(ctx context.Context, request CompleteImportedActivityRequest, result coreloop.CompleteImportedActivityResult) error {
	if err := s.Store.SaveAthleteProfile(ctx, request.AthleteProfile); err != nil {
		return fmt.Errorf("save athlete profile: %w", err)
	}
	if err := s.Store.SaveTrainingGoal(ctx, request.TrainingGoal); err != nil {
		return fmt.Errorf("save training goal: %w", err)
	}
	if err := s.Store.SavePlanWeek(ctx, result.PlanWeek); err != nil {
		return fmt.Errorf("save plan week: %w", err)
	}
	for _, workout := range result.PlanWeek.Workouts {
		if err := s.Store.SavePlannedWorkout(ctx, workout); err != nil {
			return fmt.Errorf("save planned workout %q: %w", workout.ID, err)
		}
	}
	if err := s.Store.SaveImportedActivity(ctx, result.ImportedActivity); err != nil {
		return fmt.Errorf("save imported activity: %w", err)
	}
	if err := s.Store.SaveWorkoutMatch(ctx, result.WorkoutMatch); err != nil {
		return fmt.Errorf("save workout match: %w", err)
	}
	if err := s.Store.SaveWorkoutResult(ctx, result.WorkoutResult); err != nil {
		return fmt.Errorf("save workout result: %w", err)
	}
	if result.AdaptationEvent != nil {
		if err := s.Store.SaveAdaptationEvent(ctx, *result.AdaptationEvent); err != nil {
			return fmt.Errorf("save adaptation event: %w", err)
		}
	}
	return nil
}

func replaceWorkout(week domain.PlanWeek, updatedWorkout domain.PlannedWorkout) domain.PlanWeek {
	for i, workout := range week.Workouts {
		if workout.ID == updatedWorkout.ID {
			week.Workouts[i] = updatedWorkout
			return week
		}
	}
	return week
}
