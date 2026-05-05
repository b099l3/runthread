package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
)

func TestTrainingGoalToCreateParams(t *testing.T) {
	goalID := uuid.NewString()
	athleteID := uuid.NewString()
	targetDate := date(2026, time.October, 18)
	goal := domain.TrainingGoal{
		ID:             goalID,
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     targetDate,
		TargetDistance: domain.Distance{Meters: 21097.5},
		TargetDuration: 90 * time.Minute,
		Notes:          "Half marathon tune-up",
	}

	params, err := trainingGoalToCreateParams(goal)
	if err != nil {
		t.Fatalf("trainingGoalToCreateParams returned error: %v", err)
	}

	if params.ID.String() != goalID {
		t.Fatalf("ID = %q, want %q", params.ID.String(), goalID)
	}
	if params.AthleteID.String() != athleteID {
		t.Fatalf("AthleteID = %q, want %q", params.AthleteID.String(), athleteID)
	}
	if params.Type != string(domain.GoalTypeRace) {
		t.Fatalf("Type = %q, want race", params.Type)
	}
	if !params.TargetDate.Valid || !params.TargetDate.Time.Equal(targetDate) {
		t.Fatalf("TargetDate = %#v, want valid %s", params.TargetDate, targetDate)
	}
	if !params.TargetDistanceMeters.Valid || params.TargetDistanceMeters.Float64 != 21097.5 {
		t.Fatalf("TargetDistanceMeters = %#v, want valid 21097.5", params.TargetDistanceMeters)
	}
	if !params.TargetDurationSeconds.Valid || params.TargetDurationSeconds.Int32 != 5400 {
		t.Fatalf("TargetDurationSeconds = %#v, want valid 5400", params.TargetDurationSeconds)
	}
	if !params.Notes.Valid || params.Notes.String != "Half marathon tune-up" {
		t.Fatalf("Notes = %#v, want valid note", params.Notes)
	}
}

func TestTrainingGoalFromDB(t *testing.T) {
	goalID := uuid.New()
	athleteID := uuid.New()
	targetDate := date(2026, time.November, 8)

	goal := trainingGoalFromDB(postgresdb.TrainingGoal{
		ID:                    goalID,
		AthleteID:             athleteID,
		Type:                  string(domain.GoalTypeDistance),
		TargetDate:            sql.NullTime{Time: targetDate, Valid: true},
		TargetDistanceMeters:  sql.NullFloat64{Float64: 10000, Valid: true},
		TargetDurationSeconds: sql.NullInt32{Int32: 3600, Valid: true},
		Notes:                 sql.NullString{String: "10k", Valid: true},
	})

	if goal.ID != goalID.String() {
		t.Fatalf("ID = %q, want %q", goal.ID, goalID.String())
	}
	if goal.AthleteID != athleteID.String() {
		t.Fatalf("AthleteID = %q, want %q", goal.AthleteID, athleteID.String())
	}
	if goal.Type != domain.GoalTypeDistance {
		t.Fatalf("Type = %q, want distance", goal.Type)
	}
	if !goal.TargetDate.Equal(targetDate) {
		t.Fatalf("TargetDate = %s, want %s", goal.TargetDate, targetDate)
	}
	if goal.TargetDistance.Meters != 10000 {
		t.Fatalf("TargetDistance.Meters = %v, want 10000", goal.TargetDistance.Meters)
	}
	if goal.TargetDuration != time.Hour {
		t.Fatalf("TargetDuration = %s, want 1h", goal.TargetDuration)
	}
	if goal.Notes != "10k" {
		t.Fatalf("Notes = %q, want 10k", goal.Notes)
	}
}

func TestPlanWeekToCreateParams(t *testing.T) {
	weekID := uuid.NewString()
	planID := uuid.NewString()
	startsOn := date(2026, time.June, 1)

	athleteID := uuid.NewString()
	goalID := uuid.NewString()

	params, err := planWeekToCreateParams(domain.PlanWeek{
		ID:        weekID,
		AthleteID: athleteID,
		GoalID:    goalID,
		PlanID:    planID,
		WeekIndex: 3,
		StartsOn:  startsOn,
		Focus:     domain.WeekFocusBuild,
	})
	if err != nil {
		t.Fatalf("planWeekToCreateParams returned error: %v", err)
	}

	if params.ID.String() != weekID {
		t.Fatalf("ID = %q, want %q", params.ID.String(), weekID)
	}
	if params.AthleteID.String() != athleteID {
		t.Fatalf("AthleteID = %s, want %s", params.AthleteID, athleteID)
	}
	if !params.GoalID.Valid || params.GoalID.UUID.String() != goalID {
		t.Fatalf("GoalID = %#v, want valid %s", params.GoalID, goalID)
	}
	if params.PlanID.String() != planID {
		t.Fatalf("PlanID = %q, want %q", params.PlanID.String(), planID)
	}
	if params.WeekIndex != 3 {
		t.Fatalf("WeekIndex = %d, want 3", params.WeekIndex)
	}
	if !params.StartsOn.Equal(startsOn) {
		t.Fatalf("StartsOn = %s, want %s", params.StartsOn, startsOn)
	}
	if !params.Focus.Valid || params.Focus.String != string(domain.WeekFocusBuild) {
		t.Fatalf("Focus = %#v, want build", params.Focus)
	}
}

func TestPlanWeekFromDBDoesNotInventWorkouts(t *testing.T) {
	weekID := uuid.New()
	athleteID := uuid.New()
	goalID := uuid.New()
	planID := uuid.New()
	startsOn := date(2026, time.June, 8)

	week := planWeekFromDB(postgresdb.PlanWeek{
		ID:        weekID,
		AthleteID: athleteID,
		GoalID:    uuid.NullUUID{UUID: goalID, Valid: true},
		PlanID:    planID,
		WeekIndex: 4,
		StartsOn:  startsOn,
		Focus:     sql.NullString{String: string(domain.WeekFocusRecovery), Valid: true},
	})

	if week.ID != weekID.String() {
		t.Fatalf("ID = %q, want %q", week.ID, weekID.String())
	}
	if week.AthleteID != athleteID.String() {
		t.Fatalf("AthleteID = %q, want %q", week.AthleteID, athleteID.String())
	}
	if week.GoalID != goalID.String() {
		t.Fatalf("GoalID = %q, want %q", week.GoalID, goalID.String())
	}
	if week.PlanID != planID.String() {
		t.Fatalf("PlanID = %q, want %q", week.PlanID, planID.String())
	}
	if week.WeekIndex != 4 {
		t.Fatalf("WeekIndex = %d, want 4", week.WeekIndex)
	}
	if !week.StartsOn.Equal(startsOn) {
		t.Fatalf("StartsOn = %s, want %s", week.StartsOn, startsOn)
	}
	if week.Focus != domain.WeekFocusRecovery {
		t.Fatalf("Focus = %q, want recovery", week.Focus)
	}
	if week.Workouts != nil {
		t.Fatalf("Workouts = %#v, want nil because workouts are stored separately", week.Workouts)
	}
}

func TestPlannedWorkoutToCreateParams(t *testing.T) {
	workoutID := uuid.NewString()
	planID := uuid.NewString()
	weekID := uuid.NewString()
	scheduledFor := date(2026, time.June, 9)
	workout := domain.PlannedWorkout{
		ID:             workoutID,
		PlanID:         planID,
		PlanWeekID:     weekID,
		ScheduledFor:   scheduledFor,
		Type:           domain.WorkoutTypeWorkout,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: 8000},
		TargetDuration: 42 * time.Minute,
		Intensity: domain.IntensityTarget{
			Kind:        domain.IntensityKindTempo,
			Description: "20 minutes steady",
		},
		Notes: "Keep it controlled",
	}

	params, err := plannedWorkoutToCreateParams(workout)
	if err != nil {
		t.Fatalf("plannedWorkoutToCreateParams returned error: %v", err)
	}

	if params.ID.String() != workoutID {
		t.Fatalf("ID = %q, want %q", params.ID.String(), workoutID)
	}
	if params.PlanID.String() != planID {
		t.Fatalf("PlanID = %q, want %q", params.PlanID.String(), planID)
	}
	if params.PlanWeekID.String() != weekID {
		t.Fatalf("PlanWeekID = %q, want %q", params.PlanWeekID.String(), weekID)
	}
	if !params.ScheduledFor.Equal(scheduledFor) {
		t.Fatalf("ScheduledFor = %s, want %s", params.ScheduledFor, scheduledFor)
	}
	if params.Type != string(domain.WorkoutTypeWorkout) {
		t.Fatalf("Type = %q, want workout", params.Type)
	}
	if params.Status != string(domain.PlannedWorkoutStatusScheduled) {
		t.Fatalf("Status = %q, want scheduled", params.Status)
	}
	if !params.TargetDistanceMeters.Valid || params.TargetDistanceMeters.Float64 != 8000 {
		t.Fatalf("TargetDistanceMeters = %#v, want valid 8000", params.TargetDistanceMeters)
	}
	if !params.TargetDurationSeconds.Valid || params.TargetDurationSeconds.Int32 != 2520 {
		t.Fatalf("TargetDurationSeconds = %#v, want valid 2520", params.TargetDurationSeconds)
	}
	if !params.IntensityKind.Valid || params.IntensityKind.String != string(domain.IntensityKindTempo) {
		t.Fatalf("IntensityKind = %#v, want tempo", params.IntensityKind)
	}
	if !params.IntensityDescription.Valid || params.IntensityDescription.String != "20 minutes steady" {
		t.Fatalf("IntensityDescription = %#v, want description", params.IntensityDescription)
	}
	if !params.Notes.Valid || params.Notes.String != "Keep it controlled" {
		t.Fatalf("Notes = %#v, want note", params.Notes)
	}
}

func TestPlannedWorkoutFromDB(t *testing.T) {
	workoutID := uuid.New()
	planID := uuid.New()
	weekID := uuid.New()
	scheduledFor := date(2026, time.June, 14)

	workout := plannedWorkoutFromDB(postgresdb.PlannedWorkout{
		ID:                    workoutID,
		PlanID:                planID,
		PlanWeekID:            weekID,
		ScheduledFor:          scheduledFor,
		Type:                  string(domain.WorkoutTypeLongRun),
		Status:                string(domain.PlannedWorkoutStatusCompleted),
		TargetDistanceMeters:  sql.NullFloat64{Float64: 14000, Valid: true},
		TargetDurationSeconds: sql.NullInt32{Int32: 5400, Valid: true},
		IntensityKind:         sql.NullString{String: string(domain.IntensityKindEasy), Valid: true},
		IntensityDescription:  sql.NullString{String: "Conversational", Valid: true},
		Notes:                 sql.NullString{String: "Fuel early", Valid: true},
	})

	if workout.ID != workoutID.String() {
		t.Fatalf("ID = %q, want %q", workout.ID, workoutID.String())
	}
	if workout.PlanID != planID.String() {
		t.Fatalf("PlanID = %q, want %q", workout.PlanID, planID.String())
	}
	if workout.PlanWeekID != weekID.String() {
		t.Fatalf("PlanWeekID = %q, want %q", workout.PlanWeekID, weekID.String())
	}
	if !workout.ScheduledFor.Equal(scheduledFor) {
		t.Fatalf("ScheduledFor = %s, want %s", workout.ScheduledFor, scheduledFor)
	}
	if workout.Type != domain.WorkoutTypeLongRun {
		t.Fatalf("Type = %q, want long_run", workout.Type)
	}
	if workout.Status != domain.PlannedWorkoutStatusCompleted {
		t.Fatalf("Status = %q, want completed", workout.Status)
	}
	if workout.TargetDistance.Meters != 14000 {
		t.Fatalf("TargetDistance.Meters = %v, want 14000", workout.TargetDistance.Meters)
	}
	if workout.TargetDuration != 90*time.Minute {
		t.Fatalf("TargetDuration = %s, want 90m", workout.TargetDuration)
	}
	if workout.Intensity.Kind != domain.IntensityKindEasy {
		t.Fatalf("Intensity.Kind = %q, want easy", workout.Intensity.Kind)
	}
	if workout.Intensity.Description != "Conversational" {
		t.Fatalf("Intensity.Description = %q, want Conversational", workout.Intensity.Description)
	}
	if workout.Notes != "Fuel early" {
		t.Fatalf("Notes = %q, want Fuel early", workout.Notes)
	}
}

func TestNullableHelpersUseNullForZeroValues(t *testing.T) {
	if nullableTime(time.Time{}).Valid {
		t.Fatal("nullableTime zero is valid, want null")
	}
	if nullableFloat64(0).Valid {
		t.Fatal("nullableFloat64 zero is valid, want null")
	}
	if nullableDurationSeconds(0).Valid {
		t.Fatal("nullableDurationSeconds zero is valid, want null")
	}
}

func TestNullableUUID(t *testing.T) {
	if value, err := nullableUUID(""); err != nil || value.Valid {
		t.Fatalf("nullableUUID empty = %#v, %v; want invalid nil error", value, err)
	}

	id := uuid.New()
	value, err := nullableUUID(id.String())
	if err != nil {
		t.Fatalf("nullableUUID returned error: %v", err)
	}
	if !value.Valid || value.UUID != id {
		t.Fatalf("nullableUUID = %#v, want valid %s", value, id)
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
