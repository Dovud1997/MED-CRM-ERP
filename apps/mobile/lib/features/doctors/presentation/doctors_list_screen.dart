import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/doctors/data/doctors_repository.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';

final _doctorsQueryProvider = StateProvider<String>((ref) => '');
final _doctorsSpecialtyProvider = StateProvider<String?>((ref) => null);

final _doctorsListProvider =
    FutureProvider.autoDispose<List<Doctor>>((ref) async {
  final query = ref.watch(_doctorsQueryProvider);
  final specialty = ref.watch(_doctorsSpecialtyProvider);
  return ref.watch(doctorsRepositoryProvider).listDoctors(
        query: query,
        specialty: specialty,
      );
});

final _specialtiesProvider =
    FutureProvider.autoDispose<List<Specialty>>((ref) async {
  return ref.watch(doctorsRepositoryProvider).listSpecialties();
});

class DoctorsListScreen extends ConsumerWidget {
  const DoctorsListScreen({super.key, this.autofocusSearch = false});

  final bool autofocusSearch;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final doctorsAsync = ref.watch(_doctorsListProvider);
    final specialtiesAsync = ref.watch(_specialtiesProvider);
    final selectedSpecialty = ref.watch(_doctorsSpecialtyProvider);

    return Scaffold(
      backgroundColor: AppColors.page,
      appBar: AppBar(
        title: Text(autofocusSearch ? l10n.navSearch : l10n.navFavorites),
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
            child: TextField(
              autofocus: autofocusSearch,
              decoration: InputDecoration(
                prefixIcon: const Icon(Icons.search_rounded),
                hintText: l10n.searchDoctorHint,
              ),
              onChanged: (v) =>
                  ref.read(_doctorsQueryProvider.notifier).state = v,
            ),
          ),
          SizedBox(
            height: 52,
            child: specialtiesAsync.when(
              loading: () => const SizedBox.shrink(),
              error: (_, __) => const SizedBox.shrink(),
              data: (items) {
                return ListView(
                  scrollDirection: Axis.horizontal,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  children: [
                    Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: FilterChip(
                        label: Text(l10n.navDoctors),
                        selected: selectedSpecialty == null,
                        onSelected: (_) => ref
                            .read(_doctorsSpecialtyProvider.notifier)
                            .state = null,
                      ),
                    ),
                    ...items.map(
                      (s) => Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: FilterChip(
                          label: Text(s.name),
                          selected: selectedSpecialty == s.name,
                          onSelected: (_) => ref
                              .read(_doctorsSpecialtyProvider.notifier)
                              .state = s.name,
                        ),
                      ),
                    ),
                  ],
                );
              },
            ),
          ),
          Expanded(
            child: AsyncBody<List<Doctor>>(
              value: doctorsAsync.when(
                data: AsyncValueLike.data,
                error: (e, _) => AsyncValueLike.error(e),
                loading: AsyncValueLike.loading,
              ),
              onRetry: () => ref.invalidate(_doctorsListProvider),
              builder: (doctors) {
                return ListView.separated(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
                  itemCount: doctors.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 10),
                  itemBuilder: (context, index) {
                    final doctor = doctors[index];
                    return Material(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(18),
                      child: InkWell(
                        borderRadius: BorderRadius.circular(18),
                        onTap: () =>
                            context.push('/patient/doctors/${doctor.id}'),
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Row(
                            children: [
                              CircleAvatar(
                                radius: 28,
                                backgroundColor: AppColors.softBlue,
                                child: Text(
                                  doctor.lastName.isNotEmpty
                                      ? doctor.lastName[0]
                                      : '?',
                                  style: const TextStyle(
                                    color: AppColors.primary,
                                    fontWeight: FontWeight.w700,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 14),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      doctor.fullName,
                                      style: Theme.of(context)
                                          .textTheme
                                          .titleMedium
                                          ?.copyWith(
                                            fontWeight: FontWeight.w600,
                                          ),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      doctor.specialtyLabel,
                                      style: Theme.of(context)
                                          .textTheme
                                          .bodyMedium
                                          ?.copyWith(color: AppColors.muted),
                                    ),
                                    if (doctor.branchName != null) ...[
                                      const SizedBox(height: 2),
                                      Text(
                                        doctor.branchName!,
                                        style: Theme.of(context)
                                            .textTheme
                                            .bodySmall,
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                              const Icon(
                                Icons.chevron_right_rounded,
                                color: AppColors.muted,
                              ),
                            ],
                          ),
                        ),
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
