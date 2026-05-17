package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/config"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/providers/strava"
	"github.com/runthread/runthread/services/api/internal/repository"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
	"github.com/runthread/runthread/services/api/internal/rpc/runthread/v1/runthreadv1connect"
	"github.com/runthread/runthread/services/api/internal/startup"
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

func TestStravaStartProviderConnectionReturnsOAuthURLWhenConfigured(t *testing.T) {
	stravaAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer stravaAPI.Close()

	store := repository.NewInMemoryStore()
	services, err := app.NewServices(store)
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	cfg := config.Config{
		StravaClientID:         "client-1",
		StravaClientSecret:     "secret-1",
		StravaOAuthRedirectURI: "runthread://provider/strava/callback",
		StravaAPIBaseURL:       stravaAPI.URL,
	}
	runtime, err := composeStravaRuntime(cfg, store)
	if err != nil {
		t.Fatalf("composeStravaRuntime returned error: %v", err)
	}
	services.ProviderConnect.ProviderStarters = map[string]app.ProviderConnectionStarter{
		app.ProviderStrava: testStravaConnectionStarter(runtime.OAuth, cfg.StravaOAuthRedirectURI),
	}

	server := httptest.NewServer(newMux(services, muxOptions{
		strava: runtime,
	}))
	defer server.Close()

	client := runthreadv1connect.NewRunthreadServiceClient(server.Client(), server.URL)
	response, err := client.StartProviderConnection(context.Background(), connect.NewRequest(&rpcv1.StartProviderConnectionRequest{
		AthleteId: "athlete-1",
		Provider:  rpcv1.Provider_PROVIDER_STRAVA,
	}))
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	if !response.Msg.GetOauthReady() {
		t.Fatal("oauth_ready = false, want true")
	}
	if !strings.HasPrefix(response.Msg.GetAuthorizationUrl(), stravaAPI.URL+"/oauth/authorize") {
		t.Fatalf("authorization_url = %q, want Strava API base", response.Msg.GetAuthorizationUrl())
	}
}

func TestStravaOAuthCallbackConnectsPendingConnection(t *testing.T) {
	stravaAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm returned error: %v", err)
			}
			if r.Form.Get("code") != "auth-code-1" {
				t.Fatalf("code = %q, want auth-code-1", r.Form.Get("code"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access-token",
				"refresh_token": "refresh-token",
				"expires_at":    time.Date(2026, time.June, 9, 16, 0, 0, 0, time.UTC).Unix(),
				"scope":         "activity:read_all",
				"athlete": map[string]any{
					"id": 12345,
				},
			})
		case "/api/v3/athlete/activities":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 98765},
			})
		case "/api/v3/activities/98765":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                98765,
				"athlete":           map[string]any{"id": 12345},
				"sport_type":        "Run",
				"name":              "Morning Run",
				"start_date":        "2026-06-09T07:30:00Z",
				"elapsed_time":      2460,
				"moving_time":       2460,
				"distance":          6800,
				"average_heartrate": 150,
			})
		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer stravaAPI.Close()

	store := repository.NewInMemoryStore()
	if err := startup.SeedDemoData(context.Background(), store); err != nil {
		t.Fatalf("SeedDemoData returned error: %v", err)
	}
	currentGoal := testGoal("athlete-1")
	currentGoal.ID = "goal-current-strava"
	currentGoal.Notes = "Current goal used by Strava completion."
	if err := store.SaveTrainingGoal(context.Background(), currentGoal); err != nil {
		t.Fatalf("SaveTrainingGoal returned error: %v", err)
	}
	cfg := config.Config{
		StravaClientID:         "client-1",
		StravaClientSecret:     "secret-1",
		StravaOAuthRedirectURI: "runthread://provider/strava/callback",
		StravaAPIBaseURL:       stravaAPI.URL,
	}
	runtime, err := composeStravaRuntime(cfg, store)
	if err != nil {
		t.Fatalf("composeStravaRuntime returned error: %v", err)
	}
	services, err := app.NewServices(store)
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}
	services.ProviderConnect.ProviderStarters = map[string]app.ProviderConnectionStarter{
		app.ProviderStrava: testStravaConnectionStarter(runtime.OAuth, cfg.StravaOAuthRedirectURI),
	}
	server := httptest.NewServer(newMux(services, muxOptions{
		strava: runtime,
	}))
	defer server.Close()

	client := runthreadv1connect.NewRunthreadServiceClient(server.Client(), server.URL)
	start, err := client.StartProviderConnection(context.Background(), connect.NewRequest(&rpcv1.StartProviderConnectionRequest{
		AthleteId: "athlete-1",
		Provider:  rpcv1.Provider_PROVIDER_STRAVA,
	}))
	if err != nil {
		t.Fatalf("StartProviderConnection returned error: %v", err)
	}

	response, err := server.Client().Get(server.URL + "/providers/strava/oauth/callback?state=" + start.Msg.GetState() + "&code=auth-code-1")
	if err != nil {
		t.Fatalf("GET callback returned error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d body = %q, want 200", response.StatusCode, string(body))
	}

	status, err := client.GetProviderConnectionStatus(context.Background(), connect.NewRequest(&rpcv1.GetProviderConnectionStatusRequest{
		AthleteId: "athlete-1",
		Provider:  rpcv1.Provider_PROVIDER_STRAVA,
	}))
	if err != nil {
		t.Fatalf("GetProviderConnectionStatus returned error: %v", err)
	}
	if status.Msg.GetConnection().GetStatus() != rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_CONNECTED {
		t.Fatalf("status = %s, want connected", status.Msg.GetConnection().GetStatus())
	}
	if status.Msg.GetConnection().GetProviderUserId() != "12345" {
		t.Fatalf("provider user = %q, want 12345", status.Msg.GetConnection().GetProviderUserId())
	}
	activities, err := store.ListProviderActivitiesByAthlete(context.Background(), "athlete-1")
	if err != nil {
		t.Fatalf("ListProviderActivitiesByAthlete returned error: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("provider activities = %d, want 1", len(activities))
	}
	if activities[0].ProviderActivityID != "98765" {
		t.Fatalf("provider activity id = %q, want 98765", activities[0].ProviderActivityID)
	}
	if activities[0].ImportedActivityID == "" {
		t.Fatal("expected imported activity id")
	}
	imported, err := store.GetImportedActivity(context.Background(), activities[0].ImportedActivityID)
	if err != nil {
		t.Fatalf("GetImportedActivity returned error: %v", err)
	}
	if imported.AthleteID != "athlete-1" {
		t.Fatalf("imported athlete = %q, want Runthread athlete", imported.AthleteID)
	}
	matches, err := store.GetWorkoutMatch(context.Background(), "match-generated-2-easy-"+imported.ID)
	if err != nil {
		t.Fatalf("GetWorkoutMatch returned error: %v", err)
	}
	if matches.Status != domain.WorkoutMatchStatusMatched {
		t.Fatalf("match status = %q, want matched", matches.Status)
	}
	result, err := store.GetWorkoutResult(context.Background(), "result-"+matches.ID)
	if err != nil {
		t.Fatalf("GetWorkoutResult returned error: %v", err)
	}
	if result.ImportedActivityID != imported.ID {
		t.Fatalf("result imported activity = %q, want %q", result.ImportedActivityID, imported.ID)
	}
	week, err := store.GetPlanWeek(context.Background(), "generated-week")
	if err != nil {
		t.Fatalf("GetPlanWeek returned error: %v", err)
	}
	if week.GoalID != currentGoal.ID {
		t.Fatalf("plan week goal id = %q, want current goal", week.GoalID)
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

func testStravaConnectionStarter(oauth *strava.OAuthService, redirectURI string) app.ProviderConnectionStarter {
	return strava.ConnectionStarter{
		OAuth:       oauth,
		RedirectURI: redirectURI,
		Scopes:      []string{"activity:read_all"},
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

func durationToSeconds(duration time.Duration) int64 {
	return int64(duration / time.Second)
}
