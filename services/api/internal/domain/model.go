package domain

import (
	"errors"
	"fmt"
	"time"
)

type AthleteProfile struct {
	ID                    string
	DisplayName           string
	ExperienceLevel       ExperienceLevel
	CurrentWeeklyDistance Distance
	PreferredRunDays      []time.Weekday
	Constraints           []string
}

func (p AthleteProfile) Validate() error {
	if p.ID == "" {
		return errors.New("athlete profile id is required")
	}
	if p.ExperienceLevel != "" && !p.ExperienceLevel.Valid() {
		return fmt.Errorf("invalid experience level %q", p.ExperienceLevel)
	}
	if p.CurrentWeeklyDistance.Meters < 0 {
		return errors.New("current weekly distance cannot be negative")
	}
	return nil
}

type ExperienceLevel string

const (
	ExperienceLevelBeginner     ExperienceLevel = "beginner"
	ExperienceLevelIntermediate ExperienceLevel = "intermediate"
	ExperienceLevelAdvanced     ExperienceLevel = "advanced"
)

func (l ExperienceLevel) Valid() bool {
	switch l {
	case ExperienceLevelBeginner, ExperienceLevelIntermediate, ExperienceLevelAdvanced:
		return true
	default:
		return false
	}
}

type TrainingGoal struct {
	ID             string
	AthleteID      string
	Type           GoalType
	TargetDate     time.Time
	TargetDistance Distance
	TargetDuration time.Duration
	Notes          string
}

func (g TrainingGoal) Validate() error {
	if g.ID == "" {
		return errors.New("training goal id is required")
	}
	if g.AthleteID == "" {
		return errors.New("training goal athlete id is required")
	}
	if !g.Type.Valid() {
		return fmt.Errorf("invalid goal type %q", g.Type)
	}
	if g.TargetDistance.Meters < 0 {
		return errors.New("target distance cannot be negative")
	}
	if g.TargetDuration < 0 {
		return errors.New("target duration cannot be negative")
	}
	return nil
}

type GoalType string

const (
	GoalTypeGeneralFitness GoalType = "general_fitness"
	GoalTypeRace           GoalType = "race"
	GoalTypeDistance       GoalType = "distance"
	GoalTypeTime           GoalType = "time"
)

func (t GoalType) Valid() bool {
	switch t {
	case GoalTypeGeneralFitness, GoalTypeRace, GoalTypeDistance, GoalTypeTime:
		return true
	default:
		return false
	}
}

type TrainingPlan struct {
	ID        string
	AthleteID string
	GoalID    string
	Status    PlanStatus
	StartsOn  time.Time
	EndsOn    time.Time
	Weeks     []PlanWeek
}

func (p TrainingPlan) Validate() error {
	if p.ID == "" {
		return errors.New("training plan id is required")
	}
	if p.AthleteID == "" {
		return errors.New("training plan athlete id is required")
	}
	if p.GoalID == "" {
		return errors.New("training plan goal id is required")
	}
	if !p.Status.Valid() {
		return fmt.Errorf("invalid plan status %q", p.Status)
	}
	if p.StartsOn.IsZero() {
		return errors.New("training plan start date is required")
	}
	if p.EndsOn.IsZero() {
		return errors.New("training plan end date is required")
	}
	if p.EndsOn.Before(p.StartsOn) {
		return errors.New("training plan end date cannot be before start date")
	}
	for i, week := range p.Weeks {
		if err := week.Validate(); err != nil {
			return fmt.Errorf("invalid plan week %d: %w", i, err)
		}
	}
	return nil
}

type PlanStatus string

const (
	PlanStatusDraft     PlanStatus = "draft"
	PlanStatusActive    PlanStatus = "active"
	PlanStatusCompleted PlanStatus = "completed"
	PlanStatusArchived  PlanStatus = "archived"
)

func (s PlanStatus) Valid() bool {
	switch s {
	case PlanStatusDraft, PlanStatusActive, PlanStatusCompleted, PlanStatusArchived:
		return true
	default:
		return false
	}
}

type PlanWeek struct {
	ID        string
	AthleteID string
	GoalID    string
	PlanID    string
	WeekIndex int
	StartsOn  time.Time
	Focus     WeekFocus
	Workouts  []PlannedWorkout
}

func (w PlanWeek) Validate() error {
	if w.ID == "" {
		return errors.New("plan week id is required")
	}
	if w.AthleteID == "" {
		return errors.New("plan week athlete id is required")
	}
	if w.PlanID == "" {
		return errors.New("plan week plan id is required")
	}
	if w.WeekIndex < 1 {
		return errors.New("plan week index must be at least 1")
	}
	if w.StartsOn.IsZero() {
		return errors.New("plan week start date is required")
	}
	if w.Focus != "" && !w.Focus.Valid() {
		return fmt.Errorf("invalid week focus %q", w.Focus)
	}
	for i, workout := range w.Workouts {
		if err := workout.Validate(); err != nil {
			return fmt.Errorf("invalid planned workout %d: %w", i, err)
		}
	}
	return nil
}

type WeekFocus string

const (
	WeekFocusBase     WeekFocus = "base"
	WeekFocusBuild    WeekFocus = "build"
	WeekFocusRecovery WeekFocus = "recovery"
	WeekFocusPeak     WeekFocus = "peak"
	WeekFocusTaper    WeekFocus = "taper"
)

func (f WeekFocus) Valid() bool {
	switch f {
	case WeekFocusBase, WeekFocusBuild, WeekFocusRecovery, WeekFocusPeak, WeekFocusTaper:
		return true
	default:
		return false
	}
}

type PlannedWorkout struct {
	ID             string
	PlanID         string
	PlanWeekID     string
	ScheduledFor   time.Time
	Type           WorkoutType
	Status         PlannedWorkoutStatus
	TargetDistance Distance
	TargetDuration time.Duration
	Intensity      IntensityTarget
	Notes          string
}

func (w PlannedWorkout) Validate() error {
	if w.ID == "" {
		return errors.New("planned workout id is required")
	}
	if w.PlanID == "" {
		return errors.New("planned workout plan id is required")
	}
	if w.PlanWeekID == "" {
		return errors.New("planned workout week id is required")
	}
	if w.ScheduledFor.IsZero() {
		return errors.New("planned workout scheduled date is required")
	}
	if !w.Type.Valid() {
		return fmt.Errorf("invalid workout type %q", w.Type)
	}
	if !w.Status.Valid() {
		return fmt.Errorf("invalid planned workout status %q", w.Status)
	}
	if w.TargetDistance.Meters < 0 {
		return errors.New("target distance cannot be negative")
	}
	if w.TargetDuration < 0 {
		return errors.New("target duration cannot be negative")
	}
	if w.requiresTrainingTarget() && w.TargetDistance.Meters == 0 && w.TargetDuration == 0 {
		return errors.New("planned workout needs target distance or duration")
	}
	if err := w.Intensity.Validate(); err != nil {
		return fmt.Errorf("invalid intensity target: %w", err)
	}
	return nil
}

func (w PlannedWorkout) requiresTrainingTarget() bool {
	switch w.Type {
	case WorkoutTypeRest, WorkoutTypeStrength:
		return false
	default:
		return true
	}
}

type WorkoutType string

const (
	WorkoutTypeEasy     WorkoutType = "easy"
	WorkoutTypeLongRun  WorkoutType = "long_run"
	WorkoutTypeWorkout  WorkoutType = "workout"
	WorkoutTypeRecovery WorkoutType = "recovery"
	WorkoutTypeRace     WorkoutType = "race"
	WorkoutTypeRest     WorkoutType = "rest"
	WorkoutTypeStrength WorkoutType = "strength"
	WorkoutTypeRide     WorkoutType = "ride"
)

func (t WorkoutType) Valid() bool {
	switch t {
	case WorkoutTypeEasy, WorkoutTypeLongRun, WorkoutTypeWorkout, WorkoutTypeRecovery, WorkoutTypeRace, WorkoutTypeRest, WorkoutTypeStrength, WorkoutTypeRide:
		return true
	default:
		return false
	}
}

type PlannedWorkoutStatus string

const (
	PlannedWorkoutStatusScheduled PlannedWorkoutStatus = "scheduled"
	PlannedWorkoutStatusCompleted PlannedWorkoutStatus = "completed"
	PlannedWorkoutStatusMissed    PlannedWorkoutStatus = "missed"
	PlannedWorkoutStatusSkipped   PlannedWorkoutStatus = "skipped"
	PlannedWorkoutStatusMoved     PlannedWorkoutStatus = "moved"
)

func (s PlannedWorkoutStatus) Valid() bool {
	switch s {
	case PlannedWorkoutStatusScheduled, PlannedWorkoutStatusCompleted, PlannedWorkoutStatusMissed, PlannedWorkoutStatusSkipped, PlannedWorkoutStatusMoved:
		return true
	default:
		return false
	}
}

type ImportedActivity struct {
	ID              string
	AthleteID       string
	Type            ActivityType
	StartedAt       time.Time
	Duration        time.Duration
	Distance        Distance
	AveragePace     Pace
	AverageHeartBPM int
}

func (a ImportedActivity) Validate() error {
	if a.ID == "" {
		return errors.New("imported activity id is required")
	}
	if a.AthleteID == "" {
		return errors.New("imported activity athlete id is required")
	}
	if !a.Type.Valid() {
		return fmt.Errorf("invalid activity type %q", a.Type)
	}
	if a.StartedAt.IsZero() {
		return errors.New("imported activity start time is required")
	}
	if a.Duration <= 0 {
		return errors.New("imported activity duration must be positive")
	}
	if a.Distance.Meters < 0 {
		return errors.New("imported activity distance cannot be negative")
	}
	if a.AveragePace.SecondsPerKilometer < 0 {
		return errors.New("average pace cannot be negative")
	}
	if a.AverageHeartBPM < 0 {
		return errors.New("average heart bpm cannot be negative")
	}
	return nil
}

type ActivityType string

const (
	ActivityTypeRun       ActivityType = "run"
	ActivityTypeTrailRun  ActivityType = "trail_run"
	ActivityTypeTreadmill ActivityType = "treadmill"
	ActivityTypeWalk      ActivityType = "walk"
	ActivityTypeOther     ActivityType = "other"
	ActivityTypeRide      ActivityType = "ride"
)

func (t ActivityType) Valid() bool {
	switch t {
	case ActivityTypeRun, ActivityTypeTrailRun, ActivityTypeTreadmill, ActivityTypeWalk, ActivityTypeOther, ActivityTypeRide:
		return true
	default:
		return false
	}
}

type WorkoutMatch struct {
	ID                 string
	PlannedWorkoutID   string
	ImportedActivityID string
	Status             WorkoutMatchStatus
	Confidence         MatchConfidence
	MatchedBy          MatchSource
	MatchedAt          time.Time
	Notes              string
}

func (m WorkoutMatch) Validate() error {
	if m.ID == "" {
		return errors.New("workout match id is required")
	}
	if m.PlannedWorkoutID == "" {
		return errors.New("workout match planned workout id is required")
	}
	if m.ImportedActivityID == "" {
		return errors.New("workout match imported activity id is required")
	}
	if !m.Status.Valid() {
		return fmt.Errorf("invalid workout match status %q", m.Status)
	}
	if !m.Confidence.Valid() {
		return fmt.Errorf("invalid match confidence %q", m.Confidence)
	}
	if !m.MatchedBy.Valid() {
		return fmt.Errorf("invalid match source %q", m.MatchedBy)
	}
	if m.MatchedAt.IsZero() {
		return errors.New("workout match time is required")
	}
	return nil
}

type WorkoutMatchStatus string

const (
	WorkoutMatchStatusMatched   WorkoutMatchStatus = "matched"
	WorkoutMatchStatusUncertain WorkoutMatchStatus = "uncertain"
	WorkoutMatchStatusRejected  WorkoutMatchStatus = "rejected"
)

func (s WorkoutMatchStatus) Valid() bool {
	switch s {
	case WorkoutMatchStatusMatched, WorkoutMatchStatusUncertain, WorkoutMatchStatusRejected:
		return true
	default:
		return false
	}
}

type MatchConfidence string

const (
	MatchConfidenceHigh   MatchConfidence = "high"
	MatchConfidenceMedium MatchConfidence = "medium"
	MatchConfidenceLow    MatchConfidence = "low"
)

func (c MatchConfidence) Valid() bool {
	switch c {
	case MatchConfidenceHigh, MatchConfidenceMedium, MatchConfidenceLow:
		return true
	default:
		return false
	}
}

type MatchSource string

const (
	MatchSourceAutomatic MatchSource = "automatic"
	MatchSourceManual    MatchSource = "manual"
)

func (s MatchSource) Valid() bool {
	switch s {
	case MatchSourceAutomatic, MatchSourceManual:
		return true
	default:
		return false
	}
}

type WorkoutResult struct {
	ID                 string
	PlannedWorkoutID   string
	ImportedActivityID string
	Outcome            WorkoutOutcome
	CompletedAt        time.Time
	Distance           Distance
	Duration           time.Duration
	Notes              string
}

func (r WorkoutResult) Validate() error {
	if r.ID == "" {
		return errors.New("workout result id is required")
	}
	if r.PlannedWorkoutID == "" {
		return errors.New("workout result planned workout id is required")
	}
	if !r.Outcome.Valid() {
		return fmt.Errorf("invalid workout outcome %q", r.Outcome)
	}
	if r.Distance.Meters < 0 {
		return errors.New("workout result distance cannot be negative")
	}
	if r.Duration < 0 {
		return errors.New("workout result duration cannot be negative")
	}
	if r.CompletedAt.IsZero() && r.Outcome.requiresCompletionTime() {
		return errors.New("workout result completion time is required")
	}
	if r.ImportedActivityID == "" && r.Outcome.requiresImportedActivity() {
		return errors.New("workout result imported activity id is required")
	}
	return nil
}

type WorkoutOutcome string

const (
	WorkoutOutcomeCompletedAsPlanned WorkoutOutcome = "completed_as_planned"
	WorkoutOutcomePartiallyCompleted WorkoutOutcome = "partially_completed"
	WorkoutOutcomeMissed             WorkoutOutcome = "missed"
	WorkoutOutcomeSkipped            WorkoutOutcome = "skipped"
	WorkoutOutcomeMoved              WorkoutOutcome = "moved"
	WorkoutOutcomeOverperformed      WorkoutOutcome = "overperformed"
	WorkoutOutcomeUnderperformed     WorkoutOutcome = "underperformed"
)

func (o WorkoutOutcome) Valid() bool {
	switch o {
	case WorkoutOutcomeCompletedAsPlanned, WorkoutOutcomePartiallyCompleted, WorkoutOutcomeMissed, WorkoutOutcomeSkipped, WorkoutOutcomeMoved, WorkoutOutcomeOverperformed, WorkoutOutcomeUnderperformed:
		return true
	default:
		return false
	}
}

func (o WorkoutOutcome) requiresCompletionTime() bool {
	switch o {
	case WorkoutOutcomeCompletedAsPlanned, WorkoutOutcomePartiallyCompleted, WorkoutOutcomeMoved, WorkoutOutcomeOverperformed, WorkoutOutcomeUnderperformed:
		return true
	default:
		return false
	}
}

func (o WorkoutOutcome) requiresImportedActivity() bool {
	switch o {
	case WorkoutOutcomeCompletedAsPlanned, WorkoutOutcomePartiallyCompleted, WorkoutOutcomeOverperformed, WorkoutOutcomeUnderperformed:
		return true
	default:
		return false
	}
}

type AdaptationEvent struct {
	ID        string
	PlanID    string
	AthleteID string
	Type      AdaptationType
	Reason    string
	Summary   string
	CreatedAt time.Time
	Changes   []PlanChange
}

func (e AdaptationEvent) Validate() error {
	if e.ID == "" {
		return errors.New("adaptation event id is required")
	}
	if e.PlanID == "" {
		return errors.New("adaptation event plan id is required")
	}
	if e.AthleteID == "" {
		return errors.New("adaptation event athlete id is required")
	}
	if !e.Type.Valid() {
		return fmt.Errorf("invalid adaptation type %q", e.Type)
	}
	if e.Reason == "" {
		return errors.New("adaptation event reason is required")
	}
	if e.Summary == "" {
		return errors.New("adaptation event summary is required")
	}
	if e.CreatedAt.IsZero() {
		return errors.New("adaptation event creation time is required")
	}
	for i, change := range e.Changes {
		if err := change.Validate(); err != nil {
			return fmt.Errorf("invalid plan change %d: %w", i, err)
		}
	}
	return nil
}

type AdaptationType string

const (
	AdaptationTypeMissedWorkout     AdaptationType = "missed_workout"
	AdaptationTypePartialCompletion AdaptationType = "partial_completion"
	AdaptationTypeOverperformance   AdaptationType = "overperformance"
	AdaptationTypeUnderperformance  AdaptationType = "underperformance"
	AdaptationTypeScheduleChange    AdaptationType = "schedule_change"
)

func (t AdaptationType) Valid() bool {
	switch t {
	case AdaptationTypeMissedWorkout, AdaptationTypePartialCompletion, AdaptationTypeOverperformance, AdaptationTypeUnderperformance, AdaptationTypeScheduleChange:
		return true
	default:
		return false
	}
}

type PlanChange struct {
	PlannedWorkoutID string
	Type             PlanChangeType
	Description      string
}

func (c PlanChange) Validate() error {
	if c.Type == "" && c.Description == "" && c.PlannedWorkoutID == "" {
		return errors.New("plan change cannot be empty")
	}
	if c.Type != "" && !c.Type.Valid() {
		return fmt.Errorf("invalid plan change type %q", c.Type)
	}
	return nil
}

type PlanChangeType string

const (
	PlanChangeTypeWorkoutMoved    PlanChangeType = "workout_moved"
	PlanChangeTypeWorkoutRemoved  PlanChangeType = "workout_removed"
	PlanChangeTypeWorkoutAdded    PlanChangeType = "workout_added"
	PlanChangeTypeWorkoutAdjusted PlanChangeType = "workout_adjusted"
	PlanChangeTypePlanNoteAdded   PlanChangeType = "plan_note_added"
)

func (t PlanChangeType) Valid() bool {
	switch t {
	case PlanChangeTypeWorkoutMoved, PlanChangeTypeWorkoutRemoved, PlanChangeTypeWorkoutAdded, PlanChangeTypeWorkoutAdjusted, PlanChangeTypePlanNoteAdded:
		return true
	default:
		return false
	}
}

type Distance struct {
	Meters float64
}

type Pace struct {
	SecondsPerKilometer int
}

type IntensityTarget struct {
	Kind        IntensityKind
	Description string
}

func (t IntensityTarget) Validate() error {
	if t.Kind == "" && t.Description == "" {
		return nil
	}
	if !t.Kind.Valid() {
		return fmt.Errorf("invalid intensity kind %q", t.Kind)
	}
	return nil
}

type IntensityKind string

const (
	IntensityKindEasy      IntensityKind = "easy"
	IntensityKindSteady    IntensityKind = "steady"
	IntensityKindTempo     IntensityKind = "tempo"
	IntensityKindIntervals IntensityKind = "intervals"
	IntensityKindRacePace  IntensityKind = "race_pace"
	IntensityKindHeartRate IntensityKind = "heart_rate"
	IntensityKindPerceived IntensityKind = "perceived_effort"
)

func (k IntensityKind) Valid() bool {
	switch k {
	case IntensityKindEasy, IntensityKindSteady, IntensityKindTempo, IntensityKindIntervals, IntensityKindRacePace, IntensityKindHeartRate, IntensityKindPerceived:
		return true
	default:
		return false
	}
}
