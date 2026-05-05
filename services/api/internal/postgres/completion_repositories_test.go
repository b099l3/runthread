package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
)

func TestImportedActivityToCreateParams(t *testing.T) {
	activityID := uuid.NewString()
	athleteID := uuid.NewString()
	startedAt := testDateTime(2026, time.June, 2, 7, 30)

	params, err := importedActivityToCreateParams(domain.ImportedActivity{
		ID:              activityID,
		AthleteID:       athleteID,
		Type:            domain.ActivityTypeRun,
		StartedAt:       startedAt,
		Duration:        44 * time.Minute,
		Distance:        domain.Distance{Meters: 7900},
		AveragePace:     domain.Pace{SecondsPerKilometer: 334},
		AverageHeartBPM: 146,
	})
	if err != nil {
		t.Fatalf("importedActivityToCreateParams returned error: %v", err)
	}

	if params.ID.String() != activityID {
		t.Fatalf("ID = %q, want %q", params.ID.String(), activityID)
	}
	if params.AthleteID.String() != athleteID {
		t.Fatalf("AthleteID = %q, want %q", params.AthleteID.String(), athleteID)
	}
	if params.Type != string(domain.ActivityTypeRun) {
		t.Fatalf("Type = %q, want run", params.Type)
	}
	if !params.StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %s, want %s", params.StartedAt, startedAt)
	}
	if params.DurationSeconds != 2640 {
		t.Fatalf("DurationSeconds = %d, want 2640", params.DurationSeconds)
	}
	if !params.DistanceMeters.Valid || params.DistanceMeters.Float64 != 7900 {
		t.Fatalf("DistanceMeters = %#v, want valid 7900", params.DistanceMeters)
	}
	if !params.AveragePaceSecondsPerKilometer.Valid || params.AveragePaceSecondsPerKilometer.Int32 != 334 {
		t.Fatalf("AveragePaceSecondsPerKilometer = %#v, want valid 334", params.AveragePaceSecondsPerKilometer)
	}
	if !params.AverageHeartBpm.Valid || params.AverageHeartBpm.Int32 != 146 {
		t.Fatalf("AverageHeartBpm = %#v, want valid 146", params.AverageHeartBpm)
	}
}

func TestImportedActivityFromDB(t *testing.T) {
	activityID := uuid.New()
	athleteID := uuid.New()
	startedAt := testDateTime(2026, time.June, 2, 8, 15)

	activity := importedActivityFromDB(postgresdb.ImportedActivity{
		ID:                             activityID,
		AthleteID:                      athleteID,
		Type:                           string(domain.ActivityTypeTrailRun),
		StartedAt:                      startedAt,
		DurationSeconds:                3600,
		DistanceMeters:                 sql.NullFloat64{Float64: 10000, Valid: true},
		AveragePaceSecondsPerKilometer: sql.NullInt32{Int32: 360, Valid: true},
		AverageHeartBpm:                sql.NullInt32{Int32: 150, Valid: true},
	})

	if activity.ID != activityID.String() {
		t.Fatalf("ID = %q, want %q", activity.ID, activityID.String())
	}
	if activity.AthleteID != athleteID.String() {
		t.Fatalf("AthleteID = %q, want %q", activity.AthleteID, athleteID.String())
	}
	if activity.Type != domain.ActivityTypeTrailRun {
		t.Fatalf("Type = %q, want trail_run", activity.Type)
	}
	if !activity.StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %s, want %s", activity.StartedAt, startedAt)
	}
	if activity.Duration != time.Hour {
		t.Fatalf("Duration = %s, want 1h", activity.Duration)
	}
	if activity.Distance.Meters != 10000 {
		t.Fatalf("Distance.Meters = %v, want 10000", activity.Distance.Meters)
	}
	if activity.AveragePace.SecondsPerKilometer != 360 {
		t.Fatalf("AveragePace = %d, want 360", activity.AveragePace.SecondsPerKilometer)
	}
	if activity.AverageHeartBPM != 150 {
		t.Fatalf("AverageHeartBPM = %d, want 150", activity.AverageHeartBPM)
	}
}

func TestWorkoutMatchMappings(t *testing.T) {
	matchID := uuid.NewString()
	workoutID := uuid.NewString()
	activityID := uuid.NewString()
	matchedAt := testDateTime(2026, time.June, 2, 9, 0)

	params, err := workoutMatchToCreateParams(domain.WorkoutMatch{
		ID:                 matchID,
		PlannedWorkoutID:   workoutID,
		ImportedActivityID: activityID,
		Status:             domain.WorkoutMatchStatusMatched,
		Confidence:         domain.MatchConfidenceHigh,
		MatchedBy:          domain.MatchSourceAutomatic,
		MatchedAt:          matchedAt,
		Notes:              "Matched on date and distance",
	})
	if err != nil {
		t.Fatalf("workoutMatchToCreateParams returned error: %v", err)
	}
	if params.ID.String() != matchID || params.PlannedWorkoutID.String() != workoutID || params.ImportedActivityID.String() != activityID {
		t.Fatalf("ids = %s/%s/%s, want %s/%s/%s", params.ID, params.PlannedWorkoutID, params.ImportedActivityID, matchID, workoutID, activityID)
	}
	if params.Status != string(domain.WorkoutMatchStatusMatched) || params.Confidence != string(domain.MatchConfidenceHigh) || params.MatchedBy != string(domain.MatchSourceAutomatic) {
		t.Fatalf("match fields = %q/%q/%q", params.Status, params.Confidence, params.MatchedBy)
	}
	if !params.MatchedAt.Equal(matchedAt) {
		t.Fatalf("MatchedAt = %s, want %s", params.MatchedAt, matchedAt)
	}
	if !params.Notes.Valid || params.Notes.String != "Matched on date and distance" {
		t.Fatalf("Notes = %#v, want valid note", params.Notes)
	}

	match := workoutMatchFromDB(postgresdb.WorkoutMatch{
		ID:                 params.ID,
		PlannedWorkoutID:   params.PlannedWorkoutID,
		ImportedActivityID: params.ImportedActivityID,
		Status:             params.Status,
		Confidence:         params.Confidence,
		MatchedBy:          params.MatchedBy,
		MatchedAt:          params.MatchedAt,
		Notes:              params.Notes,
	})
	if match.ID != matchID || match.PlannedWorkoutID != workoutID || match.ImportedActivityID != activityID {
		t.Fatalf("mapped ids = %s/%s/%s", match.ID, match.PlannedWorkoutID, match.ImportedActivityID)
	}
	if match.Status != domain.WorkoutMatchStatusMatched || match.Confidence != domain.MatchConfidenceHigh || match.MatchedBy != domain.MatchSourceAutomatic {
		t.Fatalf("mapped match fields = %q/%q/%q", match.Status, match.Confidence, match.MatchedBy)
	}
}

func TestWorkoutResultMappings(t *testing.T) {
	resultID := uuid.NewString()
	workoutID := uuid.NewString()
	activityID := uuid.NewString()
	completedAt := testDateTime(2026, time.June, 2, 8, 30)

	params, err := workoutResultToCreateParams(domain.WorkoutResult{
		ID:                 resultID,
		PlannedWorkoutID:   workoutID,
		ImportedActivityID: activityID,
		Outcome:            domain.WorkoutOutcomeCompletedAsPlanned,
		CompletedAt:        completedAt,
		Distance:           domain.Distance{Meters: 7900},
		Duration:           44 * time.Minute,
		Notes:              "Solid run",
	})
	if err != nil {
		t.Fatalf("workoutResultToCreateParams returned error: %v", err)
	}
	if params.ID.String() != resultID || params.PlannedWorkoutID.String() != workoutID {
		t.Fatalf("ids = %s/%s, want %s/%s", params.ID, params.PlannedWorkoutID, resultID, workoutID)
	}
	if !params.ImportedActivityID.Valid || params.ImportedActivityID.UUID.String() != activityID {
		t.Fatalf("ImportedActivityID = %#v, want valid %s", params.ImportedActivityID, activityID)
	}
	if params.Outcome != string(domain.WorkoutOutcomeCompletedAsPlanned) {
		t.Fatalf("Outcome = %q, want completed_as_planned", params.Outcome)
	}
	if !params.CompletedAt.Valid || !params.CompletedAt.Time.Equal(completedAt) {
		t.Fatalf("CompletedAt = %#v, want valid %s", params.CompletedAt, completedAt)
	}
	if !params.DistanceMeters.Valid || params.DistanceMeters.Float64 != 7900 {
		t.Fatalf("DistanceMeters = %#v, want valid 7900", params.DistanceMeters)
	}
	if !params.DurationSeconds.Valid || params.DurationSeconds.Int32 != 2640 {
		t.Fatalf("DurationSeconds = %#v, want valid 2640", params.DurationSeconds)
	}

	result := workoutResultFromDB(postgresdb.WorkoutResult{
		ID:                 params.ID,
		PlannedWorkoutID:   params.PlannedWorkoutID,
		ImportedActivityID: params.ImportedActivityID,
		Outcome:            params.Outcome,
		CompletedAt:        params.CompletedAt,
		DistanceMeters:     params.DistanceMeters,
		DurationSeconds:    params.DurationSeconds,
		Notes:              params.Notes,
	})
	if result.ID != resultID || result.PlannedWorkoutID != workoutID || result.ImportedActivityID != activityID {
		t.Fatalf("mapped ids = %s/%s/%s", result.ID, result.PlannedWorkoutID, result.ImportedActivityID)
	}
	if result.Outcome != domain.WorkoutOutcomeCompletedAsPlanned {
		t.Fatalf("Outcome = %q, want completed_as_planned", result.Outcome)
	}
	if result.Duration != 44*time.Minute {
		t.Fatalf("Duration = %s, want 44m", result.Duration)
	}
}

func TestWorkoutResultToCreateParamsAllowsMissedWithoutImportedActivity(t *testing.T) {
	params, err := workoutResultToCreateParams(domain.WorkoutResult{
		ID:               uuid.NewString(),
		PlannedWorkoutID: uuid.NewString(),
		Outcome:          domain.WorkoutOutcomeMissed,
	})
	if err != nil {
		t.Fatalf("workoutResultToCreateParams returned error: %v", err)
	}
	if params.ImportedActivityID.Valid {
		t.Fatalf("ImportedActivityID = %#v, want null", params.ImportedActivityID)
	}
}

func TestAdaptationEventMappings(t *testing.T) {
	eventID := uuid.New()
	planID := uuid.NewString()
	athleteID := uuid.NewString()
	workoutID := uuid.NewString()
	createdAt := testDateTime(2026, time.June, 3, 12, 0)
	event := domain.AdaptationEvent{
		ID:        eventID.String(),
		PlanID:    planID,
		AthleteID: athleteID,
		Type:      domain.AdaptationTypeUnderperformance,
		Reason:    "Workout was shorter than planned.",
		Summary:   "Keep the next workout conservative.",
		CreatedAt: createdAt,
		Changes: []domain.PlanChange{
			{
				PlannedWorkoutID: workoutID,
				Type:             domain.PlanChangeTypeWorkoutAdjusted,
				Description:      "Reduce next workout slightly.",
			},
		},
	}

	params, err := adaptationEventToCreateParams(event)
	if err != nil {
		t.Fatalf("adaptationEventToCreateParams returned error: %v", err)
	}
	if params.ID != eventID {
		t.Fatalf("ID = %s, want %s", params.ID, eventID)
	}
	if params.PlanID.String() != planID || params.AthleteID.String() != athleteID {
		t.Fatalf("ids = %s/%s, want %s/%s", params.PlanID, params.AthleteID, planID, athleteID)
	}
	if params.Type != string(domain.AdaptationTypeUnderperformance) {
		t.Fatalf("Type = %q, want underperformance", params.Type)
	}

	changeParams, err := planChangeToCreateParams(eventID, event.Changes[0], 0)
	if err != nil {
		t.Fatalf("planChangeToCreateParams returned error: %v", err)
	}
	if changeParams.ID == uuid.Nil {
		t.Fatal("change ID is nil, want generated UUID")
	}
	if changeParams.AdaptationEventID != eventID {
		t.Fatalf("AdaptationEventID = %s, want %s", changeParams.AdaptationEventID, eventID)
	}
	if !changeParams.PlannedWorkoutID.Valid || changeParams.PlannedWorkoutID.UUID.String() != workoutID {
		t.Fatalf("PlannedWorkoutID = %#v, want valid %s", changeParams.PlannedWorkoutID, workoutID)
	}
	if changeParams.Type != string(domain.PlanChangeTypeWorkoutAdjusted) || changeParams.Position != 0 {
		t.Fatalf("change type/position = %q/%d", changeParams.Type, changeParams.Position)
	}

	mapped := adaptationEventFromDB(postgresdb.AdaptationEvent{
		ID:        params.ID,
		PlanID:    params.PlanID,
		AthleteID: params.AthleteID,
		Type:      params.Type,
		Reason:    params.Reason,
		Summary:   params.Summary,
		CreatedAt: params.CreatedAt,
	}, []postgresdb.AdaptationEventChange{
		{
			ID:                changeParams.ID,
			AdaptationEventID: changeParams.AdaptationEventID,
			PlannedWorkoutID:  changeParams.PlannedWorkoutID,
			Type:              changeParams.Type,
			Description:       changeParams.Description,
			Position:          changeParams.Position,
		},
	})
	if mapped.ID != event.ID || mapped.PlanID != event.PlanID || mapped.AthleteID != event.AthleteID {
		t.Fatalf("mapped ids = %s/%s/%s", mapped.ID, mapped.PlanID, mapped.AthleteID)
	}
	if len(mapped.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(mapped.Changes))
	}
	if mapped.Changes[0].PlannedWorkoutID != workoutID || mapped.Changes[0].Type != domain.PlanChangeTypeWorkoutAdjusted {
		t.Fatalf("mapped change = %#v", mapped.Changes[0])
	}
}

func TestOptionalNumericHelpersUseNullForZeroValues(t *testing.T) {
	if nullableInt32(0).Valid {
		t.Fatal("nullableInt32 zero is valid, want null")
	}
	if int32FromNull(sql.NullInt32{}) != 0 {
		t.Fatal("int32FromNull null did not return zero")
	}
	if uuidStringFromNull(uuid.NullUUID{}) != "" {
		t.Fatal("uuidStringFromNull null did not return empty string")
	}
}

func testDateTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}
