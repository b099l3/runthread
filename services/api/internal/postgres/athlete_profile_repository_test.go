package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
)

func TestAthleteProfileToCreateParams(t *testing.T) {
	id := uuid.NewString()
	profile := domain.AthleteProfile{
		ID:                    id,
		DisplayName:           "Maya",
		ExperienceLevel:       domain.ExperienceLevelIntermediate,
		CurrentWeeklyDistance: domain.Distance{Meters: 42000},
		PreferredRunDays:      []time.Weekday{time.Monday, time.Wednesday, time.Sunday},
		Constraints:           []string{"no Tuesdays"},
	}

	params, err := athleteProfileToCreateParams(profile)
	if err != nil {
		t.Fatalf("athleteProfileToCreateParams returned error: %v", err)
	}

	if params.ID.String() != id {
		t.Fatalf("ID = %q, want %q", params.ID.String(), id)
	}
	if !params.DisplayName.Valid || params.DisplayName.String != "Maya" {
		t.Fatalf("DisplayName = %#v, want valid Maya", params.DisplayName)
	}
	if !params.ExperienceLevel.Valid || params.ExperienceLevel.String != string(domain.ExperienceLevelIntermediate) {
		t.Fatalf("ExperienceLevel = %#v, want valid intermediate", params.ExperienceLevel)
	}
	if params.CurrentWeeklyDistanceMeters != 42000 {
		t.Fatalf("CurrentWeeklyDistanceMeters = %v, want 42000", params.CurrentWeeklyDistanceMeters)
	}
	assertInt16s(t, params.PreferredRunDays, []int16{1, 3, 0})
	assertStrings(t, params.Constraints, []string{"no Tuesdays"})
}

func TestAthleteProfileToCreateParamsUsesNullsForEmptyOptionalStrings(t *testing.T) {
	profile := domain.AthleteProfile{
		ID: uuid.NewString(),
	}

	params, err := athleteProfileToCreateParams(profile)
	if err != nil {
		t.Fatalf("athleteProfileToCreateParams returned error: %v", err)
	}

	if params.DisplayName.Valid {
		t.Fatalf("DisplayName.Valid = true, want false")
	}
	if params.ExperienceLevel.Valid {
		t.Fatalf("ExperienceLevel.Valid = true, want false")
	}
}

func TestAthleteProfileToUpdateParamsRejectsInvalidID(t *testing.T) {
	_, err := athleteProfileToUpdateParams(domain.AthleteProfile{
		ID:                    "not-a-uuid",
		CurrentWeeklyDistance: domain.Distance{Meters: 1},
	})
	if err == nil {
		t.Fatal("athleteProfileToUpdateParams returned nil error, want invalid UUID error")
	}
}

func TestAthleteProfileFromDB(t *testing.T) {
	id := uuid.New()
	row := postgresdb.AthleteProfile{
		ID:                          id,
		DisplayName:                 sql.NullString{String: "Rin", Valid: true},
		ExperienceLevel:             sql.NullString{String: string(domain.ExperienceLevelBeginner), Valid: true},
		CurrentWeeklyDistanceMeters: 18000,
		PreferredRunDays:            []int16{2, 4, 6},
		Constraints:                 []string{"keep long run short"},
	}

	profile, err := athleteProfileFromDB(row)
	if err != nil {
		t.Fatalf("athleteProfileFromDB returned error: %v", err)
	}

	if profile.ID != id.String() {
		t.Fatalf("ID = %q, want %q", profile.ID, id.String())
	}
	if profile.DisplayName != "Rin" {
		t.Fatalf("DisplayName = %q, want Rin", profile.DisplayName)
	}
	if profile.ExperienceLevel != domain.ExperienceLevelBeginner {
		t.Fatalf("ExperienceLevel = %q, want beginner", profile.ExperienceLevel)
	}
	if profile.CurrentWeeklyDistance.Meters != 18000 {
		t.Fatalf("CurrentWeeklyDistance.Meters = %v, want 18000", profile.CurrentWeeklyDistance.Meters)
	}
	assertWeekdays(t, profile.PreferredRunDays, []time.Weekday{time.Tuesday, time.Thursday, time.Saturday})
	assertStrings(t, profile.Constraints, []string{"keep long run short"})
}

func TestAthleteProfileFromDBUsesEmptyStringsForNulls(t *testing.T) {
	profile, err := athleteProfileFromDB(postgresdb.AthleteProfile{
		ID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("athleteProfileFromDB returned error: %v", err)
	}

	if profile.DisplayName != "" {
		t.Fatalf("DisplayName = %q, want empty", profile.DisplayName)
	}
	if profile.ExperienceLevel != "" {
		t.Fatalf("ExperienceLevel = %q, want empty", profile.ExperienceLevel)
	}
}

func assertInt16s(t *testing.T, got, want []int16) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(%v) = %d, want %d", got, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func assertWeekdays(t *testing.T, got, want []time.Weekday) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(%v) = %d, want %d", got, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(%v) = %d, want %d", got, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
