package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
	"github.com/runthread/runthread/services/api/internal/rpc/runthread/v1/runthreadv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHealthzRoute(t *testing.T) {
	server := httptest.NewServer(newMux(testServices(t)))
	defer server.Close()

	response, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz returned error: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q, want health json", string(body))
	}
}

func TestCompleteImportedActivityConnectEndpoint(t *testing.T) {
	server := httptest.NewServer(newMux(testServices(t)))
	defer server.Close()

	profile := testProfile()
	goal := testGoal(profile.ID)
	targetWeekDate := testDate(2026, time.June, 3)
	expectedWorkout := firstWorkoutOfType(t, profile, goal, targetWeekDate, domain.WorkoutTypeEasy)

	client := runthreadv1connect.NewRunthreadServiceClient(server.Client(), server.URL)
	response, err := client.CompleteImportedActivity(context.Background(), connect.NewRequest(&rpcv1.CompleteImportedActivityRequest{
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
		t.Fatalf("workout result outcome = %s, want completed as planned", msg.GetWorkoutResult().GetOutcome())
	}
	if msg.GetAdaptationEvent() != nil {
		t.Fatalf("adaptation event = %#v, want nil", msg.GetAdaptationEvent())
	}
}

func TestGetCurrentPlanWeekConnectEndpoint(t *testing.T) {
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

	server := httptest.NewServer(newMux(services))
	defer server.Close()

	targetWeekDate := testDate(2026, time.June, 3)
	client := runthreadv1connect.NewRunthreadServiceClient(server.Client(), server.URL)
	response, err := client.GetCurrentPlanWeek(ctx, connect.NewRequest(&rpcv1.GetCurrentPlanWeekRequest{
		AthleteId:      profile.ID,
		GoalId:         goal.ID,
		TargetWeekDate: targetWeekDate.Format(dateLayout),
	}))
	if err != nil {
		t.Fatalf("GetCurrentPlanWeek returned error: %v", err)
	}

	week := response.Msg.GetPlanWeek()
	if week.GetAthleteId() != profile.ID {
		t.Fatalf("plan week athlete ID = %q, want %q", week.GetAthleteId(), profile.ID)
	}
	if week.GetGoalId() != goal.ID {
		t.Fatalf("plan week goal ID = %q, want %q", week.GetGoalId(), goal.ID)
	}
	if len(week.GetWorkouts()) != 7 {
		t.Fatalf("plan week workouts = %d, want 7", len(week.GetWorkouts()))
	}
	if week.GetStartsOn() != "2026-06-01" {
		t.Fatalf("plan week starts_on = %q, want 2026-06-01", week.GetStartsOn())
	}
}

func TestProviderConnectionConnectEndpoints(t *testing.T) {
	server := httptest.NewServer(newMux(testServices(t)))
	defer server.Close()

	ctx := context.Background()
	client := runthreadv1connect.NewRunthreadServiceClient(server.Client(), server.URL)
	startResponse, err := client.StartProviderConnection(ctx, connect.NewRequest(&rpcv1.StartProviderConnectionRequest{
		AthleteId:   "athlete-1",
		Provider:    rpcv1.Provider_PROVIDER_GARMIN,
		RedirectUri: "runthread://provider/garmin/callback",
	}))
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	startMsg := startResponse.Msg
	if startMsg.GetConnection().GetAthleteId() != "athlete-1" {
		t.Fatalf("connection athlete ID = %q, want athlete-1", startMsg.GetConnection().GetAthleteId())
	}
	if startMsg.GetConnection().GetProvider() != rpcv1.Provider_PROVIDER_GARMIN {
		t.Fatalf("connection provider = %s, want garmin", startMsg.GetConnection().GetProvider())
	}
	if startMsg.GetConnection().GetStatus() != rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_PENDING {
		t.Fatalf("connection status = %s, want pending", startMsg.GetConnection().GetStatus())
	}
	if startMsg.GetOauthReady() {
		t.Fatal("oauth_ready = true, want false")
	}
	if startMsg.GetAuthorizationUrl() != "" {
		t.Fatalf("authorization_url = %q, want empty placeholder", startMsg.GetAuthorizationUrl())
	}

	statusResponse, err := client.GetProviderConnectionStatus(ctx, connect.NewRequest(&rpcv1.GetProviderConnectionStatusRequest{
		AthleteId: "athlete-1",
		Provider:  rpcv1.Provider_PROVIDER_GARMIN,
	}))
	if err != nil {
		t.Fatalf("GetProviderConnectionStatus returned error: %v", err)
	}

	statusMsg := statusResponse.Msg
	if !statusMsg.GetHasConnection() {
		t.Fatal("has_connection = false, want true")
	}
	if statusMsg.GetConnection().GetId() != startMsg.GetConnection().GetId() {
		t.Fatalf("status connection ID = %q, want started connection %q", statusMsg.GetConnection().GetId(), startMsg.GetConnection().GetId())
	}
	if statusMsg.GetConnection().GetStatus() != rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_PENDING {
		t.Fatalf("status connection status = %s, want pending", statusMsg.GetConnection().GetStatus())
	}
}

const dateLayout = "2006-01-02"

func testServices(t *testing.T) app.Services {
	t.Helper()

	services, err := app.NewServices(repository.NewInMemoryStore())
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	return services
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

func durationToSeconds(duration time.Duration) int64 {
	return int64(duration / time.Second)
}
