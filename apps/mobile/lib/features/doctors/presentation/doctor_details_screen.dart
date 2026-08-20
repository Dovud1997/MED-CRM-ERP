import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/doctors/data/doctors_repository.dart';
import 'package:clinicos_mobile/features/favorites/data/favorites_controller.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

final _doctorProvider =
    FutureProvider.autoDispose.family<Doctor?, String>((ref, id) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).getDoctor(id);
  } catch (_) {
    return null;
  }
});

final _scheduleProvider =
    FutureProvider.autoDispose.family<DoctorSchedule?, String>((ref, id) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).getSchedule(id);
  } catch (_) {
    return null;
  }
});

class DoctorDetailsScreen extends ConsumerWidget {
  const DoctorDetailsScreen({super.key, required this.doctorId});

  final String doctorId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final locale = l10n.locale.toString();
    final doctorAsync = ref.watch(_doctorProvider(doctorId));
    final scheduleAsync = ref.watch(_scheduleProvider(doctorId));
    final isFav = ref.watch(favoritesControllerProvider).contains(doctorId);

    return Scaffold(
      backgroundColor: AppColors.page,
      body: AsyncBody<Doctor?>(
        value: doctorAsync.when(
          data: AsyncValueLike.data,
          error: (e, _) => AsyncValueLike.error(e),
          loading: AsyncValueLike.loading,
        ),
        onRetry: () => ref.invalidate(_doctorProvider(doctorId)),
        builder: (doctor) {
          if (doctor == null) {
            return ClinicEmptyState(message: l10n.errorNotFound);
          }
          final today = DateTime.now();
          final schedule = scheduleAsync.asData?.value;
          final todayDay = schedule?.dayForDate(today);
          String scheduleHint = l10n.empty;
          if (todayDay == null || !todayDay.isWorking) {
            scheduleHint = l10n.notWorkingToday;
          } else {
            final breakPart = todayDay.breakFrom.isNotEmpty
                ? ' · ${l10n.breakLabel} ${todayDay.breakFrom}–${todayDay.breakTo}'
                : '';
            scheduleHint =
                '${l10n.workingToday}: ${todayDay.start}–${todayDay.end}$breakPart';
          }

          return CustomScrollView(
            slivers: [
              SliverAppBar(
                expandedHeight: 220,
                pinned: true,
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                actions: [
                  IconButton(
                    onPressed: () => ref
                        .read(favoritesControllerProvider.notifier)
                        .toggle(doctorId),
                    icon: Icon(
                      isFav ? Icons.favorite_rounded : Icons.favorite_border,
                      color: isFav ? const Color(0xFFFF8A80) : Colors.white,
                    ),
                  ),
                ],
                flexibleSpace: FlexibleSpaceBar(
                  background: Container(
                    color: const Color(0xFFDCEBFF),
                    alignment: Alignment.center,
                    child: Text(
                      doctor.lastName.isNotEmpty
                          ? doctor.lastName[0].toUpperCase()
                          : 'D',
                      style: GoogleFonts.inter(
                        fontSize: 84,
                        fontWeight: FontWeight.w700,
                        color: AppColors.primary.withValues(alpha: 0.45),
                      ),
                    ),
                  ),
                ),
              ),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(20, 20, 20, 100),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        doctor.fullName,
                        style: GoogleFonts.inter(
                          fontSize: 22,
                          fontWeight: FontWeight.w700,
                          color: AppColors.ink,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        doctor.specialtyLabel.isEmpty
                            ? l10n.doctor
                            : doctor.specialtyLabel,
                        style: GoogleFonts.inter(
                          fontSize: 14,
                          color: AppColors.muted,
                        ),
                      ),
                      if (doctor.branchName != null) ...[
                        const SizedBox(height: 4),
                        Text(
                          doctor.branchName!,
                          style: GoogleFonts.inter(
                            fontSize: 13,
                            color: AppColors.muted,
                          ),
                        ),
                      ],
                      const SizedBox(height: 16),
                      Row(
                        children: [
                          _StatChip(
                            icon: Icons.star_rounded,
                            label: '4.5',
                            caption: l10n.rating,
                          ),
                          const SizedBox(width: 10),
                          _StatChip(
                            icon: Icons.work_outline_rounded,
                            label: '8+',
                            caption: l10n.yearsShort,
                          ),
                          const SizedBox(width: 10),
                          _StatChip(
                            icon: Icons.groups_outlined,
                            label: '120+',
                            caption: l10n.patientsCount,
                          ),
                        ],
                      ),
                      const SizedBox(height: 22),
                      ClinicSectionHeader(title: l10n.aboutDoctor),
                      const SizedBox(height: 10),
                      ClinicCard(
                        child: Text(
                          '${doctor.fullName} — ${doctor.specialtyLabel.isEmpty ? l10n.doctor : doctor.specialtyLabel}.',
                          style: GoogleFonts.inter(
                            height: 1.45,
                            color: AppColors.ink,
                          ),
                        ),
                      ),
                      const SizedBox(height: 22),
                      ClinicSectionHeader(title: l10n.availability),
                      const SizedBox(height: 10),
                      ClinicCard(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              DateFormat.yMMMMd(locale).format(today),
                              style: GoogleFonts.inter(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              scheduleHint,
                              style: GoogleFonts.inter(color: AppColors.muted),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 12),
          child: ClinicPrimaryButton(
            label: l10n.bookAppointment,
            onPressed: () => context.push('/patient/doctors/$doctorId/book'),
          ),
        ),
      ),
    );
  }
}

class _StatChip extends StatelessWidget {
  const _StatChip({
    required this.icon,
    required this.label,
    required this.caption,
  });

  final IconData icon;
  final String label;
  final String caption;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: AppColors.softBlue,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          children: [
            Icon(icon, color: AppColors.primary, size: 20),
            const SizedBox(height: 4),
            Text(
              label,
              style: GoogleFonts.inter(
                fontWeight: FontWeight.w700,
                color: AppColors.ink,
              ),
            ),
            Text(
              caption,
              style: GoogleFonts.inter(fontSize: 11, color: AppColors.muted),
            ),
          ],
        ),
      ),
    );
  }
}
