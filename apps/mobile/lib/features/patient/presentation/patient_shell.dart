import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:clinicos_mobile/core/auth/app_role.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/localization/locale_controller.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/features/patient/presentation/patient_home_tab.dart';

class PatientShell extends StatelessWidget {
  const PatientShell({super.key, required this.navigationShell});

  final StatefulNavigationShell navigationShell;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final index = navigationShell.currentIndex;

    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.06),
              blurRadius: 18,
              offset: const Offset(0, -4),
            ),
          ],
        ),
        child: SafeArea(
          top: false,
          child: SizedBox(
            height: 64,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _NavIcon(
                  icon: Icons.home_outlined,
                  selectedIcon: Icons.home_rounded,
                  selected: index == 0,
                  tooltip: l10n.navHome,
                  onTap: () => navigationShell.goBranch(0),
                ),
                _NavIcon(
                  icon: Icons.favorite_border_rounded,
                  selectedIcon: Icons.favorite_rounded,
                  selected: index == 1,
                  tooltip: l10n.navFavorites,
                  onTap: () => navigationShell.goBranch(1),
                ),
                _NavIcon(
                  icon: Icons.search_rounded,
                  selectedIcon: Icons.search_rounded,
                  selected: index == 2,
                  tooltip: l10n.navSearch,
                  onTap: () => navigationShell.goBranch(2),
                ),
                _NavIcon(
                  icon: Icons.calendar_month_outlined,
                  selectedIcon: Icons.calendar_month_rounded,
                  selected: index == 3,
                  tooltip: l10n.navAppointments,
                  onTap: () => navigationShell.goBranch(3),
                ),
                _NavIcon(
                  icon: Icons.person_outline_rounded,
                  selectedIcon: Icons.person_rounded,
                  selected: index == 4,
                  tooltip: l10n.navProfile,
                  onTap: () => navigationShell.goBranch(4),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _NavIcon extends StatelessWidget {
  const _NavIcon({
    required this.icon,
    required this.selectedIcon,
    required this.selected,
    required this.tooltip,
    required this.onTap,
  });

  final IconData icon;
  final IconData selectedIcon;
  final bool selected;
  final String tooltip;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: tooltip,
      child: InkWell(
        onTap: onTap,
        customBorder: const CircleBorder(),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: selected ? AppColors.primary : Colors.transparent,
            shape: BoxShape.circle,
          ),
          child: Icon(
            selected ? selectedIcon : icon,
            size: 24,
            color: selected ? Colors.white : AppColors.muted,
          ),
        ),
      ),
    );
  }
}

class PatientHomePage extends StatelessWidget {
  const PatientHomePage({super.key});

  @override
  Widget build(BuildContext context) => const PatientHomeTab();
}

/// Behance-inspired patient profile: blue header + account settings rows.
class PatientProfilePage extends ConsumerWidget {
  const PatientProfilePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);
    final locale = ref.watch(localeControllerProvider);
    final name = auth.user?.displayName ?? 'Алексей Иванов';
    final email = 'patient@clinic.local';
    final items = _profileItems(l10n);

    return Scaffold(
      backgroundColor: AppColors.page,
      body: CustomScrollView(
        slivers: [
          SliverToBoxAdapter(
            child: Container(
              width: double.infinity,
              padding: EdgeInsets.only(
                top: MediaQuery.paddingOf(context).top + 18,
                left: 20,
                right: 20,
                bottom: 28,
              ),
              decoration: const BoxDecoration(
                color: AppColors.primary,
                borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    l10n.myProfile,
                    style: GoogleFonts.inter(
                      color: Colors.white,
                      fontSize: 22,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 22),
                  Row(
                    children: [
                      Container(
                        width: 64,
                        height: 64,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: Colors.white.withValues(alpha: 0.2),
                          border: Border.all(color: Colors.white, width: 2),
                        ),
                        alignment: Alignment.center,
                        child: Text(
                          _initials(name),
                          style: GoogleFonts.inter(
                            color: Colors.white,
                            fontSize: 22,
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
                              name,
                              style: GoogleFonts.inter(
                                color: Colors.white,
                                fontSize: 18,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              email,
                              style: GoogleFonts.inter(
                                color: Colors.white.withValues(alpha: 0.85),
                                fontSize: 13,
                                fontWeight: FontWeight.w400,
                              ),
                            ),
                          ],
                        ),
                      ),
                      IconButton(
                        onPressed: () =>
                            context.push('/patient/profile/personal'),
                        style: IconButton.styleFrom(
                          backgroundColor: Colors.white.withValues(alpha: 0.15),
                        ),
                        icon: const Icon(Icons.edit_outlined, color: Colors.white),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.fromLTRB(20, 22, 20, 12),
            sliver: SliverToBoxAdapter(
              child: Text(
                l10n.accountSettings,
                style: GoogleFonts.inter(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: AppColors.ink,
                ),
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            sliver: SliverList.separated(
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, index) {
                final item = items[index];
                return _SettingsTile(
                  icon: item.icon,
                  title: item.title,
                  onTap: () {
                    if (item.route == null) return;
                    if (item.route!.startsWith('/patient/profile/') ||
                        item.route == '/patient/medical') {
                      context.push(item.route!);
                    } else {
                      context.go(item.route!);
                    }
                  },
                );
              },
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.fromLTRB(20, 22, 20, 8),
            sliver: SliverToBoxAdapter(
              child: Text(
                l10n.language,
                style: GoogleFonts.inter(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: AppColors.ink,
                ),
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            sliver: SoftSliver(
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  _LangChip(
                    label: l10n.languageRussian,
                    selected: locale.languageCode == 'ru',
                    onTap: () => ref
                        .read(localeControllerProvider.notifier)
                        .setLocale(const Locale('ru')),
                  ),
                  _LangChip(
                    label: l10n.languageUzbek,
                    selected: locale.languageCode == 'uz',
                    onTap: () => ref
                        .read(localeControllerProvider.notifier)
                        .setLocale(const Locale('uz')),
                  ),
                  _LangChip(
                    label: l10n.languageEnglish,
                    selected: locale.languageCode == 'en',
                    onTap: () => ref
                        .read(localeControllerProvider.notifier)
                        .setLocale(const Locale('en')),
                  ),
                ],
              ),
            ),
          ),
          if (kSkipLogin)
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(20, 24, 20, 8),
              sliver: SoftSliver(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      l10n.uiPreview,
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.w600,
                        color: AppColors.muted,
                      ),
                    ),
                    const SizedBox(height: 10),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        for (final role in const [
                          AppRole.patient,
                          AppRole.doctor,
                          AppRole.owner,
                          AppRole.accountant,
                        ])
                          FilterChip(
                            selected: auth.role == role,
                            label: Text(_roleLabel(l10n, role)),
                            selectedColor: AppColors.softBlue,
                            checkmarkColor: AppColors.primary,
                            labelStyle: GoogleFonts.inter(
                              color: auth.role == role
                                  ? AppColors.primary
                                  : AppColors.ink,
                              fontWeight: FontWeight.w500,
                            ),
                            onSelected: (_) {
                              ref
                                  .read(authControllerProvider.notifier)
                                  .setPreviewRole(role);
                              context.go(switch (role) {
                                AppRole.doctor => '/doctor',
                                AppRole.owner => '/owner',
                                AppRole.accountant => '/accountant',
                                _ => '/patient/profile',
                              });
                            },
                          ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          const SliverToBoxAdapter(child: SizedBox(height: 28)),
        ],
      ),
    );
  }

  static String _initials(String name) {
    final parts = name.trim().split(RegExp(r'\s+'));
    if (parts.isEmpty) return '?';
    if (parts.length == 1) return parts.first.substring(0, 1).toUpperCase();
    return (parts.first[0] + parts.last[0]).toUpperCase();
  }

  static String _roleLabel(AppLocalizations l10n, AppRole role) =>
      switch (role) {
        AppRole.patient => l10n.rolePatient,
        AppRole.doctor => l10n.roleDoctor,
        AppRole.owner => l10n.roleOwner,
        AppRole.accountant => l10n.roleAccountant,
        AppRole.unsupported => l10n.roleOther,
      };

  static List<_ProfileItem> _profileItems(AppLocalizations l10n) => [
        _ProfileItem(
          Icons.person_outline_rounded,
          l10n.personalInformation,
          route: '/patient/profile/personal',
        ),
        _ProfileItem(
          Icons.calendar_month_outlined,
          l10n.bookingHistory,
          route: '/patient/appointments',
        ),
        _ProfileItem(
          Icons.credit_card_outlined,
          l10n.myCards,
          route: '/patient/profile/cards',
        ),
        _ProfileItem(
          Icons.science_outlined,
          l10n.myTests,
          route: '/patient/medical',
        ),
        _ProfileItem(
          Icons.notifications_none_rounded,
          l10n.notifications,
          route: '/patient/profile/notifications',
        ),
        _ProfileItem(
          Icons.lock_outline_rounded,
          l10n.security,
          route: '/patient/profile/security',
        ),
        _ProfileItem(
          Icons.help_outline_rounded,
          l10n.helpCenter,
          route: '/patient/profile/help',
        ),
      ];
}

class _ProfileItem {
  const _ProfileItem(this.icon, this.title, {this.route});
  final IconData icon;
  final String title;
  final String? route;
}

class _LangChip extends StatelessWidget {
  const _LangChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return FilterChip(
      selected: selected,
      label: Text(label),
      selectedColor: AppColors.softBlue,
      checkmarkColor: AppColors.primary,
      labelStyle: GoogleFonts.inter(
        color: selected ? AppColors.primary : AppColors.ink,
        fontWeight: FontWeight.w600,
      ),
      onSelected: (_) => onTap(),
    );
  }
}

class _SettingsTile extends StatelessWidget {
  const _SettingsTile({
    required this.icon,
    required this.title,
    required this.onTap,
  });

  final IconData icon;
  final String title;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
          child: Row(
            children: [
              Container(
                width: 42,
                height: 42,
                decoration: BoxDecoration(
                  color: AppColors.softBlue,
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Icon(icon, color: AppColors.primary, size: 22),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  title,
                  style: GoogleFonts.inter(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: AppColors.ink,
                  ),
                ),
              ),
              const Icon(Icons.chevron_right, color: AppColors.muted),
            ],
          ),
        ),
      ),
    );
  }
}

class SoftSliver extends StatelessWidget {
  const SoftSliver({super.key, required this.child});
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return SliverToBoxAdapter(child: child);
  }
}
