import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import 'dart:io';
import 'dart:convert';
import 'dart:async';
import 'package:shared_preferences/shared_preferences.dart';

class NebulaScreen extends StatefulWidget {
  const NebulaScreen({super.key});

  @override
  State<NebulaScreen> createState() => _NebulaScreenState();
}

class _NebulaScreenState extends State<NebulaScreen> {
  String _storageSize = '5.0 GB';
  double _allocatedGB = 5.0;
  String _hardwareType = 'Mobile';
  String _bonus = '0.0';
  Timer? _timer;
  bool _isDesktop = false;

  @override
  void initState() {
    super.initState();
    _checkDesktopSidecar();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _checkDesktopSidecar() async {
    if (Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
      _isDesktop = true;
      _pollSidecar();
      _timer = Timer.periodic(const Duration(seconds: 5), (_) => _pollSidecar());
    } else {
      _loadMobileConfig();
    }
  }

  Future<void> _loadMobileConfig() async {
    final prefs = await SharedPreferences.getInstance();
    double savedGB = prefs.getDouble('nebula_allocated_gb') ?? 5.0;
    if (mounted) {
      setState(() {
        _allocatedGB = savedGB;
        _storageSize = '${_allocatedGB.toStringAsFixed(1)} GB';
        _bonus = '+0.0 NBL/bloco'; // Celular não ganha bônus de HDD/SSD por padrão
      });
    }
  }

  Future<void> _saveMobileConfig(double gb) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setDouble('nebula_allocated_gb', gb);
  }

  Future<void> _pollSidecar() async {
    try {
      final request = await HttpClient().getUrl(Uri.parse('http://localhost:8080/api/status'));
      final response = await request.close();
      if (response.statusCode == 200) {
        final content = await response.transform(utf8.decoder).join();
        final data = json.decode(content);
        if (mounted) {
          setState(() {
            _storageSize = '${data['cloud_allocated_gb']} GB';
            _hardwareType = data['storage_type'];
            
            // Cálculos da UI (SSD = +0.5, HDD = +0.1)
            double multiplier = _hardwareType == 'SSD' ? 0.5 : 0.1;
            _bonus = '+${(data['cloud_allocated_gb'] * multiplier).toStringAsFixed(1)} NBL/bloco';
          });
        }
      }
    } catch (e) {
      // Sidecar não está rodando ainda ou erro de conexão
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      child: Column(
        children: [
          const SizedBox(height: 24),
          _buildHeroCard(),
          const SizedBox(height: 16),
          _buildStorageStats(),
          const SizedBox(height: 16),
          _buildShardList(),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildHeroCard() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: GlassCard(
        borderColor: MeshColors.neonCyan.withOpacity(0.3),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: const BoxDecoration(
                color: MeshColors.surfaceLight,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.cloud_sync,
                size: 48,
                color: MeshColors.neonCyan,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              'Nebula Cloud',
              style: GoogleFonts.inter(
                color: MeshColors.textPrimary,
                fontSize: 24,
                fontWeight: FontWeight.w800,
                letterSpacing: 1,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Armazenamento Descentralizado do Ecossistema Nebula',
              textAlign: TextAlign.center,
              style: GoogleFonts.inter(
                color: MeshColors.textMuted,
                fontSize: 13,
                height: 1.5,
              ),
            ),
            if (!_isDesktop) ...[
              const SizedBox(height: 24),
              Text(
                'Alocar Espaço (GB)',
                style: GoogleFonts.inter(
                  color: MeshColors.textPrimary,
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                ),
              ),
              Slider(
                value: _allocatedGB,
                min: 1.0,
                max: 100.0,
                divisions: 99,
                activeColor: MeshColors.neonCyan,
                inactiveColor: MeshColors.surfaceLight,
                onChanged: (value) {
                  setState(() {
                    _allocatedGB = value;
                    _storageSize = '${_allocatedGB.toStringAsFixed(1)} GB';
                  });
                },
                onChangeEnd: (value) {
                  _saveMobileConfig(value);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Configuração de espaço salva!')),
                  );
                },
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildStorageStats() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: [
          Expanded(
            child: _buildStatTile(
              '🗄️',
              'Espaço Contribuído',
              _storageSize,
              MeshColors.neonGreen,
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _buildStatTile(
              _hardwareType == 'SSD' ? '⚡' : '💾',
              'Hardware & Bônus',
              '$_hardwareType\n$_bonus',
              MeshColors.neonViolet,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatTile(String emoji, String label, String value, Color color) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: MeshColors.glass,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Column(
        children: [
          Text(emoji, style: const TextStyle(fontSize: 24)),
          const SizedBox(height: 8),
          Text(
            value,
            style: GoogleFonts.jetBrainsMono(
              color: color,
              fontSize: 18,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: GoogleFonts.inter(
              color: MeshColors.textMuted,
              fontSize: 11,
              fontWeight: FontWeight.w500,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Widget _buildShardList() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: GlassCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.data_usage, color: MeshColors.neonCyan, size: 20),
                const SizedBox(width: 8),
                Text(
                  'Ledger Distribuído (Reed-Solomon)',
                  style: GoogleFonts.inter(
                    color: MeshColors.textPrimary,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            _buildShardItem('Ledger Block #1', '256 KB', MeshColors.neonGreen),
            _buildShardItem('Ledger Block #2', '256 KB', MeshColors.neonGreen),
            _buildShardItem('Ledger Block #3 (Paridade)', '256 KB', MeshColors.neonViolet),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: MeshColors.surfaceLight.withOpacity(0.5),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                children: [
                  const Icon(Icons.info_outline, color: MeshColors.textMuted, size: 16),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'O Ledger completo é gerenciado pelo Nó do PC. Seu smartphone armazena apenas fragmentos criptografados para garantir a disponibilidade descentralizada.',
                      style: GoogleFonts.inter(
                        color: MeshColors.textMuted,
                        fontSize: 11,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildShardItem(String name, String size, Color color) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              Icon(Icons.insert_drive_file, color: color, size: 16),
              const SizedBox(width: 8),
              Text(
                name,
                style: GoogleFonts.inter(
                  color: MeshColors.textSecondary,
                  fontSize: 13,
                ),
              ),
            ],
          ),
          Text(
            size,
            style: GoogleFonts.jetBrainsMono(
              color: MeshColors.textMuted,
              fontSize: 12,
            ),
          ),
        ],
      ),
    );
  }
}
