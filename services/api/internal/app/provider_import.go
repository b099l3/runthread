package app

import (
	"context"
	"fmt"

	"github.com/runthread/runthread/services/api/internal/adaptation"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/matching"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type ProviderImporter interface {
	ImportActivity(context.Context, providerimport.ImportRequest) (providerimport.ImportResult, error)
}

type ProviderImportService struct {
	Importer         ProviderImporter
	Store            repository.Store
	Matcher          matching.Matcher
	AdaptationEngine adaptation.Engine
}

type CompleteProviderImportRequest struct {
	Import           providerimport.ImportRequest
	PlanWeekID       string
	PlannedWorkoutID string
	PlanWeek         domain.PlanWeek
	PlannedWorkout   domain.PlannedWorkout
	ResultID         string
	Outcome          domain.WorkoutOutcome
}

type CompleteProviderImportResponse struct {
	ProviderImport  providerimport.ImportResult
	PlanWeek        domain.PlanWeek
	WorkoutMatch    domain.WorkoutMatch
	UpdatedWorkout  domain.PlannedWorkout
	WorkoutResult   domain.WorkoutResult
	AdaptationEvent *domain.AdaptationEvent
}

func NewProviderImportService(store repository.Store, providerStore repository.ProviderStore) (ProviderImportService, error) {
	if store == nil {
		return ProviderImportService{}, fmt.Errorf("repository store is required")
	}
	importer, err := providerimport.NewService(providerStore, store)
	if err != nil {
		return ProviderImportService{}, err
	}
	return ProviderImportService{
		Importer:         importer,
		Store:            store,
		Matcher:          matching.NewMatcher(),
		AdaptationEngine: adaptation.NewEngine(),
	}, nil
}

func (s ProviderImportService) CompleteProviderImport(ctx context.Context, request CompleteProviderImportRequest) (CompleteProviderImportResponse, error) {
	if s.Importer == nil {
		return CompleteProviderImportResponse{}, fmt.Errorf("provider importer is required")
	}
	if s.Store == nil {
		return CompleteProviderImportResponse{}, fmt.Errorf("repository store is required")
	}

	importResult, err := s.Importer.ImportActivity(ctx, request.Import)
	if err != nil {
		return CompleteProviderImportResponse{ProviderImport: importResult}, err
	}
	if importResult.ImportedActivity == nil {
		return CompleteProviderImportResponse{ProviderImport: importResult}, fmt.Errorf("provider import did not produce an imported activity")
	}

	week, err := s.planWeek(ctx, request)
	if err != nil {
		return CompleteProviderImportResponse{ProviderImport: importResult}, err
	}
	workout, match, err := s.workoutAndMatch(ctx, request, week, *importResult.ImportedActivity)
	if err != nil {
		return CompleteProviderImportResponse{ProviderImport: importResult, PlanWeek: week}, err
	}
	if match.Status != domain.WorkoutMatchStatusMatched {
		return CompleteProviderImportResponse{ProviderImport: importResult, PlanWeek: week, WorkoutMatch: match}, fmt.Errorf("provider imported activity did not confidently match a planned workout: %s", match.Status)
	}

	outcome := request.Outcome
	if outcome == "" {
		outcome = domain.WorkoutOutcomeCompletedAsPlanned
	}
	updatedWorkout, result, err := domain.MarkWorkoutCompleted(workout, domain.WorkoutCompletion{
		ResultID:           request.ResultID,
		ImportedActivityID: importResult.ImportedActivity.ID,
		CompletedAt:        importResult.ImportedActivity.StartedAt.Add(importResult.ImportedActivity.Duration),
		Distance:           importResult.ImportedActivity.Distance,
		Duration:           importResult.ImportedActivity.Duration,
		Outcome:            outcome,
		Notes:              "Created by provider import orchestration.",
	})
	if err != nil {
		return CompleteProviderImportResponse{ProviderImport: importResult, PlanWeek: week, WorkoutMatch: match}, fmt.Errorf("mark workout completed: %w", err)
	}

	updatedWeek := replaceWorkout(week, updatedWorkout)
	adaptationEvent, err := s.AdaptationEngine.AdaptWorkoutResult(adaptation.WorkoutResultInput{
		AthleteID: updatedWeek.AthleteID,
		PlanWeek:  updatedWeek,
		Result:    result,
	})
	if err != nil {
		return CompleteProviderImportResponse{}, fmt.Errorf("adapt workout result: %w", err)
	}

	if err := s.persistProviderCompletion(ctx, updatedWeek, updatedWorkout, match, result, adaptationEvent); err != nil {
		return CompleteProviderImportResponse{}, err
	}

	return CompleteProviderImportResponse{
		ProviderImport:  importResult,
		PlanWeek:        updatedWeek,
		WorkoutMatch:    match,
		UpdatedWorkout:  updatedWorkout,
		WorkoutResult:   result,
		AdaptationEvent: adaptationEvent,
	}, nil
}

func (s ProviderImportService) planWeek(ctx context.Context, request CompleteProviderImportRequest) (domain.PlanWeek, error) {
	if request.PlanWeek.ID != "" {
		if err := request.PlanWeek.Validate(); err != nil {
			return domain.PlanWeek{}, fmt.Errorf("invalid plan week: %w", err)
		}
		return request.PlanWeek, nil
	}
	if request.PlanWeekID == "" {
		return domain.PlanWeek{}, fmt.Errorf("plan week or plan week id is required")
	}
	week, err := s.Store.GetPlanWeek(ctx, request.PlanWeekID)
	if err != nil {
		return domain.PlanWeek{}, fmt.Errorf("get plan week: %w", err)
	}
	return week, nil
}

func (s ProviderImportService) workoutAndMatch(ctx context.Context, request CompleteProviderImportRequest, week domain.PlanWeek, activity domain.ImportedActivity) (domain.PlannedWorkout, domain.WorkoutMatch, error) {
	if request.PlannedWorkout.ID != "" {
		match, err := s.Matcher.MatchActivity(request.PlannedWorkout, activity)
		return request.PlannedWorkout, match, err
	}
	if request.PlannedWorkoutID != "" {
		workout, err := plannedWorkoutFromWeekOrStore(ctx, s.Store, week, request.PlannedWorkoutID)
		if err != nil {
			return domain.PlannedWorkout{}, domain.WorkoutMatch{}, err
		}
		match, err := s.Matcher.MatchActivity(workout, activity)
		return workout, match, err
	}
	return bestProviderImportMatch(s.Matcher, week, activity)
}

func plannedWorkoutFromWeekOrStore(ctx context.Context, store repository.Store, week domain.PlanWeek, id string) (domain.PlannedWorkout, error) {
	for _, workout := range week.Workouts {
		if workout.ID == id {
			return workout, nil
		}
	}
	workout, err := store.GetPlannedWorkout(ctx, id)
	if err != nil {
		return domain.PlannedWorkout{}, fmt.Errorf("get planned workout: %w", err)
	}
	return workout, nil
}

func bestProviderImportMatch(matcher matching.Matcher, week domain.PlanWeek, activity domain.ImportedActivity) (domain.PlannedWorkout, domain.WorkoutMatch, error) {
	var bestWorkout domain.PlannedWorkout
	var bestMatch domain.WorkoutMatch
	found := false
	for _, workout := range week.Workouts {
		match, err := matcher.MatchActivity(workout, activity)
		if err != nil {
			continue
		}
		if !found || providerMatchRank(match) > providerMatchRank(bestMatch) {
			bestWorkout = workout
			bestMatch = match
			found = true
		}
	}
	if !found {
		return domain.PlannedWorkout{}, domain.WorkoutMatch{}, fmt.Errorf("no matchable planned workouts found")
	}
	return bestWorkout, bestMatch, nil
}

func providerMatchRank(match domain.WorkoutMatch) int {
	switch match.Status {
	case domain.WorkoutMatchStatusMatched:
		return 300 + providerConfidenceRank(match.Confidence)
	case domain.WorkoutMatchStatusUncertain:
		return 200 + providerConfidenceRank(match.Confidence)
	case domain.WorkoutMatchStatusRejected:
		return 100 + providerConfidenceRank(match.Confidence)
	default:
		return 0
	}
}

func providerConfidenceRank(confidence domain.MatchConfidence) int {
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

func (s ProviderImportService) persistProviderCompletion(ctx context.Context, week domain.PlanWeek, workout domain.PlannedWorkout, match domain.WorkoutMatch, result domain.WorkoutResult, adaptationEvent *domain.AdaptationEvent) error {
	if err := s.Store.SaveWorkoutMatch(ctx, match); err != nil {
		return fmt.Errorf("save workout match: %w", err)
	}
	if err := s.Store.SaveWorkoutResult(ctx, result); err != nil {
		return fmt.Errorf("save workout result: %w", err)
	}
	if err := s.Store.SavePlannedWorkout(ctx, workout); err != nil {
		return fmt.Errorf("save planned workout: %w", err)
	}
	if err := s.Store.SavePlanWeek(ctx, week); err != nil {
		return fmt.Errorf("save plan week: %w", err)
	}
	if adaptationEvent != nil {
		if err := s.Store.SaveAdaptationEvent(ctx, *adaptationEvent); err != nil {
			return fmt.Errorf("save adaptation event: %w", err)
		}
	}
	return nil
}
