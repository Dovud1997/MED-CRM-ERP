import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/appointments/data/appointments_repository.dart';
import 'package:clinicos_mobile/shared/demo/demo_data.dart';
import 'package:clinicos_mobile/shared/models/appointment.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

final appointmentsListProvider =
    FutureProvider.autoDispose<List<Appointment>>((ref) async {
  try {
    final active = await ref.watch(appointmentsRepositoryProvider).listActive();
    final today = DateTime.now();
    final recent = <Appointment>[];
    for (var i = 0; i < 7; i++) {
      final day = today.subtract(Duration(days: i));
      recent.addAll(
        await ref.watch(appointmentsRepositoryProvider).listByDate(day),
      );
    }
    final byId = <String, Appointment>{};
    for (final a in [...active, ...recent]) {
      byId[a.id] = a;
    }
    final all = byId.values.toList()
      ..sort((a, b) => b.startsAt.compareTo(a.startsAt));
    if (all.isNotEmpty) return all;
  } catch (_) {}
  return DemoData.appointments();
});

class AppointmentsScreen extends ConsumerWidget {
  const AppointmentsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final async = ref.watch(appointmentsListProvider);

    return DefaultTabController(
      length: 3,
      child: Scaffold(
        backgroundColor: AppColors.page,
        appBar: AppBar(
          title: Text(l10n.navAppointments),
          bottom: TabBar(
            labelColor: AppColors.primary,
            unselectedLabelColor: AppColors.muted,
            indicatorColor: AppColors.primary,
            tabs: [
              Tab(text: l10n.tabUpcoming),
              Tab(text: l10n.tabPast),
              Tab(text: l10n.tabCancelled),
            ],
          ),
        ),
        body: AsyncBody<List<Appointment>>(
          value: async.when(
            data: AsyncValueLike.data,
            error: (e, _) => AsyncValueLike.error(e),
            loading: AsyncValueLike.loading,
          ),
          onRetry: () => ref.invalidate(appointmentsListProvider),
          builder: (items) {
            final upcoming = items.where((a) => a.isUpcoming).toList();
            final past =
                items.where((a) => a.isPast && !a.isCancelled).toList();
            final cancelled = items.where((a) => a.isCancelled).toList();
            return TabBarView(
              children: [
                _AppointmentList(items: upcoming, canCancel: true),
                _AppointmentList(items: past),
                _AppointmentList(items: cancelled),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _AppointmentList extends ConsumerWidget {
  const _AppointmentList({required this.items, this.canCancel = false});

  final List<Appointment> items;
  final bool canCancel;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final locale = l10n.locale.toString();
    if (items.isEmpty) {
      return ClinicEmptyState(
        message: l10n.empty,
        icon: Icons.event_outlined,
      );
    }
    return ListView.separated(
      padding: const EdgeInsets.all(20),
      itemCount: items.length,
      separatorBuilder: (_, __) => const SizedBox(height: 12),
      itemBuilder: (context, index) {
        final a = items[index];
        return ClinicCard(
          onTap: () => context.push('/patient/appointments/${a.id}'),
          child: Row(
            children: [
              DoctorAvatar(name: a.employeeName ?? l10n.doctor),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      DateFormat('d MMM yyyy · HH:mm', locale)
                          .format(a.startsAt),
                      style: GoogleFonts.inter(
                        fontSize: 12,
                        color: AppColors.muted,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      a.employeeName ?? l10n.doctor,
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.w700,
                        fontSize: 15,
                      ),
                    ),
                    Text(
                      [
                        if (a.specialty != null) a.specialty!,
                      ].join(' · '),
                      style: GoogleFonts.inter(
                        fontSize: 13,
                        color: AppColors.muted,
                      ),
                    ),
                  ],
                ),
              ),
              _StatusChip(status: a.status, label: l10n.statusText(a.status)),
            ],
          ),
        );
      },
    );
  }
}

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.status, required this.label});
  final String status;
  final String label;

  @override
  Widget build(BuildContext context) {
    final s = status.toLowerCase();
    final Color bg;
    final Color fg;
    if (s.contains('cancel')) {
      bg = const Color(0xFFFFEBEE);
      fg = const Color(0xFFC62828);
    } else if (s.contains('complete') || s == 'done') {
      bg = const Color(0xFFE8F5E9);
      fg = AppColors.success;
    } else {
      bg = AppColors.softBlue;
      fg = AppColors.primary;
    }
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: GoogleFonts.inter(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: fg,
        ),
      ),
    );
  }
}

class AppointmentDetailScreen extends ConsumerWidget {
  const AppointmentDetailScreen({super.key, required this.appointmentId});

  final String appointmentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final locale = l10n.locale.toString();
    final async = ref.watch(appointmentsListProvider);

    return ClinicPageScaffold(
      title: l10n.appointmentDetails,
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => ClinicEmptyState(message: l10n.errorGeneric),
        data: (items) {
          Appointment? found;
          for (final a in items) {
            if (a.id == appointmentId) found = a;
          }
          if (found == null) {
            return ClinicEmptyState(message: l10n.errorNotFound);
          }
          final a = found;
          return ListView(
            padding: const EdgeInsets.all(20),
            children: [
              ClinicCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        DoctorAvatar(
                          name: a.employeeName ?? l10n.doctor,
                          size: 64,
                        ),
                        const SizedBox(width: 14),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                a.employeeName ?? l10n.doctor,
                                style: GoogleFonts.inter(
                                  fontWeight: FontWeight.w700,
                                  fontSize: 18,
                                ),
                              ),
                              if (a.specialty != null)
                                Text(
                                  a.specialty!,
                                  style: GoogleFonts.inter(
                                    color: AppColors.muted,
                                  ),
                                ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 18),
                    _InfoRow(
                      label: l10n.selectDate,
                      value: DateFormat('d MMMM yyyy', locale)
                          .format(a.startsAt),
                    ),
                    _InfoRow(
                      label: l10n.selectTime,
                      value: DateFormat('HH:mm', locale).format(a.startsAt),
                    ),
                    _InfoRow(
                      label: l10n.statusLabel,
                      value: l10n.statusText(a.status),
                    ),
                    if (a.branchName != null)
                      _InfoRow(label: l10n.navClinic, value: a.branchName!),
                  ],
                ),
              ),
              if (a.isUpcoming) ...[
                const SizedBox(height: 20),
                OutlinedButton(
                  onPressed: () async {
                    await ref
                        .read(appointmentsRepositoryProvider)
                        .cancel(a.id);
                    ref.invalidate(appointmentsListProvider);
                    if (context.mounted) context.pop();
                  },
                  style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.red,
                    minimumSize: const Size.fromHeight(52),
                    side: const BorderSide(color: Colors.red),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(28),
                    ),
                  ),
                  child: Text(l10n.cancelAppointment),
                ),
              ],
            ],
          );
        },
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: GoogleFonts.inter(color: AppColors.muted),
            ),
          ),
          Text(
            value,
            style: GoogleFonts.inter(fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }
}
