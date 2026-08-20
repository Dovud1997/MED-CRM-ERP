import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/doctors/data/doctors_repository.dart';
import 'package:clinicos_mobile/features/favorites/data/favorites_controller.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

final _allDoctorsProvider =
    FutureProvider.autoDispose<List<Doctor>>((ref) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).listDoctors();
  } catch (_) {
    return const [];
  }
});

class FavoritesScreen extends ConsumerWidget {
  const FavoritesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final favoriteIds = ref.watch(favoritesControllerProvider);
    final doctorsAsync = ref.watch(_allDoctorsProvider);

    return ClinicPageScaffold(
      title: l10n.navFavorites,
      body: doctorsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => ClinicEmptyState(
          message: l10n.errorGeneric,
          actionLabel: l10n.retry,
          onAction: () => ref.invalidate(_allDoctorsProvider),
        ),
        data: (doctors) {
          final favs =
              doctors.where((d) => favoriteIds.contains(d.id)).toList();
          if (favs.isEmpty) {
            return ClinicEmptyState(
              message: l10n.noFavorites,
              icon: Icons.favorite_border_rounded,
              actionLabel: l10n.navSearch,
              onAction: () => context.go('/patient/search'),
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
            itemCount: favs.length,
            separatorBuilder: (_, __) => const SizedBox(height: 12),
            itemBuilder: (context, index) {
              final doctor = favs[index];
              return ClinicCard(
                onTap: () => context.push('/patient/doctors/${doctor.id}'),
                child: Row(
                  children: [
                    DoctorAvatar(name: doctor.fullName),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            doctor.fullName,
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w700,
                              fontSize: 15,
                              color: AppColors.ink,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            doctor.specialtyLabel,
                            style: GoogleFonts.inter(
                              fontSize: 13,
                              color: AppColors.muted,
                            ),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      onPressed: () => ref
                          .read(favoritesControllerProvider.notifier)
                          .toggle(doctor.id),
                      icon: const Icon(
                        Icons.favorite_rounded,
                        color: Color(0xFFFF3B30),
                      ),
                    ),
                  ],
                ),
              );
            },
          );
        },
      ),
    );
  }
}
