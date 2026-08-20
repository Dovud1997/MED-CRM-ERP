import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/core/network/api_client.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/demo/demo_data.dart';
import 'package:clinicos_mobile/shared/models/appointment.dart';

final appointmentsRepositoryProvider = Provider<AppointmentsRepository>((ref) {
  return AppointmentsRepository(ref.watch(apiClientProvider));
});

class AppointmentsRepository {
  AppointmentsRepository(this._api);

  final ApiClient _api;

  Future<List<Appointment>> listActive() async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/appointments',
        query: {'scope': 'active'},
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      final items = (data['items'] as List<dynamic>? ?? const [])
          .map((e) => Appointment.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList();
      if (items.isNotEmpty) return items;
    } catch (_) {
      if (!kSkipLogin) rethrow;
    }
    return DemoData.appointments().where((a) => a.isUpcoming).toList();
  }

  Future<List<Appointment>> listByDate(DateTime day) async {
    try {
      final date =
          '${day.year.toString().padLeft(4, '0')}-${day.month.toString().padLeft(2, '0')}-${day.day.toString().padLeft(2, '0')}';
      final data = await _api.get<Map<String, dynamic>>(
        '/appointments',
        query: {'date': date},
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      return (data['items'] as List<dynamic>? ?? const [])
          .map((e) => Appointment.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList();
    } catch (_) {
      if (!kSkipLogin) rethrow;
      return DemoData.appointments().where((a) {
        return a.startsAt.year == day.year &&
            a.startsAt.month == day.month &&
            a.startsAt.day == day.day;
      }).toList();
    }
  }

  Future<AppointmentsDashboard> dashboard() async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/appointments/dashboard',
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      return AppointmentsDashboard.fromJson(data);
    } catch (_) {
      if (!kSkipLogin) rethrow;
      final upcoming = DemoData.appointments().where((a) => a.isUpcoming).toList();
      return AppointmentsDashboard(
        today: upcoming.length,
        waiting: upcoming.length,
        completed: 1,
        doctorsOnDuty: DemoData.doctors.length,
        upcoming: upcoming,
      );
    }
  }

  Future<Appointment> create({
    required String patientId,
    required String employeeId,
    required String branchId,
    required DateTime startsAt,
    int durationMinutes = 30,
    String reason = '',
    String notes = '',
  }) async {
    try {
      final data = await _api.post<Map<String, dynamic>>(
        '/appointments',
        data: {
          'patientId': patientId,
          'employeeId': employeeId,
          'branchId': branchId,
          'startsAt': startsAt.toUtc().toIso8601String(),
          'durationMinutes': durationMinutes,
          'reason': reason,
          'notes': notes,
        },
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      return Appointment(
        id: data['id'] as String,
        startsAt: DateTime.parse(data['startsAt'] as String).toLocal(),
        endsAt: DateTime.parse(data['endsAt'] as String).toLocal(),
        status: data['status'] as String? ?? 'scheduled',
        employeeId: employeeId,
        branchId: branchId,
        patientId: patientId,
      );
    } catch (_) {
      if (!kSkipLogin) rethrow;
      return Appointment(
        id: 'demo-local-${startsAt.millisecondsSinceEpoch}',
        startsAt: startsAt,
        endsAt: startsAt.add(Duration(minutes: durationMinutes)),
        status: 'scheduled',
        employeeId: employeeId,
        branchId: branchId,
        patientId: patientId,
        employeeName: DemoData.doctorById(employeeId)?.fullName,
        specialty: DemoData.doctorById(employeeId)?.specialtyLabel,
        branchName: 'ONA VA BOLA',
      );
    }
  }

  Future<void> cancel(String id) async {
    try {
      await _api.patch(
        '/appointments/$id/status',
        data: {'status': 'cancelled'},
      );
    } catch (_) {
      if (!kSkipLogin) rethrow;
    }
  }
}
