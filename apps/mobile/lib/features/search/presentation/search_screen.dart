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

final _searchQueryProvider = StateProvider<String>((ref) => '');
final _searchSpecialtyProvider = StateProvider<String?>((ref) => null);

final _searchDoctorsProvider =
    FutureProvider.autoDispose<List<Doctor>>((ref) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).listDoctors(
          query: ref.watch(_searchQueryProvider),
          specialty: ref.watch(_searchSpecialtyProvider),
        );
  } catch (_) {
    return const [];
  }
});

final _searchSpecialtiesProvider =
    FutureProvider.autoDispose<List<Specialty>>((ref) async {
  try {
    return await ref.watch(doctorsRepositoryProvider).listSpecialties();
  } catch (_) {
    return const [];
  }
});

class SearchScreen extends ConsumerWidget {
  const SearchScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final doctorsAsync = ref.watch(_searchDoctorsProvider);
    final specialtiesAsync = ref.watch(_searchSpecialtiesProvider);
    final selected = ref.watch(_searchSpecialtyProvider);
    final favorites = ref.watch(favoritesControllerProvider);

    final specialtyNames = specialtiesAsync.valueOrNull?.isNotEmpty == true
        ? specialtiesAsync.valueOrNull!.map((s) => s.name).toList()
        : l10n.fallbackSpecialties;

    return Scaffold(
      backgroundColor: AppColors.page,
      appBar: AppBar(title: Text(l10n.navSearch)),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 8, 20, 0),
            child: TextField(
              autofocus: true,
              onChanged: (v) =>
                  ref.read(_searchQueryProvider.notifier).state = v,
              decoration: InputDecoration(
                hintText: l10n.searchDoctorHint,
                prefixIcon: const Icon(Icons.search_rounded),
                filled: true,
                fillColor: Colors.white,
              ),
            ),
          ),
          const SizedBox(height: 12),
          SizedBox(
            height: 44,
            child: ListView.separated(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              scrollDirection: Axis.horizontal,
              itemCount: specialtyNames.length + 1,
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemBuilder: (context, index) {
                if (index == 0) {
                  final isSelected = selected == null;
                  return _Chip(
                    label: l10n.seeAll,
                    selected: isSelected,
                    onTap: () =>
                        ref.read(_searchSpecialtyProvider.notifier).state =
                            null,
                  );
                }
                final label = specialtyNames[index - 1];
                return _Chip(
                  label: label,
                  selected: selected == label,
                  onTap: () =>
                      ref.read(_searchSpecialtyProvider.notifier).state =
                          label,
                );
              },
            ),
          ),
          const SizedBox(height: 8),
          Expanded(
            child: doctorsAsync.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => ClinicEmptyState(
                message: l10n.errorGeneric,
                actionLabel: l10n.retry,
                onAction: () => ref.invalidate(_searchDoctorsProvider),
              ),
              data: (doctors) {
                if (doctors.isEmpty) {
                  return ClinicEmptyState(message: l10n.empty);
                }
                return ListView.separated(
                  padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
                  itemCount: doctors.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 12),
                  itemBuilder: (context, index) {
                    final doctor = doctors[index];
                    final isFav = favorites.contains(doctor.id);
                    return ClinicCard(
                      onTap: () =>
                          context.push('/patient/doctors/${doctor.id}'),
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
                                  ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  [
                                    if (doctor.specialtyLabel.isNotEmpty)
                                      doctor.specialtyLabel,
                                    if (doctor.branchName != null)
                                      doctor.branchName!,
                                  ].join(' · '),
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
                            icon: Icon(
                              isFav
                                  ? Icons.favorite_rounded
                                  : Icons.favorite_border_rounded,
                              color: isFav
                                  ? const Color(0xFFFF3B30)
                                  : AppColors.muted,
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
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
          child: Text(
            label,
            style: GoogleFonts.inter(
              fontWeight: FontWeight.w600,
              fontSize: 13,
              color: selected ? Colors.white : AppColors.ink,
            ),
          ),
        ),
      ),
    );
  }
}
