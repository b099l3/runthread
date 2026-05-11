import 'package:flutter/material.dart';

import 'src/api/runthread_api.dart';
import 'src/app.dart';
import 'src/demo/demo_fallback_api.dart';

void main() {
  const configuredApiBaseUrl = String.fromEnvironment(
    'RUNTHREAD_API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );
  final apiBaseUrl = _apiBaseUrl(configuredApiBaseUrl);

  runApp(
    RunthreadApp(
      api: DemoFallbackRunthreadApi(
        primary: HttpRunthreadApi(baseUrl: apiBaseUrl),
      ),
    ),
  );
}

Uri _apiBaseUrl(String configured) {
  final fallback = Uri.parse('http://localhost:8080');
  final parsed = Uri.tryParse(configured.trim());
  if (parsed == null || !parsed.hasScheme || parsed.host.isEmpty) {
    return fallback;
  }
  return parsed;
}
