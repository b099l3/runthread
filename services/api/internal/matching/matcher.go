package matching

import (
	"fmt"
	"math"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

type Matcher struct {
	Now func() time.Time
}

func NewMatcher() Matcher {
	return Matcher{Now: time.Now}
}

func MatchActivity(workout domain.PlannedWorkout, activity domain.ImportedActivity) (domain.WorkoutMatch, error) {
	return NewMatcher().MatchActivity(workout, activity)
}

func (m Matcher) MatchActivity(workout domain.PlannedWorkout, activity domain.ImportedActivity) (domain.WorkoutMatch, error) {
	if err := workout.Validate(); err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("invalid planned workout: %w", err)
	}
	if err := activity.Validate(); err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("invalid imported activity: %w", err)
	}

	matchedAt := m.now()
	match := domain.WorkoutMatch{
		ID:                 matchID(workout.ID, activity.ID),
		PlannedWorkoutID:   workout.ID,
		ImportedActivityID: activity.ID,
		MatchedBy:          domain.MatchSourceAutomatic,
		MatchedAt:          matchedAt,
	}

	if !compatibleTypes(workout.Type, activity.Type) {
		match.Status = domain.WorkoutMatchStatusRejected
		match.Confidence = domain.MatchConfidenceLow
		match.Notes = "Rejected because activity type does not match the planned workout."
		return validateMatch(match)
	}

	dateGap := calendarDayGap(workout.ScheduledFor, activity.StartedAt)
	if dateGap > 2 {
		match.Status = domain.WorkoutMatchStatusRejected
		match.Confidence = domain.MatchConfidenceLow
		match.Notes = "Rejected because activity date is too far from the planned workout."
		return validateMatch(match)
	}

	distanceRatio := closenessRatio(workout.TargetDistance.Meters, activity.Distance.Meters)
	durationRatio := closenessRatio(workout.TargetDuration.Seconds(), activity.Duration.Seconds())
	closeSignals := 0
	if dateGap == 0 {
		closeSignals++
	}
	if distanceRatio >= 0.85 {
		closeSignals++
	}
	if durationRatio >= 0.85 {
		closeSignals++
	}

	switch {
	case closeSignals >= 3:
		match.Status = domain.WorkoutMatchStatusMatched
		match.Confidence = domain.MatchConfidenceHigh
		match.Notes = "Matched on workout type, date, distance, and duration."
	case closeSignals >= 2 && dateGap <= 1:
		match.Status = domain.WorkoutMatchStatusUncertain
		match.Confidence = domain.MatchConfidenceMedium
		match.Notes = "Possible match; at least two signals line up, but the match is not exact."
	case dateGap <= 1 && (distanceRatio >= 0.85 || durationRatio >= 0.85):
		match.Status = domain.WorkoutMatchStatusUncertain
		match.Confidence = domain.MatchConfidenceMedium
		match.Notes = "Possible match; date is close and one effort signal lines up."
	default:
		match.Status = domain.WorkoutMatchStatusRejected
		match.Confidence = domain.MatchConfidenceLow
		match.Notes = "Rejected because too few matching signals line up."
	}

	return validateMatch(match)
}

func ManualMatch(workout domain.PlannedWorkout, activity domain.ImportedActivity, matchedAt time.Time, notes string) (domain.WorkoutMatch, error) {
	if err := workout.Validate(); err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("invalid planned workout: %w", err)
	}
	if err := activity.Validate(); err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("invalid imported activity: %w", err)
	}
	if matchedAt.IsZero() {
		return domain.WorkoutMatch{}, fmt.Errorf("manual match time is required")
	}

	match := domain.WorkoutMatch{
		ID:                 matchID(workout.ID, activity.ID),
		PlannedWorkoutID:   workout.ID,
		ImportedActivityID: activity.ID,
		Status:             domain.WorkoutMatchStatusMatched,
		Confidence:         domain.MatchConfidenceHigh,
		MatchedBy:          domain.MatchSourceManual,
		MatchedAt:          matchedAt,
		Notes:              notes,
	}
	return validateMatch(match)
}

func (m Matcher) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

func compatibleTypes(workoutType domain.WorkoutType, activityType domain.ActivityType) bool {
	switch workoutType {
	case domain.WorkoutTypeEasy, domain.WorkoutTypeLongRun, domain.WorkoutTypeWorkout, domain.WorkoutTypeRecovery:
		return activityType == domain.ActivityTypeRun || activityType == domain.ActivityTypeTrailRun || activityType == domain.ActivityTypeTreadmill
	case domain.WorkoutTypeRace:
		return activityType == domain.ActivityTypeRun || activityType == domain.ActivityTypeTrailRun
	default:
		return false
	}
}

func calendarDayGap(a, b time.Time) int {
	aDate := midnightUTC(a)
	bDate := midnightUTC(b)
	diff := aDate.Sub(bDate)
	if diff < 0 {
		diff = -diff
	}
	return int(diff.Hours() / 24)
}

func midnightUTC(value time.Time) time.Time {
	year, month, day := value.In(time.UTC).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func closenessRatio(expected, actual float64) float64 {
	if expected <= 0 || actual <= 0 {
		return 0
	}
	ratio := actual / expected
	if ratio > 1 {
		ratio = 1 / ratio
	}
	return math.Max(0, ratio)
}

func validateMatch(match domain.WorkoutMatch) (domain.WorkoutMatch, error) {
	if err := match.Validate(); err != nil {
		return domain.WorkoutMatch{}, fmt.Errorf("invalid workout match: %w", err)
	}
	return match, nil
}

func matchID(workoutID string, activityID string) string {
	return fmt.Sprintf("match-%s-%s", workoutID, activityID)
}
