package handler

import (
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
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
