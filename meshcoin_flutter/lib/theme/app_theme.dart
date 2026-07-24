import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

/// ═══════════════════════════════════════════════════════════════
/// MeshCoin Design System — Neon Cyberpunk Theme
/// ═══════════════════════════════════════════════════════════════

class MeshColors {
  // Core
  static const Color background = Color(0xFF0A0E1A);
  static const Color surface = Color(0xFF111827);
  static const Color surfaceLight = Color(0xFF1E293B);
  static const Color glass = Color(0x991E293B); // 60% opacity

  // Neon Palette
  static const Color neonCyan = Color(0xFF00F0FF);
  static const Color neonViolet = Color(0xFF8B5CF6);
  static const Color neonGreen = Color(0xFF10B981);
  static const Color neonGold = Color(0xFFF59E0B);
  static const Color neonRed = Color(0xFFEF4444);
  static const Color neonPink = Color(0xFFEC4899);

  // Text
  static const Color textPrimary = Color(0xFFF1F5F9);
  static const Color textSecondary = Color(0xFF94A3B8);
  static const Color textMuted = Color(0xFF64748B);

  // Gradients
  static const LinearGradient neonGradient = LinearGradient(
    colors: [neonCyan, neonViolet],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  static const LinearGradient darkGradient = LinearGradient(
    colors: [Color(0xFF0A0E1A), Color(0xFF111827)],
    begin: Alignment.topCenter,
    end: Alignment.bottomCenter,
  );

  static const LinearGradient goldGradient = LinearGradient(
    colors: [Color(0xFFF59E0B), Color(0xFFD97706)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
}

class MeshTheme {
  static ThemeData get darkTheme {
    return ThemeData(
      brightness: Brightness.dark,
      scaffoldBackgroundColor: MeshColors.background,
      primaryColor: MeshColors.neonCyan,
      colorScheme: const ColorScheme.dark(
        primary: MeshColors.neonCyan,
        secondary: MeshColors.neonViolet,
        surface: MeshColors.surface,
        error: MeshColors.neonRed,
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        titleTextStyle: GoogleFonts.inter(
          fontSize: 20,
          fontWeight: FontWeight.w700,
          color: MeshColors.textPrimary,
        ),
      ),
      textTheme: TextTheme(
        headlineLarge: GoogleFonts.inter(
          fontSize: 32,
          fontWeight: FontWeight.w800,
          color: MeshColors.textPrimary,
        ),
        headlineMedium: GoogleFonts.inter(
          fontSize: 24,
          fontWeight: FontWeight.w700,
          color: MeshColors.textPrimary,
        ),
        headlineSmall: GoogleFonts.inter(
          fontSize: 20,
          fontWeight: FontWeight.w600,
          color: MeshColors.textPrimary,
        ),
        bodyLarge: GoogleFonts.inter(
          fontSize: 16,
          fontWeight: FontWeight.w400,
          color: MeshColors.textPrimary,
        ),
        bodyMedium: GoogleFonts.inter(
          fontSize: 14,
          fontWeight: FontWeight.w400,
          color: MeshColors.textSecondary,
        ),
        bodySmall: GoogleFonts.inter(
          fontSize: 12,
          fontWeight: FontWeight.w400,
          color: MeshColors.textMuted,
        ),
        labelLarge: GoogleFonts.inter(
          fontSize: 14,
          fontWeight: FontWeight.w600,
          color: MeshColors.textPrimary,
        ),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: MeshColors.surface,
        selectedItemColor: MeshColors.neonCyan,
        unselectedItemColor: MeshColors.textMuted,
        type: BottomNavigationBarType.fixed,
        elevation: 0,
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: MeshColors.surfaceLight,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: MeshColors.neonCyan, width: 1.5),
        ),
        hintStyle: GoogleFonts.inter(color: MeshColors.textMuted),
        labelStyle: GoogleFonts.inter(color: MeshColors.textSecondary),
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: MeshColors.neonCyan,
          foregroundColor: MeshColors.background,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          textStyle: GoogleFonts.inter(fontSize: 16, fontWeight: FontWeight.w700),
          elevation: 0,
        ),
      ),
      cardTheme: CardThemeData(
        color: MeshColors.surface,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      ),
    );
  }
}

/// Componente Glassmorphism reutilizável
class GlassCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry? padding;
  final EdgeInsetsGeometry? margin;
  final Color? borderColor;

  const GlassCard({
    super.key,
    required this.child,
    this.padding,
    this.margin,
    this.borderColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: margin ?? const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      padding: padding ?? const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: MeshColors.glass,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: borderColor ?? MeshColors.surfaceLight.withOpacity(0.5),
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: (borderColor ?? MeshColors.neonCyan).withOpacity(0.05),
            blurRadius: 20,
            spreadRadius: 0,
          ),
        ],
      ),
      child: child,
    );
  }
}

/// Botão com gradiente neon
class NeonButton extends StatelessWidget {
  final String label;
  final IconData? icon;
  final VoidCallback onPressed;
  final LinearGradient? gradient;
  final bool isSmall;

  const NeonButton({
    super.key,
    required this.label,
    this.icon,
    required this.onPressed,
    this.gradient,
    this.isSmall = false,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onPressed,
      child: Container(
        padding: EdgeInsets.symmetric(
          horizontal: isSmall ? 16 : 24,
          vertical: isSmall ? 10 : 14,
        ),
        decoration: BoxDecoration(
          gradient: gradient ?? MeshColors.neonGradient,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: MeshColors.neonCyan.withOpacity(0.3),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (icon != null) ...[
              Icon(icon, color: MeshColors.background, size: isSmall ? 18 : 20),
              const SizedBox(width: 8),
            ],
            Text(
              label,
              style: GoogleFonts.inter(
                color: MeshColors.background,
                fontWeight: FontWeight.w700,
                fontSize: isSmall ? 13 : 16,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Indicador de status com ponto colorido
class StatusDot extends StatelessWidget {
  final Color color;
  final String label;

  const StatusDot({super.key, required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(
            color: color,
            shape: BoxShape.circle,
            boxShadow: [
              BoxShadow(color: color.withOpacity(0.5), blurRadius: 6),
            ],
          ),
        ),
        const SizedBox(width: 6),
        Text(label, style: GoogleFonts.inter(color: MeshColors.textSecondary, fontSize: 12)),
      ],
    );
  }
}
