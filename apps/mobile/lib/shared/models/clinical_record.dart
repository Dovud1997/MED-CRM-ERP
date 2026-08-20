class ClinicalRecord {
  const ClinicalRecord({
    required this.bloodGroup,
    this.heightCm,
    this.weightKg,
    this.allergies = const [],
    this.history = const [],
    this.vaccinations = const [],
    this.labResults = const [],
    this.imagingStudies = const [],
    this.diagnoses = const [],
    this.orders = const [],
  });

  final String bloodGroup;
  final double? heightCm;
  final double? weightKg;
  final List<Allergy> allergies;
  final List<HistoryEntry> history;
  final List<Vaccination> vaccinations;
  final List<LabResult> labResults;
  final List<ImagingStudy> imagingStudies;
  final List<Diagnosis> diagnoses;
  final List<ClinicalOrder> orders;

  factory ClinicalRecord.fromJson(Map<String, dynamic> json) {
    return ClinicalRecord(
      bloodGroup: json['bloodGroup'] as String? ?? 'unknown',
      heightCm: (json['heightCm'] as num?)?.toDouble(),
      weightKg: (json['weightKg'] as num?)?.toDouble(),
      allergies: _mapList(json['allergies'], Allergy.fromJson),
      history: _mapList(json['history'], HistoryEntry.fromJson),
      vaccinations: _mapList(json['vaccinations'], Vaccination.fromJson),
      labResults: _mapList(json['labResults'], LabResult.fromJson),
      imagingStudies: _mapList(json['imagingStudies'], ImagingStudy.fromJson),
      diagnoses: _mapList(json['diagnoses'], Diagnosis.fromJson),
      orders: _mapList(json['orders'], ClinicalOrder.fromJson),
    );
  }

  static List<T> _mapList<T>(
    dynamic raw,
    T Function(Map<String, dynamic>) map,
  ) {
    return (raw as List<dynamic>? ?? const [])
        .map((e) => map(Map<String, dynamic>.from(e as Map)))
        .toList();
  }
}

class Allergy {
  const Allergy({
    required this.id,
    required this.allergen,
    required this.severity,
    required this.isActive,
    this.reaction,
  });

  final String id;
  final String allergen;
  final String severity;
  final bool isActive;
  final String? reaction;

  factory Allergy.fromJson(Map<String, dynamic> json) => Allergy(
        id: json['id'] as String,
        allergen: json['allergen'] as String? ?? '',
        severity: json['severity'] as String? ?? 'unknown',
        isActive: json['isActive'] as bool? ?? true,
        reaction: json['reaction'] as String?,
      );
}

class HistoryEntry {
  const HistoryEntry({
    required this.id,
    required this.type,
    required this.occurredAt,
    this.complaints,
    this.diagnosis,
    this.treatment,
    this.notes,
    this.author,
    this.branch,
  });

  final String id;
  final String type;
  final DateTime occurredAt;
  final String? complaints;
  final String? diagnosis;
  final String? treatment;
  final String? notes;
  final String? author;
  final String? branch;

  String get title {
    if (diagnosis != null && diagnosis!.trim().isNotEmpty) return diagnosis!;
    if (complaints != null && complaints!.trim().isNotEmpty) return complaints!;
    return type;
  }

  factory HistoryEntry.fromJson(Map<String, dynamic> json) => HistoryEntry(
        id: json['id'] as String,
        type: json['type'] as String? ?? 'note',
        occurredAt: DateTime.parse(json['occurredAt'] as String).toLocal(),
        complaints: json['complaints'] as String?,
        diagnosis: json['diagnosis'] as String?,
        treatment: json['treatment'] as String?,
        notes: json['notes'] as String?,
        author: json['author'] as String?,
        branch: json['branch'] as String?,
      );
}

class Vaccination {
  const Vaccination({
    required this.id,
    required this.name,
    required this.administeredOn,
  });

  final String id;
  final String name;
  final String administeredOn;

  factory Vaccination.fromJson(Map<String, dynamic> json) => Vaccination(
        id: json['id'] as String,
        name: json['name'] as String? ?? '',
        administeredOn: json['administeredOn'] as String? ?? '',
      );
}

class LabResult {
  const LabResult({
    required this.id,
    required this.type,
    required this.collectedOn,
    required this.testName,
    required this.result,
    this.unit,
    this.referenceRange,
    this.notes,
    this.attachmentId,
    this.attachmentName,
  });

  final String id;
  final String type;
  final String collectedOn;
  final String testName;
  final String result;
  final String? unit;
  final String? referenceRange;
  final String? notes;
  final String? attachmentId;
  final String? attachmentName;

  factory LabResult.fromJson(Map<String, dynamic> json) => LabResult(
        id: json['id'] as String,
        type: json['type'] as String? ?? 'other',
        collectedOn: json['collectedOn'] as String? ?? '',
        testName: json['testName'] as String? ?? '',
        result: json['result'] as String? ?? '',
        unit: json['unit'] as String?,
        referenceRange: json['referenceRange'] as String?,
        notes: json['notes'] as String?,
        attachmentId: json['attachmentId'] as String?,
        attachmentName: json['attachmentName'] as String?,
      );
}

class ImagingStudy {
  const ImagingStudy({
    required this.id,
    required this.modality,
    required this.performedOn,
    required this.bodyArea,
    required this.diagnosis,
    this.conclusion,
    this.fileName,
  });

  final String id;
  final String modality;
  final String performedOn;
  final String bodyArea;
  final String diagnosis;
  final String? conclusion;
  final String? fileName;

  factory ImagingStudy.fromJson(Map<String, dynamic> json) => ImagingStudy(
        id: json['id'] as String,
        modality: json['modality'] as String? ?? '',
        performedOn: json['performedOn'] as String? ?? '',
        bodyArea: json['bodyArea'] as String? ?? '',
        diagnosis: json['diagnosis'] as String? ?? '',
        conclusion: json['conclusion'] as String?,
        fileName: json['fileName'] as String?,
      );
}

class Diagnosis {
  const Diagnosis({
    required this.id,
    required this.name,
    required this.status,
    required this.diagnosedOn,
    this.icd10Code,
    this.notes,
    this.author,
  });

  final String id;
  final String name;
  final String status;
  final String diagnosedOn;
  final String? icd10Code;
  final String? notes;
  final String? author;

  factory Diagnosis.fromJson(Map<String, dynamic> json) => Diagnosis(
        id: json['id'] as String,
        name: json['name'] as String? ?? '',
        status: json['status'] as String? ?? 'ACTIVE',
        diagnosedOn: json['diagnosedOn'] as String? ?? '',
        icd10Code: json['icd10Code'] as String?,
        notes: json['notes'] as String?,
        author: json['author'] as String?,
      );
}

class ClinicalOrder {
  const ClinicalOrder({
    required this.id,
    required this.orderType,
    required this.title,
    required this.status,
    required this.startOn,
    this.dosage,
    this.frequency,
    this.durationDays,
    this.instructions,
    this.endOn,
    this.author,
  });

  final String id;
  final String orderType;
  final String title;
  final String status;
  final String startOn;
  final String? dosage;
  final String? frequency;
  final int? durationDays;
  final String? instructions;
  final String? endOn;
  final String? author;

  bool get isMedication => orderType == 'MEDICATION';
  bool get isActive => status == 'ACTIVE';

  factory ClinicalOrder.fromJson(Map<String, dynamic> json) => ClinicalOrder(
        id: json['id'] as String,
        orderType: json['orderType'] as String? ?? '',
        title: json['title'] as String? ?? '',
        status: json['status'] as String? ?? '',
        startOn: json['startOn'] as String? ?? '',
        dosage: json['dosage'] as String?,
        frequency: json['frequency'] as String?,
        durationDays: (json['durationDays'] as num?)?.toInt(),
        instructions: json['instructions'] as String?,
        endOn: json['endOn'] as String?,
        author: json['author'] as String?,
      );
}
