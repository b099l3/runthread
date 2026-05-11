package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRunthreadServiceCompleteImportedActivity(t *testing.T) {
	ctx := context.Background()
	profile := testProfile()
	goal := testGoal(profile.ID)
	targetWeekDate := testDate(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	services, err := app.NewServices(repository.NewInMemoryStore())
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	handler := NewRunthreadService(services)

	response, err := handler.CompleteImportedActivity(ctx, connect.NewRequest(&rpcv1.CompleteImportedActivityRequest{
		AthleteProfile: protoProfile(profile),
		TrainingGoal:   protoGoal(goal),
		TargetWeekDate: targetWeekDate.Format(dateLayout),
		ImportedActivity: &rpcv1.ImportedActivity{
			Id:                             "activity-" + expectedWorkout.ID,
			AthleteId:                      profile.ID,
			Type:                           rpcv1.ActivityType_ACTIVITY_TYPE_RUN,
			StartedAt:                      timestamppb.New(expectedWorkout.ScheduledFor.Add(7 * time.Hour)),
			DurationSeconds:                durationToSeconds(expectedWorkout.TargetDuration),
			DistanceMeters:                 expectedWorkout.TargetDistance.Meters,
			AveragePaceSecondsPerKilometer: 330,
			AverageHeartBpm:                145,
		},
		ResultId: "result-1",
		Outcome:  rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_COMPLETED_AS_PLANNED,
	}))
	if err != nil {
		t.Fatalf("CompleteImportedActivity returned error: %v", err)
	}

	msg := response.Msg
	if msg.GetWorkoutMatch().GetStatus() != rpcv1.WorkoutMatchStatus_WORKOUT_MATCH_STATUS_MATCHED {
		t.Fatalf("match status = %s, want matched", msg.GetWorkoutMatch().GetStatus())
	}
	if msg.GetUpdatedWorkout().GetStatus() != rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_COMPLETED {
		t.Fatalf("updated workout status = %s, want completed", msg.GetUpdatedWorkout().GetStatus())
	}
	if msg.GetWorkoutResult().GetOutcome() != rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_COMPLETED_AS_PLANNED {
		t.Fatalf("result outcome = %s, want completed as planned", msg.GetWorkoutResult().GetOutcome())
	}
	if msg.GetAdaptationEvent() != nil {
		t.Fatalf("expected no adaptation event, got %#v", msg.GetAdaptationEvent())
	}
}

func TestRunthreadServiceGetCurrentPlanWeek(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	profile := testProfile()
	goal := testGoal(profile.ID)
	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		t.Fatalf("SaveAthleteProfile returned error: %v", err)
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		t.Fatalf("SaveTrainingGoal returned error: %v", err)
	}
	services, err := app.NewServices(store)
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	handler := NewRunthreadService(services)

	response, err := handler.GetCurrentPlanWeek(ctx, connect.NewRequest(&rpcv1.GetCurrentPlanWeekRequest{
		AthleteId:      profile.ID,
		GoalId:         goal.ID,
		TargetWeekDate: "2026-06-03",
	}))
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}

	if response.Msg.GetPlanWeek().GetAthleteId() != profile.ID {
		t.Fatalf("plan week athlete ID = %q, want %q", response.Msg.GetPlanWeek().GetAthleteId(), profile.ID)
	}
	if len(response.Msg.GetPlanWeek().GetWorkouts()) != 7 {
		t.Fatalf("plan week workouts = %d, want 7", len(response.Msg.GetPlanWeek().GetWorkouts()))
	}
}

func TestRunthreadServiceStartProviderConnection(t *testing.T) {
	services, err := app.NewServices(repository.NewInMemoryStore())
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	handler := NewRunthreadService(services)

	response, err := handler.StartProviderConnection(context.Background(), connect.NewRequest(&rpcv1.StartProviderConnectionRequest{
		AthleteId:   "athlete-1",
		Provider:    rpcv1.Provider_PROVIDER_GARMIN,
		RedirectUri: "runthread://provider/garmin/callback",
	}))
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if response.Msg.GetConnection().GetAthleteId() != "athlete-1" {
		t.Fatalf("connection athlete id = %q, want athlete-1", response.Msg.GetConnection().GetAthleteId())
	}
	if response.Msg.GetConnection().GetStatus() != rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_PENDING {
		t.Fatalf("connection status = %s, want pending", response.Msg.GetConnection().GetStatus())
	}
	if response.Msg.GetOauthReady() {
		t.Fatal("expected oauth ready false")
	}
	if response.Msg.GetAuthorizationUrl() != "" {
		t.Fatalf("authorization url = %q, want empty", response.Msg.GetAuthorizationUrl())
	}
}

func TestRunthreadServiceGetProviderConnectionStatus(t *testing.T) {
	ctx := context.Background()
	store := repository.NewInMemoryStore()
	if err := store.SaveProviderConnection(ctx, repository.ProviderConnection{
		ID:             "connection-1",
		AthleteID:      "athlete-1",
		Provider:       app.ProviderGarmin,
		ProviderUserID: "garmin-user-1",
		Status:         repository.ProviderConnectionStatusConnected,
		ConnectedAt:    testDate(2026, time.June, 5),
	}); err != nil {
		t.Fatalf("SaveProviderConnection returned error: %v", err)
	}
	services, err := app.NewServices(store)
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	handler := NewRunthreadService(services)

	response, err := handler.GetProviderConnectionStatus(ctx, connect.NewRequest(&rpcv1.GetProviderConnectionStatusRequest{
		AthleteId: "athlete-1",
		Provider:  rpcv1.Provider_PROVIDER_GARMIN,
	}))
	if err != nil {
		t.Fatalf("GetProviderConnectionStatus returned error: %v", err)
	}

	if !response.Msg.GetHasConnection() {
		t.Fatal("expected has connection")
	}
	if response.Msg.GetConnection().GetId() != "connection-1" {
		t.Fatalf("connection id = %q, want connection-1", response.Msg.GetConnection().GetId())
	}
}

func TestRunthreadServiceCompleteImportedActivityRejectsBadRequest(t *testing.T) {
	services, err := app.NewServices(repository.NewInMemoryStore())
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	handler := NewRunthreadService(services)

	_, err = handler.CompleteImportedActivity(context.Background(), connect.NewRequest(&rpcv1.CompleteImportedActivityRequest{}))
	if err == nil {
		t.Fatal("expected error")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("error code = %s, want invalid_argument", connectErr.Code())
	}
}

func testProfile() domain.AthleteProfile {
	return domain.AthleteProfile{
		ID:                    "athlete-1",
		DisplayName:           "Maya",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
}

func testGoal(athleteID string) domain.TrainingGoal {
	return domain.TrainingGoal{
		ID:             "goal-1",
		AthleteID:      athleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     testDate(2026, time.October, 18),
		TargetDistance: domain.Distance{Meters: 21097.5},
	}
}

func firstWorkoutOfType(t *testing.T, profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time, workoutType domain.WorkoutType) domain.PlannedWorkout {
	t.Helper()

	week, err := planning.GenerateWeek(profile, goal, targetWeekDate)
	if err != nil {
		t.Fatalf("GenerateWeek returned error: %v", err)
	}
	for _, workout := range week.Workouts {
		if workout.Type == workoutType {
			return workout
		}
	}
	t.Fatalf("expected generated workout type %q", workoutType)
	return domain.PlannedWorkout{}
}

func protoProfile(profile domain.AthleteProfile) *rpcv1.AthleteProfile {
	days := make([]int32, 0, len(profile.PreferredRunDays))
	for _, day := range profile.PreferredRunDays {
		days = append(days, int32(day))
	}
	return &rpcv1.AthleteProfile{
		Id:                          profile.ID,
		DisplayName:                 profile.DisplayName,
		ExperienceLevel:             rpcv1.ExperienceLevel_EXPERIENCE_LEVEL_BEGINNER,
		CurrentWeeklyDistanceMeters: profile.CurrentWeeklyDistance.Meters,
		PreferredRunDays:            days,
	}
}

func protoGoal(goal domain.TrainingGoal) *rpcv1.TrainingGoal {
	return &rpcv1.TrainingGoal{
		Id:                    goal.ID,
		AthleteId:             goal.AthleteID,
		Type:                  rpcv1.GoalType_GOAL_TYPE_RACE,
		TargetDate:            goal.TargetDate.Format(dateLayout),
		TargetDistanceMeters:  goal.TargetDistance.Meters,
		TargetDurationSeconds: durationToSeconds(goal.TargetDuration),
		Notes:                 goal.Notes,
	}
}

func testDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
