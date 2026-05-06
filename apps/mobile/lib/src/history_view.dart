import 'package:flutter/material.dart';

import 'models/plan_week.dart';

class HistoryView extends StatelessWidget {
  const HistoryView({required this.currentPlanWeek, super.key});

  final CurrentPlanWeek currentPlanWeek;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
      children: [
        Text(
          'History',
          style: Theme.of(
            context,
          ).textTheme.headlineMedium?.copyWith(fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 6),
        Text(
          'Recent activity and plan changes',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        const SizedBox(height: 20),
        if (currentPlanWeek.planWeek.isDemo) ...[
          const _DemoNotice(),
          const SizedBox(height: 14),
        ],
        _ActivityHistory(activities: currentPlanWeek.importedActivities),
        const SizedBox(height: 14),
        _AdaptationHistory(events: currentPlanWeek.adaptationEvents),
      ],
    );
  }
}

class _DemoNotice extends StatelessWidget {
  const _DemoNotice();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
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
                'Demo data · local backend not ready',
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

class _ActivityHistory extends StatelessWidget {
  const _ActivityHistory({required this.activities});

  final List<ImportedActivity> activities;

  @override
  Widget build(BuildContext context) {
    return _HistorySection(
      icon: Icons.directions_run,
      title: 'Recent activity',
      emptyText: 'No imported activities yet.',
      isEmpty: activities.isEmpty,
      children: [
        for (final activity in activities) ...[
          Text(
            activity.type.isEmpty ? 'Imported activity' : activity.type,
            style: Theme.of(
              context,
            ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 4),
          Text(
            [
              if (activity.distanceLabel.isNotEmpty) activity.distanceLabel,
              if (activity.durationLabel.isNotEmpty) activity.durationLabel,
            ].join(' · '),
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
          ),
          if (activity != activities.last) const SizedBox(height: 12),
        ],
      ],
    );
  }
}

class _AdaptationHistory extends StatelessWidget {
  const _AdaptationHistory({required this.events});

  final List<AdaptationEvent> events;

  @override
  Widget build(BuildContext context) {
    return _HistorySection(
      icon: Icons.tune,
      title: 'Adaptation history',
      emptyText: 'No adaptations yet.',
      isEmpty: events.isEmpty,
      children: [
        for (final event in events) ...[
          Text(
            event.summary.isEmpty ? event.type : event.summary,
            style: Theme.of(
              context,
            ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
          ),
          if (event.reason.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              event.reason,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ],
          for (final change in event.changes) ...[
            const SizedBox(height: 6),
            Text(
              change.description.isEmpty ? change.type : change.description,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ],
          if (event != events.last) const SizedBox(height: 12),
        ],
      ],
    );
  }
}

class _HistorySection extends StatelessWidget {
  const _HistorySection({
    required this.icon,
    required this.title,
    required this.emptyText,
    required this.isEmpty,
    required this.children,
  });

  final IconData icon;
  final String title;
  final String emptyText;
  final bool isEmpty;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 20, color: colorScheme.primary),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            if (isEmpty)
              Text(
                emptyText,
                style: textTheme.bodyMedium?.copyWith(
                  color: colorScheme.onSurfaceVariant,
                ),
              )
            else
              ...children,
          ],
        ),
      ),
    );
  }
}
