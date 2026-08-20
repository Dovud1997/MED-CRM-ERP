/// Runtime configuration. Brand values should later come from backend white-label.
class AppConfig {
  const AppConfig({
    required this.apiBaseUrl,
    required this.defaultOrganizationId,
    this.enableLogging = false,
  });

  final String apiBaseUrl;
  final String defaultOrganizationId;
  final bool enableLogging;

  /// Matches compose / .env.example default org for local seed.
  static const development = AppConfig(
    apiBaseUrl: String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://10.0.2.2:80/api/v1',
    ),
    defaultOrganizationId: String.fromEnvironment(
      'DEFAULT_ORGANIZATION_ID',
      defaultValue: '10000000-0000-4000-8000-000000000001',
    ),
    enableLogging: true,
  );
}
