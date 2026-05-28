import 'dart:async';

import 'package:flutter/material.dart';

import 'api/runthread_api.dart';
import 'deep_links.dart';
import 'models/distance_unit.dart';
import 'models/provider_connection.dart';

typedef UrlOpener = Future<bool> Function(Uri uri);

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({
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
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen>
    with WidgetsBindingObserver {
  late Future<ProviderConnectionStatusView> _status;
  late DistanceUnit _distanceUnit;
  bool _connecting = false;
  bool _disconnecting = false;
  String? _message;
  StreamSubscription<Uri>? _deepLinkSubscription;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _distanceUnit = widget.distanceUnit;
    _status = _loadStatus();
    _deepLinkSubscription = widget.deepLinks?.listen(_handleDeepLink);
  }

  @override
  void didUpdateWidget(covariant SettingsScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.distanceUnit != oldWidget.distanceUnit) {
      _distanceUnit = widget.distanceUnit;
    }
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
      _refresh();
    }
  }

  Future<ProviderConnectionStatusView> _loadStatus() {
    return widget.api.getProviderConnectionStatus().catchError(
      (_) => ProviderConnectionStatusView.notConnected(statusUnavailable: true),
    );
  }

  void _refresh() {
    setState(() {
      _status = _loadStatus();
    });
  }

  void _handleDeepLink(Uri uri) {
    if (!mounted || !isStravaConnectionDeepLink(uri)) {
      return;
    }
    setState(() {
      _message = uri.pathSegments.length >= 2 && uri.pathSegments[1] == 'error'
          ? 'Strava authorization did not complete.'
          : 'Strava connection updated.';
      _status = _loadStatus();
    });
  }

  void _setDistanceUnit(DistanceUnit unit) {
    setState(() {
      _distanceUnit = unit;
    });
    widget.onDistanceUnitChanged(unit);
  }

  Future<void> _connectStrava() async {
    setState(() {
      _connecting = true;
      _message = null;
    });
    try {
      final result = await widget.api.startProviderConnection();
      final authorizationUri = Uri.tryParse(result.authorizationUrl);
      if (!result.oauthReady ||
          result.authorizationUrl.isEmpty ||
          authorizationUri == null) {
        setState(() {
          _message = 'Strava authorization is not ready yet.';
        });
        return;
      }
      final opened = await widget.openUrl(authorizationUri);
      if (!opened) {
        setState(() {
          _message = 'Could not open Strava authorization.';
        });
        return;
      }
      setState(() {
        _message = 'Return here after authorizing Strava.';
        _status = Future.value(
          ProviderConnectionStatusView(
            hasConnection: true,
            connection: result.connection,
          ),
        );
      });
    } catch (error) {
      setState(() {
        _message = error.toString();
      });
    } finally {
      if (mounted) {
        setState(() {
          _connecting = false;
        });
      }
    }
  }

  Future<void> _disconnectStrava(String providerConnectionId) async {
    setState(() {
      _disconnecting = true;
      _message = null;
    });
    try {
      final result = await widget.api.disconnectProviderConnection(
        providerConnectionId: providerConnectionId,
      );
      setState(() {
        _message = 'Strava disconnected.';
        _status = Future.value(
          ProviderConnectionStatusView(
            hasConnection: true,
            connection: result.connection,
          ),
        );
      });
    } catch (error) {
      setState(() {
        _message = error.toString();
      });
    } finally {
      if (mounted) {
        setState(() {
          _disconnecting = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: SafeArea(
        child: FutureBuilder<ProviderConnectionStatusView>(
          future: _status,
          builder: (context, snapshot) {
            final status =
                snapshot.data ??
                ProviderConnectionStatusView.notConnected(
                  statusUnavailable: snapshot.hasError,
                );
            return ListView(
              padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
              children: [
                Text(
                  'Preferences',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Choose how distances are shown in the app.',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 16),
                _DistanceUnitTile(
                  distanceUnit: _distanceUnit,
                  onDistanceUnitChanged: _setDistanceUnit,
                ),
                const SizedBox(height: 24),
                Text(
                  'Connections',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Manage activity providers for automatic workout completion.',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 16),
                _ProviderConnectionTile(
                  providerName: 'Strava',
                  status: status,
                  loading: snapshot.connectionState == ConnectionState.waiting,
                  connecting: _connecting,
                  disconnecting: _disconnecting,
                  message: _message,
                  onRefresh: _refresh,
                  onConnect: _connectStrava,
                  onDisconnect: _disconnectStrava,
                ),
                const SizedBox(height: 12),
                const _DisabledProviderTile(
                  providerName: 'Garmin',
                  message: 'Garmin access is pending and is disabled for beta.',
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _DistanceUnitTile extends StatelessWidget {
  const _DistanceUnitTile({
    required this.distanceUnit,
    required this.onDistanceUnitChanged,
  });

  final DistanceUnit distanceUnit;
  final ValueChanged<DistanceUnit> onDistanceUnitChanged;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Row(
          children: [
            Icon(Icons.straighten, size: 20, color: colorScheme.primary),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'Distance units',
                style: textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
            SegmentedButton<DistanceUnit>(
              showSelectedIcon: false,
              segments: const [
                ButtonSegment(
                  value: DistanceUnit.kilometers,
                  label: Text('km'),
                ),
                ButtonSegment(value: DistanceUnit.miles, label: Text('mi')),
              ],
              selected: {distanceUnit},
              onSelectionChanged: (selection) {
                onDistanceUnitChanged(selection.single);
              },
            ),
          ],
        ),
      ),
    );
  }
}

class _ProviderConnectionTile extends StatelessWidget {
  const _ProviderConnectionTile({
    required this.providerName,
    required this.status,
    required this.loading,
    required this.connecting,
    required this.disconnecting,
    required this.onRefresh,
    required this.onConnect,
    required this.onDisconnect,
    this.message,
  });

  final String providerName;
  final ProviderConnectionStatusView status;
  final bool loading;
  final bool connecting;
  final bool disconnecting;
  final String? message;
  final VoidCallback onRefresh;
  final VoidCallback onConnect;
  final ValueChanged<String> onDisconnect;

  bool get _canConnect {
    if (loading || connecting || disconnecting) {
      return false;
    }
    if (!status.hasConnection || status.connection == null) {
      return true;
    }
    return switch (status.connection!.status) {
      ProviderConnectionStatus.error ||
      ProviderConnectionStatus.disconnected ||
      ProviderConnectionStatus.unspecified => true,
      ProviderConnectionStatus.pending ||
      ProviderConnectionStatus.syncing ||
      ProviderConnectionStatus.connected => false,
    };
  }

  bool get _showConnect {
    if (!status.hasConnection || status.connection == null) {
      return true;
    }
    return switch (status.connection!.status) {
      ProviderConnectionStatus.error ||
      ProviderConnectionStatus.disconnected ||
      ProviderConnectionStatus.unspecified => true,
      ProviderConnectionStatus.pending ||
      ProviderConnectionStatus.syncing ||
      ProviderConnectionStatus.connected => false,
    };
  }

  bool get _showDisconnect {
    if (!status.hasConnection || status.connection == null) {
      return false;
    }
    return switch (status.connection!.status) {
      ProviderConnectionStatus.pending ||
      ProviderConnectionStatus.connected ||
      ProviderConnectionStatus.syncing ||
      ProviderConnectionStatus.error => true,
      ProviderConnectionStatus.disconnected ||
      ProviderConnectionStatus.unspecified => false,
    };
  }

  bool get _canDisconnect {
    return _showDisconnect &&
        !loading &&
        !connecting &&
        !disconnecting &&
        status.connection!.id.isNotEmpty;
  }

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
                  Icons.directions_run,
                  size: 20,
                  color: colorScheme.primary,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    providerName,
                    style: textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                _StatusPill(label: loading ? 'Loading' : status.statusLabel),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              loading
                  ? 'Checking Strava connection status.'
                  : status.description,
              style: textTheme.bodyMedium?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            if (message != null) ...[
              const SizedBox(height: 8),
              Text(
                message!,
                style: textTheme.bodySmall?.copyWith(
                  color: colorScheme.primary,
                ),
              ),
            ],
            const SizedBox(height: 12),
            Wrap(
              spacing: 10,
              runSpacing: 8,
              children: [
                if (_showConnect)
                  FilledButton.icon(
                    onPressed: _canConnect ? onConnect : null,
                    icon: connecting
                        ? const SizedBox.square(
                            dimension: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.link),
                    label: Text(
                      connecting ? 'Opening Strava' : 'Connect Strava',
                    ),
                  ),
                if (_showDisconnect)
                  OutlinedButton.icon(
                    onPressed: _canDisconnect
                        ? () => onDisconnect(status.connection!.id)
                        : null,
                    icon: disconnecting
                        ? const SizedBox.square(
                            dimension: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.link_off),
                    label: Text(disconnecting ? 'Disconnecting' : 'Disconnect'),
                  ),
                OutlinedButton.icon(
                  onPressed: loading ? null : onRefresh,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _DisabledProviderTile extends StatelessWidget {
  const _DisabledProviderTile({
    required this.providerName,
    required this.message,
  });

  final String providerName;
  final String message;

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
                  color: colorScheme.onSurfaceVariant,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    providerName,
                    style: textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                const _StatusPill(label: 'Disabled'),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: textTheme.bodyMedium?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        color: colorScheme.surfaceContainerHighest,
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        child: Text(
          label,
          style: Theme.of(context).textTheme.labelMedium?.copyWith(
            color: colorScheme.onSurfaceVariant,
            fontWeight: FontWeight.w700,
          ),
        ),
      ),
    );
  }
}
