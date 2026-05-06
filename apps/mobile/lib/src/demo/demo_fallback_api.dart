import '../api/runthread_api.dart';
import '../models/plan_week.dart';
import 'demo_plan_week.dart';

class DemoFallbackRunthreadApi implements RunthreadApi {
  const DemoFallbackRunthreadApi({required this.primary});

  final RunthreadApi primary;

  @override
  Future<CurrentPlanWeek> getCurrentPlanWeek() async {
    try {
      return await primary.getCurrentPlanWeek();
    } catch (_) {
      return demoCurrentPlanWeek();
    }
  }
}
