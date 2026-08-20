class Appointment {
  const Appointment({
    required this.id,
    required this.startsAt,
    required this.endsAt,
    required this.status,
    this.reason,
    this.notes,
    this.patientId,
    this.patientName,
    this.employeeId,
    this.employeeName,
    this.specialty,
    this.branchId,
    this.branchName,
  });

  final String id;
  final DateTime startsAt;
  final DateTime endsAt;
  final String status;
  final String? reason;
  final String? notes;
  final String? patientId;
  final String? patientName;
  final String? employeeId;
  final String? employeeName;
  final String? specialty;
  final String? branchId;
  final String? branchName;

  bool get isUpcoming =>
      startsAt.isAfter(DateTime.now()) &&
      !const {'completed', 'cancelled', 'no_show'}.contains(status);

  bool get isPast =>
      endsAt.isBefore(DateTime.now()) ||
      const {'completed', 'no_show'}.contains(status);

  bool get isCancelled => status == 'cancelled';

  factory Appointment.fromJson(Map<String, dynamic> json) {
    final patient = json['patient'];
    final employee = json['employee'];
    final branch = json['branch'];

    String? patientId;
    String? patientName;
    if (patient is Map) {
      patientId = patient['id']?.toString();
      patientName = patient['name']?.toString();
    } else if (patient is String) {
      patientName = patient;
    }

    String? employeeId;
    String? employeeName;
    String? specialty;
    if (employee is Map) {
      employeeId = employee['id']?.toString();
      employeeName = employee['name']?.toString();
      specialty = employee['specialty']?.toString();
    } else if (employee is String) {
      employeeName = employee;
    }

    String? branchId;
    String? branchName;
    if (branch is Map) {
      branchId = branch['id']?.toString();
      branchName = branch['name']?.toString();
    } else if (branch is String) {
      branchName = branch;
    }

    final starts = DateTime.parse(json['startsAt'] as String).toLocal();
    final endsRaw = json['endsAt'];
    final ends = endsRaw == null
        ? starts.add(const Duration(minutes: 30))
        : DateTime.parse(endsRaw as String).toLocal();

    return Appointment(
      id: json['id'] as String,
      startsAt: starts,
      endsAt: ends,
      status: json['status'] as String? ?? 'scheduled',
      reason: json['reason'] as String?,
      notes: json['notes'] as String?,
      patientId: patientId,
      patientName: patientName,
      employeeId: employeeId,
      employeeName: employeeName ?? json['employee']?.toString(),
      specialty: specialty ?? json['specialty'] as String?,
      branchId: branchId,
      branchName: branchName,
    );
  }
}

class AppointmentsDashboard {
  const AppointmentsDashboard({
    required this.today,
    required this.waiting,
    required this.completed,
    required this.doctorsOnDuty,
    required this.upcoming,
  });

  final int today;
  final int waiting;
  final int completed;
  final int doctorsOnDuty;
  final List<Appointment> upcoming;

  factory AppointmentsDashboard.fromJson(Map<String, dynamic> json) {
    final upcoming = (json['upcoming'] as List<dynamic>? ?? const [])
        .map((e) {
          final map = Map<String, dynamic>.from(e as Map);
          // Dashboard upcoming uses flat fields.
          return Appointment(
            id: map['id'] as String,
            startsAt: DateTime.parse(map['startsAt'] as String).toLocal(),
            endsAt: DateTime.parse(map['startsAt'] as String)
                .toLocal()
                .add(const Duration(minutes: 30)),
            status: map['status'] as String? ?? 'scheduled',
            reason: map['reason'] as String?,
            patientName: map['patient'] as String?,
            employeeName: map['employee'] as String?,
            specialty: map['specialty'] as String?,
            branchName: map['branch'] as String?,
          );
        })
        .toList();
    return AppointmentsDashboard(
      today: (json['today'] as num?)?.toInt() ?? 0,
      waiting: (json['waiting'] as num?)?.toInt() ?? 0,
      completed: (json['completed'] as num?)?.toInt() ?? 0,
      doctorsOnDuty: (json['doctorsOnDuty'] as num?)?.toInt() ?? 0,
      upcoming: upcoming,
    );
  }
}
