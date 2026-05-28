import '../api/runthread_api.dart';
import '../models/plan_week.dart';
import '../models/provider_connection.dart';
import 'demo_plan_week.dart';

class DemoFallbackRunthreadApi implements RunthreadApi {
  const DemoFallbackRunthreadApi({required this.primary});

  final RunthreadApi primary;
  static Object? lastFallbackError;

  @override
  Future<CurrentPlanWeek> getCurrentPlanWeek({DateTime? targetWeekDate}) async {
    try {
      final response = await primary.getCurrentPlanWeek(
        targetWeekDate: targetWeekDate,
      );
      lastFallbackError = null;
      return response;
    } catch (error) {
      lastFallbackError = error;
      return demoCurrentPlanWeek();
    }
  }

  @override
  Future<ProviderConnectionStatusView> getProviderConnectionStatus() async {
    try {
      return await primary.getProviderConnectionStatus();
    } catch (_) {
      return ProviderConnectionStatusView.notConnected(statusUnavailable: true);
    }
  }

  @override
  Future<StartProviderConnectionResult> startProviderConnection() {
    return primary.startProviderConnection();
  }

  @override
  Future<DisconnectProviderConnectionResult> disconnectProviderConnection({
    String? providerConnectionId,
  }) {
    return primary.disconnectProviderConnection(
      providerConnectionId: providerConnectionId,
    );
  }
}
