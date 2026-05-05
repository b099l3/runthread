package coreloop

import (
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/planning"
)

type ActivityImporter func() (domain.ImportedActivity, error)

type Service struct {
	Planner          planning.WeeklyPlanner
	Matcher          matching.Matcher
	AdaptationEngine adaptation.Engine
}

type CompleteImportedActivityInput struct {
	AthleteProfile domain.AthleteProfile
	TrainingGoal   domain.TrainingGoal
	TargetWeekDate time.Time
	ImportActivity ActivityImporter
	ResultID       string
	Outcome        domain.WorkoutOutcome
}

type CompleteImportedActivityResult struct {
	PlanWeek         domain.PlanWeek
	ImportedActivity domain.ImportedActivity
	WorkoutMatch     domain.WorkoutMatch
	UpdatedWorkout   domain.PlannedWorkout
	WorkoutResult    domain.WorkoutResult
	AdaptationEvent  *domain.AdaptationEvent
}

func NewService() Service {
	return Service{
		Planner:          planning.NewWeeklyPlanner(),
		Matcher:          matching.NewMatcher(),
		AdaptationEngine: adaptation.NewEngine(),
	}
}

func (s Service) CompleteImportedActivity(input CompleteImportedActivityInput) (CompleteImportedActivityResult, error) {
	if input.ImportActivity == nil {
		return CompleteImportedActivityResult{}, fmt.Errorf("activity importer is required")
	}

	week, err := s.Planner.GenerateWeek(input.AthleteProfile, input.TrainingGoal, input.TargetWeekDate)
	if err != nil {
		return CompleteImportedActivityResult{}, fmt.Errorf("generate plan week: %w", err)
	}

	activity, err := input.ImportActivity()
	if err != nil {
		return CompleteImportedActivityResult{}, fmt.Errorf("import activity: %w", err)
	}

	match, workout, err := s.bestMatch(week, activity)
	if err != nil {
		return CompleteImportedActivityResult{}, err
	}
	if match.Status != domain.WorkoutMatchStatusMatched {
		return CompleteImportedActivityResult{}, fmt.Errorf("activity did not confidently match a planned workout: %s", match.Status)
	}

	outcome := input.Outcome
	if outcome == "" {
		outcome = domain.WorkoutOutcomeCompletedAsPlanned
	}

	updatedWorkout, workoutResult, err := domain.MarkWorkoutCompleted(workout, domain.WorkoutCompletion{
		ResultID:           input.ResultID,
		ImportedActivityID: activity.ID,
		CompletedAt:        activity.StartedAt.Add(activity.Duration),
		Distance:           activity.Distance,
		Duration:           activity.Duration,
		Outcome:            outcome,
		Notes:              "Created by in-memory core-loop orchestration.",
	})
	if err != nil {
		return CompleteImportedActivityResult{}, fmt.Errorf("mark workout completed: %w", err)
	}

	adaptationEvent, err := s.AdaptationEngine.AdaptWorkoutResult(adaptation.WorkoutResultInput{
		AthleteID: input.AthleteProfile.ID,
		PlanWeek:  week,
		Result:    workoutResult,
	})
	if err != nil {
		return CompleteImportedActivityResult{}, fmt.Errorf("adapt workout result: %w", err)
	}

	return CompleteImportedActivityResult{
		PlanWeek:         week,
		ImportedActivity: activity,
		WorkoutMatch:     match,
		UpdatedWorkout:   updatedWorkout,
		WorkoutResult:    workoutResult,
		AdaptationEvent:  adaptationEvent,
	}, nil
}

func (s Service) bestMatch(week domain.PlanWeek, activity domain.ImportedActivity) (domain.WorkoutMatch, domain.PlannedWorkout, error) {
	var bestMatch domain.WorkoutMatch
	var bestWorkout domain.PlannedWorkout
	found := false

	for _, workout := range week.Workouts {
		match, err := s.Matcher.MatchActivity(workout, activity)
		if err != nil {
			continue
		}
		if betterMatch(match, bestMatch, !found) {
			bestMatch = match
			bestWorkout = workout
			found = true
		}
	}

	if !found {
		return domain.WorkoutMatch{}, domain.PlannedWorkout{}, fmt.Errorf("no matchable planned workouts found")
	}
	return bestMatch, bestWorkout, nil
}

func betterMatch(candidate domain.WorkoutMatch, current domain.WorkoutMatch, noCurrent bool) bool {
	if noCurrent {
		return true
	}
	return matchRank(candidate) > matchRank(current)
}

func matchRank(match domain.WorkoutMatch) int {
	switch match.Status {
	case domain.WorkoutMatchStatusMatched:
		return 300 + confidenceRank(match.Confidence)
	case domain.WorkoutMatchStatusUncertain:
		return 200 + confidenceRank(match.Confidence)
	case domain.WorkoutMatchStatusRejected:
		return 100 + confidenceRank(match.Confidence)
	default:
		return 0
	}
}

func confidenceRank(confidence domain.MatchConfidence) int {
	switch confidence {
	case domain.MatchConfidenceHigh:
		return 3
	case domain.MatchConfidenceMedium:
		return 2
	case domain.MatchConfidenceLow:
		return 1
	default:
		return 0
	}
}
