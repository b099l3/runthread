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
	trainingGoalOrder  []string
	planWeeks          map[string]domain.PlanWeek
	plannedWorkouts    map[string]domain.PlannedWorkout
	importedActivities map[string]domain.ImportedActivity
	workoutMatches     map[string]domain.WorkoutMatch
	workoutResults     map[string]domain.WorkoutResult
	adaptationEvents   map[string]domain.AdaptationEvent

	providerConnections      map[string]ProviderConnection
	providerActivities       map[string]ProviderActivity
	providerActivityPayloads map[string]ProviderActivityPayload
	providerImportEvents     map[string]ProviderImportEvent
}

var _ Store = (*InMemoryStore)(nil)
var _ ProviderStore = (*InMemoryStore)(nil)
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

		providerConnections:      make(map[string]ProviderConnection),
		providerActivities:       make(map[string]ProviderActivity),
		providerActivityPayloads: make(map[string]ProviderActivityPayload),
		providerImportEvents:     make(map[string]ProviderImportEvent),
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
	if _, exists := s.trainingGoals[goal.ID]; !exists {
		s.trainingGoalOrder = append(s.trainingGoalOrder, goal.ID)
	}
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

func (s *InMemoryStore) GetCurrentTrainingGoal(ctx context.Context, athleteID string) (domain.TrainingGoal, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrainingGoal{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.trainingGoalOrder) - 1; i >= 0; i-- {
		goal := s.trainingGoals[s.trainingGoalOrder[i]]
		if goal.AthleteID == athleteID {
			return goal, nil
		}
	}
	return domain.TrainingGoal{}, ErrNotFound
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

func (s *InMemoryStore) SaveProviderConnection(ctx context.Context, connection ProviderConnection) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := connection.Validate(); err != nil {
		return fmt.Errorf("invalid provider connection: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerConnections[connection.ID] = connection
	return nil
}

func (s *InMemoryStore) GetProviderConnection(ctx context.Context, id string) (ProviderConnection, error) {
	if err := ctx.Err(); err != nil {
		return ProviderConnection{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	connection, ok := s.providerConnections[id]
	if !ok {
		return ProviderConnection{}, ErrNotFound
	}
	return connection, nil
}

func (s *InMemoryStore) ListProviderConnectionsByAthlete(ctx context.Context, athleteID string) ([]ProviderConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var connections []ProviderConnection
	for _, connection := range s.providerConnections {
		if connection.AthleteID == athleteID {
			connections = append(connections, connection)
		}
	}
	return connections, nil
}

func (s *InMemoryStore) ListProviderConnectionsByStatus(ctx context.Context, status ProviderConnectionStatus) ([]ProviderConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var connections []ProviderConnection
	for _, connection := range s.providerConnections {
		if connection.Status == status {
			connections = append(connections, connection)
		}
	}
	return connections, nil
}

func (s *InMemoryStore) SaveProviderActivity(ctx context.Context, activity ProviderActivity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := activity.Validate(); err != nil {
		return fmt.Errorf("invalid provider activity: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerActivities[activity.ID] = activity
	return nil
}

func (s *InMemoryStore) GetProviderActivity(ctx context.Context, id string) (ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return ProviderActivity{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	activity, ok := s.providerActivities[id]
	if !ok {
		return ProviderActivity{}, ErrNotFound
	}
	return activity, nil
}

func (s *InMemoryStore) GetProviderActivityByProviderID(ctx context.Context, providerConnectionID string, providerActivityID string) (ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return ProviderActivity{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, activity := range s.providerActivities {
		if activity.ProviderConnectionID == providerConnectionID && activity.ProviderActivityID == providerActivityID {
			return activity, nil
		}
	}
	return ProviderActivity{}, ErrNotFound
}

func (s *InMemoryStore) ListProviderActivitiesByAthlete(ctx context.Context, athleteID string) ([]ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var activities []ProviderActivity
	for _, activity := range s.providerActivities {
		if activity.AthleteID == athleteID {
			activities = append(activities, activity)
		}
	}
	return activities, nil
}

func (s *InMemoryStore) ListProviderActivitiesByStatus(ctx context.Context, status ProviderActivityStatus) ([]ProviderActivity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var activities []ProviderActivity
	for _, activity := range s.providerActivities {
		if activity.Status == status {
			activities = append(activities, activity)
		}
	}
	return activities, nil
}

func (s *InMemoryStore) SaveProviderActivityPayload(ctx context.Context, payload ProviderActivityPayload) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := payload.Validate(); err != nil {
		return fmt.Errorf("invalid provider activity payload: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerActivityPayloads[payload.ID] = cloneProviderActivityPayload(payload)
	return nil
}

func (s *InMemoryStore) GetProviderActivityPayload(ctx context.Context, id string) (ProviderActivityPayload, error) {
	if err := ctx.Err(); err != nil {
		return ProviderActivityPayload{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	payload, ok := s.providerActivityPayloads[id]
	if !ok {
		return ProviderActivityPayload{}, ErrNotFound
	}
	return cloneProviderActivityPayload(payload), nil
}

func (s *InMemoryStore) ListProviderActivityPayloads(ctx context.Context, providerActivityID string) ([]ProviderActivityPayload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var payloads []ProviderActivityPayload
	for _, payload := range s.providerActivityPayloads {
		if payload.ProviderActivityID == providerActivityID {
			payloads = append(payloads, cloneProviderActivityPayload(payload))
		}
	}
	return payloads, nil
}

func (s *InMemoryStore) SaveProviderImportEvent(ctx context.Context, event ProviderImportEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid provider import event: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerImportEvents[event.ID] = event
	return nil
}

func (s *InMemoryStore) GetProviderImportEvent(ctx context.Context, id string) (ProviderImportEvent, error) {
	if err := ctx.Err(); err != nil {
		return ProviderImportEvent{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	event, ok := s.providerImportEvents[id]
	if !ok {
		return ProviderImportEvent{}, ErrNotFound
	}
	return event, nil
}

func (s *InMemoryStore) ListProviderImportEventsByConnection(ctx context.Context, providerConnectionID string) ([]ProviderImportEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var events []ProviderImportEvent
	for _, event := range s.providerImportEvents {
		if event.ProviderConnectionID == providerConnectionID {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *InMemoryStore) ListProviderImportEventsByStatus(ctx context.Context, status ProviderImportEventStatus) ([]ProviderImportEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var events []ProviderImportEvent
	for _, event := range s.providerImportEvents {
		if event.Status == status {
			events = append(events, event)
		}
	}
	return events, nil
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

func cloneProviderActivityPayload(payload ProviderActivityPayload) ProviderActivityPayload {
	payload.Payload = append([]byte(nil), payload.Payload...)
	return payload
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
