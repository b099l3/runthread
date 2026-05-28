class ProviderConnectionStatusView {
  const ProviderConnectionStatusView({
    required this.hasConnection,
    this.connection,
    this.statusUnavailable = false,
  });

  final bool hasConnection;
  final ProviderConnection? connection;
  final bool statusUnavailable;

  factory ProviderConnectionStatusView.notConnected({
    bool statusUnavailable = false,
  }) {
    return ProviderConnectionStatusView(
      hasConnection: false,
      statusUnavailable: statusUnavailable,
    );
  }

  factory ProviderConnectionStatusView.fromJson(Map<String, dynamic> json) {
    final hasConnection = json['hasConnection'] as bool? ?? false;
    final connectionJson = json['connection'];
    return ProviderConnectionStatusView(
      hasConnection: hasConnection,
      connection: connectionJson is Map<String, dynamic>
          ? ProviderConnection.fromJson(connectionJson)
          : null,
    );
  }

  String get providerLabel => connection?.providerLabel ?? 'Strava';

  String get statusLabel {
    if (!hasConnection || connection == null) {
      return 'Not connected';
    }
    return connection!.statusLabel;
  }

  String get description {
    if (statusUnavailable) {
      return 'Connection status is unavailable. Check the backend and try again.';
    }
    if (!hasConnection || connection == null) {
      return 'Run completion will come from imported Strava activity once provider access is ready.';
    }
    return switch (connection!.status) {
      ProviderConnectionStatus.pending =>
        'Connection has started. Strava authorization remains disabled until provider access is ready.',
      ProviderConnectionStatus.connected =>
        'Runthread is ready to receive imported Strava activity for workout completion.',
      ProviderConnectionStatus.syncing =>
        'Strava activity import is in progress. Plan changes will appear after matching finishes.',
      ProviderConnectionStatus.error =>
        connection!.lastError.isEmpty
            ? 'Strava connection needs attention before activity can import.'
            : connection!.lastError,
      ProviderConnectionStatus.disconnected =>
        'Strava is disconnected. Runthread is not importing activity from this provider.',
      ProviderConnectionStatus.unspecified =>
        'Run completion will come from imported Strava activity once provider access is ready.',
    };
  }
}

class StartProviderConnectionResult {
  const StartProviderConnectionResult({
    required this.connection,
    required this.authorizationUrl,
    required this.state,
    required this.oauthReady,
  });

  final ProviderConnection connection;
  final String authorizationUrl;
  final String state;
  final bool oauthReady;

  factory StartProviderConnectionResult.fromJson(Map<String, dynamic> json) {
    final connectionJson = json['connection'];
    return StartProviderConnectionResult(
      connection: connectionJson is Map<String, dynamic>
          ? ProviderConnection.fromJson(connectionJson)
          : const ProviderConnection(
              id: '',
              athleteId: '',
              provider: Provider.unspecified,
              status: ProviderConnectionStatus.unspecified,
              lastError: '',
            ),
      authorizationUrl: json['authorizationUrl'] as String? ?? '',
      state: json['state'] as String? ?? '',
      oauthReady: json['oauthReady'] as bool? ?? false,
    );
  }
}

class DisconnectProviderConnectionResult {
  const DisconnectProviderConnectionResult({required this.connection});

  final ProviderConnection connection;

  factory DisconnectProviderConnectionResult.fromJson(
    Map<String, dynamic> json,
  ) {
    final connectionJson = json['connection'];
    return DisconnectProviderConnectionResult(
      connection: connectionJson is Map<String, dynamic>
          ? ProviderConnection.fromJson(connectionJson)
          : const ProviderConnection(
              id: '',
              athleteId: '',
              provider: Provider.unspecified,
              status: ProviderConnectionStatus.unspecified,
              lastError: '',
            ),
    );
  }
}

class ProviderConnection {
  const ProviderConnection({
    required this.id,
    required this.athleteId,
    required this.provider,
    required this.status,
    required this.lastError,
  });

  final String id;
  final String athleteId;
  final Provider provider;
  final ProviderConnectionStatus status;
  final String lastError;

  factory ProviderConnection.fromJson(Map<String, dynamic> json) {
    return ProviderConnection(
      id: json['id'] as String? ?? '',
      athleteId: json['athleteId'] as String? ?? '',
      provider: _providerFromJson(json['provider'] as String?),
      status: _statusFromJson(json['status'] as String?),
      lastError: json['lastError'] as String? ?? '',
    );
  }

  String get providerLabel => switch (provider) {
    Provider.garmin => 'Garmin',
    Provider.strava => 'Strava',
    Provider.unspecified => 'Provider',
  };

  String get statusLabel => switch (status) {
    ProviderConnectionStatus.pending => 'Pending',
    ProviderConnectionStatus.connected => 'Connected',
    ProviderConnectionStatus.syncing => 'Syncing',
    ProviderConnectionStatus.error => 'Needs attention',
    ProviderConnectionStatus.disconnected => 'Disconnected',
    ProviderConnectionStatus.unspecified => 'Not connected',
  };
}

enum Provider { unspecified, garmin, strava }

enum ProviderConnectionStatus {
  unspecified,
  pending,
  connected,
  syncing,
  error,
  disconnected,
}

Provider _providerFromJson(String? value) {
  return switch (value) {
    'PROVIDER_GARMIN' => Provider.garmin,
    'PROVIDER_STRAVA' => Provider.strava,
    _ => Provider.unspecified,
  };
}

ProviderConnectionStatus _statusFromJson(String? value) {
  return switch (value) {
    'PROVIDER_CONNECTION_STATUS_PENDING' => ProviderConnectionStatus.pending,
    'PROVIDER_CONNECTION_STATUS_CONNECTED' =>
      ProviderConnectionStatus.connected,
    'PROVIDER_CONNECTION_STATUS_SYNCING' => ProviderConnectionStatus.syncing,
    'PROVIDER_CONNECTION_STATUS_ERROR' => ProviderConnectionStatus.error,
    'PROVIDER_CONNECTION_STATUS_DISCONNECTED' =>
      ProviderConnectionStatus.disconnected,
    _ => ProviderConnectionStatus.unspecified,
  };
}
