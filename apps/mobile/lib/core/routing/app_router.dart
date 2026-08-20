import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:clinicos_mobile/core/auth/app_role.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/features/auth/presentation/login_screen.dart';
import 'package:clinicos_mobile/features/appointments/presentation/appointments_screen.dart';
import 'package:clinicos_mobile/features/appointments/presentation/book_appointment_screen.dart';
import 'package:clinicos_mobile/features/doctors/presentation/doctor_details_screen.dart';
import 'package:clinicos_mobile/features/favorites/presentation/favorites_screen.dart';
import 'package:clinicos_mobile/features/medical_record/presentation/medical_record_screen.dart';
import 'package:clinicos_mobile/features/patient/presentation/patient_shell.dart';
import 'package:clinicos_mobile/features/patient/presentation/profile_subpages.dart';
import 'package:clinicos_mobile/features/search/presentation/search_screen.dart';
import 'package:clinicos_mobile/features/doctor_dashboard/presentation/doctor_shell.dart';
import 'package:clinicos_mobile/features/owner_dashboard/presentation/owner_shell.dart';
import 'package:clinicos_mobile/features/accounting/presentation/accountant_shell.dart';
import 'package:clinicos_mobile/shared/widgets/splash_screen.dart';

final _rootNavigatorKey = GlobalKey<NavigatorState>();

final goRouterProvider = Provider<GoRouter>((ref) {
  final auth = ref.watch(authControllerProvider);

  return GoRouter(
    navigatorKey: _rootNavigatorKey,
    initialLocation: kSkipLogin ? '/patient' : '/splash',
    refreshListenable: _AuthListenable(ref),
    redirect: (context, state) {
      final location = state.matchedLocation;

      if (kSkipLogin) {
        if (location == '/splash' || location == '/login') {
          return _homeFor(auth.role);
        }
        if (!_allowed(auth.role, location)) {
          return _homeFor(auth.role);
        }
        return null;
      }

      final loggingIn = location == '/login';
      final splashing = location == '/splash';

      if (auth.bootstrapping) {
        return splashing ? null : '/splash';
      }
      if (!auth.isAuthenticated) {
        return loggingIn ? null : '/login';
      }

      final home = _homeFor(auth.role);
      if (splashing || loggingIn) return home;
      if (!_allowed(auth.role, location)) return home;
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, __) => const SplashScreen()),
      GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) {
          return PatientShell(navigationShell: navigationShell);
        },
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/patient',
                builder: (_, __) => const PatientHomePage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/patient/favorites',
                builder: (_, __) => const FavoritesScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/patient/search',
                builder: (_, __) => const SearchScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/patient/appointments',
                builder: (_, __) => const AppointmentsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/patient/profile',
                builder: (_, __) => const PatientProfilePage(),
              ),
            ],
          ),
        ],
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/doctors/:doctorId',
        builder: (_, state) => DoctorDetailsScreen(
          doctorId: state.pathParameters['doctorId']!,
        ),
        routes: [
          GoRoute(
            path: 'book',
            builder: (_, state) => BookAppointmentScreen(
              doctorId: state.pathParameters['doctorId']!,
            ),
          ),
        ],
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/appointments/:appointmentId',
        builder: (_, state) => AppointmentDetailScreen(
          appointmentId: state.pathParameters['appointmentId']!,
        ),
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/medical',
        builder: (_, __) => const MedicalRecordScreen(),
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/profile/personal',
        builder: (_, __) => const PersonalInfoScreen(),
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/profile/cards',
        builder: (_, __) => const MyCardsScreen(),
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/profile/notifications',
        builder: (_, __) => const NotificationsScreen(),
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/profile/security',
        builder: (_, __) => const SecurityScreen(),
      ),
      GoRoute(
        parentNavigatorKey: _rootNavigatorKey,
        path: '/patient/profile/help',
        builder: (_, __) => const HelpCenterScreen(),
      ),
      GoRoute(path: '/doctor', builder: (_, __) => const DoctorShell()),
      GoRoute(path: '/owner', builder: (_, __) => const OwnerShell()),
      GoRoute(path: '/accountant', builder: (_, __) => const AccountantShell()),
    ],
  );
});

String _homeFor(AppRole role) => switch (role) {
      AppRole.patient => '/patient',
      AppRole.doctor => '/doctor',
      AppRole.owner => '/owner',
      AppRole.accountant => '/accountant',
      AppRole.unsupported => '/patient',
    };

bool _allowed(AppRole role, String location) {
  if (location.startsWith('/patient') || location == '/login') {
    return role == AppRole.patient ||
        role == AppRole.owner ||
        role == AppRole.doctor ||
        kSkipLogin;
  }
  if (location.startsWith('/doctor')) {
    return role == AppRole.doctor || kSkipLogin;
  }
  if (location.startsWith('/owner')) {
    return role == AppRole.owner || kSkipLogin;
  }
  if (location.startsWith('/accountant')) {
    return role == AppRole.accountant || kSkipLogin;
  }
  return true;
}

class _AuthListenable extends ChangeNotifier {
  _AuthListenable(this._ref) {
    _ref.listen<AuthState>(authControllerProvider, (_, __) => notifyListeners());
  }

  final Ref _ref;
}
