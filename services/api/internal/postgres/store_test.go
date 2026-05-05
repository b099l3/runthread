package postgres

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestNewStoreRequiresDB(t *testing.T) {
	_, err := NewStore(nil)
	if err == nil {
		t.Fatal("NewStore returned nil error, want db required error")
	}
}

func TestNewStoreComposesRepositories(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://example")
	if err != nil {
		t.Fatalf("open db handle: %v", err)
	}
	defer db.Close()

	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	if store.AthleteProfiles == nil {
		t.Fatal("AthleteProfiles repository is nil")
	}
	if store.TrainingGoals == nil {
		t.Fatal("TrainingGoals repository is nil")
	}
	if store.PlanWeeks == nil {
		t.Fatal("PlanWeeks repository is nil")
	}
	if store.PlannedWorkouts == nil {
		t.Fatal("PlannedWorkouts repository is nil")
	}
	if store.ImportedActivities == nil {
		t.Fatal("ImportedActivities repository is nil")
	}
	if store.WorkoutMatches == nil {
		t.Fatal("WorkoutMatches repository is nil")
	}
	if store.WorkoutResults == nil {
		t.Fatal("WorkoutResults repository is nil")
	}
	if store.AdaptationEvents == nil {
		t.Fatal("AdaptationEvents repository is nil")
	}
}
