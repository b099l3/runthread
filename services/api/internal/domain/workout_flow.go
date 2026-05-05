package domain

import (
	"errors"
	"fmt"
	"time"
)

type WorkoutCompletion struct {
	ResultID           string
	ImportedActivityID string
	CompletedAt        time.Time
	Distance           Distance
	Duration           time.Duration
	Outcome            WorkoutOutcome
	Notes              string
}

type WorkoutMove struct {
	ResultID        string
	NewScheduledFor time.Time
	MovedAt         time.Time
	Notes           string
}

func MarkWorkoutCompleted(workout PlannedWorkout, completion WorkoutCompletion) (PlannedWorkout, WorkoutResult, error) {
	if err := workout.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid planned workout: %w", err)
	}
	if !workout.canTransition() {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("planned workout %q is already %s", workout.ID, workout.Status)
	}

	outcome := completion.Outcome
	if outcome == "" {
		outcome = WorkoutOutcomeCompletedAsPlanned
	}
	if !isCompletionOutcome(outcome) {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid completion outcome %q", outcome)
	}

	result := WorkoutResult{
		ID:                 completion.ResultID,
		PlannedWorkoutID:   workout.ID,
		ImportedActivityID: completion.ImportedActivityID,
		Outcome:            outcome,
		CompletedAt:        completion.CompletedAt,
		Distance:           completion.Distance,
		Duration:           completion.Duration,
		Notes:              completion.Notes,
	}
	if err := result.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid workout result: %w", err)
	}

	updated := workout
	updated.Status = PlannedWorkoutStatusCompleted
	return updated, result, nil
}

func MarkWorkoutMissed(workout PlannedWorkout, resultID string, notes string) (PlannedWorkout, WorkoutResult, error) {
	return markWorkoutWithoutActivity(workout, resultID, PlannedWorkoutStatusMissed, WorkoutOutcomeMissed, notes)
}

func MarkWorkoutSkipped(workout PlannedWorkout, resultID string, notes string) (PlannedWorkout, WorkoutResult, error) {
	return markWorkoutWithoutActivity(workout, resultID, PlannedWorkoutStatusSkipped, WorkoutOutcomeSkipped, notes)
}

func MoveWorkout(workout PlannedWorkout, move WorkoutMove) (PlannedWorkout, WorkoutResult, error) {
	if err := workout.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid planned workout: %w", err)
	}
	if !workout.canTransition() {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("planned workout %q is already %s", workout.ID, workout.Status)
	}
	if move.NewScheduledFor.IsZero() {
		return PlannedWorkout{}, WorkoutResult{}, errors.New("new scheduled date is required")
	}
	if move.MovedAt.IsZero() {
		return PlannedWorkout{}, WorkoutResult{}, errors.New("move time is required")
	}

	updated := workout
	updated.ScheduledFor = move.NewScheduledFor
	updated.Status = PlannedWorkoutStatusMoved

	result := WorkoutResult{
		ID:               move.ResultID,
		PlannedWorkoutID: workout.ID,
		Outcome:          WorkoutOutcomeMoved,
		CompletedAt:      move.MovedAt,
		Notes:            move.Notes,
	}
	if err := result.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid workout result: %w", err)
	}
	if err := updated.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid moved workout: %w", err)
	}
	return updated, result, nil
}

func markWorkoutWithoutActivity(workout PlannedWorkout, resultID string, status PlannedWorkoutStatus, outcome WorkoutOutcome, notes string) (PlannedWorkout, WorkoutResult, error) {
	if err := workout.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid planned workout: %w", err)
	}
	if !workout.canTransition() {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("planned workout %q is already %s", workout.ID, workout.Status)
	}

	result := WorkoutResult{
		ID:               resultID,
		PlannedWorkoutID: workout.ID,
		Outcome:          outcome,
		Notes:            notes,
	}
	if err := result.Validate(); err != nil {
		return PlannedWorkout{}, WorkoutResult{}, fmt.Errorf("invalid workout result: %w", err)
	}

	updated := workout
	updated.Status = status
	return updated, result, nil
}

func (w PlannedWorkout) canTransition() bool {
	switch w.Status {
	case PlannedWorkoutStatusScheduled, PlannedWorkoutStatusMoved:
		return true
	default:
		return false
	}
}

func isCompletionOutcome(outcome WorkoutOutcome) bool {
	switch outcome {
	case WorkoutOutcomeCompletedAsPlanned, WorkoutOutcomePartiallyCompleted, WorkoutOutcomeOverperformed, WorkoutOutcomeUnderperformed:
		return true
	default:
		return false
	}
}
