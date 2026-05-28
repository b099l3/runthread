import 'dart:convert';
import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:runthread_mobile/src/api/runthread_api.dart';
import 'package:runthread_mobile/src/app.dart';
import 'package:runthread_mobile/src/demo/demo_fallback_api.dart';
import 'package:runthread_mobile/src/mobile_config.dart';
import 'package:runthread_mobile/src/models/distance_unit.dart';
import 'package:runthread_mobile/src/models/plan_week.dart';
import 'package:runthread_mobile/src/models/provider_connection.dart';
import 'package:runthread_mobile/src/week_dates.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test('mobile config uses defaults when env is empty', () {
    final config = MobileConfig.fromEnv(const {});

    expect(config.apiBaseUrl, Uri.parse('http://localhost:8080'));
    expect(config.athleteId, 'athlete-1');
    expect(config.goalId, 'goal-1');
    expect(
      config.stravaRedirectUri,
      'http://localhost:8080/providers/strava/oauth/callback',
    );
  });

  test('mobile config uses non-secret env values', () {
    final config = MobileConfig.fromEnv(const {
      'RUNTHREAD_API_BASE_URL': 'http://10.0.2.2:8080',
      'RUNTHREAD_ATHLETE_ID': '11111111-1111-1111-1111-111111111111',
      'RUNTHREAD_GOAL_ID': '22222222-2222-2222-2222-222222222222',
      'STRAVA_OAUTH_REDIRECT_URI':
          'http://10.0.2.2:8080/providers/strava/oauth/callback',
    });

    expect(config.apiBaseUrl, Uri.parse('http://10.0.2.2:8080'));
    expect(config.athleteId, '11111111-1111-1111-1111-111111111111');
    expect(config.goalId, '22222222-2222-2222-2222-222222222222');
    expect(
      config.stravaRedirectUri,
      'http://10.0.2.2:8080/providers/strava/oauth/callback',
    );
  });

  test('mobile config falls back for invalid or missing API base URL', () {
    expect(
      MobileConfig.fromEnv(const {
        'RUNTHREAD_API_BASE_URL': 'localhost:8080',
      }).apiBaseUrl,
      Uri.parse('http://localhost:8080'),
    );
    expect(
      MobileConfig.fromEnv(const {'RUNTHREAD_API_BASE_URL': '   '}).apiBaseUrl,
      Uri.parse('http://localhost:8080'),
    );
  });

  test('current week date helpers use the containing Monday week', () {
    final saturday = DateTime(2026, 5, 16, 13, 45);

    expect(currentWeekStart(now: saturday), DateTime(2026, 5, 11));
    expect(currentWeekTargetDate(now: saturday), '2026-05-16');
    expect(compactWeekRange(DateTime(2026, 5, 25)), 'May 25-31');
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
    expect(find.text('Jun 1-7'), findsOneWidget);
    expect(find.byTooltip('Previous week'), findsOneWidget);
    expect(find.byTooltip('Next week'), findsOneWidget);
    expect(find.text('Today'), findsOneWidget);
    expect(find.text('Easy'), findsNWidgets(2));
    expect(find.text('Rest'), findsNWidgets(3));
    expect(find.text('Long'), findsOneWidget);
    expect(find.text('Strength'), findsOneWidget);
    expect(find.textContaining('5.0 km'), findsOneWidget);
    expect(find.textContaining('8.0 km'), findsWidgets);
    expect(find.text('Strava'), findsNothing);
    expect(find.text('Connect Strava'), findsNothing);
    expect(find.text('Plan changes'), findsOneWidget);
    expect(find.text('No adaptations this week.'), findsOneWidget);
  });

  testWidgets('week selector reloads previous and next weeks', (tester) async {
    final api = FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek());
    await tester.pumpWidget(RunthreadApp(api: api));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Previous week'));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Next week'));
    await tester.pumpAndSettle();

    expect(api.getCurrentPlanWeekCalls, 3);
    expect(api.requestedWeekStarts.length, 3);
    expect(
      api.requestedWeekStarts[1],
      api.requestedWeekStarts[0].subtract(const Duration(days: 7)),
    );
    expect(api.requestedWeekStarts[2], api.requestedWeekStarts[0]);
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

  testWidgets('settings screen shows Strava and disabled Garmin', (
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

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    expect(find.text('Connections'), findsOneWidget);
    expect(find.text('Strava'), findsOneWidget);
    expect(find.text('Garmin'), findsOneWidget);
    expect(find.text('Disabled'), findsOneWidget);
    expect(
      find.text(
        'Run completion will come from imported Strava activity once provider access is ready.',
      ),
      findsOneWidget,
    );
    expect(find.widgetWithText(FilledButton, 'Connect Strava'), findsOneWidget);
  });

  testWidgets('settings distance unit toggle updates plan detail and history', (
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

    expect(find.textContaining('8.0 km'), findsWidgets);

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();
    expect(find.text('Distance units'), findsOneWidget);

    await tester.tap(find.text('mi'));
    await tester.pumpAndSettle();
    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(find.textContaining('5.0 mi'), findsWidgets);
    expect(find.textContaining('8.0 km'), findsNothing);

    await tester.tap(find.text('Long'));
    await tester.pumpAndSettle();
    expect(find.text('Target distance'), findsOneWidget);
    expect(find.text('5.0 mi'), findsWidgets);

    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.tap(find.byIcon(Icons.history_outlined));
    await tester.pumpAndSettle();

    expect(find.text('Sun, Jun 7 · 5.0 mi · 50 min'), findsOneWidget);

    final preferences = await SharedPreferences.getInstance();
    expect(
      preferences.getString(distanceUnitPreferenceKey),
      DistanceUnit.miles.storageValue,
    );
  });

  testWidgets('app starts with persisted mile distance unit', (tester) async {
    SharedPreferences.setMockInitialValues({
      distanceUnitPreferenceKey: DistanceUnit.miles.storageValue,
    });
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

    expect(find.textContaining('3.1 mi'), findsOneWidget);
    expect(find.textContaining('5.0 mi'), findsOneWidget);
    expect(find.textContaining('8.0 km'), findsNothing);
  });

  testWidgets('settings Strava row shows pending status without connect button', (
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

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    expect(find.text('Strava'), findsOneWidget);
    expect(find.text('Pending'), findsOneWidget);
    expect(
      find.text(
        'Connection has started. Strava authorization remains disabled until provider access is ready.',
      ),
      findsOneWidget,
    );
    expect(find.widgetWithText(FilledButton, 'Connect Strava'), findsNothing);
    expect(find.widgetWithText(OutlinedButton, 'Disconnect'), findsOneWidget);
  });

  for (final status in [
    ProviderConnectionStatus.syncing,
    ProviderConnectionStatus.error,
    ProviderConnectionStatus.disconnected,
  ]) {
    testWidgets('settings Strava row renders ${status.name} status', (
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
            providerConnection: testProviderConnectionStatus(status),
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Settings'));
      await tester.pumpAndSettle();

      expect(find.text('Strava'), findsOneWidget);
      expect(find.text(_statusLabel(status)), findsOneWidget);
      if (status == ProviderConnectionStatus.syncing) {
        expect(
          find.widgetWithText(FilledButton, 'Connect Strava'),
          findsNothing,
        );
        expect(
          find.widgetWithText(OutlinedButton, 'Disconnect'),
          findsOneWidget,
        );
      } else if (status == ProviderConnectionStatus.disconnected) {
        expect(
          find.widgetWithText(FilledButton, 'Connect Strava'),
          findsOneWidget,
        );
        expect(find.widgetWithText(OutlinedButton, 'Disconnect'), findsNothing);
      } else {
        expect(
          find.widgetWithText(FilledButton, 'Connect Strava'),
          findsOneWidget,
        );
        expect(
          find.widgetWithText(OutlinedButton, 'Disconnect'),
          findsOneWidget,
        );
      }
    });
  }

  testWidgets(
    'settings Strava row shows connected status without connect button',
    (tester) async {
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

      await tester.tap(find.byTooltip('Settings'));
      await tester.pumpAndSettle();

      expect(find.text('Strava'), findsOneWidget);
      expect(find.text('Connected'), findsOneWidget);
      expect(
        find.text(
          'Runthread is ready to receive imported Strava activity for workout completion.',
        ),
        findsOneWidget,
      );
      expect(find.widgetWithText(FilledButton, 'Connect Strava'), findsNothing);
      expect(find.widgetWithText(OutlinedButton, 'Disconnect'), findsOneWidget);
    },
  );

  testWidgets('settings Strava disconnect updates status', (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final api = FakeRunthreadApi(
      currentPlanWeek: testCurrentPlanWeek(),
      providerConnection: testProviderConnectionStatus(
        ProviderConnectionStatus.connected,
      ),
      disconnectConnectionResult: testDisconnectConnectionResult(),
    );
    await tester.pumpWidget(RunthreadApp(api: api));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(OutlinedButton, 'Disconnect'));
    await tester.pumpAndSettle();

    expect(api.disconnectProviderConnectionCalls, 1);
    expect(api.lastDisconnectedProviderConnectionId, 'connection-1');
    expect(find.text('Strava disconnected.'), findsOneWidget);
    expect(find.text('Disconnected'), findsOneWidget);
    expect(find.widgetWithText(FilledButton, 'Connect Strava'), findsOneWidget);
    expect(find.widgetWithText(OutlinedButton, 'Disconnect'), findsNothing);
  });

  testWidgets('settings Strava disconnect shows backend failure', (
    tester,
  ) async {
    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeek(),
          providerConnection: testProviderConnectionStatus(
            ProviderConnectionStatus.connected,
          ),
          disconnectConnectionError: const RunthreadApiException(
            'disconnect failed',
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(OutlinedButton, 'Disconnect'));
    await tester.pumpAndSettle();

    expect(find.text('disconnect failed'), findsOneWidget);
  });

  testWidgets('settings Strava connect opens authorization URL', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    Uri? openedUrl;
    final api = FakeRunthreadApi(
      currentPlanWeek: testCurrentPlanWeek(),
      startConnectionResult: testStartConnectionResult(
        'https://www.strava.com/oauth/authorize?client_id=client-1',
      ),
    );
    await tester.pumpWidget(
      RunthreadApp(
        api: api,
        openUrl: (uri) async {
          openedUrl = uri;
          return true;
        },
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Connect Strava'));
    await tester.pumpAndSettle();

    expect(api.startProviderConnectionCalls, 1);
    expect(openedUrl.toString(), contains('strava.com/oauth/authorize'));
    expect(find.text('Return here after authorizing Strava.'), findsOneWidget);
  });

  testWidgets('settings Strava connect shows URL launch failure', (
    tester,
  ) async {
    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeek(),
          startConnectionResult: testStartConnectionResult(
            'https://www.strava.com/oauth/authorize',
          ),
        ),
        openUrl: (_) async => false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Connect Strava'));
    await tester.pumpAndSettle();

    expect(find.text('Could not open Strava authorization.'), findsOneWidget);
  });

  testWidgets('settings Strava connect shows backend failure', (tester) async {
    await tester.pumpWidget(
      RunthreadApp(
        api: FakeRunthreadApi(
          currentPlanWeek: testCurrentPlanWeek(),
          startConnectionError: const RunthreadApiException('start failed'),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Connect Strava'));
    await tester.pumpAndSettle();

    expect(find.text('start failed'), findsOneWidget);
  });

  testWidgets('app resume refreshes current plan', (tester) async {
    final api = FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek());
    await tester.pumpWidget(RunthreadApp(api: api));
    await tester.pumpAndSettle();

    expect(api.getCurrentPlanWeekCalls, 1);

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pumpAndSettle();

    expect(api.getCurrentPlanWeekCalls, 2);
  });

  testWidgets('Strava deep link refreshes current plan', (tester) async {
    final deepLinks = StreamController<Uri>.broadcast();
    addTearDown(deepLinks.close);
    final api = FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek());
    await tester.pumpWidget(
      RunthreadApp(api: api, deepLinks: deepLinks.stream),
    );
    await tester.pumpAndSettle();

    expect(api.getCurrentPlanWeekCalls, 1);

    deepLinks.add(Uri.parse('runthread://provider/strava/connected'));
    await tester.pumpAndSettle();

    expect(api.getCurrentPlanWeekCalls, 2);
  });

  testWidgets('settings resume refreshes provider status', (tester) async {
    final api = FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek());
    await tester.pumpWidget(RunthreadApp(api: api));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    expect(api.getProviderConnectionStatusCalls, 1);

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pumpAndSettle();

    expect(api.getProviderConnectionStatusCalls, 2);
  });

  testWidgets('settings Strava deep link refreshes provider status', (
    tester,
  ) async {
    final deepLinks = StreamController<Uri>.broadcast();
    addTearDown(deepLinks.close);
    final api = FakeRunthreadApi(currentPlanWeek: testCurrentPlanWeek());
    await tester.pumpWidget(
      RunthreadApp(api: api, deepLinks: deepLinks.stream),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Settings'));
    await tester.pumpAndSettle();

    expect(api.getProviderConnectionStatusCalls, 1);

    deepLinks.add(Uri.parse('runthread://provider/strava/connected'));
    await tester.pumpAndSettle();

    expect(api.getProviderConnectionStatusCalls, 2);
    expect(find.text('Strava connection updated.'), findsOneWidget);
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

  testWidgets(
    'weekly plan screen shows imported activities on matching day card',
    (tester) async {
      tester.view.physicalSize = const Size(900, 1800);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(
        RunthreadApp(
          api: FakeRunthreadApi(
            currentPlanWeek: testCurrentPlanWeekWithUnmatchedActivities(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Week activity'), findsNothing);
      expect(find.text('Run'), findsOneWidget);
      expect(find.text('5.0 km · 30 min'), findsOneWidget);
      expect(find.text('Imported activity on this day'), findsOneWidget);
      expect(find.text('8.0 km · 50 min'), findsNothing);
    },
  );

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
    expect(find.text('Sun, Jun 7 · 8.0 km · 50 min'), findsOneWidget);
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

  test('start provider connection parses protobuf JSON', () {
    final result = StartProviderConnectionResult.fromJson({
      'connection': {
        'id': 'connection-1',
        'athleteId': 'athlete-1',
        'provider': 'PROVIDER_STRAVA',
        'status': 'PROVIDER_CONNECTION_STATUS_PENDING',
        'lastError': '',
      },
      'authorizationUrl': 'https://www.strava.com/oauth/authorize',
      'state': 'state-1',
      'oauthReady': true,
    });

    expect(result.oauthReady, isTrue);
    expect(result.authorizationUrl, contains('strava.com'));
    expect(result.connection.provider, Provider.strava);
  });

  test('disconnect provider connection parses protobuf JSON', () {
    final result = DisconnectProviderConnectionResult.fromJson({
      'connection': {
        'id': 'connection-1',
        'athleteId': 'athlete-1',
        'provider': 'PROVIDER_STRAVA',
        'status': 'PROVIDER_CONNECTION_STATUS_DISCONNECTED',
        'lastError': '',
      },
    });

    expect(result.connection.status, ProviderConnectionStatus.disconnected);
    expect(result.connection.provider, Provider.strava);
  });

  test('HTTP API starts Strava connection with configured athlete', () async {
    late Map<String, dynamic> body;
    final api = HttpRunthreadApi(
      baseUrl: Uri.parse('http://localhost:8080'),
      athleteId: '11111111-1111-1111-1111-111111111111',
      stravaRedirectUri:
          'http://localhost:8080/providers/strava/oauth/callback',
      client: MockClient((request) async {
        body = jsonDecode(request.body) as Map<String, dynamic>;
        return http.Response(
          jsonEncode({
            'connection': {
              'id': 'connection-1',
              'athleteId': body['athleteId'],
              'provider': 'PROVIDER_STRAVA',
              'status': 'PROVIDER_CONNECTION_STATUS_PENDING',
              'lastError': '',
            },
            'authorizationUrl': 'https://www.strava.com/oauth/authorize',
            'state': 'state-1',
            'oauthReady': true,
          }),
          200,
        );
      }),
    );

    final result = await api.startProviderConnection();

    expect(body['athleteId'], '11111111-1111-1111-1111-111111111111');
    expect(body['provider'], 'PROVIDER_STRAVA');
    expect(
      body['redirectUri'],
      'http://localhost:8080/providers/strava/oauth/callback',
    );
    expect(result.oauthReady, isTrue);
  });

  test('HTTP API current plan sends configured athlete and goal', () async {
    late Map<String, dynamic> body;
    final api = HttpRunthreadApi(
      baseUrl: Uri.parse('http://localhost:8080'),
      athleteId: '11111111-1111-1111-1111-111111111111',
      goalId: '22222222-2222-2222-2222-222222222222',
      stravaRedirectUri:
          'http://localhost:8080/providers/strava/oauth/callback',
      client: MockClient((request) async {
        body = jsonDecode(request.body) as Map<String, dynamic>;
        return http.Response(
          jsonEncode({
            'planWeek': {
              'id': '33333333-3333-3333-3333-333333333333',
              'startsOn': '2026-06-01',
              'focus': 'WEEK_FOCUS_BASE',
              'workouts': const [],
            },
            'importedActivities': const [],
            'workoutMatches': const [],
            'workoutResults': const [],
            'adaptationEvents': const [],
          }),
          200,
        );
      }),
    );

    await api.getCurrentPlanWeek(targetWeekDate: DateTime(2026, 5, 18));

    expect(body['athleteId'], '11111111-1111-1111-1111-111111111111');
    expect(body['goalId'], '22222222-2222-2222-2222-222222222222');
    expect(body['targetWeekDate'], '2026-05-18');
  });

  test(
    'HTTP API disconnects Strava connection with configured athlete',
    () async {
      late Map<String, dynamic> body;
      final api = HttpRunthreadApi(
        baseUrl: Uri.parse('http://localhost:8080'),
        athleteId: '11111111-1111-1111-1111-111111111111',
        stravaRedirectUri:
            'http://localhost:8080/providers/strava/oauth/callback',
        client: MockClient((request) async {
          body = jsonDecode(request.body) as Map<String, dynamic>;
          return http.Response(
            jsonEncode({
              'connection': {
                'id': body['providerConnectionId'],
                'athleteId': body['athleteId'],
                'provider': 'PROVIDER_STRAVA',
                'status': 'PROVIDER_CONNECTION_STATUS_DISCONNECTED',
                'lastError': '',
              },
            }),
            200,
          );
        }),
      );

      final result = await api.disconnectProviderConnection(
        providerConnectionId: 'connection-1',
      );

      expect(body['athleteId'], '11111111-1111-1111-1111-111111111111');
      expect(body['provider'], 'PROVIDER_STRAVA');
      expect(body['providerConnectionId'], 'connection-1');
      expect(result.connection.status, ProviderConnectionStatus.disconnected);
    },
  );

  test('HTTP API start Strava connection throws for backend error', () async {
    final api = HttpRunthreadApi(
      baseUrl: Uri.parse('http://localhost:8080'),
      athleteId: 'athlete-1',
      stravaRedirectUri:
          'http://localhost:8080/providers/strava/oauth/callback',
      client: MockClient((_) async => http.Response('failed', 500)),
    );

    expect(api.startProviderConnection, throwsA(isA<RunthreadApiException>()));
  });

  test('HTTP API current plan times out quickly when backend stalls', () {
    final api = HttpRunthreadApi(
      baseUrl: Uri.parse('http://localhost:8080'),
      athleteId: 'athlete-1',
      stravaRedirectUri:
          'http://localhost:8080/providers/strava/oauth/callback',
      requestTimeout: const Duration(milliseconds: 10),
      client: MockClient((_) async {
        await Future<void>.delayed(const Duration(seconds: 1));
        return http.Response('{}', 200);
      }),
    );

    expect(api.getCurrentPlanWeek, throwsA(isA<RunthreadApiException>()));
  });

  test('HTTP API disconnect Strava connection throws for backend error', () {
    final api = HttpRunthreadApi(
      baseUrl: Uri.parse('http://localhost:8080'),
      athleteId: 'athlete-1',
      stravaRedirectUri:
          'http://localhost:8080/providers/strava/oauth/callback',
      client: MockClient((_) async => http.Response('failed', 500)),
    );

    expect(
      () => api.disconnectProviderConnection(
        providerConnectionId: 'connection-1',
      ),
      throwsA(isA<RunthreadApiException>()),
    );
  });
}

class FakeRunthreadApi implements RunthreadApi {
  FakeRunthreadApi({
    this.currentPlanWeek,
    this.providerConnection,
    this.startConnectionResult,
    this.startConnectionError,
    this.disconnectConnectionResult,
    this.disconnectConnectionError,
    this.error,
  });

  final CurrentPlanWeek? currentPlanWeek;
  final ProviderConnectionStatusView? providerConnection;
  final StartProviderConnectionResult? startConnectionResult;
  final Object? startConnectionError;
  final DisconnectProviderConnectionResult? disconnectConnectionResult;
  final Object? disconnectConnectionError;
  final Object? error;
  int getCurrentPlanWeekCalls = 0;
  int getProviderConnectionStatusCalls = 0;
  int startProviderConnectionCalls = 0;
  int disconnectProviderConnectionCalls = 0;
  String? lastDisconnectedProviderConnectionId;
  final List<DateTime> requestedWeekStarts = [];

  @override
  Future<CurrentPlanWeek> getCurrentPlanWeek({DateTime? targetWeekDate}) async {
    getCurrentPlanWeekCalls++;
    requestedWeekStarts.add(currentWeekStart(now: targetWeekDate));
    if (error != null) {
      throw error!;
    }
    return currentPlanWeek!;
  }

  @override
  Future<ProviderConnectionStatusView> getProviderConnectionStatus() async {
    getProviderConnectionStatusCalls++;
    if (error != null) {
      throw error!;
    }
    return providerConnection ?? ProviderConnectionStatusView.notConnected();
  }

  @override
  Future<StartProviderConnectionResult> startProviderConnection() async {
    startProviderConnectionCalls++;
    if (startConnectionError != null) {
      throw startConnectionError!;
    }
    if (error != null) {
      throw error!;
    }
    return startConnectionResult ?? testStartConnectionResult('');
  }

  @override
  Future<DisconnectProviderConnectionResult> disconnectProviderConnection({
    String? providerConnectionId,
  }) async {
    disconnectProviderConnectionCalls++;
    lastDisconnectedProviderConnectionId = providerConnectionId;
    if (disconnectConnectionError != null) {
      throw disconnectConnectionError!;
    }
    if (error != null) {
      throw error!;
    }
    return disconnectConnectionResult ?? testDisconnectConnectionResult();
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

StartProviderConnectionResult testStartConnectionResult(
  String authorizationUrl,
) {
  return StartProviderConnectionResult(
    connection: const ProviderConnection(
      id: 'connection-1',
      athleteId: 'athlete-1',
      provider: Provider.strava,
      status: ProviderConnectionStatus.pending,
      lastError: '',
    ),
    authorizationUrl: authorizationUrl,
    state: 'state-1',
    oauthReady: true,
  );
}

DisconnectProviderConnectionResult testDisconnectConnectionResult() {
  return const DisconnectProviderConnectionResult(
    connection: ProviderConnection(
      id: 'connection-1',
      athleteId: 'athlete-1',
      provider: Provider.strava,
      status: ProviderConnectionStatus.disconnected,
      lastError: '',
    ),
  );
}

String _statusLabel(ProviderConnectionStatus status) {
  return switch (status) {
    ProviderConnectionStatus.pending => 'Pending',
    ProviderConnectionStatus.connected => 'Connected',
    ProviderConnectionStatus.syncing => 'Syncing',
    ProviderConnectionStatus.error => 'Needs attention',
    ProviderConnectionStatus.disconnected => 'Disconnected',
    ProviderConnectionStatus.unspecified => 'Not connected',
  };
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

CurrentPlanWeek testCurrentPlanWeekWithUnmatchedActivities() {
  return CurrentPlanWeek(
    planWeek: testPlanWeek(),
    importedActivities: [
      ImportedActivity(
        id: 'activity-week',
        type: 'Run',
        startedAt: DateTime(2026, 6, 2, 8),
        durationSeconds: 1800,
        distanceMeters: 5000,
      ),
      ImportedActivity(
        id: 'activity-other-week',
        type: 'Run',
        startedAt: DateTime(2026, 6, 14, 8),
        durationSeconds: 3000,
        distanceMeters: 8000,
      ),
    ],
    workoutMatches: const [],
    workoutResults: const [],
    adaptationEvents: const [],
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
