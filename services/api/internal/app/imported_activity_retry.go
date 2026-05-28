package app

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type ImportedActivityRetryService struct {
	Store repository.Store
	Now   func() time.Time
}

type RetryImportedActivitiesRequest struct {
	AthleteID      string
	GoalID         string
	TargetWeekDate time.Time
	Limit          int
}

type RetryImportedActivitiesResponse struct {
	Scanned                  int
	AlreadyCompleted         int
	Matched                  int
	Completed                int
	SkippedUnmatched         int
	SkippedCompletedWorkout  int
	SkippedOutsideTargetWeek int
	Failed                   int
	Failures                 []RetryImportedActivityFailure
}

type RetryImportedActivityFailure struct {
	ImportedActivityID string
	Error              string
}

func (s ImportedActivityRetryService) RetryImportedActivities(ctx context.Context, request RetryImportedActivitiesRequest) (RetryImportedActivitiesResponse, error) {
	if s.Store == nil {
		return RetryImportedActivitiesResponse{}, fmt.Errorf("repository store is required")
	}
	if request.AthleteID == "" {
		return RetryImportedActivitiesResponse{}, fmt.Errorf("athlete id is required")
	}

	activities, err := s.Store.ListImportedActivitiesByAthlete(ctx, request.AthleteID)
	if err != nil {
		return RetryImportedActivitiesResponse{}, fmt.Errorf("list imported activities: %w", err)
	}

	goalID := request.GoalID
	if goalID == "" {
		goal, err := s.Store.GetCurrentTrainingGoal(ctx, request.AthleteID)
		if err != nil {
			return RetryImportedActivitiesResponse{}, fmt.Errorf("get current training goal: %w", err)
		}
		goalID = goal.ID
	}

	currentPlan := CurrentPlanWeekService{
		Store:   s.Store,
		Planner: planning.NewWeeklyPlanner(),
	}
	matcher := ProviderActivityMatchService{Store: s.Store, Now: s.now}
	completer := ProviderActivityCompletionService{Store: s.Store, Now: s.now}

	var response RetryImportedActivitiesResponse
	for _, activity := range activities {
		if request.Limit > 0 && response.Scanned >= request.Limit {
			break
		}
		if !request.TargetWeekDate.IsZero() && !sameWeek(activity.StartedAt, request.TargetWeekDate) {
			response.SkippedOutsideTargetWeek++
			continue
		}
		response.Scanned++

		plan, err := currentPlan.GetCurrentPlanWeek(ctx, GetCurrentPlanWeekRequest{
			AthleteID:      activity.AthleteID,
			GoalID:         goalID,
			TargetWeekDate: activity.StartedAt,
		})
		if err != nil {
			response.addFailure(activity.ID, fmt.Errorf("load plan week: %w", err))
			continue
		}
		if hasWorkoutResultForImportedActivity(plan.WorkoutResults, activity.ID) {
			response.AlreadyCompleted++
			continue
		}

		match := matchedWorkoutForImportedActivity(plan.WorkoutMatches, activity.ID)
		workout := plannedWorkoutForMatch(plan.PlanWeek, match)
		if match.ID == "" || workout.ID == "" || match.MatchedBy != domain.MatchSourceManual {
			matchResponse, err := matcher.MatchProviderActivity(ctx, MatchProviderActivityRequest{
				ImportedActivity: activity,
				PlanWeek:         plan.PlanWeek,
			})
			if err != nil {
				response.addFailure(activity.ID, fmt.Errorf("match activity: %w", err))
				continue
			}
			match = matchResponse.WorkoutMatch
			workout = matchResponse.PlannedWorkout
			response.Matched++
		}
		if match.Status != domain.WorkoutMatchStatusMatched {
			response.SkippedUnmatched++
			continue
		}
		if workout.Status == domain.PlannedWorkoutStatusCompleted || hasWorkoutResultForPlannedWorkout(plan.WorkoutResults, workout.ID) {
			response.SkippedCompletedWorkout++
			continue
		}

		_, err = completer.CompleteMatchedProviderActivity(ctx, CompleteMatchedProviderActivityRequest{
			WorkoutMatch:     match,
			ImportedActivity: activity,
			PlanWeek:         plan.PlanWeek,
			PlannedWorkout:   workout,
		})
		if err != nil {
			response.addFailure(activity.ID, fmt.Errorf("complete activity: %w", err))
			continue
		}
		response.Completed++
	}

	return response, nil
}

func (r *RetryImportedActivitiesResponse) addFailure(importedActivityID string, err error) {
	r.Failed++
	r.Failures = append(r.Failures, RetryImportedActivityFailure{
		ImportedActivityID: importedActivityID,
		Error:              err.Error(),
	})
}

func hasWorkoutResultForImportedActivity(results []domain.WorkoutResult, importedActivityID string) bool {
	for _, result := range results {
		if result.ImportedActivityID == importedActivityID {
			return true
		}
	}
	return false
}

func hasWorkoutResultForPlannedWorkout(results []domain.WorkoutResult, plannedWorkoutID string) bool {
	for _, result := range results {
		if result.PlannedWorkoutID == plannedWorkoutID {
			return true
		}
	}
	return false
}

func matchedWorkoutForImportedActivity(matches []domain.WorkoutMatch, importedActivityID string) domain.WorkoutMatch {
	for _, match := range matches {
		if match.ImportedActivityID == importedActivityID && match.Status == domain.WorkoutMatchStatusMatched {
			return match
		}
	}
	return domain.WorkoutMatch{}
}

func plannedWorkoutForMatch(week domain.PlanWeek, match domain.WorkoutMatch) domain.PlannedWorkout {
	if match.ID == "" {
		return domain.PlannedWorkout{}
	}
	for _, workout := range week.Workouts {
		if workout.ID == match.PlannedWorkoutID {
			return workout
		}
	}
	return domain.PlannedWorkout{}
}

func sameWeek(left time.Time, right time.Time) bool {
	return startOfRetryWeek(left).Equal(startOfRetryWeek(right))
}

func startOfRetryWeek(value time.Time) time.Time {
	year, month, day := value.In(time.UTC).Date()
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -daysSinceMonday)
}

func (s ImportedActivityRetryService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
