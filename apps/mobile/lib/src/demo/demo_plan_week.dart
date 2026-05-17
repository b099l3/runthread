import '../models/plan_week.dart';
import '../week_dates.dart';

CurrentPlanWeek demoCurrentPlanWeek() {
  return CurrentPlanWeek(
    planWeek: demoPlanWeek(),
    importedActivities: const [],
    workoutMatches: const [],
    workoutResults: const [],
    adaptationEvents: const [],
  );
}

PlanWeek demoPlanWeek() {
  final startsOn = currentWeekStart();
  return PlanWeek(
    id: 'demo-week',
    startsOn: startsOn,
    focus: 'Base',
    isDemo: true,
    workouts: [
      _workout(
        id: 'demo-rest-monday',
        scheduledFor: startsOn,
        type: 'Rest',
        notes: 'Rest day.',
      ),
      _workout(
        id: 'demo-easy-tuesday',
        scheduledFor: startsOn.add(const Duration(days: 1)),
        type: 'Easy',
        distance: 5200,
        duration: 1900,
        intensity: 'easy',
        notes: 'Easy conversational run.',
      ),
      _workout(
        id: 'demo-rest-wednesday',
        scheduledFor: startsOn.add(const Duration(days: 2)),
        type: 'Rest',
        notes: 'Rest day.',
      ),
      _workout(
        id: 'demo-easy-thursday',
        scheduledFor: startsOn.add(const Duration(days: 3)),
        type: 'Easy',
        distance: 5200,
        duration: 1900,
        intensity: 'easy',
        notes: 'Easy conversational run.',
      ),
      _workout(
        id: 'demo-rest-friday',
        scheduledFor: startsOn.add(const Duration(days: 4)),
        type: 'Rest',
        notes: 'Rest day.',
      ),
      _workout(
        id: 'demo-strength-saturday',
        scheduledFor: startsOn.add(const Duration(days: 5)),
        type: 'Strength',
        duration: 1800,
        intensity: 'perceived',
        notes: 'Optional strength and mobility.',
      ),
      _workout(
        id: 'demo-long-sunday',
        scheduledFor: startsOn.add(const Duration(days: 6)),
        type: 'Long',
        distance: 8000,
        duration: 3000,
        intensity: 'easy',
        notes: 'Long run at a relaxed effort.',
      ),
    ],
  );
}

PlannedWorkout _workout({
  required String id,
  required DateTime scheduledFor,
  required String type,
  required String notes,
  double distance = 0,
  int duration = 0,
  String intensity = '',
}) {
  return PlannedWorkout(
    id: id,
    scheduledFor: scheduledFor,
    type: type,
    status: 'Scheduled',
    targetDistanceMeters: distance,
    targetDurationSeconds: duration,
    intensity: IntensityTarget(kind: intensity, description: ''),
    notes: notes,
  );
}
