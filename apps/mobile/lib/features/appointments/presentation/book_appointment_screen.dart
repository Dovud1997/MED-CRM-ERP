import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/network/api_exception.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/appointments/data/appointments_repository.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/features/doctors/data/doctors_repository.dart';
import 'package:clinicos_mobile/features/patient/data/patients_repository.dart';
import 'package:clinicos_mobile/shared/models/doctor.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

class BookAppointmentScreen extends ConsumerStatefulWidget {
  const BookAppointmentScreen({super.key, required this.doctorId});

  final String doctorId;

  @override
  ConsumerState<BookAppointmentScreen> createState() =>
      _BookAppointmentScreenState();
}

class _BookAppointmentScreenState extends ConsumerState<BookAppointmentScreen> {
  DateTime _day = DateTime.now();
  DateTime? _slot;
  Doctor? _doctor;
  List<DateTime> _slots = const [];
  bool _loadingSlots = true;
  bool _submitting = false;
  String? _error;
  PatientSummary? _selectedPatient;
  final _patientQuery = TextEditingController();
  List<PatientSummary> _patientHits = const [];

  @override
  void initState() {
    super.initState();
    _bootstrap();
  }

  @override
  void dispose() {
    _patientQuery.dispose();
    super.dispose();
  }

  Future<void> _bootstrap() async {
    try {
      final doctor =
          await ref.read(doctorsRepositoryProvider).getDoctor(widget.doctorId);
      if (!mounted) return;
      setState(() => _doctor = doctor);
      await _loadSlots();
    } catch (_) {
      if (!mounted) return;
      setState(() => _loadingSlots = false);
    }
  }

  Future<void> _loadSlots() async {
    setState(() {
      _loadingSlots = true;
      _slot = null;
    });
    try {
      final slots = await ref.read(doctorsRepositoryProvider).availabilitySlots(
            employeeId: widget.doctorId,
            day: _day,
          );
      if (!mounted) return;
      setState(() {
        _slots = slots;
        _loadingSlots = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _slots = const [];
        _loadingSlots = false;
      });
    }
  }

  Future<void> _searchPatients(String q) async {
    if (q.trim().length < 2) {
      setState(() => _patientHits = const []);
      return;
    }
    try {
      final hits = await ref.read(patientsRepositoryProvider).search(q.trim());
      if (!mounted) return;
      setState(() => _patientHits = hits);
    } catch (_) {
      if (!mounted) return;
      setState(() => _patientHits = const []);
    }
  }

  Future<void> _confirmAndSubmit() async {
    final l10n = AppLocalizations.of(context);
    final doctor = _doctor;
    final slot = _slot;
    if (doctor == null || slot == null) return;
    final locale = l10n.locale.toString();
    final ok = await showModalBottomSheet<bool>(
      context: context,
      showDragHandle: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (context) {
        return Padding(
          padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l10n.bookingConfirmTitle,
                style: GoogleFonts.inter(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 14),
              Text(
                doctor.fullName,
                style: GoogleFonts.inter(fontWeight: FontWeight.w600),
              ),
              Text(
                doctor.specialtyLabel,
                style: GoogleFonts.inter(color: AppColors.muted),
              ),
              const SizedBox(height: 8),
              Text(
                DateFormat('d MMMM yyyy · HH:mm', locale).format(slot),
                style: GoogleFonts.inter(fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 20),
              ClinicPrimaryButton(
                label: l10n.confirmBooking,
                onPressed: () => Navigator.pop(context, true),
              ),
              const SizedBox(height: 8),
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: Text(l10n.cancel),
              ),
            ],
          ),
        );
      },
    );
    if (ok == true) await _submit();
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context);
    final doctor = _doctor;
    final slot = _slot;
    if (doctor == null || slot == null) return;

    final authPatientId = ref.read(authControllerProvider).user?.patientId;
    final patientId = authPatientId ??
        _selectedPatient?.id ??
        (kSkipLogin ? 'demo-patient' : null);
    if (patientId == null || doctor.branchId == null) {
      setState(() => _error = l10n.errorValidation);
      return;
    }

    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await ref.read(appointmentsRepositoryProvider).create(
            patientId: patientId,
            employeeId: doctor.id,
            branchId: doctor.branchId!,
            startsAt: slot,
          );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.bookingSuccess)),
      );
      context.go('/patient/appointments');
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _submitting = false;
        _error = e is ApiException ? e.message : l10n.errorGeneric;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final locale = l10n.locale.toString();
    final doctor = _doctor;
    final needsPatientPicker =
        !kSkipLogin &&
        ref.watch(authControllerProvider).user?.patientId == null;
    final days = List.generate(14, (i) {
      final d = DateTime.now().add(Duration(days: i));
      return DateTime(d.year, d.month, d.day);
    });

    return Scaffold(
      backgroundColor: AppColors.page,
      appBar: AppBar(title: Text(l10n.bookAppointment)),
      body: doctor == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 120),
              children: [
                ClinicCard(
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
                                fontSize: 16,
                              ),
                            ),
                            Text(
                              doctor.specialtyLabel,
                              style: GoogleFonts.inter(color: AppColors.muted),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                if (needsPatientPicker) ...[
                  const SizedBox(height: 20),
                  ClinicSectionHeader(title: l10n.selectPatient),
                  const SizedBox(height: 10),
                  TextField(
                    controller: _patientQuery,
                    decoration: InputDecoration(
                      hintText: l10n.searchPatientHint,
                      prefixIcon: const Icon(Icons.search),
                    ),
                    onChanged: _searchPatients,
                  ),
                  if (_selectedPatient != null) ...[
                    const SizedBox(height: 8),
                    Text(
                      _selectedPatient!.fullName,
                      style: GoogleFonts.inter(fontWeight: FontWeight.w600),
                    ),
                  ],
                  ..._patientHits.map(
                    (p) => ListTile(
                      title: Text(p.fullName),
                      onTap: () => setState(() {
                        _selectedPatient = p;
                        _patientHits = const [];
                        _patientQuery.text = p.fullName;
                      }),
                    ),
                  ),
                ],
                const SizedBox(height: 22),
                ClinicSectionHeader(title: l10n.selectDate),
                const SizedBox(height: 12),
                SizedBox(
                  height: 78,
                  child: ListView.separated(
                    scrollDirection: Axis.horizontal,
                    itemCount: days.length,
                    separatorBuilder: (_, __) => const SizedBox(width: 10),
                    itemBuilder: (context, index) {
                      final d = days[index];
                      final selected = d.year == _day.year &&
                          d.month == _day.month &&
                          d.day == _day.day;
                      return InkWell(
                        onTap: () {
                          setState(() => _day = d);
                          _loadSlots();
                        },
                        borderRadius: BorderRadius.circular(16),
                        child: Container(
                          width: 64,
                          decoration: BoxDecoration(
                            color: selected ? AppColors.primary : Colors.white,
                            borderRadius: BorderRadius.circular(16),
                            border: selected
                                ? null
                                : Border.all(color: AppColors.line),
                          ),
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text(
                                DateFormat.E(locale).format(d),
                                style: GoogleFonts.inter(
                                  fontSize: 12,
                                  color: selected
                                      ? Colors.white70
                                      : AppColors.muted,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                '${d.day}',
                                style: GoogleFonts.inter(
                                  fontWeight: FontWeight.w700,
                                  fontSize: 18,
                                  color:
                                      selected ? Colors.white : AppColors.ink,
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
                const SizedBox(height: 22),
                ClinicSectionHeader(title: l10n.selectTime),
                const SizedBox(height: 12),
                if (_loadingSlots)
                  const Center(child: CircularProgressIndicator())
                else if (_slots.isEmpty)
                  ClinicEmptyState(message: l10n.noSlots)
                else
                  Wrap(
                    spacing: 10,
                    runSpacing: 10,
                    children: _slots.map((slot) {
                      final selected = _slot == slot;
                      return InkWell(
                        onTap: () => setState(() => _slot = slot),
                        borderRadius: BorderRadius.circular(14),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                          decoration: BoxDecoration(
                            color: selected ? AppColors.primary : Colors.white,
                            borderRadius: BorderRadius.circular(14),
                            border: selected
                                ? null
                                : Border.all(color: AppColors.line),
                          ),
                          child: Text(
                            DateFormat('HH:mm', locale).format(slot),
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w600,
                              color: selected ? Colors.white : AppColors.ink,
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    _error!,
                    style: GoogleFonts.inter(color: Colors.red),
                  ),
                ],
              ],
            ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 12),
          child: ClinicPrimaryButton(
            label: l10n.confirmBooking,
            loading: _submitting,
            onPressed: _slot == null ? null : _confirmAndSubmit,
          ),
        ),
      ),
    );
  }
}
