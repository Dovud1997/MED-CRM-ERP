import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

class PersonalInfoScreen extends ConsumerStatefulWidget {
  const PersonalInfoScreen({super.key});

  @override
  ConsumerState<PersonalInfoScreen> createState() => _PersonalInfoScreenState();
}

class _PersonalInfoScreenState extends ConsumerState<PersonalInfoScreen> {
  late final TextEditingController _name;
  late final TextEditingController _phone;
  late final TextEditingController _email;
  late final TextEditingController _dob;

  @override
  void initState() {
    super.initState();
    final user = ref.read(authControllerProvider).user;
    _name = TextEditingController(
      text: (user?.displayName ?? 'Алексей Иванов').replaceAll(' (demo)', ''),
    );
    _phone = TextEditingController(text: '+998 90 123 45 67');
    _email = TextEditingController(text: 'patient@clinic.local');
    _dob = TextEditingController(text: '12.03.1992');
  }

  @override
  void dispose() {
    _name.dispose();
    _phone.dispose();
    _email.dispose();
    _dob.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.personalInformation,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          TextField(
            controller: _name,
            decoration: InputDecoration(labelText: l10n.fullName),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _phone,
            decoration: InputDecoration(labelText: l10n.phone),
            keyboardType: TextInputType.phone,
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _email,
            decoration: InputDecoration(labelText: l10n.email),
            keyboardType: TextInputType.emailAddress,
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _dob,
            decoration: InputDecoration(labelText: l10n.dateOfBirth),
          ),
          const SizedBox(height: 24),
          ClinicPrimaryButton(
            label: l10n.save,
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text(l10n.done)),
              );
            },
          ),
        ],
      ),
    );
  }
}

class MyCardsScreen extends StatefulWidget {
  const MyCardsScreen({super.key});

  @override
  State<MyCardsScreen> createState() => _MyCardsScreenState();
}

class _MyCardsScreenState extends State<MyCardsScreen> {
  final _cards = [
    ('Uzcard', '**** 4582'),
    ('Humo', '**** 9910'),
  ];

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.myCards,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          ..._cards.map(
            (c) => Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: Container(
                height: 140,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(22),
                  gradient: LinearGradient(
                    colors: c.$1 == 'Uzcard'
                        ? const [Color(0xFF0564F2), Color(0xFF3B82F6)]
                        : const [Color(0xFF0F766E), Color(0xFF14B8A6)],
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      c.$1,
                      style: GoogleFonts.inter(
                        color: Colors.white70,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const Spacer(),
                    Text(
                      c.$2,
                      style: GoogleFonts.inter(
                        color: Colors.white,
                        fontSize: 22,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 2,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          ClinicPrimaryButton(
            label: l10n.add,
            onPressed: () {
              setState(() => _cards.add(('Visa', '**** 1200')));
            },
          ),
        ],
      ),
    );
  }
}

class NotificationsScreen extends StatefulWidget {
  const NotificationsScreen({super.key});

  @override
  State<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends State<NotificationsScreen> {
  bool _push = true;
  bool _reminders = true;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final items = [
      (l10n.appointmentReminders, l10n.notifyApptReminderBody),
      (l10n.myTests, l10n.notifyLabReadyBody),
      (l10n.navDoctors, l10n.notifyNewSlotBody),
    ];
    return ClinicPageScaffold(
      title: l10n.notifications,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          ClinicCard(
            child: Column(
              children: [
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(l10n.pushNotifications),
                  value: _push,
                  activeColor: AppColors.primary,
                  onChanged: (v) => setState(() => _push = v),
                ),
                const Divider(height: 1),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(l10n.appointmentReminders),
                  value: _reminders,
                  activeColor: AppColors.primary,
                  onChanged: (v) => setState(() => _reminders = v),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          Text(
            l10n.notifications,
            style: GoogleFonts.inter(
              fontWeight: FontWeight.w700,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 12),
          ...items.map(
            (n) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: ClinicCard(
                child: Row(
                  children: [
                    Container(
                      width: 42,
                      height: 42,
                      decoration: BoxDecoration(
                        color: AppColors.softBlue,
                        borderRadius: BorderRadius.circular(14),
                      ),
                      child: const Icon(
                        Icons.notifications_none_rounded,
                        color: AppColors.primary,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            n.$1,
                            style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                          ),
                          Text(
                            n.$2,
                            style: GoogleFonts.inter(
                              color: AppColors.muted,
                              fontSize: 13,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class SecurityScreen extends StatelessWidget {
  const SecurityScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.security,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          ClinicCard(
            onTap: () {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text(l10n.comingSoon)),
              );
            },
            child: Row(
              children: [
                const Icon(Icons.lock_outline_rounded, color: AppColors.primary),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    l10n.changePassword,
                    style: GoogleFonts.inter(fontWeight: FontWeight.w600),
                  ),
                ),
                const Icon(Icons.chevron_right, color: AppColors.muted),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class HelpCenterScreen extends StatelessWidget {
  const HelpCenterScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final faqs = [
      (l10n.bookAppointment, l10n.searchDoctorHint),
      (l10n.myAppointments, l10n.noUpcomingAppointment),
      (
        l10n.language,
        '${l10n.languageRussian} / ${l10n.languageUzbek} / ${l10n.languageEnglish}',
      ),
    ];
    return ClinicPageScaffold(
      title: l10n.helpCenter,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text(
            l10n.faq,
            style: GoogleFonts.inter(fontWeight: FontWeight.w700, fontSize: 16),
          ),
          const SizedBox(height: 12),
          ...faqs.map(
            (f) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: ClinicCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      f.$1,
                      style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 6),
                    Text(
                      f.$2,
                      style: GoogleFonts.inter(color: AppColors.muted),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 8),
          ClinicPrimaryButton(
            label: l10n.contactSupport,
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('+998 71 000 00 00')),
              );
            },
          ),
        ],
      ),
    );
  }
}
