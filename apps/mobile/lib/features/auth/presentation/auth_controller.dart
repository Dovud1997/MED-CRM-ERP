import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/core/auth/app_role.dart';
import 'package:clinicos_mobile/core/auth/session_user.dart';
import 'package:clinicos_mobile/core/config/app_config.dart';
import 'package:clinicos_mobile/core/network/api_client.dart';
import 'package:clinicos_mobile/core/storage/token_storage.dart';
import 'package:clinicos_mobile/features/auth/data/auth_repository.dart';

/// Temporary: UI development without login. Set false when wiring auth later.
const bool kSkipLogin = true;

final appConfigProvider = Provider<AppConfig>((ref) => AppConfig.development);

final tokenStorageProvider = Provider<TokenStorage>((ref) => TokenStorage());

class SessionEvents {
  void Function()? onExpired;
}

final sessionEventsProvider = Provider<SessionEvents>((ref) => SessionEvents());

final apiClientProvider = Provider<ApiClient>((ref) {
  final config = ref.watch(appConfigProvider);
  final storage = ref.watch(tokenStorageProvider);
  final events = ref.watch(sessionEventsProvider);
  return ApiClient(
    config: config,
    tokenStorage: storage,
    onSessionExpired: () => events.onExpired?.call(),
  );
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(
    api: ref.watch(apiClientProvider),
    tokenStorage: ref.watch(tokenStorageProvider),
  );
});

SessionUser _demoUser(AppRole role) {
  final roles = switch (role) {
    AppRole.patient => const ['PATIENT'],
    AppRole.doctor => const ['DOCTOR_PEDIATRICIAN'],
    AppRole.owner => const ['OWNER'],
    AppRole.accountant => const ['ACCOUNTANT'],
    AppRole.unsupported => const <String>[],
  };
  final permissions = switch (role) {
    AppRole.owner => const ['*'],
    AppRole.doctor => const [
        'patients:read',
        'patients:write',
        'clinical:read',
        'clinical:write',
        'appointments:read',
        'appointments:write',
        'employees:read',
        'branches:read',
      ],
    AppRole.accountant => const [
        'finance:read',
        'finance:write',
        'branches:read',
        'reports:read',
      ],
    AppRole.patient => const ['mobile:patient:read'],
    AppRole.unsupported => const <String>[],
  };
  return SessionUser(
    id: 'demo-user',
    organizationId: AppConfig.development.defaultOrganizationId,
    permissions: permissions,
    roles: roles,
    displayName: switch (role) {
      AppRole.patient => 'Пациент (demo)',
      AppRole.doctor => 'Врач (demo)',
      AppRole.owner => 'Владелец (demo)',
      AppRole.accountant => 'Бухгалтер (demo)',
      AppRole.unsupported => 'Demo',
    },
  );
}

class AuthState {
  const AuthState({
    this.user,
    this.bootstrapping = true,
    this.loading = false,
    this.error,
    this.previewRole = AppRole.patient,
  });

  final SessionUser? user;
  final bool bootstrapping;
  final bool loading;
  final String? error;
  final AppRole previewRole;

  bool get isAuthenticated => user != null;

  AppRole get role {
    if (kSkipLogin) return previewRole;
    if (user == null) return AppRole.unsupported;
    final fromRoles = AppRole.fromRoleCodes(user!.roles);
    if (fromRoles != AppRole.unsupported) return fromRoles;
    return AppRole.fromPermissions(user!.permissions);
  }

  AuthState copyWith({
    SessionUser? user,
    bool? bootstrapping,
    bool? loading,
    String? error,
    AppRole? previewRole,
    bool clearUser = false,
    bool clearError = false,
  }) {
    return AuthState(
      user: clearUser ? null : (user ?? this.user),
      bootstrapping: bootstrapping ?? this.bootstrapping,
      loading: loading ?? this.loading,
      error: clearError ? null : (error ?? this.error),
      previewRole: previewRole ?? this.previewRole,
    );
  }
}

class AuthController extends StateNotifier<AuthState> {
  AuthController(this._repo) : super(const AuthState()) {
    _bootstrap();
  }

  final AuthRepository _repo;

  Future<void> _bootstrap() async {
    if (kSkipLogin) {
      state = AuthState(
        user: _demoUser(AppRole.patient),
        bootstrapping: false,
        previewRole: AppRole.patient,
      );
      return;
    }
    try {
      if (await _repo.hasSession()) {
        final user = await _repo.me();
        state = AuthState(user: user, bootstrapping: false);
      } else {
        state = const AuthState(bootstrapping: false);
      }
    } catch (_) {
      await _repo.logout();
      state = const AuthState(bootstrapping: false);
    }
  }

  void setPreviewRole(AppRole role) {
    state = state.copyWith(
      previewRole: role,
      user: _demoUser(role),
      bootstrapping: false,
    );
  }

  Future<bool> login({
    required String organizationId,
    required String login,
    required String password,
  }) async {
    if (kSkipLogin) return true;
    state = state.copyWith(loading: true, clearError: true);
    try {
      final user = await _repo.login(
        organizationId: organizationId,
        login: login,
        password: password,
      );
      state = AuthState(user: user, bootstrapping: false);
      return true;
    } catch (e) {
      state = state.copyWith(
        loading: false,
        error: e.toString(),
        bootstrapping: false,
      );
      return false;
    }
  }

  Future<void> logout() async {
    if (kSkipLogin) {
      setPreviewRole(AppRole.patient);
      return;
    }
    await _repo.logout();
    markLoggedOut();
  }

  void markLoggedOut() {
    if (kSkipLogin) {
      setPreviewRole(AppRole.patient);
      return;
    }
    state = const AuthState(bootstrapping: false);
  }
}

final authControllerProvider =
    StateNotifierProvider<AuthController, AuthState>((ref) {
  final controller = AuthController(ref.watch(authRepositoryProvider));
  final events = ref.read(sessionEventsProvider);
  events.onExpired = controller.markLoggedOut;
  ref.onDispose(() {
    if (events.onExpired == controller.markLoggedOut) {
      events.onExpired = null;
    }
  });
  return controller;
});
