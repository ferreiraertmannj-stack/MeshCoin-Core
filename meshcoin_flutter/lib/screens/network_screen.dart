import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../mesh_node.dart';
import '../tunneling/mesh_tunnel.dart';

class NetworkScreen extends StatelessWidget {
  final MeshNode meshNode;
  final MeshTunnelQueue tunnelQueue;

  const NetworkScreen({
    super.key,
    required this.meshNode,
    required this.tunnelQueue,
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const SizedBox(height: 16),
          // Node Identity
          _buildNodeIdentity(),
          const SizedBox(height: 12),
          // Modo Bridge Toggle
          _buildBridgeToggle(context),
          const SizedBox(height: 12),
          // Direct Peers
          _buildPeersList(),
          const SizedBox(height: 12),
          // Routing Table (BATMAN)
          _buildRoutingTable(),
          const SizedBox(height: 12),
          // Mesh Tunneling Queue
          _buildTunnelQueue(),
          const SizedBox(height: 12),
          // Supported Chains
          _buildSupportedChains(),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildNodeIdentity() {
    return GlassCard(
      borderColor: MeshColors.neonCyan.withOpacity(0.2),
      child: Row(
        children: [
          Container(
            width: 52,
            height: 52,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: MeshColors.neonGradient,
              boxShadow: [
                BoxShadow(
                  color: MeshColors.neonCyan.withOpacity(0.3),
                  blurRadius: 12,
                ),
              ],
            ),
            child: const Icon(Icons.hub, color: MeshColors.background, size: 26),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  meshNode.nodeName,
                  style: GoogleFonts.jetBrainsMono(
                    color: MeshColors.textPrimary,
                    fontSize: 16,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 2),
                Row(
                  children: [
                    StatusDot(
                      color: meshNode.isRunning ? MeshColors.neonGreen : MeshColors.neonRed,
                      label: meshNode.isRunning ? 'Ativo' : 'Inativo',
                    ),
                    const SizedBox(width: 12),
                    Text(
                      'Protocolo B.A.T.M.A.N.',
                      style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
                    ),
                  ],
                ),
              ],
            ),
          ),
          Column(
            children: [
              Text(
                '${meshNode.directPeers.length}',
                style: GoogleFonts.jetBrainsMono(
                  color: MeshColors.neonCyan,
                  fontSize: 24,
                  fontWeight: FontWeight.w800,
                ),
              ),
              Text(
                'peers',
                style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBridgeToggle(BuildContext context) {
    return GlassCard(
      borderColor: meshNode.isBridgeNode 
          ? MeshColors.neonGreen.withOpacity(0.3) 
          : MeshColors.surfaceLight.withOpacity(0.3),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: meshNode.isBridgeNode 
                  ? MeshColors.neonGreen.withOpacity(0.15)
                  : MeshColors.surfaceLight,
            ),
            child: Icon(
              Icons.router,
              color: meshNode.isBridgeNode ? MeshColors.neonGreen : MeshColors.textMuted,
              size: 22,
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Modo Nó-Ponte (Bridge)',
                  style: GoogleFonts.inter(
                    color: MeshColors.textPrimary,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                Text(
                  meshNode.isBridgeNode
                      ? 'Sincronizando dados mesh com a internet'
                      : 'Ative para conectar a rede mesh à internet',
                  style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
                ),
              ],
            ),
          ),
          Switch(
            value: meshNode.isBridgeNode,
            onChanged: (val) {
              meshNode.isBridgeNode = val;
              (context as Element).markNeedsBuild();
            },
            activeColor: MeshColors.neonGreen,
            activeTrackColor: MeshColors.neonGreen.withOpacity(0.3),
          ),
        ],
      ),
    );
  }

  Widget _buildPeersList() {
    List<String> peers = meshNode.directPeers.toList();

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                'Peers Diretos',
                style: GoogleFonts.inter(
                  color: MeshColors.textPrimary,
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: MeshColors.neonCyan.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  '${peers.length}',
                  style: GoogleFonts.jetBrainsMono(
                    color: MeshColors.neonCyan,
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (peers.isEmpty)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: MeshColors.glass,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  Icon(Icons.search_off, color: MeshColors.textMuted, size: 32),
                  const SizedBox(height: 8),
                  Text(
                    'Nenhum peer encontrado',
                    style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 13),
                  ),
                  Text(
                    'Conecte outro dispositivo na mesma rede',
                    style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
                  ),
                ],
              ),
            )
          else
            ...peers.map((peer) {
              String ip = peer.split(':')[0];
              String port = peer.split(':')[1];
              return Container(
                margin: const EdgeInsets.only(bottom: 6),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: MeshColors.glass,
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: MeshColors.neonGreen.withOpacity(0.15)),
                ),
                child: Row(
                  children: [
                    Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: MeshColors.neonGreen,
                        boxShadow: [BoxShadow(color: MeshColors.neonGreen.withOpacity(0.5), blurRadius: 4)],
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        ip,
                        style: GoogleFonts.jetBrainsMono(
                          color: MeshColors.textPrimary,
                          fontSize: 13,
                        ),
                      ),
                    ),
                    Text(
                      ':$port',
                      style: GoogleFonts.jetBrainsMono(
                        color: MeshColors.textMuted,
                        fontSize: 12,
                      ),
                    ),
                    const SizedBox(width: 8),
                    const Icon(Icons.wifi, color: MeshColors.neonGreen, size: 16),
                  ],
                ),
              );
            }),
        ],
      ),
    );
  }

  Widget _buildRoutingTable() {
    var routes = meshNode.router.routingTable;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                'Tabela de Roteamento BATMAN',
                style: GoogleFonts.inter(
                  color: MeshColors.textPrimary,
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(colors: [MeshColors.neonViolet, MeshColors.neonPink]),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  'BATMAN',
                  style: GoogleFonts.jetBrainsMono(color: Colors.white, fontSize: 9, fontWeight: FontWeight.w700),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (routes.isEmpty)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: MeshColors.glass,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                'Aguardando OGMs de outros nós...',
                style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 12),
                textAlign: TextAlign.center,
              ),
            )
          else
            Container(
              decoration: BoxDecoration(
                color: MeshColors.glass,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  // Header
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    decoration: BoxDecoration(
                      color: MeshColors.surfaceLight,
                      borderRadius: const BorderRadius.only(
                        topLeft: Radius.circular(12),
                        topRight: Radius.circular(12),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(flex: 3, child: Text('Destino', style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11, fontWeight: FontWeight.w600))),
                        Expanded(flex: 3, child: Text('Next Hop', style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11, fontWeight: FontWeight.w600))),
                        Expanded(flex: 1, child: Text('Hops', style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11, fontWeight: FontWeight.w600))),
                      ],
                    ),
                  ),
                  ...routes.entries.map((entry) => Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    child: Row(
                      children: [
                        Expanded(flex: 3, child: Text(entry.key, style: GoogleFonts.jetBrainsMono(color: MeshColors.neonCyan, fontSize: 11), overflow: TextOverflow.ellipsis)),
                        Expanded(flex: 3, child: Text(entry.value.nextHopIp, style: GoogleFonts.jetBrainsMono(color: MeshColors.textSecondary, fontSize: 11))),
                        Expanded(flex: 1, child: Text('${entry.value.hops}', style: GoogleFonts.jetBrainsMono(color: MeshColors.neonGold, fontSize: 11, fontWeight: FontWeight.w700))),
                      ],
                    ),
                  )),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildTunnelQueue() {
    List<TunnelTransaction> txs = tunnelQueue.all;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Mesh Tunneling',
            style: GoogleFonts.inter(
              color: MeshColors.textPrimary,
              fontSize: 16,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'Transporte transações BTC, ETH, SOL e Pix via rede mesh off-grid',
            style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
          ),
          const SizedBox(height: 8),
          if (txs.isEmpty)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: MeshColors.glass,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.sync, color: MeshColors.textMuted, size: 18),
                  const SizedBox(width: 8),
                  Text(
                    'Nenhuma transação na fila',
                    style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 12),
                  ),
                ],
              ),
            )
          else
            ...txs.map((tx) => Container(
              margin: const EdgeInsets.only(bottom: 6),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: MeshColors.glass,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: MeshColors.neonGold.withOpacity(0.2)),
              ),
              child: Row(
                children: [
                  Text(tx.chainEmoji, style: const TextStyle(fontSize: 22)),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${tx.amount} ${tx.chainSymbol}',
                          style: GoogleFonts.jetBrainsMono(
                            color: MeshColors.textPrimary,
                            fontSize: 13,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        Text(
                          tx.statusLabel,
                          style: GoogleFonts.inter(
                            color: tx.status == TunnelStatus.queued ? MeshColors.neonGold : MeshColors.neonGreen,
                            fontSize: 11,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Icon(
                    tx.status == TunnelStatus.queued ? Icons.hourglass_empty : Icons.check_circle,
                    color: tx.status == TunnelStatus.queued ? MeshColors.neonGold : MeshColors.neonGreen,
                    size: 20,
                  ),
                ],
              ),
            )),
        ],
      ),
    );
  }

  Widget _buildSupportedChains() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Chains Suportadas',
            style: GoogleFonts.inter(
              color: MeshColors.textPrimary,
              fontSize: 16,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              _buildChainChip('₿', 'Bitcoin', const Color(0xFFF7931A)),
              const SizedBox(width: 8),
              _buildChainChip('Ξ', 'Ethereum', const Color(0xFF627EEA)),
              const SizedBox(width: 8),
              _buildChainChip('◎', 'Solana', const Color(0xFF9945FF)),
              const SizedBox(width: 8),
              _buildChainChip('🇧🇷', 'Pix', const Color(0xFF32BCAD)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildChainChip(String emoji, String name, Color color) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: color.withOpacity(0.1),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: color.withOpacity(0.3)),
        ),
        child: Column(
          children: [
            Text(emoji, style: const TextStyle(fontSize: 20)),
            const SizedBox(height: 4),
            Text(
              name,
              style: GoogleFonts.inter(
                color: color,
                fontSize: 10,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
