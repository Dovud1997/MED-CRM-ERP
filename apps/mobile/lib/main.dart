import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/localization/locale_controller.dart';
import 'package:clinicos_mobile/core/routing/app_router.dart';
import 'package:clinicos_mobile/core/theme/app_theme.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const ProviderScope(child: ClinicosApp()));
}

class ClinicosApp extends ConsumerWidget {
  const ClinicosApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(goRouterProvider);
    final locale = ref.watch(localeControllerProvider);
    const brand = BrandTheme.fallback;

    return MaterialApp.router(
      title: brand.clinicName,
      theme: AppTheme.light(brand),
      darkTheme: AppTheme.dark(brand),
      themeMode: ThemeMode.system,
      routerConfig: router,
      locale: locale,
      supportedLocales: AppLocalizations.supportedLocales,
      localeResolutionCallback: (device, supported) {
        if (device == null) return const Locale('ru');
        for (final locale in supported) {
          if (locale.languageCode == device.languageCode) return locale;
        }
        return const Locale('ru');
      },
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
    );
  }
}
