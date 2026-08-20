import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class BrandTheme {
  const BrandTheme({
    required this.clinicName,
    required this.primary,
    required this.secondary,
    this.supportPhone,
    this.website,
    this.logoUrl,
  });

  final String clinicName;
  final Color primary;
  final Color secondary;
  final String? supportPhone;
  final String? website;
  final String? logoUrl;

  /// Behance "Doctor Appointment" palette.
  static const fallback = BrandTheme(
    clinicName: 'Clinicos',
    primary: Color(0xFF0564F2),
    secondary: Color(0xFF198754),
  );
}

class AppColors {
  static const primary = Color(0xFF0564F2);
  static const ink = Color(0xFF040C17);
  static const success = Color(0xFF198754);
  static const softBlue = Color(0xFFF2F7FF);
  static const page = Color(0xFFF7F9FC);
  static const muted = Color(0xFF6B7280);
  static const line = Color(0xFFE8EEF7);
}

class AppTheme {
  static ThemeData light(BrandTheme brand) {
    final scheme = ColorScheme.fromSeed(
      seedColor: brand.primary,
      primary: brand.primary,
      secondary: brand.secondary,
      surface: Colors.white,
      brightness: Brightness.light,
    );
    final textTheme = GoogleFonts.interTextTheme().apply(
      bodyColor: AppColors.ink,
      displayColor: AppColors.ink,
    );
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      scaffoldBackgroundColor: AppColors.page,
      textTheme: textTheme,
      appBarTheme: AppBarTheme(
        backgroundColor: Colors.white,
        foregroundColor: AppColors.ink,
        elevation: 0,
        centerTitle: false,
        titleTextStyle: GoogleFonts.inter(
          fontSize: 18,
          fontWeight: FontWeight.w600,
          color: AppColors.ink,
        ),
      ),
      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: Colors.white,
        elevation: 0,
        height: 72,
        indicatorColor: brand.primary,
        indicatorShape: const CircleBorder(),
        labelBehavior: NavigationDestinationLabelBehavior.alwaysHide,
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          final selected = states.contains(WidgetState.selected);
          return GoogleFonts.inter(
            fontSize: 11,
            fontWeight: selected ? FontWeight.w600 : FontWeight.w500,
            color: selected ? brand.primary : AppColors.muted,
          );
        }),
        iconTheme: WidgetStateProperty.resolveWith((states) {
          final selected = states.contains(WidgetState.selected);
          return IconThemeData(
            color: selected ? Colors.white : AppColors.muted,
            size: 24,
          );
        }),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: Colors.white,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: const BorderSide(color: AppColors.line),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: const BorderSide(color: AppColors.line),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: brand.primary, width: 1.4),
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: brand.primary,
          foregroundColor: Colors.white,
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
          textStyle: GoogleFonts.inter(fontWeight: FontWeight.w600, fontSize: 16),
        ),
      ),
    );
  }

  static ThemeData dark(BrandTheme brand) {
    final scheme = ColorScheme.fromSeed(
      seedColor: brand.primary,
      brightness: Brightness.dark,
    );
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      textTheme: GoogleFonts.interTextTheme(ThemeData.dark().textTheme),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
        ),
      ),
    );
  }
}
