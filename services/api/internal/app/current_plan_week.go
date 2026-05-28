package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
)

const generatedIDDateLayout = "2006-01-02"

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
			request.TargetWeekDate = time.Now()
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
	week = normaliseGeneratedPlanWeekIDs(week)
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

func normaliseGeneratedPlanWeekIDs(week domain.PlanWeek) domain.PlanWeek {
	if _, err := uuid.Parse(week.ID); err != nil {
		week.ID = deterministicUUID("plan-week", week.AthleteID, week.GoalID, week.StartsOn.Format(generatedIDDateLayout))
	}
	if _, err := uuid.Parse(week.PlanID); err != nil {
		week.PlanID = deterministicUUID("plan", week.AthleteID, week.GoalID, week.StartsOn.Format(generatedIDDateLayout))
	}
	for i := range week.Workouts {
		workout := &week.Workouts[i]
		originalID := workout.ID
		if _, err := uuid.Parse(workout.ID); err != nil {
			workout.ID = deterministicUUID("planned-workout", week.ID, originalID, workout.ScheduledFor.Format(generatedIDDateLayout), string(workout.Type))
		}
		workout.PlanID = week.PlanID
		workout.PlanWeekID = week.ID
	}
	return week
}

func deterministicUUID(parts ...string) string {
	name := ""
	for i, part := range parts {
		if i > 0 {
			name += ":"
		}
		name += part
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)).String()
}

func responseFromSnapshot(snapshot repository.PlanWeekSnapshot) GetCurrentPlanWeekResponse {
	return GetCurrentPlanWeekResponse{
		PlanWeek:           snapshot.PlanWeek,
		ImportedActivities: append([]domain.ImportedActivity(nil), snapshot.ImportedActivities...),
		WorkoutMatches:     latestWorkoutMatches(snapshot.WorkoutMatches),
		WorkoutResults:     append([]domain.WorkoutResult(nil), snapshot.WorkoutResults...),
		AdaptationEvents:   append([]domain.AdaptationEvent(nil), snapshot.AdaptationEvents...),
	}
}

func latestWorkoutMatches(matches []domain.WorkoutMatch) []domain.WorkoutMatch {
	latestByActivity := make(map[string]domain.WorkoutMatch, len(matches))
	for _, match := range matches {
		current, ok := latestByActivity[match.ImportedActivityID]
		if !ok || newerWorkoutMatch(match, current) {
			latestByActivity[match.ImportedActivityID] = match
		}
	}
	latest := make([]domain.WorkoutMatch, 0, len(latestByActivity))
	for _, match := range latestByActivity {
		latest = append(latest, match)
	}
	return latest
}

func newerWorkoutMatch(candidate domain.WorkoutMatch, current domain.WorkoutMatch) bool {
	if !candidate.MatchedAt.Equal(current.MatchedAt) {
		return candidate.MatchedAt.After(current.MatchedAt)
	}
	if candidate.MatchedBy != current.MatchedBy {
		return candidate.MatchedBy == domain.MatchSourceManual
	}
	if workoutMatchStatusRank(candidate.Status) != workoutMatchStatusRank(current.Status) {
		return workoutMatchStatusRank(candidate.Status) > workoutMatchStatusRank(current.Status)
	}
	return candidate.ID > current.ID
}

func workoutMatchStatusRank(status domain.WorkoutMatchStatus) int {
	switch status {
	case domain.WorkoutMatchStatusMatched:
		return 3
	case domain.WorkoutMatchStatusUncertain:
		return 2
	case domain.WorkoutMatchStatusRejected:
		return 1
	default:
		return 0
	}
}
