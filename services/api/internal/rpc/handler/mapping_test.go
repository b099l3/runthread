package handler

import (
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCompleteImportedActivityRequestToApp(t *testing.T) {
	startedAt := time.Date(2026, time.June, 2, 7, 30, 0, 0, time.UTC)

	appRequest, err := completeImportedActivityRequestToApp(&rpcv1.CompleteImportedActivityRequest{
		AthleteProfile: &rpcv1.AthleteProfile{
			Id:                          "athlete-1",
			DisplayName:                 "Maya",
			ExperienceLevel:             rpcv1.ExperienceLevel_EXPERIENCE_LEVEL_BEGINNER,
			CurrentWeeklyDistanceMeters: 20000,
			PreferredRunDays:            []int32{2, 4, 0},
			Constraints:                 []string{"no Mondays"},
		},
		TrainingGoal: &rpcv1.TrainingGoal{
			Id:                   "goal-1",
			AthleteId:            "athlete-1",
			Type:                 rpcv1.GoalType_GOAL_TYPE_RACE,
			TargetDate:           "2026-10-18",
			TargetDistanceMeters: 21097.5,
		},
		TargetWeekDate: "2026-06-03",
		ImportedActivity: &rpcv1.ImportedActivity{
			Id:                             "activity-1",
			AthleteId:                      "athlete-1",
			Type:                           rpcv1.ActivityType_ACTIVITY_TYPE_RUN,
			StartedAt:                      timestamppb.New(startedAt),
			DurationSeconds:                2700,
			DistanceMeters:                 8000,
			AveragePaceSecondsPerKilometer: 337,
			AverageHeartBpm:                145,
		},
		ResultId: "result-1",
		Outcome:  rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_COMPLETED_AS_PLANNED,
	})
	if err != nil {
		t.Fatalf("completeImportedActivityRequestToApp returned error: %v", err)
	}

	if appRequest.AthleteProfile.ID != "athlete-1" {
		t.Fatalf("AthleteProfile.ID = %q, want athlete-1", appRequest.AthleteProfile.ID)
	}
	if appRequest.AthleteProfile.ExperienceLevel != domain.ExperienceLevelBeginner {
		t.Fatalf("ExperienceLevel = %q, want beginner", appRequest.AthleteProfile.ExperienceLevel)
	}
	if appRequest.TargetWeekDate.Format(dateLayout) != "2026-06-03" {
		t.Fatalf("TargetWeekDate = %s, want 2026-06-03", appRequest.TargetWeekDate.Format(dateLayout))
	}
	if appRequest.TrainingGoal.Type != domain.GoalTypeRace {
		t.Fatalf("TrainingGoal.Type = %q, want race", appRequest.TrainingGoal.Type)
	}

	activity, err := appRequest.ImportActivity(nil)
	if err != nil {
		t.Fatalf("ImportActivity returned error: %v", err)
	}
	if activity.StartedAt != startedAt {
		t.Fatalf("activity.StartedAt = %s, want %s", activity.StartedAt, startedAt)
	}
	if activity.Duration != 45*time.Minute {
		t.Fatalf("activity.Duration = %s, want 45m", activity.Duration)
	}
	if activity.AveragePace.SecondsPerKilometer != 337 {
		t.Fatalf("activity.AveragePace = %d, want 337", activity.AveragePace.SecondsPerKilometer)
	}
}

func TestRideEnumMappings(t *testing.T) {
	if got := activityTypeToDomain(rpcv1.ActivityType_ACTIVITY_TYPE_RIDE); got != domain.ActivityTypeRide {
		t.Fatalf("activityTypeToDomain ride = %q, want %q", got, domain.ActivityTypeRide)
	}
	if got := activityTypeFromDomain(domain.ActivityTypeRide); got != rpcv1.ActivityType_ACTIVITY_TYPE_RIDE {
		t.Fatalf("activityTypeFromDomain ride = %s, want ACTIVITY_TYPE_RIDE", got)
	}
	if got := workoutTypeFromDomain(domain.WorkoutTypeRide); got != rpcv1.WorkoutType_WORKOUT_TYPE_RIDE {
		t.Fatalf("workoutTypeFromDomain ride = %s, want WORKOUT_TYPE_RIDE", got)
	}
}

func TestGetCurrentPlanWeekRequestToApp(t *testing.T) {
	appRequest, err := getCurrentPlanWeekRequestToApp(&rpcv1.GetCurrentPlanWeekRequest{
		AthleteId:      "athlete-1",
		GoalId:         "goal-1",
		TargetWeekDate: "2026-06-03",
		PlanWeekId:     "week-1",
	})
	if err != nil {
		t.Fatalf("getCurrentPlanWeekRequestToApp returned error: %v", err)
	}

	if appRequest.AthleteID != "athlete-1" {
		t.Fatalf("AthleteID = %q, want athlete-1", appRequest.AthleteID)
	}
	if appRequest.GoalID != "goal-1" {
		t.Fatalf("GoalID = %q, want goal-1", appRequest.GoalID)
	}
	if appRequest.PlanWeekID != "week-1" {
		t.Fatalf("PlanWeekID = %q, want week-1", appRequest.PlanWeekID)
	}
	if appRequest.TargetWeekDate.Format(dateLayout) != "2026-06-03" {
		t.Fatalf("TargetWeekDate = %s, want 2026-06-03", appRequest.TargetWeekDate.Format(dateLayout))
	}
}

func TestProviderConnectionRequestsToApp(t *testing.T) {
	statusRequest, err := getProviderConnectionStatusRequestToApp(&rpcv1.GetProviderConnectionStatusRequest{
		AthleteId:            "athlete-1",
		Provider:             rpcv1.Provider_PROVIDER_GARMIN,
		ProviderConnectionId: "connection-1",
	})
	if err != nil {
		t.Fatalf("getProviderConnectionStatusRequestToApp returned error: %v", err)
	}
	if statusRequest.AthleteID != "athlete-1" {
		t.Fatalf("AthleteID = %q, want athlete-1", statusRequest.AthleteID)
	}
	if statusRequest.Provider != app.ProviderGarmin {
		t.Fatalf("Provider = %q, want garmin", statusRequest.Provider)
	}
	if statusRequest.ProviderConnectionID != "connection-1" {
		t.Fatalf("ProviderConnectionID = %q, want connection-1", statusRequest.ProviderConnectionID)
	}

	startRequest, err := startProviderConnectionRequestToApp(&rpcv1.StartProviderConnectionRequest{
		AthleteId:   "athlete-1",
		Provider:    rpcv1.Provider_PROVIDER_GARMIN,
		RedirectUri: "runthread://provider/garmin/callback",
	})
	if err != nil {
		t.Fatalf("startProviderConnectionRequestToApp returned error: %v", err)
	}
	if startRequest.Provider != app.ProviderGarmin {
		t.Fatalf("Provider = %q, want garmin", startRequest.Provider)
	}
	if startRequest.RedirectURI != "runthread://provider/garmin/callback" {
		t.Fatalf("RedirectURI = %q, want callback", startRequest.RedirectURI)
	}
}

func TestStravaProviderConnectionRequestsToApp(t *testing.T) {
	statusRequest, err := getProviderConnectionStatusRequestToApp(&rpcv1.GetProviderConnectionStatusRequest{
		AthleteId:            "athlete-1",
		Provider:             rpcv1.Provider_PROVIDER_STRAVA,
		ProviderConnectionId: "connection-1",
	})
	if err != nil {
		t.Fatalf("getProviderConnectionStatusRequestToApp returned error: %v", err)
	}
	if statusRequest.Provider != app.ProviderStrava {
		t.Fatalf("Provider = %q, want strava", statusRequest.Provider)
	}

	startRequest, err := startProviderConnectionRequestToApp(&rpcv1.StartProviderConnectionRequest{
		AthleteId:   "athlete-1",
		Provider:    rpcv1.Provider_PROVIDER_STRAVA,
		RedirectUri: "runthread://provider/strava/callback",
	})
	if err != nil {
		t.Fatalf("startProviderConnectionRequestToApp returned error: %v", err)
	}
	if startRequest.Provider != app.ProviderStrava {
		t.Fatalf("Provider = %q, want strava", startRequest.Provider)
	}
	if startRequest.RedirectURI != "runthread://provider/strava/callback" {
		t.Fatalf("RedirectURI = %q, want callback", startRequest.RedirectURI)
	}
}

func TestStravaProviderConnectionResponseFromApp(t *testing.T) {
	response := getProviderConnectionStatusResponseFromApp(app.GetProviderConnectionStatusResponse{
		Connection: repository.ProviderConnection{
			ID:        "connection-1",
			AthleteID: "athlete-1",
			Provider:  app.ProviderStrava,
			Status:    repository.ProviderConnectionStatusPending,
		},
		HasConnection: true,
	})

	if response.GetConnection().GetProvider() != rpcv1.Provider_PROVIDER_STRAVA {
		t.Fatalf("provider = %s, want strava", response.GetConnection().GetProvider())
	}
}

func TestProviderConnectionResponseFromApp(t *testing.T) {
	connectedAt := time.Date(2026, time.June, 5, 9, 0, 0, 0, time.UTC)
	response := getProviderConnectionStatusResponseFromApp(app.GetProviderConnectionStatusResponse{
		Connection: repository.ProviderConnection{
			ID:             "connection-1",
			AthleteID:      "athlete-1",
			Provider:       app.ProviderGarmin,
			ProviderUserID: "garmin-user-1",
			Status:         repository.ProviderConnectionStatusConnected,
			ConnectedAt:    connectedAt,
		},
		HasConnection: true,
	})

	if !response.GetHasConnection() {
		t.Fatal("expected has connection")
	}
	if response.GetConnection().GetProvider() != rpcv1.Provider_PROVIDER_GARMIN {
		t.Fatalf("provider = %s, want garmin", response.GetConnection().GetProvider())
	}
	if response.GetConnection().GetStatus() != rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_CONNECTED {
		t.Fatalf("status = %s, want connected", response.GetConnection().GetStatus())
	}
	if response.GetConnection().GetConnectedAt().AsTime() != connectedAt {
		t.Fatalf("connected at = %s, want %s", response.GetConnection().GetConnectedAt().AsTime(), connectedAt)
	}
}

func TestCompleteImportedActivityRequestToAppRejectsInvalidDate(t *testing.T) {
	_, err := completeImportedActivityRequestToApp(&rpcv1.CompleteImportedActivityRequest{
		AthleteProfile:   &rpcv1.AthleteProfile{},
		TrainingGoal:     &rpcv1.TrainingGoal{},
		TargetWeekDate:   "06/03/2026",
		ImportedActivity: &rpcv1.ImportedActivity{StartedAt: timestamppb.Now()},
	})
	if err == nil {
		t.Fatal("expected invalid date error")
	}
}
