class SessionUser {
  const SessionUser({
    required this.id,
    required this.organizationId,
    required this.permissions,
    this.roles = const [],
    this.displayName,
    this.patientId,
    this.employeeId,
  });

  final String id;
  final String organizationId;
  final List<String> permissions;
  final List<String> roles;
  final String? displayName;
  final String? patientId;
  final String? employeeId;

  bool hasPermission(String code) =>
      permissions.contains('*') || permissions.contains(code);

  factory SessionUser.fromJson(Map<String, dynamic> json) {
    return SessionUser(
      id: json['id'] as String,
      organizationId: json['organizationId'] as String,
      permissions: _parsePermissions(json['permissions']),
      roles: (json['roles'] as List<dynamic>? ?? const [])
          .map((e) => e.toString())
          .toList(),
      displayName: json['displayName'] as String?,
      patientId: json['patientId'] as String?,
      employeeId: json['employeeId'] as String?,
    );
  }

  /// Existing Go `/auth/me` returns permissions as `map[string]bool`.
  static List<String> _parsePermissions(dynamic raw) {
    if (raw is List) {
      return raw.map((e) => e.toString()).toList();
    }
    if (raw is Map) {
      return raw.entries
          .where((e) => e.value == true || e.value == 'true')
          .map((e) => e.key.toString())
          .toList();
    }
    return const [];
  }
}
