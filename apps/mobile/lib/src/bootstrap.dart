import 'package:flutter/material.dart';
import 'package:app_links/app_links.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:url_launcher/url_launcher.dart';

import 'api/runthread_api.dart';
import 'app.dart';
import 'demo/demo_fallback_api.dart';
import 'mobile_config.dart';

Future<void> bootstrapRunthreadApp({String? envAssetPath}) async {
  WidgetsFlutterBinding.ensureInitialized();

  if (envAssetPath != null) {
    await dotenv.load(fileName: envAssetPath);
  }

  final config = MobileConfig.fromEnv(dotenv.env);
  final appLinks = AppLinks();

  runApp(
    RunthreadApp(
      api: DemoFallbackRunthreadApi(
        primary: HttpRunthreadApi(
          baseUrl: config.apiBaseUrl,
          athleteId: config.athleteId,
          goalId: config.goalId,
          stravaRedirectUri: config.stravaRedirectUri,
        ),
      ),
      openUrl: _openExternalUrl,
      deepLinks: appLinks.uriLinkStream,
    ),
  );
}

Future<bool> _openExternalUrl(Uri uri) {
  return launchUrl(uri, mode: LaunchMode.externalApplication);
}
