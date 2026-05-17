import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/plan_week.dart';
import '../models/provider_connection.dart';
import '../week_dates.dart';

abstract interface class RunthreadApi {
  Future<CurrentPlanWeek> getCurrentPlanWeek();
  Future<ProviderConnectionStatusView> getProviderConnectionStatus();
}

class HttpRunthreadApi implements RunthreadApi {
  HttpRunthreadApi({required this.baseUrl, http.Client? client})
    : _client = client ?? http.Client();

  final Uri baseUrl;
  final http.Client _client;

  static const _demoAthleteId = 'athlete-1';
  static const _demoGoalId = 'goal-1';

  @override
  Future<CurrentPlanWeek> getCurrentPlanWeek() async {
    final endpoint = baseUrl.resolve(
      '/runthread.v1.RunthreadService/GetCurrentPlanWeek',
    );
    final response = await _client.post(
      endpoint,
      headers: const {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: jsonEncode({
        'athleteId': _demoAthleteId,
        'goalId': _demoGoalId,
        'targetWeekDate': currentWeekTargetDate(),
      }),
    );

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
    final response = await _client.post(
      endpoint,
      headers: const {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: jsonEncode(const {
        'athleteId': _demoAthleteId,
        'provider': 'PROVIDER_STRAVA',
      }),
    );

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw RunthreadApiException(
        'Backend returned ${response.statusCode} for Strava connection status.',
      );
    }

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;
    return ProviderConnectionStatusView.fromJson(decoded);
  }
}

class RunthreadApiException implements Exception {
  const RunthreadApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
