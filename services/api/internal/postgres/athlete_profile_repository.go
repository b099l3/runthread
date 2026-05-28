package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/runthread/runthread/services/api/internal/domain"
	postgresdb "github.com/runthread/runthread/services/api/internal/postgres/db"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type AthleteProfileRepository struct {
	queries postgresdb.Querier
}

var _ repository.AthleteProfileRepository = (*AthleteProfileRepository)(nil)

func NewAthleteProfileRepository(db *sql.DB) *AthleteProfileRepository {
	return &AthleteProfileRepository{
		queries: postgresdb.New(db),
	}
}

func (r *AthleteProfileRepository) SaveAthleteProfile(ctx context.Context, profile domain.AthleteProfile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid athlete profile: %w", err)
	}

	updateParams, err := athleteProfileToUpdateParams(profile)
	if err != nil {
		return err
	}
	if _, err := r.queries.UpdateAthleteProfile(ctx, updateParams); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("update athlete profile: %w", err)
	}

	createParams, err := athleteProfileToCreateParams(profile)
	if err != nil {
		return err
	}
	if _, err := r.queries.CreateAthleteProfile(ctx, createParams); err != nil {
		return fmt.Errorf("create athlete profile: %w", err)
	}
	return nil
}

func (r *AthleteProfileRepository) GetAthleteProfile(ctx context.Context, id string) (domain.AthleteProfile, error) {
	if err := ctx.Err(); err != nil {
		return domain.AthleteProfile{}, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.AthleteProfile{}, fmt.Errorf("parse athlete profile id: %w", err)
	}

	row, err := r.queries.GetAthleteProfile(ctx, parsedID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AthleteProfile{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.AthleteProfile{}, fmt.Errorf("get athlete profile: %w", err)
	}

	profile, err := athleteProfileFromDB(row)
	if err != nil {
		return domain.AthleteProfile{}, err
	}
	return profile, nil
}

func athleteProfileToCreateParams(profile domain.AthleteProfile) (postgresdb.CreateAthleteProfileParams, error) {
	id, err := uuid.Parse(profile.ID)
	if err != nil {
		return postgresdb.CreateAthleteProfileParams{}, fmt.Errorf("parse athlete profile id: %w", err)
	}

	return postgresdb.CreateAthleteProfileParams{
		ID:                          id,
		DisplayName:                 nullableString(profile.DisplayName),
		ExperienceLevel:             nullableString(string(profile.ExperienceLevel)),
		CurrentWeeklyDistanceMeters: profile.CurrentWeeklyDistance.Meters,
		PreferredRunDays:            weekdaysToInt64(profile.PreferredRunDays),
		Constraints:                 append([]string(nil), profile.Constraints...),
	}, nil
}

func athleteProfileToUpdateParams(profile domain.AthleteProfile) (postgresdb.UpdateAthleteProfileParams, error) {
	id, err := uuid.Parse(profile.ID)
	if err != nil {
		return postgresdb.UpdateAthleteProfileParams{}, fmt.Errorf("parse athlete profile id: %w", err)
	}

	return postgresdb.UpdateAthleteProfileParams{
		ID:                          id,
		DisplayName:                 nullableString(profile.DisplayName),
		ExperienceLevel:             nullableString(string(profile.ExperienceLevel)),
		CurrentWeeklyDistanceMeters: profile.CurrentWeeklyDistance.Meters,
		PreferredRunDays:            weekdaysToInt64(profile.PreferredRunDays),
		Constraints:                 append([]string(nil), profile.Constraints...),
	}, nil
}

func athleteProfileFromDB(row postgresdb.AthleteProfile) (domain.AthleteProfile, error) {
	return domain.AthleteProfile{
		ID:                    row.ID.String(),
		DisplayName:           stringFromNull(row.DisplayName),
		ExperienceLevel:       domain.ExperienceLevel(stringFromNull(row.ExperienceLevel)),
		CurrentWeeklyDistance: domain.Distance{Meters: row.CurrentWeeklyDistanceMeters},
		PreferredRunDays:      int64ToWeekdays(row.PreferredRunDays),
		Constraints:           append([]string(nil), row.Constraints...),
	}, nil
}

func nullableString(value string) sql.NullString {
	return sql.NullString{
		String: value,
		Valid:  value != "",
	}
}

func stringFromNull(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func weekdaysToInt64(days []time.Weekday) []int64 {
	if len(days) == 0 {
		return nil
	}

	out := make([]int64, len(days))
	for i, day := range days {
		out[i] = int64(day)
	}
	return out
}

func int64ToWeekdays(days []int64) []time.Weekday {
	if len(days) == 0 {
		return nil
	}

	out := make([]time.Weekday, len(days))
	for i, day := range days {
		out[i] = time.Weekday(day)
	}
	return out
}
