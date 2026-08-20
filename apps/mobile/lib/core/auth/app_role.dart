enum AppRole {
  patient,
  doctor,
  owner,
  accountant,
  unsupported;

  static AppRole fromRoleCodes(Iterable<String> codes) {
    final set = codes.map((e) => e.toUpperCase()).toSet();
    if (set.contains('PATIENT')) return AppRole.patient;
    if (set.contains('OWNER') || set.contains('ADMIN')) return AppRole.owner;
    if (set.contains('ACCOUNTANT')) return AppRole.accountant;
    if (set.any((c) => c.startsWith('DOCTOR') || c == 'SPEECH_THERAPIST')) {
      return AppRole.doctor;
    }
    return AppRole.unsupported;
  }

  static AppRole fromPermissions(Iterable<String> permissions) {
    final set = permissions.toSet();
    if (set.contains('*')) return AppRole.owner;
    if (set.contains('finance:read') &&
        set.contains('finance:write') &&
        !set.contains('clinical:write') &&
        !set.contains('appointments:write')) {
      return AppRole.accountant;
    }
    if (set.contains('clinical:write') && set.contains('appointments:write')) {
      return AppRole.doctor;
    }
    if (set.contains('mobile:patient:read') || set.contains('patient:self')) {
      return AppRole.patient;
    }
    return AppRole.unsupported;
  }
}
