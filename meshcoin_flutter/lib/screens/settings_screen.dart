import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          'Configurações',
          style: GoogleFonts.inter(
            color: MeshColors.textPrimary,
            fontSize: 24,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 20),
        
        // Language Selection
        ListTile(
          leading: const Icon(Icons.language, color: MeshColors.neonCyan),
          title: Text('Idioma (Language)', style: TextStyle(color: MeshColors.textPrimary)),
          trailing: DropdownButton<String>(
            dropdownColor: MeshColors.surfaceLight,
            value: 'pt-BR',
            style: const TextStyle(color: MeshColors.neonCyan),
            underline: Container(),
            items: const [
              DropdownMenuItem(value: 'pt-BR', child: Text('Português (BR)')),
              DropdownMenuItem(value: 'en', child: Text('English')),
              DropdownMenuItem(value: 'es', child: Text('Español')),
            ],
            onChanged: (val) {
              // TODO: Implement Locale provider logic
            },
          ),
        ),
        
        const Divider(color: MeshColors.surfaceLight),
        
        // CVM Legal Notice
        ListTile(
          leading: const Icon(Icons.gavel, color: MeshColors.neonCyan),
          title: Text('Aviso Legal (Compliance)', style: TextStyle(color: MeshColors.textPrimary)),
          subtitle: Text('Termos de Uso e Utilidade do Token', style: TextStyle(color: MeshColors.textSecondary)),
          onTap: () {
            showDialog(
              context: context,
              builder: (ctx) => AlertDialog(
                backgroundColor: MeshColors.surface,
                title: Text('Aviso CVM', style: TextStyle(color: MeshColors.neonRed)),
                content: Text(
                  'A Nebula Network e o Token Nebula (NBL) configuram-se estritamente como ferramentas de UTILIDADE (Gás e Storage) dentro deste ecossistema.\n\n'
                  'Não há qualquer promessa de lucros, dividendos ou valorização, não configurando assim um valor mobiliário sob as leis da CVM (Comissão de Valores Mobiliários) do Brasil ou entidades internacionais equivalentes.',
                  style: TextStyle(color: MeshColors.textPrimary),
                ),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.pop(ctx),
                    child: Text('Ciente', style: TextStyle(color: MeshColors.neonCyan)),
                  ),
                ],
              ),
            );
          },
        ),
        
        const Divider(color: MeshColors.surfaceLight),
        
        // About
        ListTile(
          leading: const Icon(Icons.info_outline, color: MeshColors.neonCyan),
          title: Text('Sobre a Nebula Network', style: TextStyle(color: MeshColors.textPrimary)),
          subtitle: Text('Versão 3.0 (Genesis Edition)', style: TextStyle(color: MeshColors.textSecondary)),
        ),
      ],
    );
  }
}
