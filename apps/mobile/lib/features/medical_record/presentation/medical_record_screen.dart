import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/features/medical_record/data/clinical_repository.dart';
import 'package:clinicos_mobile/shared/models/clinical_record.dart';
import 'package:clinicos_mobile/shared/widgets/async_body.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

class MedicalRecordScreen extends ConsumerStatefulWidget {
  const MedicalRecordScreen({super.key});

  @override
  ConsumerState<MedicalRecordScreen> createState() =>
      _MedicalRecordScreenState();
}

class _MedicalRecordScreenState extends ConsumerState<MedicalRecordScreen> {
  final _search = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final authPatientId = ref.watch(authControllerProvider).user?.patientId;
    final patientId = ref.watch(selectedPatientIdProvider) ?? authPatientId;
    final isSelf = authPatientId != null || kSkipLogin;
    final searchAsync = ref.watch(patientSearchProvider(_query));

    return DefaultTabController(
      length: 4,
      child: Scaffold(
        backgroundColor: AppColors.page,
        appBar: AppBar(
          title: Text(l10n.myMedicalRecord),
          bottom: TabBar(
            isScrollable: true,
            labelColor: AppColors.primary,
            unselectedLabelColor: AppColors.muted,
            indicatorColor: AppColors.primary,
            tabs: [
              Tab(text: l10n.medicalOverview),
              Tab(text: l10n.medicalTimeline),
              Tab(text: l10n.medicalLabs),
              Tab(text: l10n.medicalPrescriptions),
            ],
          ),
        ),
        body: Column(
          children: [
            if (!isSelf)
              Padding(
                padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
                child: Column(
                  children: [
                    TextField(
                      controller: _search,
                      decoration: InputDecoration(
                        prefixIcon: const Icon(Icons.search),
                        hintText: l10n.searchPatient,
                      ),
                      onChanged: (v) => setState(() => _query = v),
                    ),
                    if (_query.trim().length >= 2)
                      searchAsync.when(
                        data: (items) => Column(
                          children: items
                              .take(5)
                              .map(
                                (p) => ListTile(
                                  title: Text(p.fullName),
                                  onTap: () {
                                    ref
                                        .read(selectedPatientIdProvider
                                            .notifier)
                                        .state = p.id;
                                    setState(() {
                                      _query = '';
                                      _search.text = p.fullName;
                                    });
                                  },
                                ),
                              )
                              .toList(),
                        ),
                        loading: () => const LinearProgressIndicator(),
                        error: (_, __) => const SizedBox.shrink(),
                      ),
                  ],
                ),
              ),
            Expanded(
              child: patientId == null
                  ? ClinicEmptyState(
                      message: l10n.selectPatientFirst,
                      icon: Icons.folder_shared_outlined,
                    )
                  : Consumer(
                      builder: (context, ref, _) {
                        final async =
                            ref.watch(clinicalRecordProvider(patientId));
                        return AsyncBody<ClinicalRecord>(
                          value: async.when(
                            data: AsyncValueLike.data,
                            error: (e, _) => AsyncValueLike.error(e),
                            loading: AsyncValueLike.loading,
                          ),
                          onRetry: () =>
                              ref.invalidate(clinicalRecordProvider(patientId)),
                          builder: (record) => TabBarView(
                            children: [
                              _OverviewTab(record: record),
                              _TimelineTab(entries: record.history),
                              _LabsTab(
                                labs: record.labResults,
                                imaging: record.imagingStudies,
                              ),
                              _OrdersTab(orders: record.orders),
                            ],
                          ),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}

class _OverviewTab extends StatelessWidget {
  const _OverviewTab({required this.record});
  final ClinicalRecord record;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        ClinicCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _kv(l10n.bloodGroup, record.bloodGroup),
              if (record.heightCm != null)
                _kv(l10n.height, '${record.heightCm!.toStringAsFixed(0)} cm'),
              if (record.weightKg != null)
                _kv(l10n.weight, '${record.weightKg!.toStringAsFixed(0)} kg'),
            ],
          ),
        ),
        const SizedBox(height: 12),
        ClinicCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l10n.allergies,
                style: GoogleFonts.inter(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              if (record.allergies.isEmpty)
                Text(l10n.noAllergies,
                    style: GoogleFonts.inter(color: AppColors.muted))
              else
                ...record.allergies.map(
                  (a) => Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Text(
                      '${a.allergen} · ${a.severity}',
                      style: GoogleFonts.inter(color: AppColors.ink),
                    ),
                  ),
                ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        ClinicCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l10n.diagnoses,
                style: GoogleFonts.inter(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              if (record.diagnoses.isEmpty)
                Text(l10n.empty,
                    style: GoogleFonts.inter(color: AppColors.muted))
              else
                ...record.diagnoses.take(6).map(
                      (d) => Padding(
                        padding: const EdgeInsets.only(bottom: 6),
                        child: Text(
                          d.name,
                          style: GoogleFonts.inter(fontWeight: FontWeight.w600),
                        ),
                      ),
                    ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _kv(String k, String v) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Expanded(
            child: Text(k, style: GoogleFonts.inter(color: AppColors.muted)),
          ),
          Text(v, style: GoogleFonts.inter(fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

class _TimelineTab extends StatelessWidget {
  const _TimelineTab({required this.entries});
  final List<HistoryEntry> entries;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final locale = l10n.locale.toString();
    if (entries.isEmpty) {
      return ClinicEmptyState(message: l10n.empty);
    }
    return ListView.separated(
      padding: const EdgeInsets.all(20),
      itemCount: entries.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (context, index) {
        final e = entries[index];
        return ClinicCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                DateFormat('d MMM yyyy', locale).format(e.occurredAt),
                style: GoogleFonts.inter(color: AppColors.muted, fontSize: 12),
              ),
              const SizedBox(height: 4),
              Text(
                e.title,
                style: GoogleFonts.inter(fontWeight: FontWeight.w700),
              ),
              if (e.author != null)
                Text(
                  e.author!,
                  style: GoogleFonts.inter(color: AppColors.muted),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _LabsTab extends StatelessWidget {
  const _LabsTab({required this.labs, required this.imaging});
  final List<LabResult> labs;
  final List<ImagingStudy> imaging;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    if (labs.isEmpty && imaging.isEmpty) {
      return ClinicEmptyState(message: l10n.empty);
    }
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        ...labs.map(
          (l) => Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: ClinicCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    l.testName,
                    style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                  ),
                  Text(
                    '${l.result}${l.unit != null ? ' ${l.unit}' : ''}',
                    style: GoogleFonts.inter(color: AppColors.muted),
                  ),
                ],
              ),
            ),
          ),
        ),
        ...imaging.map(
          (i) => Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: ClinicCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${i.modality} · ${i.bodyArea}',
                    style: GoogleFonts.inter(fontWeight: FontWeight.w700),
                  ),
                  Text(i.diagnosis,
                      style: GoogleFonts.inter(color: AppColors.muted)),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _OrdersTab extends StatelessWidget {
  const _OrdersTab({required this.orders});
  final List<ClinicalOrder> orders;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    if (orders.isEmpty) {
      return ClinicEmptyState(message: l10n.empty);
    }
    return ListView.separated(
      padding: const EdgeInsets.all(20),
      itemCount: orders.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (context, index) {
        final o = orders[index];
        return ClinicCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                o.title,
                style: GoogleFonts.inter(fontWeight: FontWeight.w700),
              ),
              if (o.dosage != null)
                Text(o.dosage!,
                    style: GoogleFonts.inter(color: AppColors.muted)),
              Text(o.status,
                  style: GoogleFonts.inter(
                      fontSize: 12, color: AppColors.muted)),
            ],
          ),
        );
      },
    );
  }
}
