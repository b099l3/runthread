package garminadapter

import (
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/garmin"
	"github.com/runthread/runthread/services/api/internal/providerimport"
)

type MockCompleteProviderImportInput struct {
	AthleteID            string
	ProviderConnectionID string
	Payload              garmin.MockActivityPayload
	RawPayload           []byte
	ReceivedAt           time.Time
	DeliveryID           string
	PlanWeekID           string
	PlannedWorkoutID     string
	PlanWeek             domain.PlanWeek
	PlannedWorkout       domain.PlannedWorkout
	ResultID             string
	Outcome              domain.WorkoutOutcome
}

func BuildMockCompleteProviderImportRequest(input MockCompleteProviderImportInput) (app.CompleteProviderImportRequest, error) {
	payload := input.Payload
	athleteID := input.AthleteID
	if athleteID == "" {
		athleteID = payload.AthleteID
	}
	if payload.AthleteID == "" {
		payload.AthleteID = athleteID
	}
	if payload.AthleteID != athleteID {
		return app.CompleteProviderImportRequest{}, fmt.Errorf("payload athlete %q does not match request athlete %q", payload.AthleteID, athleteID)
	}

	importedActivity, err := garmin.NormalizeMockActivity(payload)
	if err != nil {
		return app.CompleteProviderImportRequest{}, fmt.Errorf("normalise mock Garmin activity: %w", err)
	}
	importedActivity.ID = providerimport.DeterministicID("runthread:imported-activity", input.ProviderConnectionID, payload.ActivityID)

	return app.CompleteProviderImportRequest{
		Import: providerimport.ImportRequest{
			AthleteID:            athleteID,
			ProviderConnectionID: input.ProviderConnectionID,
			ProviderActivity: providerimport.ProviderActivityInput{
				ProviderActivityID:   payload.ActivityID,
				ProviderActivityType: payload.GarminActivityType,
				StartedAt:            payload.StartTime,
			},
			ImportedActivity: &importedActivity,
			RawPayload:       input.RawPayload,
			PayloadKind:      "mock_activity",
			Delivery: providerimport.DeliveryMetadata{
				EventType:  "mock_activity_import",
				DeliveryID: input.DeliveryID,
				ReceivedAt: input.ReceivedAt,
			},
		},
		PlanWeekID:       input.PlanWeekID,
		PlannedWorkoutID: input.PlannedWorkoutID,
		PlanWeek:         input.PlanWeek,
		PlannedWorkout:   input.PlannedWorkout,
		ResultID:         input.ResultID,
		Outcome:          input.Outcome,
	}, nil
}
