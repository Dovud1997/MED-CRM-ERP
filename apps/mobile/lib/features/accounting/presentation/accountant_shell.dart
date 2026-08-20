import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/auth/app_role.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

class AccountingSummary {
  const AccountingSummary({
    required this.income,
    required this.expense,
    required this.balance,
    required this.entries,
    required this.obligations,
  });

  final int income;
  final int expense;
  final int balance;
  final List<Map<String, dynamic>> entries;
  final List<Map<String, dynamic>> obligations;
}

final _accountingProvider =
    FutureProvider.autoDispose<AccountingSummary?>((ref) async {
  try {
    final api = ref.watch(apiClientProvider);
    final now = DateTime.now();
    final from = DateTime(now.year, now.month, 1);
    final to = DateTime(now.year, now.month + 1, 1);
    final data = await api.get<Map<String, dynamic>>(
      '/accounting',
      query: {
        'from': DateFormat('yyyy-MM-dd').format(from),
        'to': DateFormat('yyyy-MM-dd').format(to),
      },
      parser: (raw) => Map<String, dynamic>.from(raw as Map),
    );
    final entries = (data['entries'] as List<dynamic>? ?? const [])
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();
    final obligations = (data['obligations'] as List<dynamic>? ?? const [])
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();
    var income = 0;
    var expense = 0;
    for (final e in entries) {
      final amount = (e['amountUzs'] as num?)?.toInt() ?? 0;
      final type = e['type']?.toString().toUpperCase() ?? '';
      if (type.contains('INCOME') || type == 'CREDIT' || type == 'REVENUE' || type == 'IN') {
        income += amount;
      } else {
        expense += amount;
      }
    }
    final summary = data['summary'];
    if (summary is Map) {
      income = (summary['income'] as num?)?.toInt() ?? income;
      expense = (summary['expense'] as num?)?.toInt() ?? expense;
    }
    return AccountingSummary(
      income: income,
      expense: expense,
      balance: (summary is Map ? (summary['profit'] as num?)?.toInt() : null) ??
          (income - expense),
      entries: entries,
      obligations: obligations,
    );
  } catch (_) {
    if (!kSkipLogin) return null;
    return const AccountingSummary(
      income: 48500000,
      expense: 31200000,
      balance: 17300000,
      entries: [
        {
          'description': 'Оплата консультации',
          'type': 'INCOME',
          'occurredOn': '2026-08-09',
          'amountUzs': 350000,
          'method': 'CARD',
        },
        {
          'description': 'Закуп расходников',
          'type': 'EXPENSE',
          'occurredOn': '2026-08-08',
          'amountUzs': 1200000,
          'method': 'TRANSFER',
        },
        {
          'description': 'Лабораторные услуги',
          'type': 'INCOME',
          'occurredOn': '2026-08-07',
          'amountUzs': 780000,
          'method': 'CASH',
        },
      ],
      obligations: [
        {
          'counterparty': 'ООО MedSupply',
          'status': 'OPEN',
          'type': 'PAYABLE',
          'amountUzs': 4500000,
          'paidUzs': 1000000,
        },
        {
          'counterparty': 'Пациент Рахимов',
          'status': 'PARTIAL',
          'type': 'RECEIVABLE',
          'amountUzs': 900000,
          'paidUzs': 300000,
        },
      ],
    );
  }
});

class AccountantShell extends ConsumerStatefulWidget {
  const AccountantShell({super.key});

  @override
  ConsumerState<AccountantShell> createState() => _AccountantShellState();
}

class _AccountantShellState extends ConsumerState<AccountantShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final async = ref.watch(_accountingProvider);
    final money = NumberFormat.decimalPattern(l10n.locale.toString());

    return Scaffold(
      backgroundColor: AppColors.page,
      body: AsyncBody<AccountingSummary?>(
        value: async.when(
          data: AsyncValueLike.data,
          error: (e, _) => AsyncValueLike.error(e),
          loading: AsyncValueLike.loading,
        ),
        onRetry: () => ref.invalidate(_accountingProvider),
        builder: (data) {
          if (data == null) {
            return ClinicPageScaffold(
              title: l10n.navHome,
              body: ClinicEmptyState(message: l10n.empty),
            );
          }
          final pages = [
            _SummaryPage(data: data, money: money),
            _EntriesPage(entries: data.entries, money: money, title: l10n.navOperations),
            _EntriesPage(entries: data.entries, money: money, title: l10n.navReports),
            _DebtsPage(obligations: data.obligations, money: money),
            ClinicPageScaffold(
              title: l10n.navProfile,
              body: ListView(
                padding: const EdgeInsets.all(20),
                children: [
                  ClinicCard(
                    onTap: () {
                      ref
                          .read(authControllerProvider.notifier)
                          .setPreviewRole(AppRole.patient);
                      context.go('/patient/profile');
                    },
                    child: Text(
                      l10n.uiPreview,
                      style: GoogleFonts.inter(fontWeight: FontWeight.w600),
                    ),
                  ),
                ],
              ),
            ),
          ];
          return pages[_index];
        },
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.home_outlined),
            label: l10n.navHome,
          ),
          NavigationDestination(
            icon: const Icon(Icons.receipt_long_outlined),
            label: l10n.navOperations,
          ),
          NavigationDestination(
            icon: const Icon(Icons.bar_chart_outlined),
            label: l10n.navReports,
          ),
          NavigationDestination(
            icon: const Icon(Icons.money_off_outlined),
            label: l10n.navDebts,
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

class _SummaryPage extends ConsumerWidget {
  const _SummaryPage({required this.data, required this.money});
  final AccountingSummary data;
  final NumberFormat money;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.navHome,
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(_accountingProvider),
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            Wrap(
              spacing: 10,
              runSpacing: 10,
              children: [
                _Metric(l10n.revenue, money.format(data.income)),
                _Metric(l10n.navOperations, money.format(data.expense)),
                _Metric(l10n.navFinance, money.format(data.balance)),
                _Metric(
                  l10n.debts,
                  '${data.obligations.where((o) => o['status'] == 'OPEN' || o['status'] == 'PARTIAL').length}',
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _EntriesPage extends StatelessWidget {
  const _EntriesPage({
    required this.entries,
    required this.money,
    required this.title,
  });

  final List<Map<String, dynamic>> entries;
  final NumberFormat money;
  final String title;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: title,
      body: entries.isEmpty
          ? ClinicEmptyState(message: l10n.empty)
          : ListView.separated(
              padding: const EdgeInsets.all(20),
              itemCount: entries.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, i) {
                final e = entries[i];
                return ClinicCard(
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              e['description']?.toString() ?? '',
                              style: GoogleFonts.inter(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            Text(
                              '${e['type']} · ${e['occurredOn'] ?? ''}',
                              style: GoogleFonts.inter(
                                color: AppColors.muted,
                                fontSize: 12,
                              ),
                            ),
                          ],
                        ),
                      ),
                      Text(
                        money.format(e['amountUzs'] ?? 0),
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                      ),
                    ],
                  ),
                );
              },
            ),
    );
  }
}

class _DebtsPage extends StatelessWidget {
  const _DebtsPage({required this.obligations, required this.money});
  final List<Map<String, dynamic>> obligations;
  final NumberFormat money;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ClinicPageScaffold(
      title: l10n.navDebts,
      body: obligations.isEmpty
          ? ClinicEmptyState(message: l10n.empty)
          : ListView.separated(
              padding: const EdgeInsets.all(20),
              itemCount: obligations.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, i) {
                final o = obligations[i];
                final amount = (o['amountUzs'] as num?)?.toInt() ?? 0;
                final paid = (o['paidUzs'] as num?)?.toInt() ?? 0;
                return ClinicCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        o['counterparty']?.toString() ??
                            o['description']?.toString() ??
                            '',
                        style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                      ),
                      Text(
                        '${o['status']} · ${money.format(amount - paid)}',
                        style: GoogleFonts.inter(color: AppColors.muted),
                      ),
                    ],
                  ),
                );
              },
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
              fontSize: 18,
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
