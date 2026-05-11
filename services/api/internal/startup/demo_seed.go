package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
)

const (
	DemoAthleteID = "athlete-1"
	DemoGoalID    = "goal-1"
)

func SeedDemoData(ctx context.Context, store repository.Store) error {
	if store == nil {
		return fmt.Errorf("repository store is required")
	}
	profile := domain.AthleteProfile{
		ID:                    DemoAthleteID,
		DisplayName:           "Maya",
		ExperienceLevel:       domain.ExperienceLevelBeginner,
		CurrentWeeklyDistance: domain.Distance{Meters: 20000},
		PreferredRunDays:      []time.Weekday{time.Tuesday, time.Thursday, time.Sunday},
	}
	if err := store.SaveAthleteProfile(ctx, profile); err != nil {
		return fmt.Errorf("save demo athlete profile: %w", err)
	}

	goal := domain.TrainingGoal{
		ID:             DemoGoalID,
		AthleteID:      DemoAthleteID,
		Type:           domain.GoalTypeRace,
		TargetDate:     time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC),
		TargetDistance: domain.Distance{Meters: 21097.5},
		Notes:          "Local demo half marathon goal.",
	}
	if err := store.SaveTrainingGoal(ctx, goal); err != nil {
		return fmt.Errorf("save demo training goal: %w", err)
	}
	return nil
}
