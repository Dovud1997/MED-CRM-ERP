import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/core/network/api_client.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/features/patient/data/patients_repository.dart';
import 'package:clinicos_mobile/shared/demo/demo_data.dart';
import 'package:clinicos_mobile/shared/models/clinical_record.dart';

final selectedPatientIdProvider = StateProvider<String?>((ref) {
  final id = ref.watch(authControllerProvider).user?.patientId;
  if (id != null) return id;
  if (kSkipLogin) return 'demo-patient';
  return null;
});

final clinicalRepositoryProvider = Provider<ClinicalRepository>((ref) {
  return ClinicalRepository(ref.watch(apiClientProvider));
});

class ClinicalRepository {
  ClinicalRepository(this._api);
  final ApiClient _api;

  Future<ClinicalRecord> getRecord(String patientId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/patients/$patientId/clinical',
        parser: (raw) => Map<String, dynamic>.from(raw as Map),
      );
      return ClinicalRecord.fromJson(data);
    } catch (_) {
      if (!kSkipLogin) rethrow;
      return DemoData.clinicalRecord;
    }
  }
}

final clinicalRecordProvider =
    FutureProvider.autoDispose.family<ClinicalRecord, String>((ref, patientId) {
  return ref.watch(clinicalRepositoryProvider).getRecord(patientId);
});

final patientSearchProvider =
    FutureProvider.autoDispose.family<List<PatientSummary>, String>((ref, q) {
  if (q.trim().length < 2) return Future.value(const []);
  return ref.watch(patientsRepositoryProvider).search(q.trim());
});
