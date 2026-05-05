package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/domain"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const dateLayout = "2006-01-02"

func getCurrentPlanWeekRequestToApp(req *rpcv1.GetCurrentPlanWeekRequest) (app.GetCurrentPlanWeekRequest, error) {
	if req == nil {
		return app.GetCurrentPlanWeekRequest{}, fmt.Errorf("request is required")
	}

	targetWeekDate, err := parseOptionalDate(req.GetTargetWeekDate(), "target week date")
	if err != nil {
		return app.GetCurrentPlanWeekRequest{}, err
	}

	// TODO: Replace request-supplied athlete, goal, and week identifiers with
	// authenticated user context plus current-plan lookup once auth exists.
	return app.GetCurrentPlanWeekRequest{
		PlanWeekID:     req.GetPlanWeekId(),
		AthleteID:      req.GetAthleteId(),
		GoalID:         req.GetGoalId(),
		TargetWeekDate: targetWeekDate,
	}, nil
}

func getCurrentPlanWeekResponseFromApp(response app.GetCurrentPlanWeekResponse) *rpcv1.GetCurrentPlanWeekResponse {
	activities := make([]*rpcv1.ImportedActivity, 0, len(response.ImportedActivities))
	for _, activity := range response.ImportedActivities {
		activities = append(activities, importedActivityFromDomain(activity))
	}
	matches := make([]*rpcv1.WorkoutMatch, 0, len(response.WorkoutMatches))
	for _, match := range response.WorkoutMatches {
		matches = append(matches, workoutMatchFromDomain(match))
	}
	results := make([]*rpcv1.WorkoutResult, 0, len(response.WorkoutResults))
	for _, result := range response.WorkoutResults {
		results = append(results, workoutResultFromDomain(result))
	}
	events := make([]*rpcv1.AdaptationEvent, 0, len(response.AdaptationEvents))
	for _, event := range response.AdaptationEvents {
		events = append(events, adaptationEventFromDomain(&event))
	}

	return &rpcv1.GetCurrentPlanWeekResponse{
		PlanWeek:           planWeekFromDomain(response.PlanWeek),
		ImportedActivities: activities,
		WorkoutMatches:     matches,
		WorkoutResults:     results,
		AdaptationEvents:   events,
	}
}

func completeImportedActivityRequestToApp(req *rpcv1.CompleteImportedActivityRequest) (app.CompleteImportedActivityRequest, error) {
	if req == nil {
		return app.CompleteImportedActivityRequest{}, fmt.Errorf("request is required")
	}

	profile, err := athleteProfileToDomain(req.GetAthleteProfile())
	if err != nil {
		return app.CompleteImportedActivityRequest{}, fmt.Errorf("athlete profile: %w", err)
	}
	goal, err := trainingGoalToDomain(req.GetTrainingGoal())
	if err != nil {
		return app.CompleteImportedActivityRequest{}, fmt.Errorf("training goal: %w", err)
	}
	targetWeekDate, err := parseDate(req.GetTargetWeekDate(), "target week date")
	if err != nil {
		return app.CompleteImportedActivityRequest{}, err
	}
	activity, err := importedActivityToDomain(req.GetImportedActivity())
	if err != nil {
		return app.CompleteImportedActivityRequest{}, fmt.Errorf("imported activity: %w", err)
	}

	// TODO: Replace request-supplied athlete, goal, and activity data with
	// authenticated user context plus repository lookups once auth and persisted
	// read flows exist. This wide request keeps the first RPC contract testable.
	return app.CompleteImportedActivityRequest{
		AthleteProfile: profile,
		TrainingGoal:   goal,
		TargetWeekDate: targetWeekDate,
		ImportActivity: func(ctx context.Context) (domain.ImportedActivity, error) {
			return activity, nil
		},
		ResultID: req.GetResultId(),
		Outcome:  workoutOutcomeToDomain(req.GetOutcome()),
	}, nil
}

func completeImportedActivityResponseFromApp(response app.CompleteImportedActivityResponse) *rpcv1.CompleteImportedActivityResponse {
	return &rpcv1.CompleteImportedActivityResponse{
		PlanWeek:         planWeekFromDomain(response.PlanWeek),
		ImportedActivity: importedActivityFromDomain(response.ImportedActivity),
		WorkoutMatch:     workoutMatchFromDomain(response.WorkoutMatch),
		UpdatedWorkout:   plannedWorkoutFromDomain(response.UpdatedWorkout),
		WorkoutResult:    workoutResultFromDomain(response.WorkoutResult),
		AdaptationEvent:  adaptationEventFromDomain(response.AdaptationEvent),
	}
}

func athleteProfileToDomain(profile *rpcv1.AthleteProfile) (domain.AthleteProfile, error) {
	if profile == nil {
		return domain.AthleteProfile{}, fmt.Errorf("is required")
	}
	preferredRunDays := make([]time.Weekday, 0, len(profile.GetPreferredRunDays()))
	for _, day := range profile.GetPreferredRunDays() {
		preferredRunDays = append(preferredRunDays, time.Weekday(day))
	}
	return domain.AthleteProfile{
		ID:                    profile.GetId(),
		DisplayName:           profile.GetDisplayName(),
		ExperienceLevel:       experienceLevelToDomain(profile.GetExperienceLevel()),
		CurrentWeeklyDistance: domain.Distance{Meters: profile.GetCurrentWeeklyDistanceMeters()},
		PreferredRunDays:      preferredRunDays,
		Constraints:           append([]string(nil), profile.GetConstraints()...),
	}, nil
}

func trainingGoalToDomain(goal *rpcv1.TrainingGoal) (domain.TrainingGoal, error) {
	if goal == nil {
		return domain.TrainingGoal{}, fmt.Errorf("is required")
	}
	targetDate, err := parseOptionalDate(goal.GetTargetDate(), "target date")
	if err != nil {
		return domain.TrainingGoal{}, err
	}
	return domain.TrainingGoal{
		ID:             goal.GetId(),
		AthleteID:      goal.GetAthleteId(),
		Type:           goalTypeToDomain(goal.GetType()),
		TargetDate:     targetDate,
		TargetDistance: domain.Distance{Meters: goal.GetTargetDistanceMeters()},
		TargetDuration: secondsToDuration(goal.GetTargetDurationSeconds()),
		Notes:          goal.GetNotes(),
	}, nil
}

func importedActivityToDomain(activity *rpcv1.ImportedActivity) (domain.ImportedActivity, error) {
	if activity == nil {
		return domain.ImportedActivity{}, fmt.Errorf("is required")
	}
	startedAt := activity.GetStartedAt()
	if startedAt == nil {
		return domain.ImportedActivity{}, fmt.Errorf("started_at is required")
	}
	return domain.ImportedActivity{
		ID:              activity.GetId(),
		AthleteID:       activity.GetAthleteId(),
		Type:            activityTypeToDomain(activity.GetType()),
		StartedAt:       startedAt.AsTime(),
		Duration:        secondsToDuration(activity.GetDurationSeconds()),
		Distance:        domain.Distance{Meters: activity.GetDistanceMeters()},
		AveragePace:     domain.Pace{SecondsPerKilometer: int(activity.GetAveragePaceSecondsPerKilometer())},
		AverageHeartBPM: int(activity.GetAverageHeartBpm()),
	}, nil
}

func planWeekFromDomain(week domain.PlanWeek) *rpcv1.PlanWeek {
	workouts := make([]*rpcv1.PlannedWorkout, 0, len(week.Workouts))
	for _, workout := range week.Workouts {
		workouts = append(workouts, plannedWorkoutFromDomain(workout))
	}
	return &rpcv1.PlanWeek{
		Id:        week.ID,
		AthleteId: week.AthleteID,
		GoalId:    week.GoalID,
		PlanId:    week.PlanID,
		WeekIndex: int32(week.WeekIndex),
		StartsOn:  formatDate(week.StartsOn),
		Focus:     weekFocusFromDomain(week.Focus),
		Workouts:  workouts,
	}
}

func plannedWorkoutFromDomain(workout domain.PlannedWorkout) *rpcv1.PlannedWorkout {
	return &rpcv1.PlannedWorkout{
		Id:                    workout.ID,
		PlanId:                workout.PlanID,
		PlanWeekId:            workout.PlanWeekID,
		ScheduledFor:          formatDate(workout.ScheduledFor),
		Type:                  workoutTypeFromDomain(workout.Type),
		Status:                plannedWorkoutStatusFromDomain(workout.Status),
		TargetDistanceMeters:  workout.TargetDistance.Meters,
		TargetDurationSeconds: durationToSeconds(workout.TargetDuration),
		Intensity: &rpcv1.IntensityTarget{
			Kind:        string(workout.Intensity.Kind),
			Description: workout.Intensity.Description,
		},
		Notes: workout.Notes,
	}
}

func importedActivityFromDomain(activity domain.ImportedActivity) *rpcv1.ImportedActivity {
	return &rpcv1.ImportedActivity{
		Id:                             activity.ID,
		AthleteId:                      activity.AthleteID,
		Type:                           activityTypeFromDomain(activity.Type),
		StartedAt:                      timestamppb.New(activity.StartedAt),
		DurationSeconds:                durationToSeconds(activity.Duration),
		DistanceMeters:                 activity.Distance.Meters,
		AveragePaceSecondsPerKilometer: int64(activity.AveragePace.SecondsPerKilometer),
		AverageHeartBpm:                int32(activity.AverageHeartBPM),
	}
}

func workoutMatchFromDomain(match domain.WorkoutMatch) *rpcv1.WorkoutMatch {
	return &rpcv1.WorkoutMatch{
		Id:                 match.ID,
		PlannedWorkoutId:   match.PlannedWorkoutID,
		ImportedActivityId: match.ImportedActivityID,
		Status:             workoutMatchStatusFromDomain(match.Status),
		Confidence:         matchConfidenceFromDomain(match.Confidence),
		MatchedBy:          matchSourceFromDomain(match.MatchedBy),
		MatchedAt:          timestamppb.New(match.MatchedAt),
		Notes:              match.Notes,
	}
}

func workoutResultFromDomain(result domain.WorkoutResult) *rpcv1.WorkoutResult {
	return &rpcv1.WorkoutResult{
		Id:                 result.ID,
		PlannedWorkoutId:   result.PlannedWorkoutID,
		ImportedActivityId: result.ImportedActivityID,
		Outcome:            workoutOutcomeFromDomain(result.Outcome),
		CompletedAt:        timestamppb.New(result.CompletedAt),
		DistanceMeters:     result.Distance.Meters,
		DurationSeconds:    durationToSeconds(result.Duration),
		Notes:              result.Notes,
	}
}

func adaptationEventFromDomain(event *domain.AdaptationEvent) *rpcv1.AdaptationEvent {
	if event == nil {
		return nil
	}
	changes := make([]*rpcv1.PlanChange, 0, len(event.Changes))
	for _, change := range event.Changes {
		changes = append(changes, &rpcv1.PlanChange{
			PlannedWorkoutId: change.PlannedWorkoutID,
			Type:             planChangeTypeFromDomain(change.Type),
			Description:      change.Description,
		})
	}
	return &rpcv1.AdaptationEvent{
		Id:        event.ID,
		PlanId:    event.PlanID,
		AthleteId: event.AthleteID,
		Type:      adaptationTypeFromDomain(event.Type),
		Reason:    event.Reason,
		Summary:   event.Summary,
		CreatedAt: timestamppb.New(event.CreatedAt),
		Changes:   changes,
	}
}

func parseDate(value string, field string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("%s is required", field)
	}
	return parseOptionalDate(value, field)
}

func parseOptionalDate(value string, field string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must use YYYY-MM-DD: %w", field, err)
	}
	return parsed, nil
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(dateLayout)
}

func secondsToDuration(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}

func durationToSeconds(duration time.Duration) int64 {
	return int64(duration / time.Second)
}
