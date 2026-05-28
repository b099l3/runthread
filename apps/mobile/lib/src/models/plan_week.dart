import 'distance_unit.dart';

class CurrentPlanWeek {
  const CurrentPlanWeek({
    required this.planWeek,
    required this.importedActivities,
    required this.workoutMatches,
    required this.workoutResults,
    required this.adaptationEvents,
  });

  final PlanWeek planWeek;
  final List<ImportedActivity> importedActivities;
  final List<WorkoutMatch> workoutMatches;
  final List<WorkoutResult> workoutResults;
  final List<AdaptationEvent> adaptationEvents;

  factory CurrentPlanWeek.fromJson(Map<String, dynamic> json) {
    final planWeek = json['planWeek'];
    if (planWeek is! Map<String, dynamic>) {
      throw const FormatException('response did not include a plan week');
    }
    return CurrentPlanWeek(
      planWeek: PlanWeek.fromJson(planWeek),
      importedActivities: _listFromJson(
        json['importedActivities'],
        ImportedActivity.fromJson,
      ),
      workoutMatches: _listFromJson(
        json['workoutMatches'],
        WorkoutMatch.fromJson,
      ),
      workoutResults: _listFromJson(
        json['workoutResults'],
        WorkoutResult.fromJson,
      ),
      adaptationEvents: _listFromJson(
        json['adaptationEvents'],
        AdaptationEvent.fromJson,
      ),
    );
  }

  WorkoutCompletionState completionFor(PlannedWorkout workout) {
    final result = _firstWhereOrNull(workoutResults, (candidate) {
      return candidate.plannedWorkoutId == workout.id;
    });
    final match = _firstWhereOrNull(workoutMatches, (candidate) {
      return candidate.plannedWorkoutId == workout.id;
    });
    final activityID = result?.importedActivityId ?? match?.importedActivityId;
    final activity = activityID == null
        ? null
        : _firstWhereOrNull(importedActivities, (candidate) {
            return candidate.id == activityID;
          });
    return WorkoutCompletionState(
      importedActivity: activity,
      workoutMatch: match,
      workoutResult: result,
    );
  }
}

class AdaptationEvent {
  const AdaptationEvent({
    required this.id,
    required this.type,
    required this.reason,
    required this.summary,
    required this.changes,
  });

  final String id;
  final String type;
  final String reason;
  final String summary;
  final List<PlanChange> changes;

  factory AdaptationEvent.fromJson(Map<String, dynamic> json) {
    return AdaptationEvent(
      id: json['id'] as String? ?? '',
      type: _labelForEnum(json['type']),
      reason: json['reason'] as String? ?? '',
      summary: json['summary'] as String? ?? '',
      changes: _listFromJson(json['changes'], PlanChange.fromJson),
    );
  }
}

class PlanChange {
  const PlanChange({
    required this.plannedWorkoutId,
    required this.type,
    required this.description,
  });

  final String plannedWorkoutId;
  final String type;
  final String description;

  factory PlanChange.fromJson(Map<String, dynamic> json) {
    return PlanChange(
      plannedWorkoutId: json['plannedWorkoutId'] as String? ?? '',
      type: _labelForEnum(json['type']),
      description: json['description'] as String? ?? '',
    );
  }
}

class PlanWeek {
  const PlanWeek({
    required this.id,
    required this.startsOn,
    required this.focus,
    required this.workouts,
    this.isDemo = false,
  });

  final String id;
  final DateTime startsOn;
  final String focus;
  final List<PlannedWorkout> workouts;
  final bool isDemo;

  factory PlanWeek.fromJson(Map<String, dynamic> json) {
    final workoutsJson = json['workouts'];
    return PlanWeek(
      id: json['id'] as String? ?? '',
      startsOn: DateTime.parse(json['startsOn'] as String),
      focus: _labelForEnum(json['focus']),
      workouts: workoutsJson is List
          ? workoutsJson
                .whereType<Map<String, dynamic>>()
                .map(PlannedWorkout.fromJson)
                .toList(growable: false)
          : const [],
      isDemo: false,
    );
  }
}

class ImportedActivity {
  const ImportedActivity({
    required this.id,
    required this.type,
    required this.startedAt,
    required this.durationSeconds,
    required this.distanceMeters,
  });

  final String id;
  final String type;
  final DateTime? startedAt;
  final int durationSeconds;
  final double distanceMeters;

  factory ImportedActivity.fromJson(Map<String, dynamic> json) {
    return ImportedActivity(
      id: json['id'] as String? ?? '',
      type: _labelForEnum(json['type']),
      startedAt: _parseOptionalDateTime(json['startedAt']),
      durationSeconds: _intFromJson(json['durationSeconds']),
      distanceMeters: _doubleFromJson(json['distanceMeters']),
    );
  }

  String distanceLabel(DistanceUnit unit) {
    if (distanceMeters <= 0) {
      return '';
    }
    return distanceLabelFromMeters(distanceMeters, unit);
  }

  String get durationLabel => _durationLabel(durationSeconds);

  String get startedDateLabel {
    final value = startedAt;
    if (value == null) {
      return '';
    }
    return '${_weekdayLabel(value.weekday)}, ${_monthLabel(value.month)} ${value.day}';
  }
}

class WorkoutMatch {
  const WorkoutMatch({
    required this.id,
    required this.plannedWorkoutId,
    required this.importedActivityId,
    required this.status,
    required this.confidence,
  });

  final String id;
  final String plannedWorkoutId;
  final String importedActivityId;
  final String status;
  final String confidence;

  factory WorkoutMatch.fromJson(Map<String, dynamic> json) {
    return WorkoutMatch(
      id: json['id'] as String? ?? '',
      plannedWorkoutId: json['plannedWorkoutId'] as String? ?? '',
      importedActivityId: json['importedActivityId'] as String? ?? '',
      status: _labelForEnum(json['status']),
      confidence: _labelForEnum(json['confidence']),
    );
  }
}

class WorkoutResult {
  const WorkoutResult({
    required this.id,
    required this.plannedWorkoutId,
    required this.importedActivityId,
    required this.outcome,
    required this.distanceMeters,
    required this.durationSeconds,
  });

  final String id;
  final String plannedWorkoutId;
  final String importedActivityId;
  final String outcome;
  final double distanceMeters;
  final int durationSeconds;

  factory WorkoutResult.fromJson(Map<String, dynamic> json) {
    return WorkoutResult(
      id: json['id'] as String? ?? '',
      plannedWorkoutId: json['plannedWorkoutId'] as String? ?? '',
      importedActivityId: json['importedActivityId'] as String? ?? '',
      outcome: _labelForEnum(json['outcome']),
      distanceMeters: _doubleFromJson(json['distanceMeters']),
      durationSeconds: _intFromJson(json['durationSeconds']),
    );
  }
}

class WorkoutCompletionState {
  const WorkoutCompletionState({
    required this.importedActivity,
    required this.workoutMatch,
    required this.workoutResult,
  });

  final ImportedActivity? importedActivity;
  final WorkoutMatch? workoutMatch;
  final WorkoutResult? workoutResult;

  bool get hasActivity => importedActivity != null;
  bool get hasResult => workoutResult != null;
}

class PlannedWorkout {
  const PlannedWorkout({
    required this.id,
    required this.scheduledFor,
    required this.type,
    required this.status,
    required this.targetDistanceMeters,
    required this.targetDurationSeconds,
    required this.intensity,
    required this.notes,
  });

  final String id;
  final DateTime scheduledFor;
  final String type;
  final String status;
  final double targetDistanceMeters;
  final int targetDurationSeconds;
  final IntensityTarget intensity;
  final String notes;

  factory PlannedWorkout.fromJson(Map<String, dynamic> json) {
    return PlannedWorkout(
      id: json['id'] as String? ?? '',
      scheduledFor: DateTime.parse(json['scheduledFor'] as String),
      type: _labelForEnum(json['type']),
      status: _labelForEnum(json['status']),
      targetDistanceMeters: _doubleFromJson(json['targetDistanceMeters']),
      targetDurationSeconds: _intFromJson(json['targetDurationSeconds']),
      intensity: IntensityTarget.fromJson(
        json['intensity'] is Map<String, dynamic>
            ? json['intensity'] as Map<String, dynamic>
            : const {},
      ),
      notes: json['notes'] as String? ?? '',
    );
  }

  bool get isRun => targetDistanceMeters > 0;

  String distanceLabel(DistanceUnit unit) {
    if (!isRun) {
      return '';
    }
    return distanceLabelFromMeters(targetDistanceMeters, unit);
  }

  String get durationLabel {
    return _durationLabel(targetDurationSeconds);
  }

  String get intensityLabel {
    final values = [
      if (intensity.kind.isNotEmpty) intensity.kind,
      if (intensity.description.isNotEmpty) intensity.description,
    ];
    return values.join(' · ');
  }

  String get formattedScheduledDate {
    const months = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec',
    ];
    const weekdays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    return '${weekdays[scheduledFor.weekday - 1]}, ${months[scheduledFor.month - 1]} ${scheduledFor.day}, ${scheduledFor.year}';
  }
}

class IntensityTarget {
  const IntensityTarget({required this.kind, required this.description});

  final String kind;
  final String description;

  factory IntensityTarget.fromJson(Map<String, dynamic> json) {
    return IntensityTarget(
      kind: json['kind'] as String? ?? '',
      description: json['description'] as String? ?? '',
    );
  }
}

String _labelForEnum(Object? value) {
  final raw = value?.toString() ?? '';
  if (raw.isEmpty) {
    return '';
  }
  final token = raw.split('_').where((part) => part.isNotEmpty).last;
  return token[0].toUpperCase() + token.substring(1).toLowerCase();
}

List<T> _listFromJson<T>(
  Object? value,
  T Function(Map<String, dynamic>) parse,
) {
  if (value is! List) {
    return const [];
  }
  return value
      .whereType<Map<String, dynamic>>()
      .map(parse)
      .toList(growable: false);
}

DateTime? _parseOptionalDateTime(Object? value) {
  if (value is! String || value.isEmpty) {
    return null;
  }
  return DateTime.tryParse(value);
}

double _doubleFromJson(Object? value) {
  if (value is num) {
    return value.toDouble();
  }
  if (value is String) {
    return double.tryParse(value) ?? 0;
  }
  return 0;
}

int _intFromJson(Object? value) {
  if (value is num) {
    return value.toInt();
  }
  if (value is String) {
    return int.tryParse(value) ?? 0;
  }
  return 0;
}

String _durationLabel(int seconds) {
  if (seconds <= 0) {
    return '';
  }
  final minutes = (seconds / 60).round();
  if (minutes < 60) {
    return '$minutes min';
  }
  final hours = minutes ~/ 60;
  final remainder = minutes % 60;
  return remainder == 0 ? '${hours}h' : '${hours}h ${remainder}m';
}

String _weekdayLabel(int weekday) {
  const weekdays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  if (weekday < 1 || weekday > weekdays.length) {
    return '';
  }
  return weekdays[weekday - 1];
}

String _monthLabel(int month) {
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  if (month < 1 || month > months.length) {
    return '';
  }
  return months[month - 1];
}

T? _firstWhereOrNull<T>(Iterable<T> values, bool Function(T) test) {
  for (final value in values) {
    if (test(value)) {
      return value;
    }
  }
  return null;
}
