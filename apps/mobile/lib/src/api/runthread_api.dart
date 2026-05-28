import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/plan_week.dart';
import '../models/provider_connection.dart';
import '../week_dates.dart';

abstract interface class RunthreadApi {
  Future<CurrentPlanWeek> getCurrentPlanWeek({DateTime? targetWeekDate});
  Future<ProviderConnectionStatusView> getProviderConnectionStatus();
  Future<StartProviderConnectionResult> startProviderConnection();
  Future<DisconnectProviderConnectionResult> disconnectProviderConnection({
    String? providerConnectionId,
  });
}

class HttpRunthreadApi implements RunthreadApi {
  HttpRunthreadApi({
    required this.baseUrl,
    required this.athleteId,
    this.goalId = 'goal-1',
    required this.stravaRedirectUri,
    this.requestTimeout = const Duration(seconds: 5),
    http.Client? client,
  }) : _client = client ?? http.Client();

  final Uri baseUrl;
  final String athleteId;
  final String goalId;
  final String stravaRedirectUri;
  final Duration requestTimeout;
  final http.Client _client;

  @override
  Future<CurrentPlanWeek> getCurrentPlanWeek({DateTime? targetWeekDate}) async {
    final endpoint = baseUrl.resolve(
      '/runthread.v1.RunthreadService/GetCurrentPlanWeek',
    );
    final response = await _postJson(endpoint, {
      'athleteId': athleteId,
      'goalId': goalId,
      'targetWeekDate': currentWeekTargetDate(now: targetWeekDate),
    });

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw RunthreadApiException(
        'Backend returned ${response.statusCode}. Seeded demo athlete and goal records are required for this MVP screen.',
      );
    }

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return CurrentPlanWeek.fromJson(decoded);
  }

  @override
  Future<ProviderConnectionStatusView> getProviderConnectionStatus() async {
    final endpoint = baseUrl.resolve(
      '/runthread.v1.RunthreadService/GetProviderConnectionStatus',
    );
    final response = await _postJson(endpoint, {
      'athleteId': athleteId,
      'provider': 'PROVIDER_STRAVA',
    });

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw RunthreadApiException(
        'Backend returned ${response.statusCode} for Strava connection status.',
      );
    }

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return ProviderConnectionStatusView.fromJson(decoded);
  }

  @override
  Future<StartProviderConnectionResult> startProviderConnection() async {
    final endpoint = baseUrl.resolve(
      '/runthread.v1.RunthreadService/StartProviderConnection',
    );
    final response = await _postJson(endpoint, {
      'athleteId': athleteId,
      'provider': 'PROVIDER_STRAVA',
      'redirectUri': stravaRedirectUri,
    });

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw RunthreadApiException(
        'Backend returned ${response.statusCode} while starting Strava connection.',
      );
    }

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return StartProviderConnectionResult.fromJson(decoded);
  }

  @override
  Future<DisconnectProviderConnectionResult> disconnectProviderConnection({
    String? providerConnectionId,
  }) async {
    final endpoint = baseUrl.resolve(
      '/runthread.v1.RunthreadService/DisconnectProviderConnection',
    );
    final body = <String, dynamic>{
      'athleteId': athleteId,
      'provider': 'PROVIDER_STRAVA',
    };
    if (providerConnectionId != null && providerConnectionId.isNotEmpty) {
      body['providerConnectionId'] = providerConnectionId;
    }
    final response = await _postJson(endpoint, body);

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw RunthreadApiException(
        'Backend returned ${response.statusCode} while disconnecting Strava.',
      );
    }

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return DisconnectProviderConnectionResult.fromJson(decoded);
  }

  Future<http.Response> _postJson(Uri endpoint, Map<String, dynamic> body) {
    return _client
        .post(
          endpoint,
          headers: const {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
          },
          body: jsonEncode(body),
        )
        .timeout(
          requestTimeout,
          onTimeout: () {
            throw RunthreadApiException(
              'Backend request timed out after ${requestTimeout.inSeconds}s.',
            );
          },
        );
  }
}

class RunthreadApiException implements Exception {
  const RunthreadApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
