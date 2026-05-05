package planning

import (
	"fmt"
	"math"
	"time"

	"github.com/runthread/runthread/services/api/internal/domain"
)

const (
	mileInMeters       = 1609.344
	defaultPlanID      = "generated-plan"
	defaultPlanWeekID  = "generated-week"
	defaultEasyPace    = 6 * time.Minute
	defaultWorkoutPace = 5 * time.Minute
)

type WeeklyPlanner struct{}

func NewWeeklyPlanner() WeeklyPlanner {
	return WeeklyPlanner{}
}

func GenerateWeek(profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time) (domain.PlanWeek, error) {
	return NewWeeklyPlanner().GenerateWeek(profile, goal, targetWeekDate)
}

func (p WeeklyPlanner) GenerateWeek(profile domain.AthleteProfile, goal domain.TrainingGoal, targetWeekDate time.Time) (domain.PlanWeek, error) {
	if err := profile.Validate(); err != nil {
		return domain.PlanWeek{}, fmt.Errorf("invalid athlete profile: %w", err)
	}
	if err := goal.Validate(); err != nil {
		return domain.PlanWeek{}, fmt.Errorf("invalid training goal: %w", err)
	}
	if targetWeekDate.IsZero() {
		return domain.PlanWeek{}, fmt.Errorf("target week date is required")
	}

	startsOn := startOfWeek(targetWeekDate)
	longRunDay := chooseLongRunDay(profile)
	totalVolume := weeklyVolume(profile, goal)

	workouts := make([]domain.PlannedWorkout, 0, 7)
	for dayOffset := 0; dayOffset < 7; dayOffset++ {
		day := startsOn.AddDate(0, 0, dayOffset)
		workouts = append(workouts, plannedRest(dayOffset, day))
	}

	// Assumption: beginners need consistency before intensity, so they get easy running plus a long run.
	// Intermediate runners can handle one controlled threshold session if it is separated from the long run.
	if profile.ExperienceLevel == domain.ExperienceLevelIntermediate {
		thresholdDay := chooseThresholdDay(longRunDay)
		workouts[thresholdDay] = plannedThreshold(thresholdDay, startsOn.AddDate(0, 0, thresholdDay), totalVolume*0.20)
	}

	longRunDistance := totalVolume * longRunShare(profile)
	workouts[longRunDay] = plannedLongRun(longRunDay, startsOn.AddDate(0, 0, longRunDay), longRunDistance)

	remainingVolume := totalVolume - runningVolume(workouts)
	easyDays := chooseEasyDays(profile, longRunDay, workouts)
	distributeEasyRuns(workouts, easyDays, remainingVolume)

	if includeStrength(profile) {
		strengthDay := chooseStrengthDay(longRunDay, workouts)
		workouts[strengthDay] = plannedStrength(strengthDay, startsOn.AddDate(0, 0, strengthDay))
	}

	week := domain.PlanWeek{
		ID:        defaultPlanWeekID,
		AthleteID: profile.ID,
		GoalID:    goal.ID,
		PlanID:    defaultPlanID,
		WeekIndex: 1,
		StartsOn:  startsOn,
		Focus:     domain.WeekFocusBase,
		Workouts:  workouts,
	}
	if err := week.Validate(); err != nil {
		return domain.PlanWeek{}, fmt.Errorf("generated invalid plan week: %w", err)
	}
	return week, nil
}

func startOfWeek(date time.Time) time.Time {
	year, month, day := date.Date()
	location := date.Location()
	normalized := time.Date(year, month, day, 0, 0, 0, 0, location)
	offset := (int(normalized.Weekday()) + 6) % 7
	return normalized.AddDate(0, 0, -offset)
}

func chooseLongRunDay(profile domain.AthleteProfile) int {
	for _, day := range profile.PreferredRunDays {
		if day == time.Saturday {
			return weekdayIndex(day)
		}
	}
	for _, day := range profile.PreferredRunDays {
		if day == time.Sunday {
			return weekdayIndex(day)
		}
	}
	return weekdayIndex(time.Sunday)
}

func weeklyVolume(profile domain.AthleteProfile, goal domain.TrainingGoal) float64 {
	current := profile.CurrentWeeklyDistance.Meters
	if current <= 0 {
		current = 12 * mileInMeters
	}

	goalFloor := 0.0
	switch {
	case goal.TargetDistance.Meters >= 42195:
		goalFloor = 24 * mileInMeters
	case goal.TargetDistance.Meters >= 21097:
		goalFloor = 18 * mileInMeters
	case goal.TargetDistance.Meters >= 10000:
		goalFloor = 14 * mileInMeters
	case goal.Type == domain.GoalTypeGeneralFitness:
		goalFloor = 10 * mileInMeters
	}

	// Assumption: early plans should be conservative. A generated week can nudge toward
	// the goal floor, but it should not jump more than 5% over current training volume.
	target := math.Max(current, goalFloor)
	cap := current * 1.05
	if current < 10*mileInMeters {
		cap = current + 0.5*mileInMeters
	}
	return roundToNearest100(math.Min(target, cap))
}

func longRunShare(profile domain.AthleteProfile) float64 {
	switch profile.ExperienceLevel {
	case domain.ExperienceLevelBeginner:
		return 0.35
	case domain.ExperienceLevelIntermediate:
		return 0.32
	default:
		return 0.30
	}
}

func chooseThresholdDay(longRunDay int) int {
	for _, candidate := range []int{2, 3, 1, 4} {
		if daysApart(candidate, longRunDay) > 1 {
			return candidate
		}
	}
	return 2
}

func chooseEasyDays(profile domain.AthleteProfile, longRunDay int, workouts []domain.PlannedWorkout) []int {
	count := 2
	if profile.ExperienceLevel == domain.ExperienceLevelIntermediate {
		count = 3
	}

	preferred := make([]int, 0, len(profile.PreferredRunDays))
	for _, day := range profile.PreferredRunDays {
		index := weekdayIndex(day)
		if canPlaceEasyRun(index, longRunDay, workouts) {
			preferred = append(preferred, index)
		}
	}

	days := make([]int, 0, count)
	for _, day := range preferred {
		if !containsDay(days, day) {
			days = append(days, day)
		}
		if len(days) == count {
			return days
		}
	}

	for _, day := range []int{0, 2, 4, 3, 1, 5} {
		if !containsDay(days, day) && canPlaceEasyRun(day, longRunDay, workouts) {
			days = append(days, day)
		}
		if len(days) == count {
			return days
		}
	}
	return days
}

func canPlaceEasyRun(day int, longRunDay int, workouts []domain.PlannedWorkout) bool {
	if day == longRunDay {
		return false
	}
	return workouts[day].Type == domain.WorkoutTypeRest
}

func distributeEasyRuns(workouts []domain.PlannedWorkout, easyDays []int, remainingVolume float64) {
	if len(easyDays) == 0 || remainingVolume <= 0 {
		return
	}
	distance := roundToNearest100(remainingVolume / float64(len(easyDays)))
	for _, day := range easyDays {
		workouts[day] = plannedEasy(day, workouts[day].ScheduledFor, distance)
	}
}

func includeStrength(profile domain.AthleteProfile) bool {
	return profile.ExperienceLevel == domain.ExperienceLevelIntermediate
}

func chooseStrengthDay(longRunDay int, workouts []domain.PlannedWorkout) int {
	for _, day := range []int{1, 3, 4, 0, 5} {
		if day != longRunDay && workouts[day].Type == domain.WorkoutTypeRest && !adjacentToHardDay(day, workouts) {
			return day
		}
	}
	for day, workout := range workouts {
		if day != longRunDay && workout.Type == domain.WorkoutTypeRest {
			return day
		}
	}
	return 0
}

func plannedRest(dayOffset int, scheduledFor time.Time) domain.PlannedWorkout {
	return domain.PlannedWorkout{
		ID:           workoutID(dayOffset, domain.WorkoutTypeRest),
		PlanID:       defaultPlanID,
		PlanWeekID:   defaultPlanWeekID,
		ScheduledFor: scheduledFor,
		Type:         domain.WorkoutTypeRest,
		Status:       domain.PlannedWorkoutStatusScheduled,
		Notes:        "Rest day.",
	}
}

func plannedEasy(dayOffset int, scheduledFor time.Time, distance float64) domain.PlannedWorkout {
	return plannedRun(dayOffset, scheduledFor, domain.WorkoutTypeEasy, domain.IntensityKindEasy, distance, "Easy conversational run.")
}

func plannedLongRun(dayOffset int, scheduledFor time.Time, distance float64) domain.PlannedWorkout {
	return plannedRun(dayOffset, scheduledFor, domain.WorkoutTypeLongRun, domain.IntensityKindEasy, distance, "Long run at a relaxed effort.")
}

func plannedThreshold(dayOffset int, scheduledFor time.Time, distance float64) domain.PlannedWorkout {
	return plannedRun(dayOffset, scheduledFor, domain.WorkoutTypeWorkout, domain.IntensityKindTempo, distance, "Controlled threshold session; comfortably hard, not a race.")
}

func plannedRun(dayOffset int, scheduledFor time.Time, workoutType domain.WorkoutType, intensity domain.IntensityKind, distance float64, notes string) domain.PlannedWorkout {
	pace := defaultEasyPace
	if intensity == domain.IntensityKindTempo {
		pace = defaultWorkoutPace
	}
	duration := time.Duration(distance/defaultPaceMetersPerSecond(pace)) * time.Second
	return domain.PlannedWorkout{
		ID:             workoutID(dayOffset, workoutType),
		PlanID:         defaultPlanID,
		PlanWeekID:     defaultPlanWeekID,
		ScheduledFor:   scheduledFor,
		Type:           workoutType,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDistance: domain.Distance{Meters: roundToNearest100(distance)},
		TargetDuration: duration.Round(time.Minute),
		Intensity:      domain.IntensityTarget{Kind: intensity},
		Notes:          notes,
	}
}

func plannedStrength(dayOffset int, scheduledFor time.Time) domain.PlannedWorkout {
	return domain.PlannedWorkout{
		ID:             workoutID(dayOffset, domain.WorkoutTypeStrength),
		PlanID:         defaultPlanID,
		PlanWeekID:     defaultPlanWeekID,
		ScheduledFor:   scheduledFor,
		Type:           domain.WorkoutTypeStrength,
		Status:         domain.PlannedWorkoutStatusScheduled,
		TargetDuration: 30 * time.Minute,
		Intensity:      domain.IntensityTarget{Kind: domain.IntensityKindPerceived, Description: "Light strength and mobility."},
		Notes:          "Optional strength and mobility.",
	}
}

func workoutID(dayOffset int, workoutType domain.WorkoutType) string {
	return fmt.Sprintf("generated-%d-%s", dayOffset+1, workoutType)
}

func runningVolume(workouts []domain.PlannedWorkout) float64 {
	total := 0.0
	for _, workout := range workouts {
		if isRunWorkout(workout.Type) {
			total += workout.TargetDistance.Meters
		}
	}
	return total
}

func isRunWorkout(workoutType domain.WorkoutType) bool {
	switch workoutType {
	case domain.WorkoutTypeEasy, domain.WorkoutTypeLongRun, domain.WorkoutTypeWorkout, domain.WorkoutTypeRecovery, domain.WorkoutTypeRace:
		return true
	default:
		return false
	}
}

func isHardWorkout(workoutType domain.WorkoutType) bool {
	return workoutType == domain.WorkoutTypeWorkout || workoutType == domain.WorkoutTypeLongRun || workoutType == domain.WorkoutTypeRace
}

func adjacentToHardDay(day int, workouts []domain.PlannedWorkout) bool {
	for _, other := range []int{day - 1, day + 1} {
		if other >= 0 && other < len(workouts) && isHardWorkout(workouts[other].Type) {
			return true
		}
	}
	return false
}

func daysApart(a, b int) int {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff
}

func weekdayIndex(day time.Weekday) int {
	return (int(day) + 6) % 7
}

func containsDay(days []int, day int) bool {
	for _, candidate := range days {
		if candidate == day {
			return true
		}
	}
	return false
}

func defaultPaceMetersPerSecond(pace time.Duration) float64 {
	return 1000 / pace.Seconds()
}

func roundToNearest100(value float64) float64 {
	return math.Round(value/100) * 100
}
