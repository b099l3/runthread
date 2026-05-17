package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type TrainingGoalRepository struct {
	queries postgresdb.Querier
}

var _ repository.TrainingGoalRepository = (*TrainingGoalRepository)(nil)
var _ repository.CurrentTrainingGoalRepository = (*TrainingGoalRepository)(nil)

func NewTrainingGoalRepository(db *sql.DB) *TrainingGoalRepository {
	return &TrainingGoalRepository{
		queries: postgresdb.New(db),
	}
}

func (r *TrainingGoalRepository) SaveTrainingGoal(ctx context.Context, goal domain.TrainingGoal) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := goal.Validate(); err != nil {
		return fmt.Errorf("invalid training goal: %w", err)
	}

	updateParams, err := trainingGoalToUpdateParams(goal)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateTrainingGoal(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update training goal: %w", err)
	}

	createParams, err := trainingGoalToCreateParams(goal)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateTrainingGoal(ctx, createParams); err != nil {
		return fmt.Errorf("create training goal: %w", err)
	}
	return nil
}

func (r *TrainingGoalRepository) GetTrainingGoal(ctx context.Context, id string) (domain.TrainingGoal, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrainingGoal{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.TrainingGoal{}, fmt.Errorf("parse training goal id: %w", err)
	}

	row, err := r.queries.GetTrainingGoal(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.TrainingGoal{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.TrainingGoal{}, fmt.Errorf("get training goal: %w", err)
	}
	return trainingGoalFromDB(row), nil
}

func (r *TrainingGoalRepository) GetCurrentTrainingGoal(ctx context.Context, athleteID string) (domain.TrainingGoal, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrainingGoal{}, err
	}

	parsedAthleteID, err := uuid.Parse(athleteID)
	if err != nil {
		return domain.TrainingGoal{}, fmt.Errorf("parse training goal athlete id: %w", err)
	}

	rows, err := r.queries.ListTrainingGoalsByAthlete(ctx, parsedAthleteID)
	if err != nil {
		return domain.TrainingGoal{}, fmt.Errorf("list training goals by athlete: %w", err)
	}
	if len(rows) == 0 {
		return domain.TrainingGoal{}, repository.ErrNotFound
	}
	return trainingGoalFromDB(rows[0]), nil
}

type PlanWeekRepository struct {
	queries postgresdb.Querier
}

var _ repository.PlanWeekRepository = (*PlanWeekRepository)(nil)

func NewPlanWeekRepository(db *sql.DB) *PlanWeekRepository {
	return &PlanWeekRepository{
		queries: postgresdb.New(db),
	}
}

func (r *PlanWeekRepository) SavePlanWeek(ctx context.Context, week domain.PlanWeek) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := week.Validate(); err != nil {
		return fmt.Errorf("invalid plan week: %w", err)
	}

	updateParams, err := planWeekToUpdateParams(week)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdatePlanWeek(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update plan week: %w", err)
	}

	createParams, err := planWeekToCreateParams(week)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreatePlanWeek(ctx, createParams); err != nil {
		return fmt.Errorf("create plan week: %w", err)
	}
	return nil
}

func (r *PlanWeekRepository) GetPlanWeek(ctx context.Context, id string) (domain.PlanWeek, error) {
	if err := ctx.Err(); err != nil {
		return domain.PlanWeek{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.PlanWeek{}, fmt.Errorf("parse plan week id: %w", err)
	}

	row, err := r.queries.GetPlanWeek(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlanWeek{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.PlanWeek{}, fmt.Errorf("get plan week: %w", err)
	}
	return planWeekFromDB(row), nil
}

type PlannedWorkoutRepository struct {
	queries postgresdb.Querier
}

var _ repository.PlannedWorkoutRepository = (*PlannedWorkoutRepository)(nil)

func NewPlannedWorkoutRepository(db *sql.DB) *PlannedWorkoutRepository {
	return &PlannedWorkoutRepository{
		queries: postgresdb.New(db),
	}
}

func (r *PlannedWorkoutRepository) SavePlannedWorkout(ctx context.Context, workout domain.PlannedWorkout) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := workout.Validate(); err != nil {
		return fmt.Errorf("invalid planned workout: %w", err)
	}

	updateParams, err := plannedWorkoutToUpdateParams(workout)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdatePlannedWorkout(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update planned workout: %w", err)
	}

	createParams, err := plannedWorkoutToCreateParams(workout)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreatePlannedWorkout(ctx, createParams); err != nil {
		return fmt.Errorf("create planned workout: %w", err)
	}
	return nil
}

func (r *PlannedWorkoutRepository) GetPlannedWorkout(ctx context.Context, id string) (domain.PlannedWorkout, error) {
	if err := ctx.Err(); err != nil {
		return domain.PlannedWorkout{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.PlannedWorkout{}, fmt.Errorf("parse planned workout id: %w", err)
	}

	row, err := r.queries.GetPlannedWorkout(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlannedWorkout{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.PlannedWorkout{}, fmt.Errorf("get planned workout: %w", err)
	}
	return plannedWorkoutFromDB(row), nil
}

func trainingGoalToCreateParams(goal domain.TrainingGoal) (postgresdb.CreateTrainingGoalParams, error) {
	id, athleteID, err := parseRecordAndParentIDs(goal.ID, goal.AthleteID, "training goal")
	if err != nil {
		return postgresdb.CreateTrainingGoalParams{}, err
	}

	return postgresdb.CreateTrainingGoalParams{
		ID:                    id,
		AthleteID:             athleteID,
		Type:                  string(goal.Type),
		TargetDate:            nullableTime(goal.TargetDate),
		TargetDistanceMeters:  nullableFloat64(goal.TargetDistance.Meters),
		TargetDurationSeconds: nullableDurationSeconds(goal.TargetDuration),
		Notes:                 nullableString(goal.Notes),
	}, nil
}

func trainingGoalToUpdateParams(goal domain.TrainingGoal) (postgresdb.UpdateTrainingGoalParams, error) {
	id, err := uuid.Parse(goal.ID)
	if err != nil {
		return postgresdb.UpdateTrainingGoalParams{}, fmt.Errorf("parse training goal id: %w", err)
	}

	return postgresdb.UpdateTrainingGoalParams{
		ID:                    id,
		Type:                  string(goal.Type),
		TargetDate:            nullableTime(goal.TargetDate),
		TargetDistanceMeters:  nullableFloat64(goal.TargetDistance.Meters),
		TargetDurationSeconds: nullableDurationSeconds(goal.TargetDuration),
		Notes:                 nullableString(goal.Notes),
	}, nil
}

func trainingGoalFromDB(row postgresdb.TrainingGoal) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             row.ID.String(),
		AthleteID:      row.AthleteID.String(),
		Type:           domain.GoalType(row.Type),
		TargetDate:     timeFromNull(row.TargetDate),
		TargetDistance: domain.Distance{Meters: float64FromNull(row.TargetDistanceMeters)},
		TargetDuration: durationFromNullSeconds(row.TargetDurationSeconds),
		Notes:          stringFromNull(row.Notes),
	}
}

func planWeekToCreateParams(week domain.PlanWeek) (postgresdb.CreatePlanWeekParams, error) {
	id, planID, err := parseRecordAndParentIDs(week.ID, week.PlanID, "plan week")
	if err != nil {
		return postgresdb.CreatePlanWeekParams{}, err
	}
	athleteID, err := uuid.Parse(week.AthleteID)
	if err != nil {
		return postgresdb.CreatePlanWeekParams{}, fmt.Errorf("parse plan week athlete id: %w", err)
	}
	goalID, err := nullableUUID(week.GoalID)
	if err != nil {
		return postgresdb.CreatePlanWeekParams{}, fmt.Errorf("parse plan week goal id: %w", err)
	}

	return postgresdb.CreatePlanWeekParams{
		ID:        id,
		AthleteID: athleteID,
		GoalID:    goalID,
		PlanID:    planID,
		WeekIndex: int32(week.WeekIndex),
		StartsOn:  week.StartsOn,
		Focus:     nullableString(string(week.Focus)),
	}, nil
}

func planWeekToUpdateParams(week domain.PlanWeek) (postgresdb.UpdatePlanWeekParams, error) {
	id, err := uuid.Parse(week.ID)
	if err != nil {
		return postgresdb.UpdatePlanWeekParams{}, fmt.Errorf("parse plan week id: %w", err)
	}
	goalID, err := nullableUUID(week.GoalID)
	if err != nil {
		return postgresdb.UpdatePlanWeekParams{}, fmt.Errorf("parse plan week goal id: %w", err)
	}

	return postgresdb.UpdatePlanWeekParams{
		ID:        id,
		GoalID:    goalID,
		WeekIndex: int32(week.WeekIndex),
		StartsOn:  week.StartsOn,
		Focus:     nullableString(string(week.Focus)),
	}, nil
}

func planWeekFromDB(row postgresdb.PlanWeek) domain.PlanWeek {
	return domain.PlanWeek{
		ID:        row.ID.String(),
		AthleteID: row.AthleteID.String(),
		GoalID:    uuidStringFromNull(row.GoalID),
		PlanID:    row.PlanID.String(),
		WeekIndex: int(row.WeekIndex),
		StartsOn:  row.StartsOn,
		Focus:     domain.WeekFocus(stringFromNull(row.Focus)),
	}
}

func plannedWorkoutToCreateParams(workout domain.PlannedWorkout) (postgresdb.CreatePlannedWorkoutParams, error) {
	id, planWeekID, err := parseRecordAndParentIDs(workout.ID, workout.PlanWeekID, "planned workout")
	if err != nil {
		return postgresdb.CreatePlannedWorkoutParams{}, err
	}
	planID, err := uuid.Parse(workout.PlanID)
	if err != nil {
		return postgresdb.CreatePlannedWorkoutParams{}, fmt.Errorf("parse planned workout plan id: %w", err)
	}

	return postgresdb.CreatePlannedWorkoutParams{
		ID:                    id,
		PlanWeekID:            planWeekID,
		PlanID:                planID,
		ScheduledFor:          workout.ScheduledFor,
		Type:                  string(workout.Type),
		Status:                string(workout.Status),
		TargetDistanceMeters:  nullableFloat64(workout.TargetDistance.Meters),
		TargetDurationSeconds: nullableDurationSeconds(workout.TargetDuration),
		IntensityKind:         nullableString(string(workout.Intensity.Kind)),
		IntensityDescription:  nullableString(workout.Intensity.Description),
		Notes:                 nullableString(workout.Notes),
	}, nil
}

func plannedWorkoutToUpdateParams(workout domain.PlannedWorkout) (postgresdb.UpdatePlannedWorkoutParams, error) {
	id, err := uuid.Parse(workout.ID)
	if err != nil {
		return postgresdb.UpdatePlannedWorkoutParams{}, fmt.Errorf("parse planned workout id: %w", err)
	}

	return postgresdb.UpdatePlannedWorkoutParams{
		ID:                    id,
		ScheduledFor:          workout.ScheduledFor,
		Type:                  string(workout.Type),
		Status:                string(workout.Status),
		TargetDistanceMeters:  nullableFloat64(workout.TargetDistance.Meters),
		TargetDurationSeconds: nullableDurationSeconds(workout.TargetDuration),
		IntensityKind:         nullableString(string(workout.Intensity.Kind)),
		IntensityDescription:  nullableString(workout.Intensity.Description),
		Notes:                 nullableString(workout.Notes),
	}, nil
}

func plannedWorkoutFromDB(row postgresdb.PlannedWorkout) domain.PlannedWorkout {
	return domain.PlannedWorkout{
		ID:             row.ID.String(),
		PlanID:         row.PlanID.String(),
		PlanWeekID:     row.PlanWeekID.String(),
		ScheduledFor:   row.ScheduledFor,
		Type:           domain.WorkoutType(row.Type),
		Status:         domain.PlannedWorkoutStatus(row.Status),
		TargetDistance: domain.Distance{Meters: float64FromNull(row.TargetDistanceMeters)},
		TargetDuration: durationFromNullSeconds(row.TargetDurationSeconds),
		Intensity: domain.IntensityTarget{
			Kind:        domain.IntensityKind(stringFromNull(row.IntensityKind)),
			Description: stringFromNull(row.IntensityDescription),
		},
		Notes: stringFromNull(row.Notes),
	}
}

func parseRecordAndParentIDs(recordID string, parentID string, label string) (uuid.UUID, uuid.UUID, error) {
	parsedRecordID, err := uuid.Parse(recordID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("parse %s id: %w", label, err)
	}
	parsedParentID, err := uuid.Parse(parentID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("parse %s parent id: %w", label, err)
	}
	return parsedRecordID, parsedParentID, nil
}

func nullableUUID(value string) (uuid.NullUUID, error) {
	if value == "" {
		return uuid.NullUUID{}, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: parsed, Valid: true}, nil
}

func nullableTime(value time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  value,
		Valid: !value.IsZero(),
	}
}

func timeFromNull(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}

func nullableFloat64(value float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: value,
		Valid:   value != 0,
	}
}

func float64FromNull(value sql.NullFloat64) float64 {
	if !value.Valid {
		return 0
	}
	return value.Float64
}

func nullableDurationSeconds(value time.Duration) sql.NullInt32 {
	if value == 0 {
		return sql.NullInt32{}
	}
	seconds := value.Seconds()
	if seconds > math.MaxInt32 || seconds < math.MinInt32 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(seconds), Valid: true}
}

func durationFromNullSeconds(value sql.NullInt32) time.Duration {
	if !value.Valid {
		return 0
	}
	return time.Duration(value.Int32) * time.Second
}
