package adaptation

import (
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

type Engine struct {
	Now func() time.Time
}

type WorkoutResultInput struct {
	AthleteID string
	PlanWeek  domain.PlanWeek
	Result    domain.WorkoutResult
}

func NewEngine() Engine {
	return Engine{Now: time.Now}
}

func AdaptWorkoutResult(input WorkoutResultInput) (*domain.AdaptationEvent, error) {
	return NewEngine().AdaptWorkoutResult(input)
}

func (e Engine) AdaptWorkoutResult(input WorkoutResultInput) (*domain.AdaptationEvent, error) {
	if input.AthleteID == "" {
		return nil, fmt.Errorf("athlete id is required")
	}
	if err := input.PlanWeek.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plan week: %w", err)
	}
	if err := input.Result.Validate(); err != nil {
		return nil, fmt.Errorf("invalid workout result: %w", err)
	}

	source, found := findWorkout(input.PlanWeek, input.Result.PlannedWorkoutID)
	if !found {
		return nil, fmt.Errorf("planned workout %q not found in plan week", input.Result.PlannedWorkoutID)
	}

	event := domain.AdaptationEvent{
		ID:        eventID(input.Result.ID),
		PlanID:    input.PlanWeek.PlanID,
		AthleteID: input.AthleteID,
		CreatedAt: e.now(),
	}

	switch input.Result.Outcome {
	case domain.WorkoutOutcomeMissed:
		event.Type = domain.AdaptationTypeMissedWorkout
		event.Reason = "A planned workout was missed."
		event.Summary = "Kept the missed workout from being stacked onto the next run."
		event.Changes = []domain.PlanChange{nextWorkoutAdjustment(input.PlanWeek, source, "Keep the next run easy instead of adding the missed workload.")}
	case domain.WorkoutOutcomePartiallyCompleted:
		event.Type = domain.AdaptationTypePartialCompletion
		event.Reason = "A planned workout was only partially completed."
		event.Summary = "Protected the rest of the week by keeping the next workout conservative."
		event.Changes = []domain.PlanChange{nextWorkoutAdjustment(input.PlanWeek, source, "Keep the next workout conservative because the planned session was only partially completed.")}
	case domain.WorkoutOutcomeOverperformed:
		event.Type = domain.AdaptationTypeOverperformance
		event.Reason = "A workout was completed above the planned load."
		event.Summary = "Added a recovery note rather than increasing future training load."
		event.Changes = []domain.PlanChange{nextWorkoutAdjustment(input.PlanWeek, source, "Avoid adding extra intensity after an above-plan effort.")}
	case domain.WorkoutOutcomeUnderperformed:
		event.Type = domain.AdaptationTypeUnderperformance
		event.Reason = "A workout came in below the planned load."
		event.Summary = "Reduced pressure on the next harder session."
		event.Changes = []domain.PlanChange{nextWorkoutAdjustment(input.PlanWeek, source, "Reduce pressure on the next harder session after an underperformed workout.")}
	default:
		return nil, nil
	}

	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("invalid adaptation event: %w", err)
	}
	return &event, nil
}

func (e Engine) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now()
}

func findWorkout(week domain.PlanWeek, workoutID string) (domain.PlannedWorkout, bool) {
	for _, workout := range week.Workouts {
		if workout.ID == workoutID {
			return workout, true
		}
	}
	return domain.PlannedWorkout{}, false
}

func nextWorkoutAdjustment(week domain.PlanWeek, source domain.PlannedWorkout, description string) domain.PlanChange {
	target, found := nextAdjustableWorkout(week, source.ScheduledFor)
	if !found {
		return domain.PlanChange{
			Type:        domain.PlanChangeTypePlanNoteAdded,
			Description: description,
		}
	}
	return domain.PlanChange{
		PlannedWorkoutID: target.ID,
		Type:             domain.PlanChangeTypeWorkoutAdjusted,
		Description:      description,
	}
}

func nextAdjustableWorkout(week domain.PlanWeek, after time.Time) (domain.PlannedWorkout, bool) {
	var fallback domain.PlannedWorkout
	hasFallback := false

	for _, workout := range week.Workouts {
		if !workout.ScheduledFor.After(after) || !isRunWorkout(workout.Type) {
			continue
		}
		if isHardWorkout(workout.Type) {
			return workout, true
		}
		if !hasFallback {
			fallback = workout
			hasFallback = true
		}
	}
	return fallback, hasFallback
}

func isRunWorkout(workoutType domain.WorkoutType) bool {
	switch workoutType {
	case domain.WorkoutTypeEasy, domain.WorkoutTypeLongRun, domain.WorkoutTypeWorkout, domain.WorkoutTypeRecovery, domain.WorkoutTypeRace:
		return true
	default:
		return false
	}
}

func isHardWorkout(workoutType domain.WorkoutType) bool {
	return workoutType == domain.WorkoutTypeWorkout || workoutType == domain.WorkoutTypeLongRun || workoutType == domain.WorkoutTypeRace
}

func eventID(resultID string) string {
	return fmt.Sprintf("adaptation-%s", resultID)
}
