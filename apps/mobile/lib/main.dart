import 'package:flutter/material.dart';

import 'src/api/runthread_api.dart';
import 'src/app.dart';
import 'src/demo/demo_fallback_api.dart';

void main() {
  const apiBaseUrl = String.fromEnvironment(
    'RUNTHREAD_API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  runApp(
    RunthreadApp(
      api: DemoFallbackRunthreadApi(
        primary: HttpRunthreadApi(baseUrl: Uri.parse(apiBaseUrl)),
      ),
    ),
  );
}
