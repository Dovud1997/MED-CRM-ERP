import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:clinicos_mobile/core/auth/app_role.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/appointments/data/appointments_repository.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/models/appointment.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

class OwnerDashboardData {
  const OwnerDashboardData({
    required this.appointments,
    required this.report,
  });

  final AppointmentsDashboard appointments;
  final Map<String, dynamic> report;
}

final _ownerDashProvider =
    FutureProvider.autoDispose<OwnerDashboardData?>((ref) async {
  final appointments =
      await ref.watch(appointmentsRepositoryProvider).dashboard();
  Map<String, dynamic> report = const {};
  try {
    final api = ref.watch(apiClientProvider);
    report = await api.get<Map<String, dynamic>>(
      '/reports/summary',
      parser: (raw) => Map<String, dynamic>.from(raw as Map),
    );
  } catch (_) {
    if (kSkipLogin) {
      report = {
        'summary': {
          'newPatients': 18,
          'appointments': appointments.today,
          'cancelled': 2,
          'inpatientRevenueUzs': 12500000,
        },
      };
    }
  }
  return OwnerDashboardData(appointments: appointments, report: report);
});

class OwnerShell extends ConsumerStatefulWidget {
  const OwnerShell({super.key});

  @override
  ConsumerState<OwnerShell> createState() => _OwnerShellState();
}

class _OwnerShellState extends ConsumerState<OwnerShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final pages = [
      const _OwnerDashboardPage(),
      const _OwnerSimplePage(kind: _OwnerPageKind.analytics),
      const _OwnerSimplePage(kind: _OwnerPageKind.clinic),
      const _OwnerSimplePage(kind: _OwnerPageKind.finance),
      _OwnerProfilePage(
        onBack: () {
          ref
              .read(authControllerProvider.notifier)
              .setPreviewRole(AppRole.patient);
          context.go('/patient/profile');
        },
      ),
    ];

    return Scaffold(
      backgroundColor: AppColors.page,
      body: pages[_index],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.dashboard_outlined),
            label: l10n.navDashboard,
          ),
          NavigationDestination(
            icon: const Icon(Icons.insights_outlined),
            label: l10n.navAnalytics,
          ),
          NavigationDestination(
            icon: const Icon(Icons.apartment_outlined),
            label: l10n.navClinic,
          ),
          NavigationDestination(
            icon: const Icon(Icons.account_balance_wallet_outlined),
            label: l10n.navFinance,
          ),
          NavigationDestination(
            icon: const Icon(Icons.person_outline),
            label: l10n.navProfile,
          ),
        ],
      ),
    );
  }
}

class _OwnerDashboardPage extends ConsumerWidget {
  const _OwnerDashboardPage();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final dash = ref.watch(_ownerDashProvider);
    return ClinicPageScaffold(
      title: l10n.navDashboard,
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(_ownerDashProvider),
        child: AsyncBody<OwnerDashboardData?>(
          value: dash.when(
            data: AsyncValueLike.data,
            error: (e, _) => AsyncValueLike.error(e),
            loading: AsyncValueLike.loading,
          ),
          onRetry: () => ref.invalidate(_ownerDashProvider),
          builder: (data) {
            if (data == null) {
              return ListView(
                children: [ClinicEmptyState(message: l10n.empty)],
              );
            }
            final summary =
                Map<String, dynamic>.from(data.report['summary'] as Map? ?? {});
            return ListView(
              padding: const EdgeInsets.all(20),
              children: [
                ClinicSectionHeader(title: l10n.navToday),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    _Metric(l10n.navAppointments, '${data.appointments.today}'),
                    _Metric(l10n.tabUpcoming, '${data.appointments.waiting}'),
                    _Metric(l10n.tabPast, '${data.appointments.completed}'),
                    _Metric(
                      l10n.navDoctors,
                      '${data.appointments.doctorsOnDuty}',
                    ),
                    _Metric(
                      l10n.navPatients,
                      '${summary['newPatients'] ?? 0}',
                    ),
                    _Metric(
                      l10n.revenue,
                      '${summary['inpatientRevenueUzs'] ?? 0}',
                    ),
                  ],
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

enum _OwnerPageKind { analytics, clinic, finance }

class _OwnerSimplePage extends ConsumerWidget {
  const _OwnerSimplePage({required this.kind});
  final _OwnerPageKind kind;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final title = switch (kind) {
      _OwnerPageKind.analytics => l10n.navAnalytics,
      _OwnerPageKind.clinic => l10n.navClinic,
      _OwnerPageKind.finance => l10n.navFinance,
    };
    final subtitle = switch (kind) {
      _OwnerPageKind.analytics => l10n.clinicOverview,
      _OwnerPageKind.clinic => l10n.clinicOverview,
      _OwnerPageKind.finance => l10n.financeOverview,
    };
    final dash = ref.watch(_ownerDashProvider).valueOrNull;
    return ClinicPageScaffold(
      title: title,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          ClinicCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  subtitle,
                  style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 8),
                Text(
                  dash == null
                      ? l10n.empty
                      : '${l10n.navAppointments}: ${dash.appointments.today}',
                  style: GoogleFonts.inter(color: AppColors.muted),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _OwnerProfilePage extends StatelessWidget {
  const _OwnerProfilePage({required this.onBack});
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.navProfile,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          ClinicCard(
            onTap: onBack,
            child: Text(
              l10n.uiPreview,
              style: GoogleFonts.inter(fontWeight: FontWeight.w600),
            ),
          ),
        ],
      ),
    );
  }
}

class _Metric extends StatelessWidget {
  const _Metric(this.label, this.value);
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 150,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            value,
            style: GoogleFonts.inter(
              fontSize: 22,
              fontWeight: FontWeight.w700,
              color: AppColors.primary,
            ),
          ),
          Text(label, style: GoogleFonts.inter(color: AppColors.muted)),
        ],
      ),
    );
  }
}
