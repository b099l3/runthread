package repository

import (
	"context"
	"errors"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

var ErrNotFound = errors.New("repository record not found")

type AthleteProfileRepository interface {
	SaveAthleteProfile(context.Context, domain.AthleteProfile) error
	GetAthleteProfile(context.Context, string) (domain.AthleteProfile, error)
}

type TrainingGoalRepository interface {
	SaveTrainingGoal(context.Context, domain.TrainingGoal) error
	GetTrainingGoal(context.Context, string) (domain.TrainingGoal, error)
}

type PlanWeekRepository interface {
	SavePlanWeek(context.Context, domain.PlanWeek) error
	GetPlanWeek(context.Context, string) (domain.PlanWeek, error)
}

type PlannedWorkoutRepository interface {
	SavePlannedWorkout(context.Context, domain.PlannedWorkout) error
	GetPlannedWorkout(context.Context, string) (domain.PlannedWorkout, error)
}

type ImportedActivityRepository interface {
	SaveImportedActivity(context.Context, domain.ImportedActivity) error
	GetImportedActivity(context.Context, string) (domain.ImportedActivity, error)
}

type WorkoutMatchRepository interface {
	SaveWorkoutMatch(context.Context, domain.WorkoutMatch) error
	GetWorkoutMatch(context.Context, string) (domain.WorkoutMatch, error)
}

type WorkoutResultRepository interface {
	SaveWorkoutResult(context.Context, domain.WorkoutResult) error
	GetWorkoutResult(context.Context, string) (domain.WorkoutResult, error)
}

type AdaptationEventRepository interface {
	SaveAdaptationEvent(context.Context, domain.AdaptationEvent) error
	GetAdaptationEvent(context.Context, string) (domain.AdaptationEvent, error)
}

type ProviderConnectionRepository interface {
	SaveProviderConnection(context.Context, ProviderConnection) error
	GetProviderConnection(context.Context, string) (ProviderConnection, error)
	ListProviderConnectionsByAthlete(context.Context, string) ([]ProviderConnection, error)
	ListProviderConnectionsByStatus(context.Context, ProviderConnectionStatus) ([]ProviderConnection, error)
}

type ProviderActivityRepository interface {
	SaveProviderActivity(context.Context, ProviderActivity) error
	GetProviderActivity(context.Context, string) (ProviderActivity, error)
	GetProviderActivityByProviderID(context.Context, string, string) (ProviderActivity, error)
	ListProviderActivitiesByAthlete(context.Context, string) ([]ProviderActivity, error)
	ListProviderActivitiesByStatus(context.Context, ProviderActivityStatus) ([]ProviderActivity, error)
}

type ProviderActivityPayloadRepository interface {
	SaveProviderActivityPayload(context.Context, ProviderActivityPayload) error
	GetProviderActivityPayload(context.Context, string) (ProviderActivityPayload, error)
	ListProviderActivityPayloads(context.Context, string) ([]ProviderActivityPayload, error)
}

type ProviderImportEventRepository interface {
	SaveProviderImportEvent(context.Context, ProviderImportEvent) error
	GetProviderImportEvent(context.Context, string) (ProviderImportEvent, error)
	ListProviderImportEventsByConnection(context.Context, string) ([]ProviderImportEvent, error)
	ListProviderImportEventsByStatus(context.Context, ProviderImportEventStatus) ([]ProviderImportEvent, error)
}

type Store interface {
	AthleteProfileRepository
	TrainingGoalRepository
	PlanWeekRepository
	PlannedWorkoutRepository
	ImportedActivityRepository
	WorkoutMatchRepository
	WorkoutResultRepository
	AdaptationEventRepository
}

type ProviderStore interface {
	ProviderConnectionRepository
	ProviderActivityRepository
	ProviderActivityPayloadRepository
	ProviderImportEventRepository
}

type PlanWeekSnapshot struct {
	PlanWeek           domain.PlanWeek
	ImportedActivities []domain.ImportedActivity
	WorkoutMatches     []domain.WorkoutMatch
	WorkoutResults     []domain.WorkoutResult
	AdaptationEvents   []domain.AdaptationEvent
}

type CurrentPlanWeekReader interface {
	GetCurrentPlanWeekSnapshot(context.Context, CurrentPlanWeekQuery) (PlanWeekSnapshot, error)
}

type CurrentPlanWeekQuery struct {
	PlanWeekID     string
	AthleteID      string
	GoalID         string
	TargetWeekDate time.Time
}
