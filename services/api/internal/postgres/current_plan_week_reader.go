package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

func (s *Store) GetCurrentPlanWeekSnapshot(ctx context.Context, query repository.CurrentPlanWeekQuery) (repository.PlanWeekSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return repository.PlanWeekSnapshot{}, err
	}
	if s.queries == nil {
		return repository.PlanWeekSnapshot{}, fmt.Errorf("postgres queries are required")
	}

	week, err := s.currentPlanWeek(ctx, query)
	if err != nil {
		return repository.PlanWeekSnapshot{}, err
	}
	workouts, err := s.currentPlannedWorkouts(ctx, week.ID)
	if err != nil {
		return repository.PlanWeekSnapshot{}, err
	}
	week.Workouts = workouts

	activities, err := s.recentImportedActivities(ctx, week.AthleteID)
	if err != nil {
		return repository.PlanWeekSnapshot{}, err
	}
	matches, err := s.currentWorkoutMatches(ctx, workouts)
	if err != nil {
		return repository.PlanWeekSnapshot{}, err
	}
	results, err := s.currentWorkoutResults(ctx, workouts)
	if err != nil {
		return repository.PlanWeekSnapshot{}, err
	}
	events, err := s.currentAdaptationEvents(ctx, week.PlanID)
	if err != nil {
		return repository.PlanWeekSnapshot{}, err
	}

	return repository.PlanWeekSnapshot{
		PlanWeek:           week,
		ImportedActivities: activities,
		WorkoutMatches:     matches,
		WorkoutResults:     results,
		AdaptationEvents:   events,
	}, nil
}

func (s *Store) currentPlanWeek(ctx context.Context, query repository.CurrentPlanWeekQuery) (domain.PlanWeek, error) {
	if query.PlanWeekID != "" {
		parsedID, err := uuid.Parse(query.PlanWeekID)
		if err != nil {
			return domain.PlanWeek{}, fmt.Errorf("parse current plan week id: %w", err)
		}
		row, err := s.queries.GetPlanWeek(ctx, parsedID)
		if err != nil {
			return domain.PlanWeek{}, mapSQLError(err, "get current plan week")
		}
		return planWeekFromDB(row), nil
	}

	parsedAthleteID, err := uuid.Parse(query.AthleteID)
	if err != nil {
		return domain.PlanWeek{}, fmt.Errorf("parse current plan week athlete id: %w", err)
	}
	rows, err := s.queries.ListPlanWeeksByAthlete(ctx, parsedAthleteID)
	if err != nil {
		return domain.PlanWeek{}, fmt.Errorf("list current plan weeks by athlete: %w", err)
	}
	targetStartsOn := startOfWeek(query.TargetWeekDate)
	for _, row := range rows {
		week := planWeekFromDB(row)
		if query.GoalID != "" && week.GoalID != query.GoalID {
			continue
		}
		if !sameCalendarDate(week.StartsOn, targetStartsOn) {
			continue
		}
		return week, nil
	}
	return domain.PlanWeek{}, repository.ErrNotFound
}

func (s *Store) currentPlannedWorkouts(ctx context.Context, planWeekID string) ([]domain.PlannedWorkout, error) {
	parsedID, err := uuid.Parse(planWeekID)
	if err != nil {
		return nil, fmt.Errorf("parse current plan week workouts id: %w", err)
	}
	rows, err := s.queries.ListPlannedWorkoutsByWeek(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list current planned workouts: %w", err)
	}
	workouts := make([]domain.PlannedWorkout, 0, len(rows))
	for _, row := range rows {
		workouts = append(workouts, plannedWorkoutFromDB(row))
	}
	return workouts, nil
}

func (s *Store) recentImportedActivities(ctx context.Context, athleteID string) ([]domain.ImportedActivity, error) {
	parsedID, err := uuid.Parse(athleteID)
	if err != nil {
		return nil, fmt.Errorf("parse current imported activities athlete id: %w", err)
	}
	rows, err := s.queries.ListImportedActivitiesByAthlete(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list current imported activities: %w", err)
	}
	activities := make([]domain.ImportedActivity, 0, len(rows))
	for _, row := range rows {
		if len(activities) == 20 {
			break
		}
		activities = append(activities, importedActivityFromDB(row))
	}
	return activities, nil
}

func (s *Store) currentWorkoutMatches(ctx context.Context, workouts []domain.PlannedWorkout) ([]domain.WorkoutMatch, error) {
	var matches []domain.WorkoutMatch
	for _, workout := range workouts {
		parsedID, err := uuid.Parse(workout.ID)
		if err != nil {
			return nil, fmt.Errorf("parse current workout match planned workout id: %w", err)
		}
		rows, err := s.queries.ListWorkoutMatchesByPlannedWorkout(ctx, parsedID)
		if err != nil {
			return nil, fmt.Errorf("list current workout matches: %w", err)
		}
		for _, row := range rows {
			matches = append(matches, workoutMatchFromDB(row))
		}
	}
	return matches, nil
}

func (s *Store) currentWorkoutResults(ctx context.Context, workouts []domain.PlannedWorkout) ([]domain.WorkoutResult, error) {
	var results []domain.WorkoutResult
	for _, workout := range workouts {
		parsedID, err := uuid.Parse(workout.ID)
		if err != nil {
			return nil, fmt.Errorf("parse current workout result planned workout id: %w", err)
		}
		rows, err := s.queries.ListWorkoutResultsByPlannedWorkout(ctx, parsedID)
		if err != nil {
			return nil, fmt.Errorf("list current workout results: %w", err)
		}
		for _, row := range rows {
			results = append(results, workoutResultFromDB(row))
		}
	}
	return results, nil
}

func (s *Store) currentAdaptationEvents(ctx context.Context, planID string) ([]domain.AdaptationEvent, error) {
	parsedID, err := uuid.Parse(planID)
	if err != nil {
		return nil, fmt.Errorf("parse current adaptation events plan id: %w", err)
	}
	rows, err := s.queries.ListAdaptationEventsByPlan(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list current adaptation events: %w", err)
	}
	events := make([]domain.AdaptationEvent, 0, len(rows))
	for _, row := range rows {
		changes, err := s.queries.ListAdaptationEventChanges(ctx, row.ID)
		if err != nil {
			return nil, fmt.Errorf("list current adaptation event changes: %w", err)
		}
		events = append(events, adaptationEventFromDB(row, changes))
	}
	return events, nil
}

func mapSQLError(err error, operation string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func startOfWeek(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	year, month, day := date.Date()
	location := date.Location()
	normalized := time.Date(year, month, day, 0, 0, 0, 0, location)
	offset := (int(normalized.Weekday()) + 6) % 7
	return normalized.AddDate(0, 0, -offset)
}

func midnight(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func sameCalendarDate(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
