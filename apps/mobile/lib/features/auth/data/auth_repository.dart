import 'package:clinicos_mobile/core/auth/session_user.dart';
import 'package:clinicos_mobile/core/network/api_client.dart';
import 'package:clinicos_mobile/core/storage/token_storage.dart';

class AuthRepository {
  AuthRepository({
    required ApiClient api,
    required TokenStorage tokenStorage,
  })  : _api = api,
        _tokens = tokenStorage;

  final ApiClient _api;
  final TokenStorage _tokens;

  Future<SessionUser> login({
    required String organizationId,
    required String login,
    required String password,
  }) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/auth/login',
      data: {
        'organizationId': organizationId,
        'login': login,
        'password': password,
      },
      parser: (raw) => Map<String, dynamic>.from(raw as Map),
    );
    await _tokens.saveTokens(
      accessToken: data['accessToken'] as String,
      refreshToken: data['refreshToken'] as String,
    );
    return me();
  }

  Future<SessionUser> me() async {
    final data = await _api.get<Map<String, dynamic>>(
      '/auth/me',
      parser: (raw) => Map<String, dynamic>.from(raw as Map),
    );
    return SessionUser.fromJson(data);
  }

  Future<void> logout() async {
    final refresh = await _tokens.readRefreshToken();
    if (refresh != null) {
      try {
        await _api.post(
          '/auth/logout',
          data: {'refreshToken': refresh},
        );
      } catch (_) {
        // Always clear local tokens.
      }
    }
    await _tokens.clear();
  }

  Future<bool> hasSession() async {
    final access = await _tokens.readAccessToken();
    return access != null && access.isNotEmpty;
  }
}
