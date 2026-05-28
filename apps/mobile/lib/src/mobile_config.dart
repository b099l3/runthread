class MobileConfig {
  MobileConfig({
    required this.apiBaseUrl,
    required this.athleteId,
    required this.goalId,
    required this.stravaRedirectUri,
  });

  factory MobileConfig.fromEnv(Map<String, String> env) {
    return MobileConfig(
      apiBaseUrl: _apiBaseUrl(env[_apiBaseUrlKey]),
      athleteId: _envValue(env, _athleteIdKey, _defaultAthleteId),
      goalId: _envValue(env, _goalIdKey, _defaultGoalId),
      stravaRedirectUri: _envValue(
        env,
        _stravaRedirectUriKey,
        _defaultStravaRedirectUri,
      ),
    );
  }

  static final defaults = MobileConfig.fromEnv(const {});

  final Uri apiBaseUrl;
  final String athleteId;
  final String goalId;
  final String stravaRedirectUri;

  static const _apiBaseUrlKey = 'RUNTHREAD_API_BASE_URL';
  static const _athleteIdKey = 'RUNTHREAD_ATHLETE_ID';
  static const _goalIdKey = 'RUNTHREAD_GOAL_ID';
  static const _stravaRedirectUriKey = 'STRAVA_OAUTH_REDIRECT_URI';

  static const _defaultApiBaseUrl = 'http://localhost:8080';
  static const _defaultAthleteId = 'athlete-1';
  static const _defaultGoalId = 'goal-1';
  static const _defaultStravaRedirectUri =
      'http://localhost:8080/providers/strava/oauth/callback';

  static Uri _apiBaseUrl(String? configured) {
    final fallback = Uri.parse(_defaultApiBaseUrl);
    final parsed = Uri.tryParse((configured ?? '').trim());
    if (parsed == null || !parsed.hasScheme || parsed.host.isEmpty) {
      return fallback;
    }
    return parsed;
  }

  static String _envValue(
    Map<String, String> env,
    String key,
    String defaultValue,
  ) {
    final configured = env[key]?.trim();
    if (configured == null || configured.isEmpty) {
      return defaultValue;
    }
    return configured;
  }
}
