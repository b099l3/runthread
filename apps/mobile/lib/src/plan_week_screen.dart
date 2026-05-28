import 'dart:async';

import 'package:flutter/material.dart';

import 'api/runthread_api.dart';
import 'demo/demo_fallback_api.dart';
import 'deep_links.dart';
import 'models/distance_unit.dart';
import 'history_view.dart';
import 'models/plan_week.dart';
import 'settings_screen.dart';
import 'week_dates.dart';
import 'workout_detail_screen.dart';

class PlanWeekScreen extends StatefulWidget {
  const PlanWeekScreen({
    required this.api,
    required this.openUrl,
    required this.distanceUnit,
    required this.onDistanceUnitChanged,
    this.deepLinks,
    super.key,
  });

  final RunthreadApi api;
  final UrlOpener openUrl;
  final DistanceUnit distanceUnit;
  final ValueChanged<DistanceUnit> onDistanceUnitChanged;
  final Stream<Uri>? deepLinks;

  @override
  State<PlanWeekScreen> createState() => _PlanWeekScreenState();
}

class _PlanWeekScreenState extends State<PlanWeekScreen>
    with WidgetsBindingObserver {
  late Future<_PlanWeekScreenData> _screenData;
  int _selectedIndex = 0;
  late DateTime _selectedWeekStart;
  StreamSubscription<Uri>? _deepLinkSubscription;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _selectedWeekStart = currentWeekStart();
    _screenData = _loadScreenData();
    _deepLinkSubscription = widget.deepLinks?.listen(_handleDeepLink);
  }

  @override
  void didUpdateWidget(covariant PlanWeekScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.deepLinks != oldWidget.deepLinks) {
      _deepLinkSubscription?.cancel();
      _deepLinkSubscription = widget.deepLinks?.listen(_handleDeepLink);
    }
  }

  @override
  void dispose() {
    _deepLinkSubscription?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _reload();
    }
  }

  void _reload() {
    setState(() {
      _screenData = _loadScreenData();
    });
  }

  void _handleDeepLink(Uri uri) {
    if (!mounted || !isStravaConnectionDeepLink(uri)) {
      return;
    }
    _selectedIndex = 0;
    _reload();
  }

  Future<_PlanWeekScreenData> _loadScreenData() async {
    final currentPlanWeek = await widget.api.getCurrentPlanWeek(
      targetWeekDate: _selectedWeekStart,
    );
    return _PlanWeekScreenData(currentPlanWeek: currentPlanWeek);
  }

  void _changeWeek(int weekOffset) {
    setState(() {
      _selectedWeekStart = _selectedWeekStart.add(
        Duration(days: weekOffset * 7),
      );
      _screenData = _loadScreenData();
    });
  }

  void _goToCurrentWeek() {
    setState(() {
      _selectedWeekStart = currentWeekStart();
      _screenData = _loadScreenData();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Runthread'),
        actions: [
          IconButton(
            tooltip: 'Refresh plan',
            onPressed: _reload,
            icon: const Icon(Icons.refresh),
          ),
          IconButton(
            tooltip: 'Settings',
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => SettingsScreen(
                    api: widget.api,
                    openUrl: widget.openUrl,
                    deepLinks: widget.deepLinks,
                    distanceUnit: widget.distanceUnit,
                    onDistanceUnitChanged: widget.onDistanceUnitChanged,
                  ),
                ),
              );
            },
            icon: const Icon(Icons.settings_outlined),
          ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _selectedIndex,
        onDestinationSelected: (index) {
          setState(() {
            _selectedIndex = index;
          });
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.calendar_today_outlined),
            selectedIcon: Icon(Icons.calendar_today),
            label: 'Plan',
          ),
          NavigationDestination(
            icon: Icon(Icons.history_outlined),
            selectedIcon: Icon(Icons.history),
            label: 'History',
          ),
        ],
      ),
      body: SafeArea(
        child: FutureBuilder<_PlanWeekScreenData>(
          future: _screenData,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const _LoadingState();
            }
            if (snapshot.hasError) {
              return _ErrorState(
                message: snapshot.error.toString(),
                onRetry: _reload,
              );
            }
            final screenData = snapshot.data;
            if (screenData == null) {
              return _ErrorState(
                message: 'No plan week was returned.',
                onRetry: _reload,
              );
            }
            return _selectedIndex == 0
                ? _PlanWeekContent(
                    screenData: screenData,
                    onPreviousWeek: () => _changeWeek(-1),
                    onNextWeek: () => _changeWeek(1),
                    onCurrentWeek: _goToCurrentWeek,
                    distanceUnit: widget.distanceUnit,
                  )
                : HistoryView(
                    currentPlanWeek: screenData.currentPlanWeek,
                    distanceUnit: widget.distanceUnit,
                  );
          },
        ),
      ),
    );
  }
}

class _PlanWeekScreenData {
  const _PlanWeekScreenData({required this.currentPlanWeek});

  final CurrentPlanWeek currentPlanWeek;
}

class _PlanWeekContent extends StatelessWidget {
  const _PlanWeekContent({
    required this.screenData,
    required this.onPreviousWeek,
    required this.onNextWeek,
    required this.onCurrentWeek,
    required this.distanceUnit,
  });

  final _PlanWeekScreenData screenData;
  final VoidCallback onPreviousWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onCurrentWeek;
  final DistanceUnit distanceUnit;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final currentPlanWeek = screenData.currentPlanWeek;
    final planWeek = currentPlanWeek.planWeek;
    return ListView(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
      children: [
        Text(
          'This week',
          style: textTheme.headlineMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 6),
        _WeekSelector(
          startsOn: planWeek.startsOn,
          onPreviousWeek: onPreviousWeek,
          onNextWeek: onNextWeek,
          onCurrentWeek: onCurrentWeek,
        ),
        const SizedBox(height: 10),
        Text(
          'Starts ${_formatDate(planWeek.startsOn)} · ${planWeek.focus} focus',
          style: textTheme.bodyMedium?.copyWith(
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        if (planWeek.isDemo) ...[
          const SizedBox(height: 12),
          const _DemoNotice(),
        ],
        const SizedBox(height: 16),
        _AdaptationSummary(events: currentPlanWeek.adaptationEvents),
        const SizedBox(height: 20),
        for (final workout in planWeek.workouts) ...[
          _WorkoutTile(
            workout: workout,
            completion: currentPlanWeek.completionFor(workout),
            activities: _activitiesForDay(
              currentPlanWeek.importedActivities,
              workout.scheduledFor,
            ),
            distanceUnit: distanceUnit,
            onTap: () {
              Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => WorkoutDetailScreen(
                    workout: workout,
                    completion: currentPlanWeek.completionFor(workout),
                    distanceUnit: distanceUnit,
                  ),
                ),
              );
            },
          ),
          const SizedBox(height: 10),
        ],
      ],
    );
  }
}

class _WeekSelector extends StatelessWidget {
  const _WeekSelector({
    required this.startsOn,
    required this.onPreviousWeek,
    required this.onNextWeek,
    required this.onCurrentWeek,
  });

  final DateTime startsOn;
  final VoidCallback onPreviousWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onCurrentWeek;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Row(
      children: [
        IconButton.filledTonal(
          tooltip: 'Previous week',
          onPressed: onPreviousWeek,
          icon: const Icon(Icons.chevron_left),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            compactWeekRange(startsOn),
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: colorScheme.onSurface,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
        const SizedBox(width: 8),
        IconButton.filledTonal(
          tooltip: 'Next week',
          onPressed: onNextWeek,
          icon: const Icon(Icons.chevron_right),
        ),
        const SizedBox(width: 8),
        TextButton(onPressed: onCurrentWeek, child: const Text('Today')),
      ],
    );
  }
}

class _ImportedActivityRow extends StatelessWidget {
  const _ImportedActivityRow({
    required this.activity,
    required this.distanceUnit,
  });

  final ImportedActivity activity;
  final DistanceUnit distanceUnit;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final distanceLabel = activity.distanceLabel(distanceUnit);
    final details = [
      if (distanceLabel.isNotEmpty) distanceLabel,
      if (activity.durationLabel.isNotEmpty) activity.durationLabel,
    ];
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(
          _iconForActivity(activity.type),
          size: 18,
          color: colorScheme.onSurfaceVariant,
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                activity.type.isEmpty ? 'Imported activity' : activity.type,
                style: textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              if (details.isNotEmpty) ...[
                const SizedBox(height: 3),
                Text(
                  details.join(' · '),
                  style: textTheme.bodySmall?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}

class _AdaptationSummary extends StatelessWidget {
  const _AdaptationSummary({required this.events});

  final List<AdaptationEvent> events;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.tune, size: 20, color: colorScheme.primary),
                const SizedBox(width: 8),
                Text(
                  'Plan changes',
                  style: textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (events.isEmpty)
              Text(
                'No adaptations this week.',
                style: textTheme.bodyMedium?.copyWith(
                  color: colorScheme.onSurfaceVariant,
                ),
              )
            else
              for (final event in events) ...[
                _AdaptationEventView(event: event),
                if (event != events.last) const SizedBox(height: 12),
              ],
          ],
        ),
      ),
    );
  }
}

class _AdaptationEventView extends StatelessWidget {
  const _AdaptationEventView({required this.event});

  final AdaptationEvent event;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          event.summary.isEmpty ? event.type : event.summary,
          style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
        ),
        if (event.reason.isNotEmpty) ...[
          const SizedBox(height: 4),
          Text(
            event.reason,
            style: textTheme.bodySmall?.copyWith(
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ],
        for (final change in event.changes) ...[
          const SizedBox(height: 6),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '•',
                style: textTheme.bodySmall?.copyWith(
                  color: colorScheme.onSurfaceVariant,
                ),
              ),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  change.description.isEmpty ? change.type : change.description,
                  style: textTheme.bodySmall?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }
}

class _DemoNotice extends StatelessWidget {
  const _DemoNotice();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final fallbackError = DemoFallbackRunthreadApi.lastFallbackError;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: colorScheme.secondaryContainer,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        child: Row(
          children: [
            Icon(
              Icons.info_outline,
              size: 18,
              color: colorScheme.onSecondaryContainer,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                fallbackError == null
                    ? 'Demo data · local backend not ready'
                    : 'Demo data · ${fallbackError.toString()}',
                style: Theme.of(context).textTheme.labelLarge?.copyWith(
                  color: colorScheme.onSecondaryContainer,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _WorkoutTile extends StatelessWidget {
  const _WorkoutTile({
    required this.workout,
    required this.completion,
    required this.activities,
    required this.distanceUnit,
    required this.onTap,
  });

  final PlannedWorkout workout;
  final WorkoutCompletionState completion;
  final List<ImportedActivity> activities;
  final DistanceUnit distanceUnit;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final distanceLabel = workout.distanceLabel(distanceUnit);
    final details = [
      if (distanceLabel.isNotEmpty) distanceLabel,
      if (workout.durationLabel.isNotEmpty) workout.durationLabel,
      if (workout.intensity.kind.isNotEmpty) workout.intensity.kind,
    ];

    return Card(
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SizedBox(
                width: 48,
                child: Column(
                  children: [
                    Text(
                      _weekday(workout.scheduledFor),
                      style: textTheme.labelLarge?.copyWith(
                        color: colorScheme.primary,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 3),
                    Text(
                      workout.scheduledFor.day.toString(),
                      style: textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(
                          _iconForWorkout(workout.type),
                          size: 20,
                          color: colorScheme.primary,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            workout.type,
                            style: textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                        Text(
                          workout.status,
                          style: textTheme.labelMedium?.copyWith(
                            color: colorScheme.onSurfaceVariant,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Icon(
                          Icons.chevron_right,
                          size: 20,
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ],
                    ),
                    if (details.isNotEmpty) ...[
                      const SizedBox(height: 8),
                      Text(details.join(' · '), style: textTheme.bodyMedium),
                    ],
                    if (workout.notes.isNotEmpty) ...[
                      const SizedBox(height: 6),
                      Text(
                        workout.notes,
                        style: textTheme.bodySmall?.copyWith(
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    const SizedBox(height: 8),
                    _CompletionSummary(
                      completion: completion,
                      hasDayActivity: activities.isNotEmpty,
                    ),
                    if (activities.isNotEmpty) ...[
                      const SizedBox(height: 10),
                      for (final activity in activities) ...[
                        _ImportedActivityRow(
                          activity: activity,
                          distanceUnit: distanceUnit,
                        ),
                        if (activity != activities.last)
                          const SizedBox(height: 8),
                      ],
                    ],
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CompletionSummary extends StatelessWidget {
  const _CompletionSummary({
    required this.completion,
    required this.hasDayActivity,
  });

  final WorkoutCompletionState completion;
  final bool hasDayActivity;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final hasActivity = completion.hasActivity;
    final label = completion.hasResult
        ? 'Completed · ${completion.workoutResult!.outcome}'
        : hasActivity
        ? 'Activity imported'
        : hasDayActivity
        ? 'Imported activity on this day'
        : 'No imported activity';
    final hasVisibleActivity = hasActivity || hasDayActivity;

    return Row(
      children: [
        Icon(
          hasVisibleActivity
              ? Icons.check_circle_outline
              : Icons.radio_button_unchecked,
          size: 16,
          color: hasVisibleActivity
              ? colorScheme.primary
              : colorScheme.onSurfaceVariant,
        ),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            label,
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: hasVisibleActivity
                  ? colorScheme.primary
                  : colorScheme.onSurfaceVariant,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      ],
    );
  }
}

class _LoadingState extends StatelessWidget {
  const _LoadingState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: SizedBox(
        width: 32,
        height: 32,
        child: CircularProgressIndicator(strokeWidth: 3),
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  const _ErrorState({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.sync_problem,
              size: 36,
              color: Theme.of(context).colorScheme.error,
            ),
            const SizedBox(height: 14),
            Text(
              'Plan unavailable',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 18),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}

IconData _iconForWorkout(String type) {
  switch (type.toLowerCase()) {
    case 'rest':
      return Icons.hotel;
    case 'strength':
      return Icons.fitness_center;
    case 'ride':
      return Icons.directions_bike;
    case 'long':
      return Icons.route;
    default:
      return Icons.directions_run;
  }
}

IconData _iconForActivity(String type) {
  switch (type.toLowerCase()) {
    case 'walk':
      return Icons.directions_walk;
    case 'ride':
    case 'virtualride':
      return Icons.directions_bike;
    default:
      return Icons.directions_run;
  }
}

List<ImportedActivity> _activitiesForDay(
  List<ImportedActivity> activities,
  DateTime day,
) {
  final targetDay = DateTime(day.year, day.month, day.day);
  final dayActivities = activities.where((activity) {
    final startedAt = activity.startedAt;
    if (startedAt == null) {
      return false;
    }
    final startedDate = DateTime(
      startedAt.year,
      startedAt.month,
      startedAt.day,
    );
    return startedDate == targetDay;
  }).toList();
  dayActivities.sort((a, b) {
    final left = a.startedAt;
    final right = b.startedAt;
    if (left == null && right == null) {
      return 0;
    }
    if (left == null) {
      return 1;
    }
    if (right == null) {
      return -1;
    }
    return left.compareTo(right);
  });
  return dayActivities;
}

String _weekday(DateTime date) {
  const weekdays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  return weekdays[date.weekday - 1];
}

String _formatDate(DateTime date) {
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
  return '${months[date.month - 1]} ${date.day}, ${date.year}';
}
