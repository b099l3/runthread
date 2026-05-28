import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'api/runthread_api.dart';
import 'models/distance_unit.dart';
import 'plan_week_screen.dart';
import 'settings_screen.dart';

class RunthreadApp extends StatefulWidget {
  const RunthreadApp({
    required this.api,
    this.openUrl = _defaultOpenUrl,
    this.deepLinks,
    super.key,
  });

  final RunthreadApi api;
  final UrlOpener openUrl;
  final Stream<Uri>? deepLinks;

  @override
  State<RunthreadApp> createState() => _RunthreadAppState();
}

class _RunthreadAppState extends State<RunthreadApp> {
  DistanceUnit _distanceUnit = DistanceUnit.kilometers;

  @override
  void initState() {
    super.initState();
    _loadDistanceUnit();
  }

  Future<void> _loadDistanceUnit() async {
    final preferences = await SharedPreferences.getInstance();
    if (!mounted) {
      return;
    }
    setState(() {
      _distanceUnit = distanceUnitFromStorageValue(
        preferences.getString(distanceUnitPreferenceKey),
      );
    });
  }

  Future<void> _setDistanceUnit(DistanceUnit unit) async {
    setState(() {
      _distanceUnit = unit;
    });
    final preferences = await SharedPreferences.getInstance();
    await preferences.setString(distanceUnitPreferenceKey, unit.storageValue);
  }

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
      home: PlanWeekScreen(
        api: widget.api,
        openUrl: widget.openUrl,
        deepLinks: widget.deepLinks,
        distanceUnit: _distanceUnit,
        onDistanceUnitChanged: _setDistanceUnit,
      ),
    );
  }
}

Future<bool> _defaultOpenUrl(Uri uri) async => false;
