import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/plan_week.dart';

abstract interface class RunthreadApi {
  Future<CurrentPlanWeek> getCurrentPlanWeek();
}

class HttpRunthreadApi implements RunthreadApi {
  HttpRunthreadApi({required this.baseUrl, http.Client? client})
    : _client = client ?? http.Client();

  final Uri baseUrl;
  final http.Client _client;

  static const _demoAthleteId = 'athlete-1';
  static const _demoGoalId = 'goal-1';
  static const _demoTargetWeekDate = '2026-06-03';

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
      body: jsonEncode(const {
        'athleteId': _demoAthleteId,
        'goalId': _demoGoalId,
        'targetWeekDate': _demoTargetWeekDate,
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
}

class RunthreadApiException implements Exception {
  const RunthreadApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
