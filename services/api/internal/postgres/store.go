package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type Store struct {
	AthleteProfiles    repository.AthleteProfileRepository
	TrainingGoals      repository.TrainingGoalRepository
	PlanWeeks          repository.PlanWeekRepository
	PlannedWorkouts    repository.PlannedWorkoutRepository
	ImportedActivities repository.ImportedActivityRepository
	WorkoutMatches     repository.WorkoutMatchRepository
	WorkoutResults     repository.WorkoutResultRepository
	AdaptationEvents   repository.AdaptationEventRepository
}

var _ repository.Store = (*Store)(nil)

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres store db is required")
	}

	return &Store{
		AthleteProfiles:    NewAthleteProfileRepository(db),
		TrainingGoals:      NewTrainingGoalRepository(db),
		PlanWeeks:          NewPlanWeekRepository(db),
		PlannedWorkouts:    NewPlannedWorkoutRepository(db),
		ImportedActivities: NewImportedActivityRepository(db),
		WorkoutMatches:     NewWorkoutMatchRepository(db),
		WorkoutResults:     NewWorkoutResultRepository(db),
		AdaptationEvents:   NewAdaptationEventRepository(db),
	}, nil
}

func (s *Store) SaveAthleteProfile(ctx context.Context, profile domain.AthleteProfile) error {
	return s.AthleteProfiles.SaveAthleteProfile(ctx, profile)
}

func (s *Store) GetAthleteProfile(ctx context.Context, id string) (domain.AthleteProfile, error) {
	return s.AthleteProfiles.GetAthleteProfile(ctx, id)
}

func (s *Store) SaveTrainingGoal(ctx context.Context, goal domain.TrainingGoal) error {
	return s.TrainingGoals.SaveTrainingGoal(ctx, goal)
}

func (s *Store) GetTrainingGoal(ctx context.Context, id string) (domain.TrainingGoal, error) {
	return s.TrainingGoals.GetTrainingGoal(ctx, id)
}

func (s *Store) GetCurrentTrainingGoal(ctx context.Context, athleteID string) (domain.TrainingGoal, error) {
	currentGoals, ok := s.TrainingGoals.(repository.CurrentTrainingGoalRepository)
	if !ok {
		return domain.TrainingGoal{}, fmt.Errorf("current training goal repository is required")
	}
	return currentGoals.GetCurrentTrainingGoal(ctx, athleteID)
}

func (s *Store) SavePlanWeek(ctx context.Context, week domain.PlanWeek) error {
	return s.PlanWeeks.SavePlanWeek(ctx, week)
}

func (s *Store) GetPlanWeek(ctx context.Context, id string) (domain.PlanWeek, error) {
	return s.PlanWeeks.GetPlanWeek(ctx, id)
}

func (s *Store) SavePlannedWorkout(ctx context.Context, workout domain.PlannedWorkout) error {
	return s.PlannedWorkouts.SavePlannedWorkout(ctx, workout)
}

func (s *Store) GetPlannedWorkout(ctx context.Context, id string) (domain.PlannedWorkout, error) {
	return s.PlannedWorkouts.GetPlannedWorkout(ctx, id)
}

func (s *Store) SaveImportedActivity(ctx context.Context, activity domain.ImportedActivity) error {
	return s.ImportedActivities.SaveImportedActivity(ctx, activity)
}

func (s *Store) GetImportedActivity(ctx context.Context, id string) (domain.ImportedActivity, error) {
	return s.ImportedActivities.GetImportedActivity(ctx, id)
}

func (s *Store) SaveWorkoutMatch(ctx context.Context, match domain.WorkoutMatch) error {
	return s.WorkoutMatches.SaveWorkoutMatch(ctx, match)
}

func (s *Store) GetWorkoutMatch(ctx context.Context, id string) (domain.WorkoutMatch, error) {
	return s.WorkoutMatches.GetWorkoutMatch(ctx, id)
}

func (s *Store) SaveWorkoutResult(ctx context.Context, result domain.WorkoutResult) error {
	return s.WorkoutResults.SaveWorkoutResult(ctx, result)
}

func (s *Store) GetWorkoutResult(ctx context.Context, id string) (domain.WorkoutResult, error) {
	return s.WorkoutResults.GetWorkoutResult(ctx, id)
}

func (s *Store) SaveAdaptationEvent(ctx context.Context, event domain.AdaptationEvent) error {
	return s.AdaptationEvents.SaveAdaptationEvent(ctx, event)
}

func (s *Store) GetAdaptationEvent(ctx context.Context, id string) (domain.AdaptationEvent, error) {
	return s.AdaptationEvents.GetAdaptationEvent(ctx, id)
}
