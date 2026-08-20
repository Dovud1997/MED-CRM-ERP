import 'package:clinicos_mobile/shared/models/appointment.dart';
import 'package:clinicos_mobile/shared/models/clinical_record.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';

/// Local preview data while `kSkipLogin` / unauthenticated API returns 401.
class DemoData {
  static const doctors = [
    Doctor(
      id: 'demo-doc-1',
      userId: 'demo-u-1',
      firstName: 'Елизавета',
      lastName: 'Давидова',
      specialty: 'Кардиолог',
      position: 'Врач',
      branchId: 'demo-branch',
      branchName: 'ONA VA BOLA',
    ),
    Doctor(
      id: 'demo-doc-2',
      userId: 'demo-u-2',
      firstName: 'Натан',
      lastName: 'Хейл',
      specialty: 'Невролог',
      position: 'Врач',
      branchId: 'demo-branch',
      branchName: 'ONA VA BOLA',
    ),
    Doctor(
      id: 'demo-doc-3',
      userId: 'demo-u-3',
      firstName: 'Азиза',
      lastName: 'Каримова',
      specialty: 'Педиатр',
      position: 'Врач',
      branchId: 'demo-branch',
      branchName: 'ONA VA BOLA',
    ),
    Doctor(
      id: 'demo-doc-4',
      userId: 'demo-u-4',
      firstName: 'Жасур',
      lastName: 'Рахимов',
      specialty: 'Стоматолог',
      position: 'Врач',
      branchId: 'demo-branch',
      branchName: 'ONA VA BOLA',
    ),
    Doctor(
      id: 'demo-doc-5',
      userId: 'demo-u-5',
      firstName: 'Мария',
      lastName: 'Иванова',
      specialty: 'Терапевт',
      position: 'Врач',
      branchId: 'demo-branch',
      branchName: 'ONA VA BOLA',
    ),
  ];

  static List<Specialty> get specialties => const [
        Specialty(id: 's1', name: 'Невролог'),
        Specialty(id: 's2', name: 'Педиатр'),
        Specialty(id: 's3', name: 'Кардиолог'),
        Specialty(id: 's4', name: 'Стоматолог'),
        Specialty(id: 's5', name: 'Терапевт'),
      ];

  static List<Appointment> appointments() {
    final now = DateTime.now();
    return [
      Appointment(
        id: 'demo-appt-1',
        startsAt: now.add(const Duration(days: 1, hours: 2)),
        endsAt: now.add(const Duration(days: 1, hours: 2, minutes: 30)),
        status: 'scheduled',
        employeeId: 'demo-doc-1',
        employeeName: 'Давидова Елизавета',
        specialty: 'Кардиолог',
        branchName: 'ONA VA BOLA',
        patientName: 'Пациент (demo)',
      ),
      Appointment(
        id: 'demo-appt-2',
        startsAt: now.add(const Duration(days: 3, hours: 4)),
        endsAt: now.add(const Duration(days: 3, hours: 4, minutes: 30)),
        status: 'scheduled',
        employeeId: 'demo-doc-2',
        employeeName: 'Хейл Натан',
        specialty: 'Невролог',
        branchName: 'ONA VA BOLA',
        patientName: 'Пациент (demo)',
      ),
      Appointment(
        id: 'demo-appt-3',
        startsAt: now.subtract(const Duration(days: 5)),
        endsAt: now.subtract(const Duration(days: 5)).add(const Duration(minutes: 30)),
        status: 'completed',
        employeeId: 'demo-doc-3',
        employeeName: 'Каримова Азиза',
        specialty: 'Педиатр',
        branchName: 'ONA VA BOLA',
        patientName: 'Пациент (demo)',
      ),
    ];
  }

  static Doctor? doctorById(String id) {
    for (final d in doctors) {
      if (d.id == id) return d;
    }
    return null;
  }

  static List<Doctor> filterDoctors({String? query, String? specialty}) {
    var items = doctors.toList();
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
                d.specialtyLabel.toLowerCase().contains(q),
          )
          .toList();
    }
    return items;
  }

  static List<DateTime> demoSlots(DateTime day) {
    final base = DateTime(day.year, day.month, day.day, 9);
    return List.generate(8, (i) => base.add(Duration(minutes: i * 30)));
  }

  static ClinicalRecord get clinicalRecord {
    final now = DateTime.now();
    return ClinicalRecord(
      bloodGroup: 'A+',
      heightCm: 172,
      weightKg: 68,
      allergies: const [
        Allergy(
          id: 'a1',
          allergen: 'Пенициллин',
          severity: 'high',
          isActive: true,
          reaction: 'Сыпь',
        ),
      ],
      history: [
        HistoryEntry(
          id: 'h1',
          type: 'visit',
          occurredAt: now.subtract(const Duration(days: 12)),
          complaints: 'Головная боль',
          diagnosis: 'Мигрень',
          treatment: 'Покой, обильное питьё',
          author: 'Хейл Натан',
          branch: 'ONA VA BOLA',
        ),
        HistoryEntry(
          id: 'h2',
          type: 'visit',
          occurredAt: now.subtract(const Duration(days: 40)),
          diagnosis: 'ОРВИ',
          author: 'Каримова Азиза',
          branch: 'ONA VA BOLA',
        ),
      ],
      labResults: const [
        LabResult(
          id: 'l1',
          type: 'blood',
          collectedOn: '2026-07-20',
          testName: 'Гемоглобин',
          result: '138',
          unit: 'g/L',
          referenceRange: '120–160',
        ),
        LabResult(
          id: 'l2',
          type: 'blood',
          collectedOn: '2026-07-20',
          testName: 'Глюкоза',
          result: '5.1',
          unit: 'mmol/L',
          referenceRange: '3.9–6.1',
        ),
      ],
      imagingStudies: const [
        ImagingStudy(
          id: 'i1',
          modality: 'УЗИ',
          performedOn: '2026-06-10',
          bodyArea: 'Брюшная полость',
          diagnosis: 'Без патологии',
          conclusion: 'Органы без особенностей',
        ),
      ],
      diagnoses: const [
        Diagnosis(
          id: 'd1',
          name: 'Мигрень',
          status: 'ACTIVE',
          diagnosedOn: '2026-07-28',
          icd10Code: 'G43',
          author: 'Хейл Натан',
        ),
      ],
      orders: const [
        ClinicalOrder(
          id: 'o1',
          orderType: 'MEDICATION',
          title: 'Ибупрофен',
          status: 'ACTIVE',
          startOn: '2026-07-28',
          dosage: '200 мг',
          frequency: '2 раза в день',
          durationDays: 5,
          author: 'Хейл Натан',
        ),
      ],
    );
  }
}
