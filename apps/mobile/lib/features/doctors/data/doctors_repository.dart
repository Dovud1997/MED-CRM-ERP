import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/core/network/api_client.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/demo/demo_data.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';

final doctorsRepositoryProvider = Provider<DoctorsRepository>((ref) {
  return DoctorsRepository(ref.watch(apiClientProvider));
});

class DoctorsRepository {
  DoctorsRepository(this._api);

  final ApiClient _api;

  Future<List<Doctor>> listDoctors({String? query, String? specialty}) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/employees',
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      var items = (data['items'] as List<dynamic>? ?? const [])
          .map((e) => Map<String, dynamic>.from(e as Map))
          .where(Doctor.looksLikeDoctor)
          .map(Doctor.fromEmployeeJson)
          .where((d) => d.isActive)
          .toList();

      if (specialty != null && specialty.trim().isNotEmpty) {
        final s = specialty.toLowerCase();
        items = items
            .where((d) => d.specialtyLabel.toLowerCase().contains(s))
            .toList();
      }
      if (query != null && query.trim().isNotEmpty) {
        final q = query.toLowerCase();
        items = items
            .where(
              (d) =>
                  d.fullName.toLowerCase().contains(q) ||
                  d.specialtyLabel.toLowerCase().contains(q) ||
                  (d.branchName ?? '').toLowerCase().contains(q),
            )
            .toList();
      }
      items.sort((a, b) => a.fullName.compareTo(b.fullName));
      if (items.isNotEmpty) return items;
    } catch (_) {
      if (!kSkipLogin) rethrow;
    }
    return DemoData.filterDoctors(query: query, specialty: specialty);
  }

  Future<Doctor?> getDoctor(String id) async {
    final demo = DemoData.doctorById(id);
    try {
      final all = await listDoctors();
      for (final d in all) {
        if (d.id == id) return d;
      }
    } catch (_) {
      if (!kSkipLogin) rethrow;
    }
    return demo;
  }

  Future<List<Specialty>> listSpecialties() async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/specialists',
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      final items = (data['items'] as List<dynamic>? ?? const [])
          .map((e) => Specialty.fromJson(Map<String, dynamic>.from(e as Map)))
          .where((s) => s.name.isNotEmpty)
          .toList();
      if (items.isNotEmpty) return items;
    } catch (_) {
      if (!kSkipLogin) rethrow;
    }
    return DemoData.specialties;
  }

  Future<DoctorSchedule> getSchedule(String employeeId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/doctor-schedules/$employeeId',
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      return DoctorSchedule.fromJson(data);
    } catch (_) {
      if (!kSkipLogin) rethrow;
      return DoctorSchedule(
        employeeId: employeeId,
        configured: true,
        days: [
          for (var w = 1; w <= 5; w++)
            ScheduleDay(
              weekday: w,
              isWorking: true,
              start: '09:00',
              end: '17:00',
              breakFrom: '13:00',
              breakTo: '14:00',
            ),
          const ScheduleDay(weekday: 6, isWorking: false),
          const ScheduleDay(weekday: 7, isWorking: false),
        ],
      );
    }
  }

  Future<List<DateTime>> availabilitySlots({
    required String employeeId,
    required DateTime day,
    int slotMinutes = 30,
  }) async {
    try {
      final schedule = await getSchedule(employeeId);
      final daySchedule = schedule.dayForDate(day);
      if (daySchedule == null || !daySchedule.isWorking) {
        if (kSkipLogin && employeeId.startsWith('demo-')) {
          return DemoData.demoSlots(day);
        }
        return const [];
      }

      final start = _combine(day, daySchedule.start);
      final end = _combine(day, daySchedule.end);
      if (start == null || end == null || !end.isAfter(start)) {
        return kSkipLogin ? DemoData.demoSlots(day) : const [];
      }

      DateTime? breakStart;
      DateTime? breakEnd;
      if (daySchedule.breakFrom.isNotEmpty && daySchedule.breakTo.isNotEmpty) {
        breakStart = _combine(day, daySchedule.breakFrom);
        breakEnd = _combine(day, daySchedule.breakTo);
      }

      final busy = await _busyIntervals(employeeId, day);
      final slots = <DateTime>[];
      var cursor = start;
      final now = DateTime.now();
      while (cursor.add(Duration(minutes: slotMinutes)).compareTo(end) <= 0) {
        final slotEnd = cursor.add(Duration(minutes: slotMinutes));
        final inBreak = breakStart != null &&
            breakEnd != null &&
            cursor.isBefore(breakEnd) &&
            slotEnd.isAfter(breakStart);
        final overlapsBusy = busy.any(
          (b) => cursor.isBefore(b.$2) && slotEnd.isAfter(b.$1),
        );
        if (!inBreak && !overlapsBusy && slotEnd.isAfter(now)) {
          slots.add(cursor);
        }
        cursor = slotEnd;
      }
      if (slots.isNotEmpty) return slots;
    } catch (_) {
      if (!kSkipLogin) rethrow;
    }
    return DemoData.demoSlots(day);
  }

  Future<List<(DateTime, DateTime)>> _busyIntervals(
    String employeeId,
    DateTime day,
  ) async {
    try {
      final date = _fmtDate(day);
      final data = await _api.get<Map<String, dynamic>>(
        '/appointments',
        query: {'date': date},
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      final items = data['items'] as List<dynamic>? ?? const [];
      final result = <(DateTime, DateTime)>[];
      for (final raw in items) {
        final map = Map<String, dynamic>.from(raw as Map);
        final employee = map['employee'];
        final empId = employee is Map ? employee['id']?.toString() : null;
        if (empId != employeeId) continue;
        final status = map['status'] as String? ?? '';
        if (status == 'cancelled' || status == 'no_show') continue;
        final starts = DateTime.parse(map['startsAt'] as String).toLocal();
        final ends = DateTime.parse(map['endsAt'] as String).toLocal();
        result.add((starts, ends));
      }
      return result;
    } catch (_) {
      return const [];
    }
  }

  DateTime? _combine(DateTime day, String hhmm) {
    final parts = hhmm.split(':');
    if (parts.length != 2) return null;
    final h = int.tryParse(parts[0]);
    final m = int.tryParse(parts[1]);
    if (h == null || m == null) return null;
    return DateTime(day.year, day.month, day.day, h, m);
  }

  String _fmtDate(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
}
