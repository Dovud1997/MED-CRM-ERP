class Doctor {
  const Doctor({
    required this.id,
    required this.userId,
    required this.firstName,
    required this.lastName,
    this.middleName,
    this.position,
    this.specialty,
    this.branchId,
    this.branchName,
    this.isActive = true,
    this.publicEmail,
  });

  final String id;
  final String userId;
  final String firstName;
  final String lastName;
  final String? middleName;
  final String? position;
  final String? specialty;
  final String? branchId;
  final String? branchName;
  final bool isActive;
  final String? publicEmail;

  String get fullName =>
      [lastName, firstName, if (middleName != null && middleName!.isNotEmpty) middleName]
          .join(' ');

  String get specialtyLabel =>
      (specialty != null && specialty!.trim().isNotEmpty)
          ? specialty!.trim()
          : (position ?? '');

  factory Doctor.fromEmployeeJson(Map<String, dynamic> json) {
    return Doctor(
      id: json['id'] as String,
      userId: json['userId'] as String? ?? '',
      firstName: json['firstName'] as String? ?? '',
      lastName: json['lastName'] as String? ?? '',
      middleName: json['middleName'] as String?,
      position: json['position'] as String?,
      specialty: json['specialty'] as String?,
      branchId: json['branchId'] as String?,
      branchName: json['branch'] as String?,
      isActive: json['isActive'] as bool? ?? true,
      publicEmail: json['publicEmail'] as String?,
    );
  }

  /// Heuristic: clinical staff with specialty/position that looks like a doctor.
  static bool looksLikeDoctor(Map<String, dynamic> json) {
    if (json['isOwner'] == true) return false;
    final specialty = (json['specialty'] as String? ?? '').toLowerCase();
    final position = (json['position'] as String? ?? '').toLowerCase();
    if (specialty.isNotEmpty) return true;
    const markers = [
      'врач',
      'доктор',
      'doctor',
      'shifokor',
      'стоматолог',
      'педиатр',
      'гинеколог',
      'хирург',
      'узи',
      'логопед',
      'терапевт',
      'невролог',
      'кардиолог',
      'лор',
      'дерматолог',
      'уролог',
    ];
    return markers.any((m) => position.contains(m));
  }
}

class Specialty {
  const Specialty({
    required this.id,
    required this.name,
    this.employeeCount = 0,
  });

  final String id;
  final String name;
  final int employeeCount;

  factory Specialty.fromJson(Map<String, dynamic> json) {
    return Specialty(
      id: json['id'] as String,
      name: json['name'] as String,
      employeeCount: (json['employeeCount'] as num?)?.toInt() ?? 0,
    );
  }
}

class ScheduleDay {
  const ScheduleDay({
    required this.weekday,
    required this.isWorking,
    this.start = '',
    this.end = '',
    this.breakFrom = '',
    this.breakTo = '',
  });

  final int weekday;
  final bool isWorking;
  final String start;
  final String end;
  final String breakFrom;
  final String breakTo;

  factory ScheduleDay.fromJson(Map<String, dynamic> json) {
    return ScheduleDay(
      weekday: (json['weekday'] as num).toInt(),
      isWorking: json['isWorking'] as bool? ?? false,
      start: json['start'] as String? ?? '',
      end: json['end'] as String? ?? '',
      breakFrom: json['breakFrom'] as String? ?? '',
      breakTo: json['breakTo'] as String? ?? '',
    );
  }
}

class DoctorSchedule {
  const DoctorSchedule({
    required this.employeeId,
    required this.configured,
    required this.days,
  });

  final String employeeId;
  final bool configured;
  final List<ScheduleDay> days;

  factory DoctorSchedule.fromJson(Map<String, dynamic> json) {
    final days = (json['days'] as List<dynamic>? ?? const [])
        .map((e) => ScheduleDay.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
    return DoctorSchedule(
      employeeId: json['employeeId'] as String,
      configured: json['configured'] as bool? ?? days.isNotEmpty,
      days: days,
    );
  }

  ScheduleDay? dayForDate(DateTime date) {
    final weekday = date.weekday; // 1=Mon .. 7=Sun (ISO)
    for (final day in days) {
      if (day.weekday == weekday) return day;
    }
    return null;
  }
}
