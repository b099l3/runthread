import 'package:flutter/material.dart';

import 'api/runthread_api.dart';
import 'demo/demo_fallback_api.dart';
import 'history_view.dart';
import 'models/plan_week.dart';
import 'models/provider_connection.dart';
import 'workout_detail_screen.dart';

class PlanWeekScreen extends StatefulWidget {
  const PlanWeekScreen({required this.api, super.key});

  final RunthreadApi api;

  @override
  State<PlanWeekScreen> createState() => _PlanWeekScreenState();
}

class _PlanWeekScreenState extends State<PlanWeekScreen> {
  late Future<_PlanWeekScreenData> _screenData;
  int _selectedIndex = 0;

  @override
  void initState() {
    super.initState();
    _screenData = _loadScreenData();
  }

  void _reload() {
    setState(() {
      _screenData = _loadScreenData();
    });
  }

  Future<_PlanWeekScreenData> _loadScreenData() async {
    final currentPlanWeek = await widget.api.getCurrentPlanWeek();
    final providerConnection = await widget.api
        .getProviderConnectionStatus()
        .catchError(
          (_) => ProviderConnectionStatusView.notConnected(
            statusUnavailable: true,
          ),
        );
    return _PlanWeekScreenData(
      currentPlanWeek: currentPlanWeek,
      providerConnection: providerConnection,
    );
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
                ? _PlanWeekContent(screenData: screenData)
                : HistoryView(currentPlanWeek: screenData.currentPlanWeek);
          },
        ),
      ),
    );
  }
}

class _PlanWeekScreenData {
  const _PlanWeekScreenData({
    required this.currentPlanWeek,
    required this.providerConnection,
  });

  final CurrentPlanWeek currentPlanWeek;
  final ProviderConnectionStatusView providerConnection;
}

class _PlanWeekContent extends StatelessWidget {
  const _PlanWeekContent({required this.screenData});

  final _PlanWeekScreenData screenData;

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
        _ProviderConnectionSummary(
          providerConnection: screenData.providerConnection,
        ),
        const SizedBox(height: 12),
        _AdaptationSummary(events: currentPlanWeek.adaptationEvents),
        const SizedBox(height: 20),
        for (final workout in planWeek.workouts) ...[
          _WorkoutTile(
            workout: workout,
            completion: currentPlanWeek.completionFor(workout),
            onTap: () {
              Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => WorkoutDetailScreen(
                    workout: workout,
                    completion: currentPlanWeek.completionFor(workout),
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

class _ProviderConnectionSummary extends StatelessWidget {
  const _ProviderConnectionSummary({required this.providerConnection});

  final ProviderConnectionStatusView providerConnection;

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
                Icon(
                  Icons.watch_outlined,
                  size: 20,
                  color: colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    providerConnection.providerLabel,
                    style: textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                DecoratedBox(
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(8),
                    color: colorScheme.surfaceContainerHighest,
                  ),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    child: Text(
                      providerConnection.statusLabel,
                      style: textTheme.labelMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              providerConnection.description,
              style: textTheme.bodyMedium?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                OutlinedButton.icon(
                  onPressed: null,
                  icon: const Icon(Icons.link),
                  label: Text('Connect ${providerConnection.providerLabel}'),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Disabled for now',
                    style: textTheme.labelMedium?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
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
    required this.onTap,
  });

  final PlannedWorkout workout;
  final WorkoutCompletionState completion;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final details = [
      if (workout.distanceLabel.isNotEmpty) workout.distanceLabel,
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
                    _CompletionSummary(completion: completion),
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
  const _CompletionSummary({required this.completion});

  final WorkoutCompletionState completion;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final hasActivity = completion.hasActivity;
    final label = completion.hasResult
        ? 'Completed · ${completion.workoutResult!.outcome}'
        : hasActivity
        ? 'Activity imported'
        : 'No imported activity';

    return Row(
      children: [
        Icon(
          hasActivity
              ? Icons.check_circle_outline
              : Icons.radio_button_unchecked,
          size: 16,
          color: hasActivity
              ? colorScheme.primary
              : colorScheme.onSurfaceVariant,
        ),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            label,
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: hasActivity
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
    case 'long':
      return Icons.route;
    default:
      return Icons.directions_run;
  }
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
