import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/appointments/data/appointments_repository.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/features/doctors/data/doctors_repository.dart';
import 'package:clinicos_mobile/features/favorites/data/favorites_controller.dart';
import 'package:clinicos_mobile/shared/models/appointment.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';

final _homeDashProvider =
    FutureProvider.autoDispose<AppointmentsDashboard?>((ref) async {
  try {
    return await ref.watch(appointmentsRepositoryProvider).dashboard();
  } catch (_) {
    return null;
  }
});

final _homeDoctorsProvider =
    FutureProvider.autoDispose<List<Doctor>>((ref) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).listDoctors();
  } catch (_) {
    return const [];
  }
});

final _homeSpecialtiesProvider =
    FutureProvider.autoDispose<List<Specialty>>((ref) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).listSpecialties();
  } catch (_) {
    return const [];
  }
});

final _selectedSpecialtyProvider = StateProvider<String?>((ref) => null);

class PatientHomeTab extends ConsumerWidget {
  const PatientHomeTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final user = ref.watch(authControllerProvider).user;
    final name =
        (user?.displayName ?? 'Алексей').replaceAll(' (demo)', '');
    final dash = ref.watch(_homeDashProvider);
    final doctorsAsync = ref.watch(_homeDoctorsProvider);
    final specialtiesAsync = ref.watch(_homeSpecialtiesProvider);
    final selectedSpecialty = ref.watch(_selectedSpecialtyProvider);

    final next = dash.valueOrNull?.upcoming.isNotEmpty == true
        ? dash.valueOrNull!.upcoming.first
        : null;

    final specialties = specialtiesAsync.valueOrNull;
    final specialtyNames = (specialties != null && specialties.isNotEmpty)
        ? specialties.map((s) => s.name).take(8).toList()
        : l10n.fallbackSpecialties;

    var doctors = doctorsAsync.valueOrNull ?? const <Doctor>[];
    if (selectedSpecialty != null) {
      doctors = doctors
          .where(
            (d) => d.specialtyLabel
                .toLowerCase()
                .contains(selectedSpecialty.toLowerCase()),
          )
          .toList();
    }
    doctors = doctors.take(6).toList();

    return Scaffold(
      backgroundColor: AppColors.page,
      body: RefreshIndicator(
        color: AppColors.primary,
        onRefresh: () async {
          ref.invalidate(_homeDashProvider);
          ref.invalidate(_homeDoctorsProvider);
          ref.invalidate(_homeSpecialtiesProvider);
        },
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.fromLTRB(
                  20,
                  MediaQuery.paddingOf(context).top + 12,
                  20,
                  0,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _HomeHeader(
                      hello: l10n.hello,
                      name: name,
                      onProfile: () => context.go('/patient/profile'),
                      onNotify: () =>
                          context.push('/patient/profile/notifications'),
                    ),
                    const SizedBox(height: 18),
                    _SearchField(
                      hint: l10n.searchDoctorHint,
                      onTap: () => context.go('/patient/search'),
                    ),
                    const SizedBox(height: 22),
                    _SectionHeader(
                      title: l10n.myAppointments,
                      action: l10n.seeAll,
                      onAction: () => context.go('/patient/appointments'),
                    ),
                    const SizedBox(height: 12),
                    _AppointmentCard(
                      appointment: next,
                      onOpen: () => context.go('/patient/appointments'),
                    ),
                    const SizedBox(height: 22),
                    _SectionHeader(
                      title: l10n.doctorSpecialty,
                      action: l10n.seeAll,
                      onAction: () => context.go('/patient/search'),
                    ),
                    const SizedBox(height: 12),
                  ],
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: SizedBox(
                height: 44,
                child: ListView.separated(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  scrollDirection: Axis.horizontal,
                  itemCount: specialtyNames.length,
                  separatorBuilder: (_, __) => const SizedBox(width: 10),
                  itemBuilder: (context, index) {
                    final label = specialtyNames[index];
                    final selected = selectedSpecialty == label ||
                        (selectedSpecialty == null && index == 0);
                    return _SpecialtyChip(
                      label: label,
                      selected: selected,
                      icon: _specialtyIcon(label),
                      onTap: () {
                        ref.read(_selectedSpecialtyProvider.notifier).state =
                            label;
                      },
                    );
                  },
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(20, 22, 20, 12),
                child: _SectionHeader(
                  title: l10n.popularDoctors,
                  action: l10n.seeAll,
                  onAction: () => context.go('/patient/search'),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: SizedBox(
                height: 236,
                child: doctors.isEmpty
                    ? ListView(
                        scrollDirection: Axis.horizontal,
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        children: [
                          _DemoDoctorCard(
                            name: 'Dr. Elizabeth Davis',
                            specialty: l10n.specialtyCardiologist,
                            rating: '4.5',
                            price: '—',
                            color: const Color(0xFFE8E0F5),
                          ),
                          const SizedBox(width: 12),
                          _DemoDoctorCard(
                            name: 'Dr. Nathan Hale',
                            specialty: l10n.specialtyNeurologist,
                            rating: '4.8',
                            price: '—',
                            color: const Color(0xFFDCEBFF),
                          ),
                        ],
                      )
                    : ListView.separated(
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        scrollDirection: Axis.horizontal,
                        itemCount: doctors.length,
                        separatorBuilder: (_, __) => const SizedBox(width: 12),
                        itemBuilder: (context, index) {
                          final doctor = doctors[index];
                          return _DoctorCard(
                            doctor: doctor,
                            accent: index.isEven
                                ? const Color(0xFFE8E0F5)
                                : const Color(0xFFDCEBFF),
                            onTap: () =>
                                context.push('/patient/doctors/${doctor.id}'),
                          );
                        },
                      ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 28)),
          ],
        ),
      ),
    );
  }

  static IconData _specialtyIcon(String name) {
    final n = name.toLowerCase();
    if (n.contains('neuro') || n.contains('невро') || n.contains('nevrolog')) {
      return Icons.psychology_alt_outlined;
    }
    if (n.contains('pedia') ||
        n.contains('педиа') ||
        n.contains('pediatr') ||
        n.contains('bola')) {
      return Icons.child_care_outlined;
    }
    if (n.contains('cardio') || n.contains('кардио') || n.contains('kardiolog')) {
      return Icons.favorite_outline_rounded;
    }
    if (n.contains('dent') ||
        n.contains('стомат') ||
        n.contains('stomat') ||
        n.contains('tish')) {
      return Icons.medical_services_outlined;
    }
    if (n.contains('therap') || n.contains('терап') || n.contains('terapevt')) {
      return Icons.healing_outlined;
    }
    if (n.contains('surg') || n.contains('хирург') || n.contains('jarroh')) {
      return Icons.content_cut_rounded;
    }
    return Icons.health_and_safety_outlined;
  }
}

class _HomeHeader extends StatelessWidget {
  const _HomeHeader({
    required this.hello,
    required this.name,
    required this.onProfile,
    required this.onNotify,
  });

  final String hello;
  final String name;
  final VoidCallback onProfile;
  final VoidCallback onNotify;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        GestureDetector(
          onTap: onProfile,
          child: Container(
            width: 46,
            height: 46,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: AppColors.softBlue,
              border: Border.all(color: Colors.white, width: 2),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.06),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            alignment: Alignment.center,
            child: Text(
              name.isNotEmpty ? name.trim()[0].toUpperCase() : 'P',
              style: GoogleFonts.inter(
                fontWeight: FontWeight.w700,
                color: AppColors.primary,
                fontSize: 18,
              ),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                hello,
                style: GoogleFonts.inter(
                  fontSize: 13,
                  color: AppColors.muted,
                  fontWeight: FontWeight.w500,
                ),
              ),
              Text(
                name,
                style: GoogleFonts.inter(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                  color: AppColors.ink,
                ),
              ),
            ],
          ),
        ),
        Material(
          color: Colors.white,
          shape: const CircleBorder(),
          elevation: 0,
          child: InkWell(
            customBorder: const CircleBorder(),
            onTap: onNotify,
            child: SizedBox(
              width: 44,
              height: 44,
              child: Stack(
                alignment: Alignment.center,
                children: [
                  const Icon(Icons.notifications_none_rounded,
                      color: AppColors.ink),
                  Positioned(
                    top: 11,
                    right: 12,
                    child: Container(
                      width: 8,
                      height: 8,
                      decoration: const BoxDecoration(
                        color: Color(0xFFFF3B30),
                        shape: BoxShape.circle,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _SearchField extends StatelessWidget {
  const _SearchField({required this.hint, required this.onTap});
  final String hint;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Container(
          height: 52,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: AppColors.line),
          ),
          child: Row(
            children: [
              const Icon(Icons.search, color: AppColors.muted),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  hint,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: GoogleFonts.inter(
                    color: AppColors.muted,
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({
    required this.title,
    required this.action,
    required this.onAction,
  });

  final String title;
  final String action;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Text(
            title,
            style: GoogleFonts.inter(
              fontSize: 17,
              fontWeight: FontWeight.w700,
              color: AppColors.ink,
            ),
          ),
        ),
        TextButton(
          onPressed: onAction,
          style: TextButton.styleFrom(
            foregroundColor: AppColors.primary,
            padding: EdgeInsets.zero,
            minimumSize: Size.zero,
            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          ),
          child: Text(
            action,
            style: GoogleFonts.inter(fontWeight: FontWeight.w600, fontSize: 13),
          ),
        ),
      ],
    );
  }
}

class _AppointmentCard extends StatelessWidget {
  const _AppointmentCard({
    required this.appointment,
    required this.onOpen,
  });

  final Appointment? appointment;
  final VoidCallback onOpen;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final locale = l10n.locale.toString();
    final a = appointment;
    return Material(
      color: Colors.white,
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        onTap: onOpen,
        borderRadius: BorderRadius.circular(20),
        child: Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(20),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.04),
                blurRadius: 16,
                offset: const Offset(0, 6),
              ),
            ],
          ),
          child: a == null
              ? Row(
                  children: [
                    Container(
                      width: 56,
                      height: 56,
                      decoration: BoxDecoration(
                        color: AppColors.softBlue,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: const Icon(Icons.event_available_outlined,
                          color: AppColors.primary),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        l10n.noUpcomingAppointment,
                        style: GoogleFonts.inter(
                          fontWeight: FontWeight.w600,
                          color: AppColors.ink,
                        ),
                      ),
                    ),
                    _RoundArrow(onTap: onOpen),
                  ],
                )
              : Row(
                  children: [
                    Container(
                      width: 56,
                      height: 56,
                      decoration: BoxDecoration(
                        color: AppColors.softBlue,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      alignment: Alignment.center,
                      child: Text(
                        (a.employeeName ?? 'D').trim().isNotEmpty
                            ? (a.employeeName!.trim()[0]).toUpperCase()
                            : 'D',
                        style: GoogleFonts.inter(
                          fontWeight: FontWeight.w700,
                          fontSize: 20,
                          color: AppColors.primary,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            DateFormat('d MMM yyyy', locale).format(a.startsAt),
                            style: GoogleFonts.inter(
                              fontSize: 12,
                              color: AppColors.muted,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            a.employeeName ?? l10n.doctor,
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w700,
                              fontSize: 15,
                              color: AppColors.ink,
                            ),
                          ),
                          Text(
                            [
                              if (a.specialty != null && a.specialty!.isNotEmpty)
                                a.specialty!,
                              DateFormat('HH:mm', locale).format(a.startsAt),
                            ].join(' · '),
                            style: GoogleFonts.inter(
                              fontSize: 12,
                              color: AppColors.muted,
                            ),
                          ),
                        ],
                      ),
                    ),
                    _RoundArrow(onTap: onOpen),
                  ],
                ),
        ),
      ),
    );
  }
}

class _RoundArrow extends StatelessWidget {
  const _RoundArrow({required this.onTap});
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.primary,
      shape: const CircleBorder(),
      child: InkWell(
        customBorder: const CircleBorder(),
        onTap: onTap,
        child: const SizedBox(
          width: 36,
          height: 36,
          child: Icon(Icons.arrow_outward_rounded, color: Colors.white, size: 18),
        ),
      ),
    );
  }
}

class _SpecialtyChip extends StatelessWidget {
  const _SpecialtyChip({
    required this.label,
    required this.selected,
    required this.icon,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? AppColors.primary : Colors.white,
      borderRadius: BorderRadius.circular(28),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(28),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(28),
            border: selected ? null : Border.all(color: AppColors.line),
          ),
          child: Row(
            children: [
              Icon(
                icon,
                size: 18,
                color: selected ? Colors.white : AppColors.primary,
              ),
              const SizedBox(width: 8),
              Text(
                label,
                style: GoogleFonts.inter(
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                  color: selected ? Colors.white : AppColors.ink,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _DoctorCard extends ConsumerWidget {
  const _DoctorCard({
    required this.doctor,
    required this.accent,
    required this.onTap,
  });

  final Doctor doctor;
  final Color accent;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isFav =
        ref.watch(favoritesControllerProvider).contains(doctor.id);
    return SizedBox(
      width: 168,
      child: Material(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(22),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Stack(
                  children: [
                    Container(
                      width: double.infinity,
                      decoration: BoxDecoration(
                        color: accent,
                        borderRadius: const BorderRadius.vertical(
                          top: Radius.circular(22),
                        ),
                      ),
                      alignment: Alignment.center,
                      child: Text(
                        doctor.lastName.isNotEmpty
                            ? doctor.lastName[0].toUpperCase()
                            : 'D',
                        style: GoogleFonts.inter(
                          fontSize: 42,
                          fontWeight: FontWeight.w700,
                          color: AppColors.primary.withValues(alpha: 0.55),
                        ),
                      ),
                    ),
                    Positioned(
                      top: 10,
                      right: 10,
                      child: Material(
                        color: Colors.white,
                        shape: const CircleBorder(),
                        child: InkWell(
                          customBorder: const CircleBorder(),
                          onTap: () => ref
                              .read(favoritesControllerProvider.notifier)
                              .toggle(doctor.id),
                          child: SizedBox(
                            width: 30,
                            height: 30,
                            child: Icon(
                              isFav
                                  ? Icons.favorite_rounded
                                  : Icons.favorite_border,
                              size: 16,
                              color: const Color(0xFFFF3B30),
                            ),
                          ),
                        ),
                      ),
                    ),
                    Positioned(
                      left: 10,
                      bottom: 10,
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.92),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.star_rounded,
                                size: 14, color: Color(0xFFFFC107)),
                            const SizedBox(width: 2),
                            Text(
                              '4.5',
                              style: GoogleFonts.inter(
                                fontSize: 12,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      doctor.fullName,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.w700,
                        fontSize: 13,
                        color: AppColors.ink,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      doctor.specialtyLabel.isEmpty
                          ? AppLocalizations.of(context).doctor
                          : doctor.specialtyLabel,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: GoogleFonts.inter(
                        fontSize: 12,
                        color: AppColors.muted,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _DemoDoctorCard extends StatelessWidget {
  const _DemoDoctorCard({
    required this.name,
    required this.specialty,
    required this.rating,
    required this.price,
    required this.color,
  });

  final String name;
  final String specialty;
  final String rating;
  final String price;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 168,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(22),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Container(
                width: double.infinity,
                decoration: BoxDecoration(
                  color: color,
                  borderRadius: const BorderRadius.vertical(
                    top: Radius.circular(22),
                  ),
                ),
                alignment: Alignment.center,
                child: Text(
                  name.split(' ').last[0],
                  style: GoogleFonts.inter(
                    fontSize: 42,
                    fontWeight: FontWeight.w700,
                    color: AppColors.primary.withValues(alpha: 0.45),
                  ),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: GoogleFonts.inter(
                      fontWeight: FontWeight.w700,
                      fontSize: 13,
                    ),
                  ),
                  Text(
                    '$specialty · ★ $rating',
                    style: GoogleFonts.inter(
                      fontSize: 12,
                      color: AppColors.muted,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
