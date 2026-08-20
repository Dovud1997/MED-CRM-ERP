import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/core/network/api_client.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';

class PatientSummary {
  const PatientSummary({
    required this.id,
    required this.firstName,
    required this.lastName,
    this.middleName,
    this.birthDate,
  });

  final String id;
  final String firstName;
  final String lastName;
  final String? middleName;
  final String? birthDate;

  String get fullName =>
      [lastName, firstName, if (middleName != null && middleName!.isNotEmpty) middleName]
          .join(' ');

  factory PatientSummary.fromJson(Map<String, dynamic> json) {
    return PatientSummary(
      id: json['id'] as String,
      firstName: json['firstName'] as String? ?? '',
      lastName: json['lastName'] as String? ?? '',
      middleName: json['middleName'] as String?,
      birthDate: json['birthDate']?.toString(),
    );
  }
}

final patientsRepositoryProvider = Provider<PatientsRepository>((ref) {
  return PatientsRepository(ref.watch(apiClientProvider));
});

class PatientsRepository {
  PatientsRepository(this._api);
  final ApiClient _api;

  Future<List<PatientSummary>> search(String query) async {
    final data = await _api.get<Map<String, dynamic>>(
      '/patients',
      query: {'q': query},
      parser: (raw) => Map<String, dynamic>.from(raw as Map),
    );
    return (data['items'] as List<dynamic>? ?? const [])
        .map((e) => PatientSummary.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
  }
}
