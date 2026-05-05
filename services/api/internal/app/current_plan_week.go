package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type CurrentPlanWeekService struct {
	Store   repository.Store
	Planner planning.WeeklyPlanner
}

type GetCurrentPlanWeekRequest struct {
	PlanWeekID     string
	AthleteID      string
	GoalID         string
	TargetWeekDate time.Time
}

type GetCurrentPlanWeekResponse struct {
	PlanWeek           domain.PlanWeek
	ImportedActivities []domain.ImportedActivity
	WorkoutMatches     []domain.WorkoutMatch
	WorkoutResults     []domain.WorkoutResult
	AdaptationEvents   []domain.AdaptationEvent
}

func (s CurrentPlanWeekService) GetCurrentPlanWeek(ctx context.Context, request GetCurrentPlanWeekRequest) (GetCurrentPlanWeekResponse, error) {
	if s.Store == nil {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("repository store is required")
	}
	if request.PlanWeekID == "" {
		if request.AthleteID == "" {
			return GetCurrentPlanWeekResponse{}, fmt.Errorf("athlete id is required")
		}
		if request.TargetWeekDate.IsZero() {
			return GetCurrentPlanWeekResponse{}, fmt.Errorf("target week date is required")
		}
	}

	if reader, ok := s.Store.(repository.CurrentPlanWeekReader); ok {
		snapshot, err := reader.GetCurrentPlanWeekSnapshot(ctx, repository.CurrentPlanWeekQuery{
			PlanWeekID:     request.PlanWeekID,
			AthleteID:      request.AthleteID,
			GoalID:         request.GoalID,
			TargetWeekDate: request.TargetWeekDate,
		})
		if err == nil {
			return responseFromSnapshot(snapshot), nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return GetCurrentPlanWeekResponse{}, err
		}
	}

	if request.PlanWeekID != "" {
		week, err := s.Store.GetPlanWeek(ctx, request.PlanWeekID)
		if err == nil {
			return GetCurrentPlanWeekResponse{PlanWeek: week}, nil
		}
		return GetCurrentPlanWeekResponse{}, err
	}

	return s.generateDemoWeek(ctx, request)
}

func (s CurrentPlanWeekService) generateDemoWeek(ctx context.Context, request GetCurrentPlanWeekRequest) (GetCurrentPlanWeekResponse, error) {
	if request.GoalID == "" {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("goal id is required when no saved plan week exists")
	}

	profile, err := s.Store.GetAthleteProfile(ctx, request.AthleteID)
	if err != nil {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("get athlete profile: %w", err)
	}
	goal, err := s.Store.GetTrainingGoal(ctx, request.GoalID)
	if err != nil {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("get training goal: %w", err)
	}
	if goal.AthleteID != profile.ID {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("training goal does not belong to athlete")
	}

	planner := s.Planner
	week, err := planner.GenerateWeek(profile, goal, request.TargetWeekDate)
	if err != nil {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("generate plan week: %w", err)
	}
	if err := s.Store.SavePlanWeek(ctx, week); err != nil {
		return GetCurrentPlanWeekResponse{}, fmt.Errorf("save generated plan week: %w", err)
	}
	for _, workout := range week.Workouts {
		if err := s.Store.SavePlannedWorkout(ctx, workout); err != nil {
			return GetCurrentPlanWeekResponse{}, fmt.Errorf("save generated planned workout %q: %w", workout.ID, err)
		}
	}

	return GetCurrentPlanWeekResponse{PlanWeek: week}, nil
}

func responseFromSnapshot(snapshot repository.PlanWeekSnapshot) GetCurrentPlanWeekResponse {
	return GetCurrentPlanWeekResponse{
		PlanWeek:           snapshot.PlanWeek,
		ImportedActivities: append([]domain.ImportedActivity(nil), snapshot.ImportedActivities...),
		WorkoutMatches:     append([]domain.WorkoutMatch(nil), snapshot.WorkoutMatches...),
		WorkoutResults:     append([]domain.WorkoutResult(nil), snapshot.WorkoutResults...),
		AdaptationEvents:   append([]domain.AdaptationEvent(nil), snapshot.AdaptationEvents...),
	}
}
