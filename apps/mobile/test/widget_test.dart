import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:runthread_mobile/src/api/runthread_api.dart';
import 'package:runthread_mobile/src/app.dart';
import 'package:runthread_mobile/src/demo/demo_fallback_api.dart';
import 'package:runthread_mobile/src/models/plan_week.dart';
import 'package:runthread_mobile/src/models/provider_connection.dart';
import 'package:runthread_mobile/src/week_dates.dart';

void main() {
  test('current week date helpers use the containing Monday week', () {
    final saturday = DateTime(2026, 5, 16, 13, 45);

    expect(currentWeekStart(now: saturday), DateTime(2026, 5, 11));
    expect(currentWeekTargetDate(now: saturday), '2026-05-16');
  });

  testWidgets('weekly plan screen renders seven workouts', (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('This week'), findsOneWidget);
    expect(find.text('Easy'), findsNWidgets(2));
    expect(find.text('Rest'), findsNWidgets(3));
    expect(find.text('Long'), findsOneWidget);
    expect(find.text('Strength'), findsOneWidget);
    expect(find.textContaining('5.0 km'), findsOneWidget);
    expect(find.textContaining('8.0 km'), findsOneWidget);
    expect(find.text('Strava'), findsOneWidget);
    expect(find.text('Not connected'), findsOneWidget);
    expect(find.text('Connect Strava'), findsOneWidget);
    expect(find.text('Plan changes'), findsOneWidget);
    expect(find.text('No adaptations this week.'), findsOneWidget);
  });

  testWidgets('weekly plan screen shows loading then error', (tester) async {
    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          error: const RunthreadApiException('backend unavailable'),
        ),
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    await tester.pumpAndSettle();

    expect(find.text('Plan unavailable'), findsOneWidget);
    expect(find.text('Retry'), findsOneWidget);
    expect(find.textContaining('backend unavailable'), findsOneWidget);
  });

  testWidgets('demo fallback renders a visible demo notice', (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: DemoFallbackRunthreadApi(
          primary: FakeRunthreadApi(
            error: const RunthreadApiException('backend unavailable'),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('This week'), findsOneWidget);
    expect(find.textContaining('Demo data ·'), findsOneWidget);
    expect(find.text('Plan unavailable'), findsNothing);

    await tester.tap(find.byIcon(Icons.history_outlined));
    await tester.pumpAndSettle();

    expect(find.text('History'), findsWidgets);
    expect(find.textContaining('Demo data ·'), findsOneWidget);
  });

  testWidgets('strava connection entry point is read only', (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Strava'), findsOneWidget);
    expect(
      find.text(
        'Run completion will come from imported Strava activity once provider access is ready.',
      ),
      findsOneWidget,
    );
    expect(find.text('Disabled for now'), findsOneWidget);

    final connectButton = tester.widget<OutlinedButton>(
      find.widgetWithText(OutlinedButton, 'Connect Strava'),
    );
    expect(connectButton.onPressed, isNull);
  });

  testWidgets('strava connection entry point shows pending status', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeek(),
          providerConnection: testProviderConnectionStatus(
            ProviderConnectionStatus.pending,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Strava'), findsOneWidget);
    expect(find.text('Pending'), findsOneWidget);
    expect(
      find.text(
        'Connection has started. Strava authorization remains disabled until provider access is ready.',
      ),
      findsOneWidget,
    );
    final connectButton = tester.widget<OutlinedButton>(
      find.widgetWithText(OutlinedButton, 'Connect Strava'),
    );
    expect(connectButton.onPressed, isNull);
  });

  testWidgets('strava connection entry point shows connected status', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeek(),
          providerConnection: testProviderConnectionStatus(
            ProviderConnectionStatus.connected,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Strava'), findsOneWidget);
    expect(find.text('Connected'), findsOneWidget);
    expect(
      find.text(
        'Runthread is ready to receive imported Strava activity for workout completion.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('tapping a workout opens detail view', (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek()),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Long'));
    await tester.pumpAndSettle();

    expect(find.text('Workout'), findsOneWidget);
    expect(find.text('Sun, Jun 7, 2026'), findsOneWidget);
    expect(find.text('Status'), findsOneWidget);
    expect(find.text('Scheduled'), findsOneWidget);
    expect(find.text('Target distance'), findsOneWidget);
    expect(find.text('8.0 km'), findsOneWidget);
    expect(find.text('Target duration'), findsOneWidget);
    expect(find.text('50 min'), findsOneWidget);
    expect(find.text('Intensity'), findsOneWidget);
    expect(find.text('easy'), findsOneWidget);
    expect(find.text('Long run at a relaxed effort.'), findsOneWidget);
    expect(find.text('Mark complete'), findsOneWidget);
    expect(
      find.text(
        'Real completion will come from imported Strava activity later.',
      ),
      findsOneWidget,
    );
    final markCompleteButton = tester.widget<FilledButton>(
      find.widgetWithText(FilledButton, 'Mark complete'),
    );
    expect(markCompleteButton.onPressed, isNull);
    expect(
      find.text('No imported activity has been matched to this workout yet.'),
      findsOneWidget,
    );
  });

  testWidgets('workout detail shows imported activity completion', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeekWithCompletion(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Completed · Completed'), findsOneWidget);

    await tester.tap(find.text('Long'));
    await tester.pumpAndSettle();

    expect(find.text('Completion'), findsOneWidget);
    expect(find.text('Completed'), findsOneWidget);
    expect(find.text('Match'), findsOneWidget);
    expect(find.text('Matched'), findsOneWidget);
    expect(find.text('Activity type'), findsOneWidget);
    expect(find.text('Run'), findsOneWidget);
    expect(find.text('Activity distance'), findsOneWidget);
    expect(find.text('8.0 km'), findsNWidgets(2));
    expect(find.text('Completed from activity'), findsOneWidget);
    expect(
      find.text('This workout is completed from an imported activity match.'),
      findsOneWidget,
    );
    final completedButton = tester.widget<FilledButton>(
      find.widgetWithText(FilledButton, 'Completed from activity'),
    );
    expect(completedButton.onPressed, isNull);
  });

  testWidgets('weekly plan screen shows adaptation summary', (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeekWithAdaptation(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Plan changes'), findsOneWidget);
    expect(find.text('Long run adjusted'), findsOneWidget);
    expect(find.text('You underperformed the prior workout.'), findsOneWidget);
    expect(find.text('Reduced Sunday long run by 10%.'), findsOneWidget);
    expect(find.text('No adaptations this week.'), findsNothing);
  });

  testWidgets('history tab shows empty activity and adaptation states', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('This week'), findsOneWidget);

    await tester.tap(find.byIcon(Icons.history_outlined));
    await tester.pumpAndSettle();

    expect(find.text('History'), findsWidgets);
    expect(find.text('Recent activity'), findsOneWidget);
    expect(find.text('No imported activities yet.'), findsOneWidget);
    expect(find.text('Adaptation history'), findsOneWidget);
    expect(find.text('No adaptations yet.'), findsOneWidget);
    expect(find.text('This week'), findsNothing);
  });

  testWidgets('history tab shows imported activity and adaptations', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeekWithHistory(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.history_outlined));
    await tester.pumpAndSettle();

    expect(find.text('Recent activity'), findsOneWidget);
    expect(find.text('Run'), findsOneWidget);
    expect(find.text('8.0 km · 50 min'), findsOneWidget);
    expect(find.text('Adaptation history'), findsOneWidget);
    expect(find.text('Long run adjusted'), findsOneWidget);
    expect(find.text('You underperformed the prior workout.'), findsOneWidget);
    expect(find.text('Reduced Sunday long run by 10%.'), findsOneWidget);
  });

  test('current plan week parses protobuf JSON int64 strings', () {
    final currentPlanWeek = CurrentPlanWeek.fromJson({
      'planWeek': {
        'id': 'week-1',
        'startsOn': '2026-06-01',
        'focus': 'WEEK_FOCUS_BASE',
        'workouts': [
          {
            'id': 'workout-1',
            'scheduledFor': '2026-06-02',
            'type': 'WORKOUT_TYPE_EASY',
            'status': 'PLANNED_WORKOUT_STATUS_SCHEDULED',
            'targetDistanceMeters': 5000,
            'targetDurationSeconds': '1800',
            'intensity': {'kind': 'easy', 'description': ''},
            'notes': 'Easy run.',
          },
        ],
      },
      'importedActivities': [
        {
          'id': 'activity-1',
          'type': 'ACTIVITY_TYPE_RUN',
          'startedAt': '2026-06-02T08:00:00Z',
          'durationSeconds': '1800',
          'distanceMeters': 5000,
        },
      ],
      'workoutMatches': const [],
      'workoutResults': [
        {
          'id': 'result-1',
          'plannedWorkoutId': 'workout-1',
          'importedActivityId': 'activity-1',
          'outcome': 'WORKOUT_OUTCOME_COMPLETED_AS_PLANNED',
          'distanceMeters': 5000,
          'durationSeconds': '1800',
        },
      ],
      'adaptationEvents': const [],
    });

    expect(
      currentPlanWeek.planWeek.workouts.single.targetDurationSeconds,
      1800,
    );
    expect(currentPlanWeek.importedActivities.single.durationSeconds, 1800);
    expect(currentPlanWeek.workoutResults.single.durationSeconds, 1800);
  });

  test('provider connection status parses Strava protobuf JSON', () {
    final status = ProviderConnectionStatusView.fromJson({
      'hasConnection': true,
      'connection': {
        'id': 'connection-1',
        'athleteId': 'athlete-1',
        'provider': 'PROVIDER_STRAVA',
        'status': 'PROVIDER_CONNECTION_STATUS_CONNECTED',
        'lastError': '',
      },
    });

    expect(status.hasConnection, isTrue);
    expect(status.providerLabel, 'Strava');
    expect(status.statusLabel, 'Connected');
  });

  test('provider connection status still parses Garmin protobuf JSON', () {
    final status = ProviderConnectionStatusView.fromJson({
      'hasConnection': true,
      'connection': {
        'id': 'connection-1',
        'athleteId': 'athlete-1',
        'provider': 'PROVIDER_GARMIN',
        'status': 'PROVIDER_CONNECTION_STATUS_CONNECTED',
        'lastError': '',
      },
    });

    expect(status.providerLabel, 'Garmin');
  });
}

class FakeRunthreadApi implements RunthreadApi {
  const FakeRunthreadApi({
    this.currentPlanWeek,
    this.providerConnection,
    this.error,
  });

  final CurrentPlanWeek? currentPlanWeek;
  final ProviderConnectionStatusView? providerConnection;
  final Object? error;

  @override
  Future<CurrentPlanWeek> getCurrentPlanWeek() async {
    if (error != null) {
      throw error!;
    }
    return currentPlanWeek!;
  }

  @override
  Future<ProviderConnectionStatusView> getProviderConnectionStatus() async {
    if (error != null) {
      throw error!;
    }
    return providerConnection ?? ProviderConnectionStatusView.notConnected();
  }
}

ProviderConnectionStatusView testProviderConnectionStatus(
  ProviderConnectionStatus status,
) {
  return ProviderConnectionStatusView(
    hasConnection: true,
    connection: ProviderConnection(
      id: 'connection-1',
      athleteId: 'athlete-1',
      provider: Provider.strava,
      status: status,
      lastError: '',
    ),
  );
}

CurrentPlanWeek testCurrentPlanWeek() {
  return CurrentPlanWeek(
    planWeek: testPlanWeek(),
    importedActivities: const [],
    workoutMatches: const [],
    workoutResults: const [],
    adaptationEvents: const [],
  );
}

CurrentPlanWeek testCurrentPlanWeekWithCompletion() {
  return CurrentPlanWeek(
    planWeek: testPlanWeek(),
    importedActivities: [
      ImportedActivity(
        id: 'activity-workout-7',
        type: 'Run',
        startedAt: DateTime(2026, 6, 7, 8),
        durationSeconds: 3000,
        distanceMeters: 8000,
      ),
    ],
    workoutMatches: const [
      WorkoutMatch(
        id: 'match-1',
        plannedWorkoutId: 'workout-7',
        importedActivityId: 'activity-workout-7',
        status: 'Matched',
        confidence: 'High',
      ),
    ],
    workoutResults: const [
      WorkoutResult(
        id: 'result-1',
        plannedWorkoutId: 'workout-7',
        importedActivityId: 'activity-workout-7',
        outcome: 'Completed',
        distanceMeters: 8000,
        durationSeconds: 3000,
      ),
    ],
    adaptationEvents: const [],
  );
}

CurrentPlanWeek testCurrentPlanWeekWithAdaptation() {
  return CurrentPlanWeek(
    planWeek: testPlanWeek(),
    importedActivities: const [],
    workoutMatches: const [],
    workoutResults: const [],
    adaptationEvents: const [
      AdaptationEvent(
        id: 'adaptation-1',
        type: 'Underperformance',
        reason: 'You underperformed the prior workout.',
        summary: 'Long run adjusted',
        changes: [
          PlanChange(
            plannedWorkoutId: 'workout-7',
            type: 'Adjusted',
            description: 'Reduced Sunday long run by 10%.',
          ),
        ],
      ),
    ],
  );
}

CurrentPlanWeek testCurrentPlanWeekWithHistory() {
  return CurrentPlanWeek(
    planWeek: testPlanWeek(),
    importedActivities: [
      ImportedActivity(
        id: 'activity-workout-7',
        type: 'Run',
        startedAt: DateTime(2026, 6, 7, 8),
        durationSeconds: 3000,
        distanceMeters: 8000,
      ),
    ],
    workoutMatches: const [
      WorkoutMatch(
        id: 'match-1',
        plannedWorkoutId: 'workout-7',
        importedActivityId: 'activity-workout-7',
        status: 'Matched',
        confidence: 'High',
      ),
    ],
    workoutResults: const [
      WorkoutResult(
        id: 'result-1',
        plannedWorkoutId: 'workout-7',
        importedActivityId: 'activity-workout-7',
        outcome: 'Completed',
        distanceMeters: 8000,
        durationSeconds: 3000,
      ),
    ],
    adaptationEvents: const [
      AdaptationEvent(
        id: 'adaptation-1',
        type: 'Underperformance',
        reason: 'You underperformed the prior workout.',
        summary: 'Long run adjusted',
        changes: [
          PlanChange(
            plannedWorkoutId: 'workout-7',
            type: 'Adjusted',
            description: 'Reduced Sunday long run by 10%.',
          ),
        ],
      ),
    ],
  );
}

PlanWeek testPlanWeek() {
  final startsOn = DateTime(2026, 6, 1);
  return PlanWeek(
    id: 'week-1',
    startsOn: startsOn,
    focus: 'Base',
    workouts: [
      testWorkout(startsOn, 'Rest', 'Rest day.'),
      testWorkout(
        startsOn.add(const Duration(days: 1)),
        'Easy',
        'Easy conversational run.',
        distance: 5000,
        duration: 1800,
      ),
      testWorkout(
        startsOn.add(const Duration(days: 2)),
        'Strength',
        'Optional strength and mobility.',
        duration: 1800,
      ),
      testWorkout(
        startsOn.add(const Duration(days: 3)),
        'Easy',
        'Easy conversational run.',
        distance: 5200,
        duration: 1900,
      ),
      testWorkout(startsOn.add(const Duration(days: 4)), 'Rest', 'Rest day.'),
      testWorkout(startsOn.add(const Duration(days: 5)), 'Rest', 'Rest day.'),
      testWorkout(
        startsOn.add(const Duration(days: 6)),
        'Long',
        'Long run at a relaxed effort.',
        distance: 8000,
        duration: 3000,
      ),
    ],
  );
}

PlannedWorkout testWorkout(
  DateTime scheduledFor,
  String type,
  String notes, {
  double distance = 0,
  int duration = 0,
}) {
  return PlannedWorkout(
    id: 'workout-${scheduledFor.weekday}',
    scheduledFor: scheduledFor,
    type: type,
    status: 'Scheduled',
    targetDistanceMeters: distance,
    targetDurationSeconds: duration,
    intensity: const IntensityTarget(kind: 'easy', description: ''),
    notes: notes,
  );
}
