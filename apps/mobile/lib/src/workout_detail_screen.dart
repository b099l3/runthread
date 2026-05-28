import 'package:flutter/material.dart';

import 'models/distance_unit.dart';
import 'models/plan_week.dart';

class WorkoutDetailScreen extends StatelessWidget {
  const WorkoutDetailScreen({
    required this.workout,
    required this.completion,
    required this.distanceUnit,
    super.key,
  });

  final PlannedWorkout workout;
  final WorkoutCompletionState completion;
  final DistanceUnit distanceUnit;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final distanceLabel = workout.distanceLabel(distanceUnit);
    return Scaffold(
      appBar: AppBar(title: const Text('Workout')),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 28),
          children: [
            Text(
              workout.type,
              style: textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              workout.formattedScheduledDate,
              style: textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 18),
            _DetailSection(
              children: [
                _DetailRow(label: 'Status', value: workout.status),
                _DetailRow(
                  label: 'Target distance',
                  value: distanceLabel.isEmpty ? 'None' : distanceLabel,
                ),
                _DetailRow(
                  label: 'Target duration',
                  value: workout.durationLabel.isEmpty
                      ? 'None'
                      : workout.durationLabel,
                ),
                _DetailRow(
                  label: 'Intensity',
                  value: workout.intensityLabel.isEmpty
                      ? 'None'
                      : workout.intensityLabel,
                ),
              ],
            ),
            const SizedBox(height: 14),
            _CompletionActionSection(completion: completion),
            const SizedBox(height: 14),
            _CompletionSection(
              completion: completion,
              distanceUnit: distanceUnit,
            ),
            const SizedBox(height: 14),
            _DetailSection(
              children: [
                Text(
                  'Notes',
                  style: textTheme.labelLarge?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  workout.notes.isEmpty ? 'No notes yet.' : workout.notes,
                  style: textTheme.bodyMedium,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _CompletionActionSection extends StatelessWidget {
  const _CompletionActionSection({required this.completion});

  final WorkoutCompletionState completion;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final hasResult = completion.hasResult;

    return _DetailSection(
      children: [
        Row(
          children: [
            Icon(Icons.sync, size: 20, color: colorScheme.primary),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Completion source',
                style: textTheme.labelLarge?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Text(
          hasResult
              ? 'This workout is completed from an imported activity match.'
              : 'Real completion will come from imported Strava activity later.',
          style: textTheme.bodyMedium?.copyWith(
            color: colorScheme.onSurfaceVariant,
          ),
        ),
        const SizedBox(height: 12),
        Align(
          alignment: Alignment.centerLeft,
          child: FilledButton.icon(
            onPressed: null,
            icon: Icon(hasResult ? Icons.check_circle_outline : Icons.lock),
            label: Text(
              hasResult ? 'Completed from activity' : 'Mark complete',
            ),
          ),
        ),
      ],
    );
  }
}

class _CompletionSection extends StatelessWidget {
  const _CompletionSection({
    required this.completion,
    required this.distanceUnit,
  });

  final WorkoutCompletionState completion;
  final DistanceUnit distanceUnit;

  @override
  Widget build(BuildContext context) {
    if (!completion.hasActivity && !completion.hasResult) {
      return _DetailSection(
        children: [
          Text(
            'Imported activity',
            style: Theme.of(
              context,
            ).textTheme.labelLarge?.copyWith(fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 8),
          Text(
            'No imported activity has been matched to this workout yet.',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      );
    }

    final activity = completion.importedActivity;
    final result = completion.workoutResult;
    final match = completion.workoutMatch;
    final activityDistanceLabel = activity?.distanceLabel(distanceUnit) ?? '';

    return _DetailSection(
      children: [
        Text(
          'Imported activity',
          style: Theme.of(
            context,
          ).textTheme.labelLarge?.copyWith(fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 8),
        if (result != null)
          _DetailRow(label: 'Completion', value: result.outcome),
        if (match != null) _DetailRow(label: 'Match', value: match.status),
        if (match != null && match.confidence.isNotEmpty)
          _DetailRow(label: 'Confidence', value: match.confidence),
        if (activity != null) ...[
          _DetailRow(label: 'Activity type', value: activity.type),
          _DetailRow(
            label: 'Activity distance',
            value: activityDistanceLabel.isEmpty
                ? 'None'
                : activityDistanceLabel,
          ),
          _DetailRow(
            label: 'Activity duration',
            value: activity.durationLabel.isEmpty
                ? 'None'
                : activity.durationLabel,
          ),
        ],
      ],
    );
  }
}

class _DetailSection extends StatelessWidget {
  const _DetailSection({required this.children});

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: children,
        ),
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  const _DetailRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 7),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 128,
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: Theme.of(
                context,
              ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
            ),
          ),
        ],
      ),
    );
  }
}
