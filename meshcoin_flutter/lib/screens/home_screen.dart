import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../mesh_node.dart';
import '../tunneling/mesh_tunnel.dart';

class HomeScreen extends StatefulWidget {
  final MeshNode meshNode;
  final String address;
  final double balance;
  final MeshTunnelQueue tunnelQueue;
  final List<Map<String, dynamic>> recentTransactions;

  const HomeScreen({
    super.key,
    required this.meshNode,
    required this.address,
    required this.balance,
    required this.tunnelQueue,
    required this.recentTransactions,
  });

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> with SingleTickerProviderStateMixin {
  late AnimationController _glowController;
  late Animation<double> _glowAnimation;

  @override
  void initState() {
    super.initState();
    _glowController = AnimationController(
      duration: const Duration(seconds: 3),
      vsync: this,
    )..repeat(reverse: true);
    _glowAnimation = Tween<double>(begin: 0.3, end: 1.0).animate(
      CurvedAnimation(parent: _glowController, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _glowController.dispose();
    super.dispose();
  }

  String _getNetworkMode() {
    if (widget.meshNode.directPeers.isEmpty) return 'Isolado';
    if (widget.meshNode.isBridgeNode) return 'On-Grid';
    return 'Off-Grid Mesh';
  }

  Color _getNetworkColor() {
    if (widget.meshNode.directPeers.isEmpty) return MeshColors.neonRed;
    if (widget.meshNode.isBridgeNode) return MeshColors.neonGreen;
    return MeshColors.neonGold;
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      child: Column(
        children: [
          const SizedBox(height: 20),
          // Logo com Glow Animado
          _buildAnimatedLogo(),
          const SizedBox(height: 24),
          // Saldo Principal
          _buildBalanceCard(),
          const SizedBox(height: 16),
          // Status da Rede
          _buildNetworkStatus(),
          const SizedBox(height: 16),
          // Tunneling Status
          if (widget.tunnelQueue.pendingCount > 0)
            _buildTunnelAlert(),
          const SizedBox(height: 16),
          // Transações Recentes
          _buildRecentTransactions(),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildAnimatedLogo() {
    return AnimatedBuilder(
      animation: _glowAnimation,
      builder: (context, child) {
        return Container(
          width: 100,
          height: 100,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            boxShadow: [
              BoxShadow(
                color: MeshColors.neonCyan.withOpacity(0.3 * _glowAnimation.value),
                blurRadius: 30 * _glowAnimation.value,
                spreadRadius: 5 * _glowAnimation.value,
              ),
              BoxShadow(
                color: MeshColors.neonViolet.withOpacity(0.2 * _glowAnimation.value),
                blurRadius: 40 * _glowAnimation.value,
                spreadRadius: 10 * _glowAnimation.value,
              ),
            ],
          ),
          child: ClipOval(
            child: Image.asset(
              'assets/images/logo.png',
              width: 100,
              height: 100,
              fit: BoxFit.cover,
              errorBuilder: (context, error, stack) => Container(
                width: 100,
                height: 100,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: MeshColors.neonGradient,
                ),
                child: const Icon(Icons.currency_bitcoin, size: 50, color: MeshColors.background),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildBalanceCard() {
    return GlassCard(
      borderColor: MeshColors.neonCyan.withOpacity(0.2),
      padding: const EdgeInsets.symmetric(vertical: 28, horizontal: 20),
      child: Column(
        children: [
          Text(
            'Saldo Disponível',
            style: GoogleFonts.inter(
              color: MeshColors.textSecondary,
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                widget.balance.toStringAsFixed(4),
                style: GoogleFonts.inter(
                  color: MeshColors.textPrimary,
                  fontSize: 40,
                  fontWeight: FontWeight.w800,
                ),
              ),
              Padding(
                padding: const EdgeInsets.only(bottom: 6, left: 8),
                child: Text(
                  'MESH',
                  style: GoogleFonts.inter(
                    color: MeshColors.neonCyan,
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (widget.address != 'Não gerada')
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: MeshColors.surfaceLight,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '${widget.address.substring(0, 12)}...${widget.address.substring(widget.address.length - 8)}',
                style: GoogleFonts.jetBrainsMono(
                  color: MeshColors.textMuted,
                  fontSize: 12,
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildNetworkStatus() {
    int totalPeers = widget.meshNode.directPeers.length;
    int routedPeers = widget.meshNode.router.routingTable.length;

    return GlassCard(
      child: Row(
        children: [
          // Status Indicator
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: _getNetworkColor().withOpacity(0.15),
              border: Border.all(color: _getNetworkColor(), width: 2),
            ),
            child: Icon(
              widget.meshNode.directPeers.isEmpty ? Icons.signal_wifi_off : Icons.wifi_tethering,
              color: _getNetworkColor(),
              size: 22,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _getNetworkMode(),
                  style: GoogleFonts.inter(
                    color: _getNetworkColor(),
                    fontSize: 16,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  '$totalPeers diretos · $routedPeers rotas BATMAN',
                  style: GoogleFonts.inter(
                    color: MeshColors.textMuted,
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
          // Nó ponte toggle
          Column(
            children: [
              Icon(
                Icons.router,
                color: widget.meshNode.isBridgeNode ? MeshColors.neonGreen : MeshColors.textMuted,
                size: 20,
              ),
              Text(
                'Bridge',
                style: GoogleFonts.inter(
                  color: MeshColors.textMuted,
                  fontSize: 10,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTunnelAlert() {
    int pending = widget.tunnelQueue.pendingCount;
    return GlassCard(
      borderColor: MeshColors.neonGold.withOpacity(0.3),
      child: Row(
        children: [
          Icon(Icons.sync, color: MeshColors.neonGold, size: 24),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Mesh Tunneling',
                  style: GoogleFonts.inter(
                    color: MeshColors.neonGold,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                Text(
                  '$pending transação(ões) aguardando Bridge Node',
                  style: GoogleFonts.inter(
                    color: MeshColors.textMuted,
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
          Icon(Icons.arrow_forward_ios, color: MeshColors.textMuted, size: 14),
        ],
      ),
    );
  }

  Widget _buildRecentTransactions() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Transações Recentes',
            style: GoogleFonts.inter(
              color: MeshColors.textPrimary,
              fontSize: 18,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 12),
          if (widget.recentTransactions.isEmpty)
            GlassCard(
              margin: EdgeInsets.zero,
              child: Center(
                child: Column(
                  children: [
                    Icon(Icons.receipt_long, color: MeshColors.textMuted, size: 40),
                    const SizedBox(height: 8),
                    Text(
                      'Nenhuma transação ainda',
                      style: GoogleFonts.inter(color: MeshColors.textMuted),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Gere uma carteira e conecte à rede mesh',
                      style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 12),
                    ),
                  ],
                ),
              ),
            )
          else
            ...widget.recentTransactions.take(5).map((tx) => _buildTxItem(tx)),
        ],
      ),
    );
  }

  Widget _buildTxItem(Map<String, dynamic> tx) {
    bool isSent = tx['remetente'] == widget.address;
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: MeshColors.glass,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: MeshColors.surfaceLight.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: (isSent ? MeshColors.neonRed : MeshColors.neonGreen).withOpacity(0.15),
            ),
            child: Icon(
              isSent ? Icons.arrow_upward : Icons.arrow_downward,
              color: isSent ? MeshColors.neonRed : MeshColors.neonGreen,
              size: 20,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  isSent ? 'Enviado' : 'Recebido',
                  style: GoogleFonts.inter(
                    color: MeshColors.textPrimary,
                    fontWeight: FontWeight.w600,
                    fontSize: 14,
                  ),
                ),
                Text(
                  'Via Mesh P2P',
                  style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
                ),
              ],
            ),
          ),
          Text(
            '${isSent ? "-" : "+"}${tx['valor'] ?? '0'} MESH',
            style: GoogleFonts.jetBrainsMono(
              color: isSent ? MeshColors.neonRed : MeshColors.neonGreen,
              fontWeight: FontWeight.w700,
              fontSize: 14,
            ),
          ),
        ],
      ),
    );
  }
}
