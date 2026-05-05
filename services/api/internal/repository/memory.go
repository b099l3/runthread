package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

type InMemoryStore struct {
	mu sync.RWMutex

	athleteProfiles    map[string]domain.AthleteProfile
	trainingGoals      map[string]domain.TrainingGoal
	planWeeks          map[string]domain.PlanWeek
	plannedWorkouts    map[string]domain.PlannedWorkout
	importedActivities map[string]domain.ImportedActivity
	workoutMatches     map[string]domain.WorkoutMatch
	workoutResults     map[string]domain.WorkoutResult
	adaptationEvents   map[string]domain.AdaptationEvent
}

var _ Store = (*InMemoryStore)(nil)
var _ CurrentPlanWeekReader = (*InMemoryStore)(nil)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		athleteProfiles:    make(map[string]domain.AthleteProfile),
		trainingGoals:      make(map[string]domain.TrainingGoal),
		planWeeks:          make(map[string]domain.PlanWeek),
		plannedWorkouts:    make(map[string]domain.PlannedWorkout),
		importedActivities: make(map[string]domain.ImportedActivity),
		workoutMatches:     make(map[string]domain.WorkoutMatch),
		workoutResults:     make(map[string]domain.WorkoutResult),
		adaptationEvents:   make(map[string]domain.AdaptationEvent),
	}
}

func (s *InMemoryStore) SaveAthleteProfile(ctx context.Context, profile domain.AthleteProfile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid athlete profile: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.athleteProfiles[profile.ID] = cloneAthleteProfile(profile)
	return nil
}

func (s *InMemoryStore) GetAthleteProfile(ctx context.Context, id string) (domain.AthleteProfile, error) {
	if err := ctx.Err(); err != nil {
		return domain.AthleteProfile{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.athleteProfiles[id]
	if !ok {
		return domain.AthleteProfile{}, ErrNotFound
	}
	return cloneAthleteProfile(profile), nil
}

func (s *InMemoryStore) SaveTrainingGoal(ctx context.Context, goal domain.TrainingGoal) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := goal.Validate(); err != nil {
		return fmt.Errorf("invalid training goal: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.trainingGoals[goal.ID] = goal
	return nil
}

func (s *InMemoryStore) GetTrainingGoal(ctx context.Context, id string) (domain.TrainingGoal, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrainingGoal{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	goal, ok := s.trainingGoals[id]
	if !ok {
		return domain.TrainingGoal{}, ErrNotFound
	}
	return goal, nil
}

func (s *InMemoryStore) SavePlanWeek(ctx context.Context, week domain.PlanWeek) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := week.Validate(); err != nil {
		return fmt.Errorf("invalid plan week: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.planWeeks[week.ID] = clonePlanWeek(week)
	return nil
}

func (s *InMemoryStore) GetPlanWeek(ctx context.Context, id string) (domain.PlanWeek, error) {
	if err := ctx.Err(); err != nil {
		return domain.PlanWeek{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	week, ok := s.planWeeks[id]
	if !ok {
		return domain.PlanWeek{}, ErrNotFound
	}
	return clonePlanWeek(week), nil
}

func (s *InMemoryStore) SavePlannedWorkout(ctx context.Context, workout domain.PlannedWorkout) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := workout.Validate(); err != nil {
		return fmt.Errorf("invalid planned workout: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.plannedWorkouts[workout.ID] = workout
	return nil
}

func (s *InMemoryStore) GetPlannedWorkout(ctx context.Context, id string) (domain.PlannedWorkout, error) {
	if err := ctx.Err(); err != nil {
		return domain.PlannedWorkout{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	workout, ok := s.plannedWorkouts[id]
	if !ok {
		return domain.PlannedWorkout{}, ErrNotFound
	}
	return workout, nil
}

func (s *InMemoryStore) SaveImportedActivity(ctx context.Context, activity domain.ImportedActivity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := activity.Validate(); err != nil {
		return fmt.Errorf("invalid imported activity: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.importedActivities[activity.ID] = activity
	return nil
}

func (s *InMemoryStore) GetImportedActivity(ctx context.Context, id string) (domain.ImportedActivity, error) {
	if err := ctx.Err(); err != nil {
		return domain.ImportedActivity{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	activity, ok := s.importedActivities[id]
	if !ok {
		return domain.ImportedActivity{}, ErrNotFound
	}
	return activity, nil
}

func (s *InMemoryStore) SaveWorkoutMatch(ctx context.Context, match domain.WorkoutMatch) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := match.Validate(); err != nil {
		return fmt.Errorf("invalid workout match: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.workoutMatches[match.ID] = match
	return nil
}

func (s *InMemoryStore) GetWorkoutMatch(ctx context.Context, id string) (domain.WorkoutMatch, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkoutMatch{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	match, ok := s.workoutMatches[id]
	if !ok {
		return domain.WorkoutMatch{}, ErrNotFound
	}
	return match, nil
}

func (s *InMemoryStore) SaveWorkoutResult(ctx context.Context, result domain.WorkoutResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := result.Validate(); err != nil {
		return fmt.Errorf("invalid workout result: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.workoutResults[result.ID] = result
	return nil
}

func (s *InMemoryStore) GetWorkoutResult(ctx context.Context, id string) (domain.WorkoutResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkoutResult{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.workoutResults[id]
	if !ok {
		return domain.WorkoutResult{}, ErrNotFound
	}
	return result, nil
}

func (s *InMemoryStore) SaveAdaptationEvent(ctx context.Context, event domain.AdaptationEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid adaptation event: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.adaptationEvents[event.ID] = cloneAdaptationEvent(event)
	return nil
}

func (s *InMemoryStore) GetAdaptationEvent(ctx context.Context, id string) (domain.AdaptationEvent, error) {
	if err := ctx.Err(); err != nil {
		return domain.AdaptationEvent{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	event, ok := s.adaptationEvents[id]
	if !ok {
		return domain.AdaptationEvent{}, ErrNotFound
	}
	return cloneAdaptationEvent(event), nil
}

func (s *InMemoryStore) GetCurrentPlanWeekSnapshot(ctx context.Context, query CurrentPlanWeekQuery) (PlanWeekSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return PlanWeekSnapshot{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	week, ok := s.findPlanWeekLocked(query)
	if !ok {
		return PlanWeekSnapshot{}, ErrNotFound
	}

	workoutIDs := make(map[string]struct{}, len(week.Workouts))
	for _, workout := range week.Workouts {
		workoutIDs[workout.ID] = struct{}{}
	}

	snapshot := PlanWeekSnapshot{
		PlanWeek: clonePlanWeek(week),
	}
	startsOn := startOfWeek(query.TargetWeekDate)
	if startsOn.IsZero() {
		startsOn = midnight(week.StartsOn)
	}
	endsBefore := startsOn.AddDate(0, 0, 7)

	for _, activity := range s.importedActivities {
		if activity.AthleteID == week.AthleteID && !activity.StartedAt.Before(startsOn) && activity.StartedAt.Before(endsBefore) {
			snapshot.ImportedActivities = append(snapshot.ImportedActivities, activity)
		}
	}
	for _, match := range s.workoutMatches {
		if _, ok := workoutIDs[match.PlannedWorkoutID]; ok {
			snapshot.WorkoutMatches = append(snapshot.WorkoutMatches, match)
		}
	}
	for _, result := range s.workoutResults {
		if _, ok := workoutIDs[result.PlannedWorkoutID]; ok {
			snapshot.WorkoutResults = append(snapshot.WorkoutResults, result)
		}
	}
	for _, event := range s.adaptationEvents {
		if event.AthleteID == week.AthleteID && event.PlanID == week.PlanID {
			snapshot.AdaptationEvents = append(snapshot.AdaptationEvents, cloneAdaptationEvent(event))
		}
	}

	return snapshot, nil
}

func (s *InMemoryStore) findPlanWeekLocked(query CurrentPlanWeekQuery) (domain.PlanWeek, bool) {
	if query.PlanWeekID != "" {
		week, ok := s.planWeeks[query.PlanWeekID]
		return week, ok
	}

	targetStartsOn := startOfWeek(query.TargetWeekDate)
	for _, week := range s.planWeeks {
		if week.AthleteID != query.AthleteID {
			continue
		}
		if query.GoalID != "" && week.GoalID != query.GoalID {
			continue
		}
		if !sameCalendarDate(week.StartsOn, targetStartsOn) {
			continue
		}
		return week, true
	}
	return domain.PlanWeek{}, false
}

func cloneAthleteProfile(profile domain.AthleteProfile) domain.AthleteProfile {
	profile.PreferredRunDays = append([]time.Weekday(nil), profile.PreferredRunDays...)
	profile.Constraints = append([]string(nil), profile.Constraints...)
	return profile
}

func clonePlanWeek(week domain.PlanWeek) domain.PlanWeek {
	week.Workouts = append([]domain.PlannedWorkout(nil), week.Workouts...)
	return week
}

func cloneAdaptationEvent(event domain.AdaptationEvent) domain.AdaptationEvent {
	event.Changes = append([]domain.PlanChange(nil), event.Changes...)
	return event
}

func startOfWeek(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	normalized := midnight(date)
	offset := (int(normalized.Weekday()) + 6) % 7
	return normalized.AddDate(0, 0, -offset)
}

func midnight(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func sameCalendarDate(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return a.IsZero() && b.IsZero()
	}
	aYear, aMonth, aDay := a.Date()
	bYear, bMonth, bDay := b.Date()
	return aYear == bYear && aMonth == bMonth && aDay == bDay
}
