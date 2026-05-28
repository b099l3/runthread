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

type ImportedActivityRepository struct {
	queries postgresdb.Querier
}

var _ repository.ImportedActivityRepository = (*ImportedActivityRepository)(nil)

func NewImportedActivityRepository(db *sql.DB) *ImportedActivityRepository {
	return &ImportedActivityRepository{
		queries: postgresdb.New(db),
	}
}

func (r *ImportedActivityRepository) SaveImportedActivity(ctx context.Context, activity domain.ImportedActivity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := activity.Validate(); err != nil {
		return fmt.Errorf("invalid imported activity: %w", err)
	}

	updateParams, err := importedActivityToUpdateParams(activity)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateImportedActivity(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update imported activity: %w", err)
	}

	createParams, err := importedActivityToCreateParams(activity)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateImportedActivity(ctx, createParams); err != nil {
		return fmt.Errorf("create imported activity: %w", err)
	}
	return nil
}

func (r *ImportedActivityRepository) GetImportedActivity(ctx context.Context, id string) (domain.ImportedActivity, error) {
	if err := ctx.Err(); err != nil {
		return domain.ImportedActivity{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("parse imported activity id: %w", err)
	}

	row, err := r.queries.GetImportedActivity(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ImportedActivity{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.ImportedActivity{}, fmt.Errorf("get imported activity: %w", err)
	}
	return importedActivityFromDB(row), nil
}

func (r *ImportedActivityRepository) ListImportedActivitiesByAthlete(ctx context.Context, athleteID string) ([]domain.ImportedActivity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(athleteID)
	if err != nil {
		return nil, fmt.Errorf("parse imported activity athlete id: %w", err)
	}

	rows, err := r.queries.ListImportedActivitiesByAthlete(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("list imported activities by athlete: %w", err)
	}
	activities := make([]domain.ImportedActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, importedActivityFromDB(row))
	}
	return activities, nil
}

type WorkoutMatchRepository struct {
	queries postgresdb.Querier
}

var _ repository.WorkoutMatchRepository = (*WorkoutMatchRepository)(nil)

func NewWorkoutMatchRepository(db *sql.DB) *WorkoutMatchRepository {
	return &WorkoutMatchRepository{
		queries: postgresdb.New(db),
	}
}

func (r *WorkoutMatchRepository) SaveWorkoutMatch(ctx context.Context, match domain.WorkoutMatch) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := match.Validate(); err != nil {
		return fmt.Errorf("invalid workout match: %w", err)
	}

	updateParams, err := workoutMatchToUpdateParams(match)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateWorkoutMatch(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update workout match: %w", err)
	}

	createParams, err := workoutMatchToCreateParams(match)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateWorkoutMatch(ctx, createParams); err != nil {
		return fmt.Errorf("create workout match: %w", err)
	}
	return nil
}

func (r *WorkoutMatchRepository) GetWorkoutMatch(ctx context.Context, id string) (domain.WorkoutMatch, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkoutMatch{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("parse workout match id: %w", err)
	}

	row, err := r.queries.GetWorkoutMatch(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.WorkoutMatch{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("get workout match: %w", err)
	}
	return workoutMatchFromDB(row), nil
}

type WorkoutResultRepository struct {
	queries postgresdb.Querier
}

var _ repository.WorkoutResultRepository = (*WorkoutResultRepository)(nil)

func NewWorkoutResultRepository(db *sql.DB) *WorkoutResultRepository {
	return &WorkoutResultRepository{
		queries: postgresdb.New(db),
	}
}

func (r *WorkoutResultRepository) SaveWorkoutResult(ctx context.Context, result domain.WorkoutResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := result.Validate(); err != nil {
		return fmt.Errorf("invalid workout result: %w", err)
	}

	updateParams, err := workoutResultToUpdateParams(result)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateWorkoutResult(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update workout result: %w", err)
	}

	createParams, err := workoutResultToCreateParams(result)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateWorkoutResult(ctx, createParams); err != nil {
		return fmt.Errorf("create workout result: %w", err)
	}
	return nil
}

func (r *WorkoutResultRepository) GetWorkoutResult(ctx context.Context, id string) (domain.WorkoutResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.WorkoutResult{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.WorkoutResult{}, fmt.Errorf("parse workout result id: %w", err)
	}

	row, err := r.queries.GetWorkoutResult(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.WorkoutResult{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.WorkoutResult{}, fmt.Errorf("get workout result: %w", err)
	}
	return workoutResultFromDB(row), nil
}

type AdaptationEventRepository struct {
	db      *sql.DB
	queries *postgresdb.Queries
}

var _ repository.AdaptationEventRepository = (*AdaptationEventRepository)(nil)

func NewAdaptationEventRepository(db *sql.DB) *AdaptationEventRepository {
	return &AdaptationEventRepository{
		db:      db,
		queries: postgresdb.New(db),
	}
}

func (r *AdaptationEventRepository) SaveAdaptationEvent(ctx context.Context, event domain.AdaptationEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid adaptation event: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin adaptation event transaction: %w", err)
	}
	defer tx.Rollback()

	q := r.queries.WithTx(tx)
	updateParams, err := adaptationEventToUpdateParams(event)
	if err != nil {
		return err
	}
	if _, err := q.UpdateAdaptationEvent(ctx, updateParams); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("update adaptation event: %w", err)
		}
		createParams, err := adaptationEventToCreateParams(event)
		if err != nil {
			return err
		}
		if _, err := q.CreateAdaptationEvent(ctx, createParams); err != nil {
			return fmt.Errorf("create adaptation event: %w", err)
		}
	}

	eventID, err := uuid.Parse(event.ID)
	if err != nil {
		return fmt.Errorf("parse adaptation event id: %w", err)
	}
	if err := q.DeleteAdaptationEventChanges(ctx, eventID); err != nil {
		return fmt.Errorf("replace adaptation event changes: %w", err)
	}
	for i, change := range event.Changes {
		params, err := planChangeToCreateParams(eventID, change, i)
		if err != nil {
			return err
		}
		if _, err := q.CreateAdaptationEventChange(ctx, params); err != nil {
			return fmt.Errorf("create adaptation event change %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit adaptation event transaction: %w", err)
	}
	return nil
}

func (r *AdaptationEventRepository) GetAdaptationEvent(ctx context.Context, id string) (domain.AdaptationEvent, error) {
	if err := ctx.Err(); err != nil {
		return domain.AdaptationEvent{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.AdaptationEvent{}, fmt.Errorf("parse adaptation event id: %w", err)
	}

	row, err := r.queries.GetAdaptationEvent(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AdaptationEvent{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.AdaptationEvent{}, fmt.Errorf("get adaptation event: %w", err)
	}
	changes, err := r.queries.ListAdaptationEventChanges(ctx, parsedID)
	if err != nil {
		return domain.AdaptationEvent{}, fmt.Errorf("list adaptation event changes: %w", err)
	}
	return adaptationEventFromDB(row, changes), nil
}

func importedActivityToCreateParams(activity domain.ImportedActivity) (postgresdb.CreateImportedActivityParams, error) {
	id, athleteID, err := parseRecordAndParentIDs(activity.ID, activity.AthleteID, "imported activity")
	if err != nil {
		return postgresdb.CreateImportedActivityParams{}, err
	}
	durationSeconds, err := requiredDurationSeconds(activity.Duration, "imported activity duration")
	if err != nil {
		return postgresdb.CreateImportedActivityParams{}, err
	}

	return postgresdb.CreateImportedActivityParams{
		ID:                             id,
		AthleteID:                      athleteID,
		Type:                           string(activity.Type),
		StartedAt:                      activity.StartedAt,
		DurationSeconds:                durationSeconds,
		DistanceMeters:                 nullableFloat64(activity.Distance.Meters),
		AveragePaceSecondsPerKilometer: nullableInt32(activity.AveragePace.SecondsPerKilometer),
		AverageHeartBpm:                nullableInt32(activity.AverageHeartBPM),
	}, nil
}

func importedActivityToUpdateParams(activity domain.ImportedActivity) (postgresdb.UpdateImportedActivityParams, error) {
	params, err := importedActivityToCreateParams(activity)
	if err != nil {
		return postgresdb.UpdateImportedActivityParams{}, err
	}
	return postgresdb.UpdateImportedActivityParams{
		ID:                             params.ID,
		AthleteID:                      params.AthleteID,
		Type:                           params.Type,
		StartedAt:                      params.StartedAt,
		DurationSeconds:                params.DurationSeconds,
		DistanceMeters:                 params.DistanceMeters,
		AveragePaceSecondsPerKilometer: params.AveragePaceSecondsPerKilometer,
		AverageHeartBpm:                params.AverageHeartBpm,
	}, nil
}

func importedActivityFromDB(row postgresdb.ImportedActivity) domain.ImportedActivity {
	return domain.ImportedActivity{
		ID:              row.ID.String(),
		AthleteID:       row.AthleteID.String(),
		Type:            domain.ActivityType(row.Type),
		StartedAt:       row.StartedAt,
		Duration:        time.Duration(row.DurationSeconds) * time.Second,
		Distance:        domain.Distance{Meters: float64FromNull(row.DistanceMeters)},
		AveragePace:     domain.Pace{SecondsPerKilometer: int32FromNull(row.AveragePaceSecondsPerKilometer)},
		AverageHeartBPM: int32FromNull(row.AverageHeartBpm),
	}
}

func workoutMatchToCreateParams(match domain.WorkoutMatch) (postgresdb.CreateWorkoutMatchParams, error) {
	id, plannedWorkoutID, err := parseRecordAndParentIDs(match.ID, match.PlannedWorkoutID, "workout match")
	if err != nil {
		return postgresdb.CreateWorkoutMatchParams{}, err
	}
	importedActivityID, err := uuid.Parse(match.ImportedActivityID)
	if err != nil {
		return postgresdb.CreateWorkoutMatchParams{}, fmt.Errorf("parse workout match imported activity id: %w", err)
	}

	return postgresdb.CreateWorkoutMatchParams{
		ID:                 id,
		PlannedWorkoutID:   plannedWorkoutID,
		ImportedActivityID: importedActivityID,
		Status:             string(match.Status),
		Confidence:         string(match.Confidence),
		MatchedBy:          string(match.MatchedBy),
		MatchedAt:          match.MatchedAt,
		Notes:              nullableString(match.Notes),
	}, nil
}

func workoutMatchToUpdateParams(match domain.WorkoutMatch) (postgresdb.UpdateWorkoutMatchParams, error) {
	params, err := workoutMatchToCreateParams(match)
	if err != nil {
		return postgresdb.UpdateWorkoutMatchParams{}, err
	}
	return postgresdb.UpdateWorkoutMatchParams{
		ID:                 params.ID,
		PlannedWorkoutID:   params.PlannedWorkoutID,
		ImportedActivityID: params.ImportedActivityID,
		Status:             params.Status,
		Confidence:         params.Confidence,
		MatchedBy:          params.MatchedBy,
		MatchedAt:          params.MatchedAt,
		Notes:              params.Notes,
	}, nil
}

func workoutMatchFromDB(row postgresdb.WorkoutMatch) domain.WorkoutMatch {
	return domain.WorkoutMatch{
		ID:                 row.ID.String(),
		PlannedWorkoutID:   row.PlannedWorkoutID.String(),
		ImportedActivityID: row.ImportedActivityID.String(),
		Status:             domain.WorkoutMatchStatus(row.Status),
		Confidence:         domain.MatchConfidence(row.Confidence),
		MatchedBy:          domain.MatchSource(row.MatchedBy),
		MatchedAt:          row.MatchedAt,
		Notes:              stringFromNull(row.Notes),
	}
}

func workoutResultToCreateParams(result domain.WorkoutResult) (postgresdb.CreateWorkoutResultParams, error) {
	id, plannedWorkoutID, err := parseRecordAndParentIDs(result.ID, result.PlannedWorkoutID, "workout result")
	if err != nil {
		return postgresdb.CreateWorkoutResultParams{}, err
	}
	importedActivityID, err := nullableUUID(result.ImportedActivityID)
	if err != nil {
		return postgresdb.CreateWorkoutResultParams{}, fmt.Errorf("parse workout result imported activity id: %w", err)
	}

	return postgresdb.CreateWorkoutResultParams{
		ID:                 id,
		PlannedWorkoutID:   plannedWorkoutID,
		ImportedActivityID: importedActivityID,
		Outcome:            string(result.Outcome),
		CompletedAt:        nullableTime(result.CompletedAt),
		DistanceMeters:     nullableFloat64(result.Distance.Meters),
		DurationSeconds:    nullableDurationSeconds(result.Duration),
		Notes:              nullableString(result.Notes),
	}, nil
}

func workoutResultToUpdateParams(result domain.WorkoutResult) (postgresdb.UpdateWorkoutResultParams, error) {
	params, err := workoutResultToCreateParams(result)
	if err != nil {
		return postgresdb.UpdateWorkoutResultParams{}, err
	}
	return postgresdb.UpdateWorkoutResultParams{
		ID:                 params.ID,
		PlannedWorkoutID:   params.PlannedWorkoutID,
		ImportedActivityID: params.ImportedActivityID,
		Outcome:            params.Outcome,
		CompletedAt:        params.CompletedAt,
		DistanceMeters:     params.DistanceMeters,
		DurationSeconds:    params.DurationSeconds,
		Notes:              params.Notes,
	}, nil
}

func workoutResultFromDB(row postgresdb.WorkoutResult) domain.WorkoutResult {
	return domain.WorkoutResult{
		ID:                 row.ID.String(),
		PlannedWorkoutID:   row.PlannedWorkoutID.String(),
		ImportedActivityID: uuidStringFromNull(row.ImportedActivityID),
		Outcome:            domain.WorkoutOutcome(row.Outcome),
		CompletedAt:        timeFromNull(row.CompletedAt),
		Distance:           domain.Distance{Meters: float64FromNull(row.DistanceMeters)},
		Duration:           durationFromNullSeconds(row.DurationSeconds),
		Notes:              stringFromNull(row.Notes),
	}
}

func adaptationEventToCreateParams(event domain.AdaptationEvent) (postgresdb.CreateAdaptationEventParams, error) {
	id, planID, err := parseRecordAndParentIDs(event.ID, event.PlanID, "adaptation event")
	if err != nil {
		return postgresdb.CreateAdaptationEventParams{}, err
	}
	athleteID, err := uuid.Parse(event.AthleteID)
	if err != nil {
		return postgresdb.CreateAdaptationEventParams{}, fmt.Errorf("parse adaptation event athlete id: %w", err)
	}

	return postgresdb.CreateAdaptationEventParams{
		ID:        id,
		PlanID:    planID,
		AthleteID: athleteID,
		Type:      string(event.Type),
		Reason:    event.Reason,
		Summary:   event.Summary,
		CreatedAt: event.CreatedAt,
	}, nil
}

func adaptationEventToUpdateParams(event domain.AdaptationEvent) (postgresdb.UpdateAdaptationEventParams, error) {
	params, err := adaptationEventToCreateParams(event)
	if err != nil {
		return postgresdb.UpdateAdaptationEventParams{}, err
	}
	return postgresdb.UpdateAdaptationEventParams{
		ID:        params.ID,
		PlanID:    params.PlanID,
		AthleteID: params.AthleteID,
		Type:      params.Type,
		Reason:    params.Reason,
		Summary:   params.Summary,
		CreatedAt: params.CreatedAt,
	}, nil
}

func planChangeToCreateParams(eventID uuid.UUID, change domain.PlanChange, position int) (postgresdb.CreateAdaptationEventChangeParams, error) {
	plannedWorkoutID, err := nullableUUID(change.PlannedWorkoutID)
	if err != nil {
		return postgresdb.CreateAdaptationEventChangeParams{}, fmt.Errorf("parse adaptation event change planned workout id: %w", err)
	}
	if position > math.MaxInt32 {
		return postgresdb.CreateAdaptationEventChangeParams{}, fmt.Errorf("adaptation event change position %d exceeds int32", position)
	}

	return postgresdb.CreateAdaptationEventChangeParams{
		ID:                uuid.New(),
		AdaptationEventID: eventID,
		PlannedWorkoutID:  plannedWorkoutID,
		Type:              string(change.Type),
		Description:       change.Description,
		Position:          int32(position),
	}, nil
}

func adaptationEventFromDB(row postgresdb.AdaptationEvent, changes []postgresdb.AdaptationEventChange) domain.AdaptationEvent {
	event := domain.AdaptationEvent{
		ID:        row.ID.String(),
		PlanID:    row.PlanID.String(),
		AthleteID: row.AthleteID.String(),
		Type:      domain.AdaptationType(row.Type),
		Reason:    row.Reason,
		Summary:   row.Summary,
		CreatedAt: row.CreatedAt,
		Changes:   make([]domain.PlanChange, 0, len(changes)),
	}
	for _, change := range changes {
		event.Changes = append(event.Changes, planChangeFromDB(change))
	}
	return event
}

func planChangeFromDB(row postgresdb.AdaptationEventChange) domain.PlanChange {
	return domain.PlanChange{
		PlannedWorkoutID: uuidStringFromNull(row.PlannedWorkoutID),
		Type:             domain.PlanChangeType(row.Type),
		Description:      row.Description,
	}
}

func requiredDurationSeconds(value time.Duration, label string) (int32, error) {
	seconds := int64(value / time.Second)
	if seconds > math.MaxInt32 || seconds < math.MinInt32 {
		return 0, fmt.Errorf("%s exceeds int32 seconds", label)
	}
	return int32(seconds), nil
}

func nullableInt32(value int) sql.NullInt32 {
	if value == 0 {
		return sql.NullInt32{}
	}
	if value > math.MaxInt32 || value < math.MinInt32 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(value), Valid: true}
}

func int32FromNull(value sql.NullInt32) int {
	if !value.Valid {
		return 0
	}
	return int(value.Int32)
}

func uuidStringFromNull(value uuid.NullUUID) string {
	if !value.Valid {
		return ""
	}
	return value.UUID.String()
}
