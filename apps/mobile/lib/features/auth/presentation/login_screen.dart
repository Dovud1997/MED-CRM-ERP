import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:clinicos_mobile/core/config/app_config.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/widgets/clinic_widgets.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  late final TextEditingController _org;
  late final TextEditingController _login;
  late final TextEditingController _password;
  final _formKey = GlobalKey<FormState>();

  @override
  void initState() {
    super.initState();
    _org =
        TextEditingController(text: AppConfig.development.defaultOrganizationId);
    _login = TextEditingController();
    _password = TextEditingController();
  }

  @override
  void dispose() {
    _org.dispose();
    _login.dispose();
    _password.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);

    return Scaffold(
      backgroundColor: AppColors.page,
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.fromLTRB(24, 40, 24, 24),
          children: [
            Text(
              l10n.clinicBrand,
              style: GoogleFonts.inter(
                fontSize: 28,
                fontWeight: FontWeight.w700,
                color: AppColors.primary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              l10n.loginTitle,
              style: GoogleFonts.inter(
                fontSize: 22,
                fontWeight: FontWeight.w700,
                color: AppColors.ink,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              l10n.loginSubtitle,
              style: GoogleFonts.inter(color: AppColors.muted),
            ),
            const SizedBox(height: 28),
            Form(
              key: _formKey,
              child: Column(
                children: [
                  TextFormField(
                    controller: _org,
                    decoration: InputDecoration(labelText: l10n.organizationId),
                    validator: (v) => (v == null || v.trim().isEmpty)
                        ? l10n.errorValidation
                        : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _login,
                    decoration: InputDecoration(labelText: l10n.login),
                    validator: (v) => (v == null || v.trim().isEmpty)
                        ? l10n.errorValidation
                        : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _password,
                    obscureText: true,
                    decoration: InputDecoration(labelText: l10n.password),
                    validator: (v) =>
                        (v == null || v.length < 8) ? l10n.errorValidation : null,
                  ),
                  if (auth.error != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      auth.error!,
                      style: GoogleFonts.inter(color: Colors.red),
                    ),
                  ],
                  const SizedBox(height: 24),
                  ClinicPrimaryButton(
                    label: l10n.signIn,
                    loading: auth.loading,
                    onPressed: () {
                      if (!_formKey.currentState!.validate()) return;
                      ref.read(authControllerProvider.notifier).login(
                            organizationId: _org.text.trim(),
                            login: _login.text.trim(),
                            password: _password.text,
                          );
                    },
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
