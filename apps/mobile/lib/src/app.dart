import 'package:flutter/material.dart';

import 'api/runthread_api.dart';
import 'plan_week_screen.dart';

class RunthreadApp extends StatelessWidget {
  const RunthreadApp({required this.api, super.key});

  final RunthreadApi api;

  @override
  Widget build(BuildContext context) {
    const ink = Color(0xFF18211F);
    const surface = Color(0xFFF7F5EF);
    const primary = Color(0xFF1D6B59);

    return MaterialApp(
      title: 'Runthread',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: primary,
          brightness: Brightness.light,
          surface: surface,
        ),
        scaffoldBackgroundColor: surface,
        appBarTheme: const AppBarTheme(
          backgroundColor: surface,
          foregroundColor: ink,
          elevation: 0,
          centerTitle: false,
        ),
        cardTheme: CardThemeData(
          color: Colors.white,
          elevation: 0,
          margin: EdgeInsets.zero,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
            side: const BorderSide(color: Color(0xFFE2DFD4)),
          ),
        ),
        useMaterial3: true,
      ),
      home: PlanWeekScreen(api: api),
    );
  }
}
