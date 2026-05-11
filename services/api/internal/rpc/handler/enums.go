package handler

import (
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/repository"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
)

func experienceLevelToDomain(value rpcv1.ExperienceLevel) domain.ExperienceLevel {
	switch value {
	case rpcv1.ExperienceLevel_EXPERIENCE_LEVEL_BEGINNER:
		return domain.ExperienceLevelBeginner
	case rpcv1.ExperienceLevel_EXPERIENCE_LEVEL_INTERMEDIATE:
		return domain.ExperienceLevelIntermediate
	case rpcv1.ExperienceLevel_EXPERIENCE_LEVEL_ADVANCED:
		return domain.ExperienceLevelAdvanced
	default:
		return ""
	}
}

func goalTypeToDomain(value rpcv1.GoalType) domain.GoalType {
	switch value {
	case rpcv1.GoalType_GOAL_TYPE_GENERAL_FITNESS:
		return domain.GoalTypeGeneralFitness
	case rpcv1.GoalType_GOAL_TYPE_RACE:
		return domain.GoalTypeRace
	case rpcv1.GoalType_GOAL_TYPE_DISTANCE:
		return domain.GoalTypeDistance
	case rpcv1.GoalType_GOAL_TYPE_TIME:
		return domain.GoalTypeTime
	default:
		return ""
	}
}

func providerToApp(value rpcv1.Provider) string {
	switch value {
	case rpcv1.Provider_PROVIDER_GARMIN:
		return "garmin"
	default:
		return ""
	}
}

func providerFromRepository(value string) rpcv1.Provider {
	switch value {
	case "garmin":
		return rpcv1.Provider_PROVIDER_GARMIN
	default:
		return rpcv1.Provider_PROVIDER_UNSPECIFIED
	}
}

func providerConnectionStatusFromRepository(value repository.ProviderConnectionStatus) rpcv1.ProviderConnectionStatus {
	switch value {
	case repository.ProviderConnectionStatusPending:
		return rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_PENDING
	case repository.ProviderConnectionStatusConnected:
		return rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_CONNECTED
	case repository.ProviderConnectionStatusSyncing:
		return rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_SYNCING
	case repository.ProviderConnectionStatusError:
		return rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_ERROR
	case repository.ProviderConnectionStatusDisconnected:
		return rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_DISCONNECTED
	default:
		return rpcv1.ProviderConnectionStatus_PROVIDER_CONNECTION_STATUS_UNSPECIFIED
	}
}

func weekFocusFromDomain(value domain.WeekFocus) rpcv1.WeekFocus {
	switch value {
	case domain.WeekFocusBase:
		return rpcv1.WeekFocus_WEEK_FOCUS_BASE
	case domain.WeekFocusBuild:
		return rpcv1.WeekFocus_WEEK_FOCUS_BUILD
	case domain.WeekFocusRecovery:
		return rpcv1.WeekFocus_WEEK_FOCUS_RECOVERY
	case domain.WeekFocusPeak:
		return rpcv1.WeekFocus_WEEK_FOCUS_PEAK
	case domain.WeekFocusTaper:
		return rpcv1.WeekFocus_WEEK_FOCUS_TAPER
	default:
		return rpcv1.WeekFocus_WEEK_FOCUS_UNSPECIFIED
	}
}

func workoutTypeFromDomain(value domain.WorkoutType) rpcv1.WorkoutType {
	switch value {
	case domain.WorkoutTypeEasy:
		return rpcv1.WorkoutType_WORKOUT_TYPE_EASY
	case domain.WorkoutTypeLongRun:
		return rpcv1.WorkoutType_WORKOUT_TYPE_LONG_RUN
	case domain.WorkoutTypeWorkout:
		return rpcv1.WorkoutType_WORKOUT_TYPE_WORKOUT
	case domain.WorkoutTypeRecovery:
		return rpcv1.WorkoutType_WORKOUT_TYPE_RECOVERY
	case domain.WorkoutTypeRace:
		return rpcv1.WorkoutType_WORKOUT_TYPE_RACE
	case domain.WorkoutTypeRest:
		return rpcv1.WorkoutType_WORKOUT_TYPE_REST
	case domain.WorkoutTypeStrength:
		return rpcv1.WorkoutType_WORKOUT_TYPE_STRENGTH
	default:
		return rpcv1.WorkoutType_WORKOUT_TYPE_UNSPECIFIED
	}
}

func plannedWorkoutStatusFromDomain(value domain.PlannedWorkoutStatus) rpcv1.PlannedWorkoutStatus {
	switch value {
	case domain.PlannedWorkoutStatusScheduled:
		return rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_SCHEDULED
	case domain.PlannedWorkoutStatusCompleted:
		return rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_COMPLETED
	case domain.PlannedWorkoutStatusMissed:
		return rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_MISSED
	case domain.PlannedWorkoutStatusSkipped:
		return rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_SKIPPED
	case domain.PlannedWorkoutStatusMoved:
		return rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_MOVED
	default:
		return rpcv1.PlannedWorkoutStatus_PLANNED_WORKOUT_STATUS_UNSPECIFIED
	}
}

func activityTypeToDomain(value rpcv1.ActivityType) domain.ActivityType {
	switch value {
	case rpcv1.ActivityType_ACTIVITY_TYPE_RUN:
		return domain.ActivityTypeRun
	case rpcv1.ActivityType_ACTIVITY_TYPE_TRAIL_RUN:
		return domain.ActivityTypeTrailRun
	case rpcv1.ActivityType_ACTIVITY_TYPE_TREADMILL:
		return domain.ActivityTypeTreadmill
	case rpcv1.ActivityType_ACTIVITY_TYPE_WALK:
		return domain.ActivityTypeWalk
	case rpcv1.ActivityType_ACTIVITY_TYPE_OTHER:
		return domain.ActivityTypeOther
	default:
		return ""
	}
}

func activityTypeFromDomain(value domain.ActivityType) rpcv1.ActivityType {
	switch value {
	case domain.ActivityTypeRun:
		return rpcv1.ActivityType_ACTIVITY_TYPE_RUN
	case domain.ActivityTypeTrailRun:
		return rpcv1.ActivityType_ACTIVITY_TYPE_TRAIL_RUN
	case domain.ActivityTypeTreadmill:
		return rpcv1.ActivityType_ACTIVITY_TYPE_TREADMILL
	case domain.ActivityTypeWalk:
		return rpcv1.ActivityType_ACTIVITY_TYPE_WALK
	case domain.ActivityTypeOther:
		return rpcv1.ActivityType_ACTIVITY_TYPE_OTHER
	default:
		return rpcv1.ActivityType_ACTIVITY_TYPE_UNSPECIFIED
	}
}

func workoutMatchStatusFromDomain(value domain.WorkoutMatchStatus) rpcv1.WorkoutMatchStatus {
	switch value {
	case domain.WorkoutMatchStatusMatched:
		return rpcv1.WorkoutMatchStatus_WORKOUT_MATCH_STATUS_MATCHED
	case domain.WorkoutMatchStatusUncertain:
		return rpcv1.WorkoutMatchStatus_WORKOUT_MATCH_STATUS_UNCERTAIN
	case domain.WorkoutMatchStatusRejected:
		return rpcv1.WorkoutMatchStatus_WORKOUT_MATCH_STATUS_REJECTED
	default:
		return rpcv1.WorkoutMatchStatus_WORKOUT_MATCH_STATUS_UNSPECIFIED
	}
}

func matchConfidenceFromDomain(value domain.MatchConfidence) rpcv1.MatchConfidence {
	switch value {
	case domain.MatchConfidenceHigh:
		return rpcv1.MatchConfidence_MATCH_CONFIDENCE_HIGH
	case domain.MatchConfidenceMedium:
		return rpcv1.MatchConfidence_MATCH_CONFIDENCE_MEDIUM
	case domain.MatchConfidenceLow:
		return rpcv1.MatchConfidence_MATCH_CONFIDENCE_LOW
	default:
		return rpcv1.MatchConfidence_MATCH_CONFIDENCE_UNSPECIFIED
	}
}

func matchSourceFromDomain(value domain.MatchSource) rpcv1.MatchSource {
	switch value {
	case domain.MatchSourceAutomatic:
		return rpcv1.MatchSource_MATCH_SOURCE_AUTOMATIC
	case domain.MatchSourceManual:
		return rpcv1.MatchSource_MATCH_SOURCE_MANUAL
	default:
		return rpcv1.MatchSource_MATCH_SOURCE_UNSPECIFIED
	}
}

func workoutOutcomeToDomain(value rpcv1.WorkoutOutcome) domain.WorkoutOutcome {
	switch value {
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_COMPLETED_AS_PLANNED:
		return domain.WorkoutOutcomeCompletedAsPlanned
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_PARTIALLY_COMPLETED:
		return domain.WorkoutOutcomePartiallyCompleted
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_MISSED:
		return domain.WorkoutOutcomeMissed
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_SKIPPED:
		return domain.WorkoutOutcomeSkipped
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_MOVED:
		return domain.WorkoutOutcomeMoved
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_OVERPERFORMED:
		return domain.WorkoutOutcomeOverperformed
	case rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_UNDERPERFORMED:
		return domain.WorkoutOutcomeUnderperformed
	default:
		return ""
	}
}

func workoutOutcomeFromDomain(value domain.WorkoutOutcome) rpcv1.WorkoutOutcome {
	switch value {
	case domain.WorkoutOutcomeCompletedAsPlanned:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_COMPLETED_AS_PLANNED
	case domain.WorkoutOutcomePartiallyCompleted:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_PARTIALLY_COMPLETED
	case domain.WorkoutOutcomeMissed:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_MISSED
	case domain.WorkoutOutcomeSkipped:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_SKIPPED
	case domain.WorkoutOutcomeMoved:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_MOVED
	case domain.WorkoutOutcomeOverperformed:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_OVERPERFORMED
	case domain.WorkoutOutcomeUnderperformed:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_UNDERPERFORMED
	default:
		return rpcv1.WorkoutOutcome_WORKOUT_OUTCOME_UNSPECIFIED
	}
}

func adaptationTypeFromDomain(value domain.AdaptationType) rpcv1.AdaptationType {
	switch value {
	case domain.AdaptationTypeMissedWorkout:
		return rpcv1.AdaptationType_ADAPTATION_TYPE_MISSED_WORKOUT
	case domain.AdaptationTypePartialCompletion:
		return rpcv1.AdaptationType_ADAPTATION_TYPE_PARTIAL_COMPLETION
	case domain.AdaptationTypeOverperformance:
		return rpcv1.AdaptationType_ADAPTATION_TYPE_OVERPERFORMANCE
	case domain.AdaptationTypeUnderperformance:
		return rpcv1.AdaptationType_ADAPTATION_TYPE_UNDERPERFORMANCE
	case domain.AdaptationTypeScheduleChange:
		return rpcv1.AdaptationType_ADAPTATION_TYPE_SCHEDULE_CHANGE
	default:
		return rpcv1.AdaptationType_ADAPTATION_TYPE_UNSPECIFIED
	}
}

func planChangeTypeFromDomain(value domain.PlanChangeType) rpcv1.PlanChangeType {
	switch value {
	case domain.PlanChangeTypeWorkoutMoved:
		return rpcv1.PlanChangeType_PLAN_CHANGE_TYPE_WORKOUT_MOVED
	case domain.PlanChangeTypeWorkoutRemoved:
		return rpcv1.PlanChangeType_PLAN_CHANGE_TYPE_WORKOUT_REMOVED
	case domain.PlanChangeTypeWorkoutAdded:
		return rpcv1.PlanChangeType_PLAN_CHANGE_TYPE_WORKOUT_ADDED
	case domain.PlanChangeTypeWorkoutAdjusted:
		return rpcv1.PlanChangeType_PLAN_CHANGE_TYPE_WORKOUT_ADJUSTED
	case domain.PlanChangeTypePlanNoteAdded:
		return rpcv1.PlanChangeType_PLAN_CHANGE_TYPE_PLAN_NOTE_ADDED
	default:
		return rpcv1.PlanChangeType_PLAN_CHANGE_TYPE_UNSPECIFIED
	}
}
