import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/auth/app_role.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/appointments/data/appointments_repository.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/models/appointment.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

final _doctorDashProvider =
    FutureProvider.autoDispose<AppointmentsDashboard?>((ref) async {
  try {
    return await ref.watch(appointmentsRepositoryProvider).dashboard();
  } catch (_) {
    return null;
  }
});

class DoctorShell extends ConsumerStatefulWidget {
  const DoctorShell({super.key});

  @override
  ConsumerState<DoctorShell> createState() => _DoctorShellState();
}

class _DoctorShellState extends ConsumerState<DoctorShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final pages = [
      const _DoctorTodayPage(),
      const _DoctorSchedulePage(),
      const _DoctorPatientsPage(),
      const _DoctorMessagesPage(),
      _DoctorProfilePage(onBackToPatient: () {
        ref.read(authControllerProvider.notifier).setPreviewRole(AppRole.patient);
        context.go('/patient/profile');
      }),
    ];

    return Scaffold(
      backgroundColor: AppColors.page,
      body: pages[_index],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.today_outlined),
            label: l10n.navToday,
          ),
          NavigationDestination(
            icon: const Icon(Icons.calendar_view_week_outlined),
            label: l10n.navSchedule,
          ),
          NavigationDestination(
            icon: const Icon(Icons.groups_outlined),
            label: l10n.navPatients,
          ),
          NavigationDestination(
            icon: const Icon(Icons.chat_outlined),
            label: l10n.navMessages,
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

class _DoctorTodayPage extends ConsumerWidget {
  const _DoctorTodayPage();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final dash = ref.watch(_doctorDashProvider);
    return ClinicPageScaffold(
      title: l10n.navToday,
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(_doctorDashProvider),
        child: AsyncBody<AppointmentsDashboard?>(
          value: dash.when(
            data: AsyncValueLike.data,
            error: (e, _) => AsyncValueLike.error(e),
            loading: AsyncValueLike.loading,
          ),
          onRetry: () => ref.invalidate(_doctorDashProvider),
          builder: (data) {
            if (data == null) {
              return ListView(
                children: [ClinicEmptyState(message: l10n.empty)],
              );
            }
            return ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    _Metric(l10n.todayAppointments, '${data.today}'),
                    _Metric(l10n.tabUpcoming, '${data.waiting}'),
                    _Metric(l10n.tabPast, '${data.completed}'),
                  ],
                ),
                const SizedBox(height: 20),
                ClinicSectionHeader(title: l10n.upcomingAppointment),
                const SizedBox(height: 10),
                if (data.upcoming.isEmpty)
                  ClinicEmptyState(message: l10n.noUpcomingAppointment)
                else
                  ...data.upcoming.take(8).map(
                        (a) => Padding(
                          padding: const EdgeInsets.only(bottom: 10),
                          child: ClinicCard(
                            child: Row(
                              children: [
                                DoctorAvatar(
                                  name: a.patientName ?? a.employeeName ?? 'P',
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        DateFormat('HH:mm').format(a.startsAt),
                                        style: GoogleFonts.inter(
                                          fontWeight: FontWeight.w700,
                                        ),
                                      ),
                                      Text(
                                        a.patientName ??
                                            a.employeeName ??
                                            l10n.navPatients,
                                        style: GoogleFonts.inter(
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
                      ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _DoctorSchedulePage extends ConsumerWidget {
  const _DoctorSchedulePage();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final dash = ref.watch(_doctorDashProvider);
    return ClinicPageScaffold(
      title: l10n.navSchedule,
      body: AsyncBody<AppointmentsDashboard?>(
        value: dash.when(
          data: AsyncValueLike.data,
          error: (e, _) => AsyncValueLike.error(e),
          loading: AsyncValueLike.loading,
        ),
        builder: (data) {
          final items = data?.upcoming ?? const <Appointment>[];
          if (items.isEmpty) {
            return ClinicEmptyState(message: l10n.empty);
          }
          return ListView.separated(
            padding: const EdgeInsets.all(20),
            itemCount: items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 10),
            itemBuilder: (context, index) {
              final a = items[index];
              return ClinicCard(
                child: Text(
                  '${DateFormat('d MMM HH:mm').format(a.startsAt)} · ${a.patientName ?? a.status}',
                  style: GoogleFonts.inter(fontWeight: FontWeight.w600),
                ),
              );
            },
          );
        },
      ),
    );
  }
}

class _DoctorPatientsPage extends ConsumerWidget {
  const _DoctorPatientsPage();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final dash = ref.watch(_doctorDashProvider);
    final names = <String>{};
    for (final a in dash.valueOrNull?.upcoming ?? const <Appointment>[]) {
      if (a.patientName != null && a.patientName!.isNotEmpty) {
        names.add(a.patientName!);
      }
    }
    return ClinicPageScaffold(
      title: l10n.navPatients,
      body: names.isEmpty
          ? ClinicEmptyState(message: l10n.empty)
          : ListView.separated(
              padding: const EdgeInsets.all(20),
              itemCount: names.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, index) {
                final name = names.elementAt(index);
                return ClinicCard(
                  child: Row(
                    children: [
                      DoctorAvatar(name: name),
                      const SizedBox(width: 12),
                      Text(
                        name,
                        style: GoogleFonts.inter(fontWeight: FontWeight.w600),
                      ),
                    ],
                  ),
                );
              },
            ),
    );
  }
}

class _DoctorMessagesPage extends StatelessWidget {
  const _DoctorMessagesPage();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final threads = const [
      ('Пациент Алексей', 'Когда можно на повторный приём?'),
      ('Ресепшн', 'Пациент ждёт в холле'),
      ('Каримова Азиза', 'Передайте карту ребёнка'),
    ];
    return ClinicPageScaffold(
      title: l10n.navMessages,
      body: ListView.separated(
        padding: const EdgeInsets.all(20),
        itemCount: threads.length,
        separatorBuilder: (_, __) => const SizedBox(height: 10),
        itemBuilder: (context, index) {
          final t = threads[index];
          return ClinicCard(
            onTap: () => Navigator.of(context).push(
              MaterialPageRoute<void>(
                builder: (_) => _ChatThreadPage(
                  title: t.$1,
                  preview: t.$2,
                ),
              ),
            ),
            child: Row(
              children: [
                DoctorAvatar(name: t.$1),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        t.$1,
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                      ),
                      Text(
                        t.$2,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
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
          );
        },
      ),
    );
  }
}

class _ChatThreadPage extends StatefulWidget {
  const _ChatThreadPage({required this.title, required this.preview});
  final String title;
  final String preview;

  @override
  State<_ChatThreadPage> createState() => _ChatThreadPageState();
}

class _ChatThreadPageState extends State<_ChatThreadPage> {
  final _controller = TextEditingController();
  late final List<(bool mine, String text)> _messages;

  @override
  void initState() {
    super.initState();
    _messages = [
      (false, widget.preview),
      (true, 'Хорошо, уточню расписание и отвечу.'),
    ];
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      backgroundColor: AppColors.page,
      appBar: AppBar(title: Text(widget.title)),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(20),
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                final m = _messages[index];
                return Align(
                  alignment:
                      m.$1 ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: const EdgeInsets.only(bottom: 10),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 10,
                    ),
                    constraints: BoxConstraints(
                      maxWidth: MediaQuery.sizeOf(context).width * 0.75,
                    ),
                    decoration: BoxDecoration(
                      color: m.$1 ? AppColors.primary : Colors.white,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(
                      m.$2,
                      style: GoogleFonts.inter(
                        color: m.$1 ? Colors.white : AppColors.ink,
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _controller,
                      decoration: InputDecoration(
                        hintText: l10n.typeReply,
                        filled: true,
                        fillColor: Colors.white,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filled(
                    onPressed: () {
                      final text = _controller.text.trim();
                      if (text.isEmpty) return;
                      setState(() {
                        _messages.add((true, text));
                        _controller.clear();
                      });
                    },
                    style: IconButton.styleFrom(
                      backgroundColor: AppColors.primary,
                    ),
                    icon: const Icon(Icons.send_rounded, color: Colors.white),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _DoctorProfilePage extends StatelessWidget {
  const _DoctorProfilePage({required this.onBackToPatient});
  final VoidCallback onBackToPatient;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.navProfile,
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          ClinicCard(
            onTap: onBackToPatient,
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
