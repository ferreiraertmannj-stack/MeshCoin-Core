import 'dart:async';
import 'dart:math';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../blockchain/ledger.dart';
import '../mesh_node.dart';

class MiningScreen extends StatefulWidget {
  final Ledger ledger;
  final String minerAddress;
  final MeshNode meshNode;

  const MiningScreen({
    super.key,
    required this.ledger,
    required this.minerAddress,
    required this.meshNode,
  });

  @override
  State<MiningScreen> createState() => _MiningScreenState();
}

class _MiningScreenState extends State<MiningScreen> with TickerProviderStateMixin {
  bool _isMining = false;
  double _hashRate = 0;
  int _blocksFound = 0;
  int _totalHashes = 0;
  double _scratchpadMB = 0;
  Duration _uptime = Duration.zero;
  String _currentHash = '0' * 64;
  Timer? _miningTimer;
  Timer? _uptimeTimer;
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      duration: const Duration(milliseconds: 1500),
      vsync: this,
    );
    _pulseAnimation = Tween<double>(begin: 0.8, end: 1.2).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _miningTimer?.cancel();
    _uptimeTimer?.cancel();
    _pulseController.dispose();
    super.dispose();
  }

  void _toggleMining() {
    if (widget.minerAddress == 'Não gerada') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gere uma carteira primeiro!')),
      );
      return;
    }

    setState(() {
      _isMining = !_isMining;
    });

    if (_isMining) {
      _pulseController.repeat(reverse: true);
      _scratchpadMB = 64.0; // NeonHash RAM footprint sim

      _uptimeTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
        if (!mounted) return;
        setState(() => _uptime += const Duration(seconds: 1));
      });

      _mineLoop();
    } else {
      _pulseController.stop();
      _miningTimer?.cancel();
      _uptimeTimer?.cancel();
      setState(() => _hashRate = 0);
    }
  }

  Future<void> _mineLoop() async {
    while (_isMining && mounted) {
      // Pequena pausa para a UI respirar e pegar blocos/transações que chegam
      await Future.delayed(const Duration(milliseconds: 500));
      if (!_isMining || !mounted) break;

      int t0 = DateTime.now().millisecondsSinceEpoch;
      int hashesDone = 0;

      // Chama a mineração real do ledger (síncrona na UI, por isso blocos vazios ou diff alta podem travar)
      // Como diff=3 é rápido (média de 4096 hashes), roda quase instantâneo.
      var newBlock = widget.ledger.minePendingTransactions(
        widget.minerAddress,
        onProgress: (h) {
          hashesDone = h;
          if (mounted) {
            setState(() {
              _currentHash = List.generate(64, (_) => Random().nextInt(16).toRadixString(16)).join();
            });
          }
        },
      );

      int dt = DateTime.now().millisecondsSinceEpoch - t0;

      if (newBlock != null && mounted) {
        setState(() {
          _blocksFound++;
          _totalHashes += hashesDone;
          _hashRate = (hashesDone / max(1, dt)) * 1000;
          _currentHash = newBlock.hash;
        });

        ScaffoldMessenger.of(context).clearSnackBars();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Bloco #${newBlock.index} minerado! Recompensa recebida.'),
            backgroundColor: MeshColors.neonGreen,
            duration: const Duration(seconds: 2),
          ),
        );

        // Broadcast o novo bloco na rede Mesh!
        widget.meshNode.sendRoutedData('BROADCAST', {
          'tipo': 'NEW_BLOCK',
          'block': newBlock.toJson(),
        });
      } else {
        // Se ainda não deu 2 minutos, roda um hashing "dummy" simulando o PoW contínuo
        setState(() {
          final rng = Random();
          _hashRate = 120 + rng.nextDouble() * 80; // H/s ARM simulado
          _totalHashes += (_hashRate * 0.5).round();
          _currentHash = List.generate(64, (_) => rng.nextInt(16).toRadixString(16)).join();
        });
      }
    }
  }

  String _formatDuration(Duration d) {
    String twoDigits(int n) => n.toString().padLeft(2, '0');
    return '${twoDigits(d.inHours)}:${twoDigits(d.inMinutes.remainder(60))}:${twoDigits(d.inSeconds.remainder(60))}';
  }

  String _formatNumber(int n) {
    if (n >= 1000000) return '${(n / 1000000).toStringAsFixed(1)}M';
    if (n >= 1000) return '${(n / 1000).toStringAsFixed(1)}K';
    return n.toString();
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      child: Column(
        children: [
          const SizedBox(height: 24),
          // Mining Power Button
          _buildMiningButton(),
          const SizedBox(height: 24),
          // Hash Rate
          _buildHashRateCard(),
          const SizedBox(height: 12),
          // Stats Grid
          _buildStatsGrid(),
          const SizedBox(height: 12),
          // Current Hash
          _buildCurrentHash(),
          const SizedBox(height: 12),
          // NeonHash Info
          _buildNeonHashInfo(),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildMiningButton() {
    return Center(
      child: GestureDetector(
        onTap: _toggleMining,
        child: AnimatedBuilder(
          animation: _pulseAnimation,
          builder: (context, child) {
            double scale = _isMining ? _pulseAnimation.value : 1.0;
            return Transform.scale(
              scale: scale,
              child: Container(
                width: 140,
                height: 140,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: _isMining
                      ? MeshColors.goldGradient
                      : const LinearGradient(
                          colors: [MeshColors.surfaceLight, MeshColors.surface],
                        ),
                  border: Border.all(
                    color: _isMining ? MeshColors.neonGold : MeshColors.textMuted,
                    width: 3,
                  ),
                  boxShadow: _isMining
                      ? [
                          BoxShadow(
                            color: MeshColors.neonGold.withOpacity(0.4),
                            blurRadius: 30,
                            spreadRadius: 5,
                          ),
                        ]
                      : [],
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      _isMining ? Icons.bolt : Icons.power_settings_new,
                      color: _isMining ? MeshColors.background : MeshColors.textMuted,
                      size: 40,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      _isMining ? 'MINERANDO' : 'INICIAR',
                      style: GoogleFonts.inter(
                        color: _isMining ? MeshColors.background : MeshColors.textMuted,
                        fontSize: 12,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 2,
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildHashRateCard() {
    return GlassCard(
      borderColor: _isMining ? MeshColors.neonGold.withOpacity(0.3) : null,
      child: Column(
        children: [
          Text(
            'Hash Rate',
            style: GoogleFonts.inter(color: MeshColors.textSecondary, fontSize: 13),
          ),
          const SizedBox(height: 4),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                _isMining ? _hashRate.toStringAsFixed(1) : '0.0',
                style: GoogleFonts.jetBrainsMono(
                  color: _isMining ? MeshColors.neonGold : MeshColors.textMuted,
                  fontSize: 42,
                  fontWeight: FontWeight.w800,
                ),
              ),
              Padding(
                padding: const EdgeInsets.only(bottom: 8, left: 4),
                child: Text(
                  'H/s',
                  style: GoogleFonts.inter(
                    color: MeshColors.textMuted,
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          // Progress bar visual
          Container(
            height: 4,
            width: double.infinity,
            decoration: BoxDecoration(
              color: MeshColors.surfaceLight,
              borderRadius: BorderRadius.circular(2),
            ),
            child: FractionallySizedBox(
              alignment: Alignment.centerLeft,
              widthFactor: _isMining ? (_hashRate / 250).clamp(0, 1) : 0,
              child: Container(
                decoration: BoxDecoration(
                  gradient: MeshColors.goldGradient,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsGrid() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: [
          Expanded(child: _buildStatTile('⛏️', 'Blocos', _blocksFound.toString(), MeshColors.neonGold)),
          const SizedBox(width: 8),
          Expanded(child: _buildStatTile('🧮', 'Hashes', _formatNumber(_totalHashes), MeshColors.neonCyan)),
          const SizedBox(width: 8),
          Expanded(child: _buildStatTile('🧠', 'RAM', '${_isMining ? _scratchpadMB.toStringAsFixed(0) : "0"} MB', MeshColors.neonViolet)),
          const SizedBox(width: 8),
          Expanded(child: _buildStatTile('⏱️', 'Uptime', _formatDuration(_uptime), MeshColors.neonGreen)),
        ],
      ),
    );
  }

  Widget _buildStatTile(String emoji, String label, String value, Color color) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: MeshColors.glass,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.15)),
      ),
      child: Column(
        children: [
          Text(emoji, style: const TextStyle(fontSize: 18)),
          const SizedBox(height: 4),
          Text(
            value,
            style: GoogleFonts.jetBrainsMono(
              color: color,
              fontSize: 14,
              fontWeight: FontWeight.w700,
            ),
          ),
          Text(
            label,
            style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 10),
          ),
        ],
      ),
    );
  }

  Widget _buildCurrentHash() {
    return GlassCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.tag, color: MeshColors.neonCyan, size: 16),
              const SizedBox(width: 6),
              Text(
                'Hash Atual',
                style: GoogleFonts.inter(
                  color: MeshColors.textSecondary,
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: MeshColors.background,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(
              _currentHash,
              style: GoogleFonts.jetBrainsMono(
                color: _isMining ? MeshColors.neonGold : MeshColors.textMuted,
                fontSize: 11,
              ),
              overflow: TextOverflow.ellipsis,
              maxLines: 2,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNeonHashInfo() {
    return GlassCard(
      borderColor: MeshColors.neonViolet.withOpacity(0.2),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(colors: [MeshColors.neonViolet, MeshColors.neonPink]),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  'NeonHash',
                  style: GoogleFonts.jetBrainsMono(
                    color: Colors.white,
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                'Proof-of-Work ARM',
                style: GoogleFonts.inter(
                  color: MeshColors.textPrimary,
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            '• Memory-Hard: aloca Scratchpad de RAM pesado\n'
            '• Otimizado para ARM NEON (smartphones)\n'
            '• Resistente a ASICs e GPUs dedicadas\n'
            '• Filosofia: Uma CPU, Um Voto',
            style: GoogleFonts.inter(
              color: MeshColors.textMuted,
              fontSize: 12,
              height: 1.6,
            ),
          ),
        ],
      ),
    );
  }
}
