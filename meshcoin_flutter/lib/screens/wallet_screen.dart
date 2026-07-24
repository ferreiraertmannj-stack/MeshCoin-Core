import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:qr_flutter/qr_flutter.dart';
import '../theme/app_theme.dart';
import '../crypto/wallet_crypto.dart';
import '../mesh_node.dart';

class WalletScreen extends StatefulWidget {
  final MeshNode meshNode;
  final String address;
  final double balance;
  final Function(Map<String, String>) onWalletGenerated;
  final Function(String dest, String valor) onSendPayment;
  final List<Map<String, dynamic>> transactions;

  const WalletScreen({
    super.key,
    required this.meshNode,
    required this.address,
    required this.balance,
    required this.onWalletGenerated,
    required this.onSendPayment,
    required this.transactions,
  });

  @override
  State<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends State<WalletScreen> {
  final TextEditingController _destController = TextEditingController();
  final TextEditingController _valorController = TextEditingController();
  bool _isGenerating = false;
  bool _showReceive = false;

  Future<void> _generateWallet() async {
    setState(() => _isGenerating = true);
    await Future.delayed(const Duration(milliseconds: 800)); // UX delay
    
    try {
      Map<String, String> wallet = WalletCrypto.generateKeypair();
      await WalletCrypto.saveWallet(wallet);
      widget.onWalletGenerated(wallet);
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(Icons.check_circle, color: MeshColors.neonGreen),
                const SizedBox(width: 8),
                Text('Carteira gerada com criptografia real!', 
                  style: GoogleFonts.inter(fontWeight: FontWeight.w600)),
              ],
            ),
            backgroundColor: MeshColors.surface,
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro: $e'), backgroundColor: MeshColors.neonRed),
        );
      }
    }
    
    setState(() => _isGenerating = false);
  }

  void _sendPayment() {
    if (widget.address == 'Não gerada') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gere uma carteira primeiro!')),
      );
      return;
    }
    if (_destController.text.isEmpty || _valorController.text.isEmpty) return;

    widget.onSendPayment(_destController.text, _valorController.text);
    _destController.clear();
    _valorController.clear();

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.send, color: MeshColors.neonCyan),
            const SizedBox(width: 8),
            Text('Transação enviada via Mesh P2P!',
              style: GoogleFonts.inter(fontWeight: FontWeight.w600)),
          ],
        ),
        backgroundColor: MeshColors.surface,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SizedBox(height: 16),
          // Endereço e Ações
          _buildAddressCard(),
          const SizedBox(height: 16),
          // Ações rápidas
          _buildQuickActions(),
          const SizedBox(height: 16),
          // QR Code para Receber
          if (_showReceive && widget.address != 'Não gerada')
            _buildReceiveCard(),
          // Formulário de Envio
          _buildSendCard(),
          const SizedBox(height: 16),
          // Histórico
          _buildTransactionHistory(),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildAddressCard() {
    bool hasWallet = widget.address != 'Não gerada';

    return GlassCard(
      borderColor: hasWallet
          ? MeshColors.neonGreen.withOpacity(0.2)
          : MeshColors.neonViolet.withOpacity(0.2),
      child: Column(
        children: [
          Row(
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: hasWallet
                      ? const LinearGradient(colors: [MeshColors.neonGreen, MeshColors.neonCyan])
                      : const LinearGradient(colors: [MeshColors.neonViolet, MeshColors.neonPink]),
                ),
                child: Icon(
                  hasWallet ? Icons.account_balance_wallet : Icons.add_circle_outline,
                  color: MeshColors.background,
                  size: 24,
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      hasWallet ? 'Carteira Ativa' : 'Sem Carteira',
                      style: GoogleFonts.inter(
                        color: MeshColors.textPrimary,
                        fontSize: 16,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      hasWallet ? 'ECDSA secp256k1 · Base58Check' : 'Gere uma carteira criptográfica real',
                      style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 11),
                    ),
                  ],
                ),
              ),
              if (hasWallet)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: MeshColors.neonGreen.withOpacity(0.15),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    '${widget.balance.toStringAsFixed(2)} MESH',
                    style: GoogleFonts.jetBrainsMono(
                      color: MeshColors.neonGreen,
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
            ],
          ),
          if (hasWallet) ...[
            const SizedBox(height: 14),
            GestureDetector(
              onTap: () {
                Clipboard.setData(ClipboardData(text: widget.address));
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Endereço copiado!', style: GoogleFonts.inter()),
                    duration: const Duration(seconds: 1),
                    backgroundColor: MeshColors.surface,
                    behavior: SnackBarBehavior.floating,
                  ),
                );
              },
              child: Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: MeshColors.surfaceLight,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        widget.address,
                        style: GoogleFonts.jetBrainsMono(
                          color: MeshColors.neonCyan,
                          fontSize: 11,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Icon(Icons.copy, color: MeshColors.textMuted, size: 16),
                  ],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildQuickActions() {
    bool hasWallet = widget.address != 'Não gerada';

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: [
          Expanded(
            child: _isGenerating
                ? const Center(child: CircularProgressIndicator(color: MeshColors.neonCyan))
                : NeonButton(
                    label: hasWallet ? 'Nova Carteira' : 'Gerar Carteira',
                    icon: hasWallet ? Icons.refresh : Icons.add,
                    onPressed: _generateWallet,
                    gradient: hasWallet
                        ? const LinearGradient(colors: [MeshColors.neonViolet, MeshColors.neonPink])
                        : null,
                  ),
          ),
          if (hasWallet) ...[
            const SizedBox(width: 12),
            Expanded(
              child: NeonButton(
                label: _showReceive ? 'Fechar QR' : 'Receber',
                icon: _showReceive ? Icons.close : Icons.qr_code,
                onPressed: () => setState(() => _showReceive = !_showReceive),
                gradient: const LinearGradient(
                  colors: [MeshColors.neonGreen, Color(0xFF059669)],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildReceiveCard() {
    return GlassCard(
      borderColor: MeshColors.neonGreen.withOpacity(0.3),
      child: Column(
        children: [
          Text(
            'Escaneie para receber MESH',
            style: GoogleFonts.inter(
              color: MeshColors.textPrimary,
              fontSize: 14,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
            ),
            child: QrImageView(
              data: widget.address,
              version: QrVersions.auto,
              size: 180,
              backgroundColor: Colors.white,
              eyeStyle: const QrEyeStyle(color: Color(0xFF0A0E1A)),
              dataModuleStyle: const QrDataModuleStyle(color: Color(0xFF0A0E1A)),
            ),
          ),
          const SizedBox(height: 12),
          Text(
            widget.address,
            style: GoogleFonts.jetBrainsMono(
              color: MeshColors.neonCyan,
              fontSize: 10,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Widget _buildSendCard() {
    return GlassCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.send, color: MeshColors.neonCyan, size: 20),
              const SizedBox(width: 8),
              Text(
                'Enviar MESH',
                style: GoogleFonts.inter(
                  color: MeshColors.textPrimary,
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          TextField(
            controller: _destController,
            style: GoogleFonts.jetBrainsMono(color: MeshColors.textPrimary, fontSize: 13),
            decoration: InputDecoration(
              labelText: 'Endereço do Destinatário',
              prefixIcon: const Icon(Icons.person_outline, color: MeshColors.textMuted, size: 20),
              suffixIcon: IconButton(
                icon: const Icon(Icons.qr_code_scanner, color: MeshColors.neonCyan, size: 20),
                onPressed: () {}, // Futuro: scanner de QR
              ),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _valorController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            style: GoogleFonts.jetBrainsMono(color: MeshColors.textPrimary, fontSize: 13),
            decoration: const InputDecoration(
              labelText: 'Valor (MESH)',
              prefixIcon: Icon(Icons.monetization_on_outlined, color: MeshColors.textMuted, size: 20),
            ),
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: NeonButton(
              label: 'Enviar Transação',
              icon: Icons.arrow_forward,
              onPressed: _sendPayment,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTransactionHistory() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Histórico',
            style: GoogleFonts.inter(
              color: MeshColors.textPrimary,
              fontSize: 18,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 12),
          if (widget.transactions.isEmpty)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: MeshColors.glass,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  Icon(Icons.history, color: MeshColors.textMuted, size: 36),
                  const SizedBox(height: 8),
                  Text('Sem transações', style: GoogleFonts.inter(color: MeshColors.textMuted)),
                ],
              ),
            )
          else
            ...widget.transactions.map((tx) {
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
                      width: 36,
                      height: 36,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: (isSent ? MeshColors.neonRed : MeshColors.neonGreen).withOpacity(0.15),
                      ),
                      child: Icon(
                        isSent ? Icons.arrow_upward : Icons.arrow_downward,
                        color: isSent ? MeshColors.neonRed : MeshColors.neonGreen,
                        size: 18,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            isSent ? 'Enviado' : 'Recebido',
                            style: GoogleFonts.inter(color: MeshColors.textPrimary, fontWeight: FontWeight.w600, fontSize: 13),
                          ),
                          Text(
                            '${(tx['destinatario'] ?? tx['remetente'] ?? '').toString().substring(0, 16)}...',
                            style: GoogleFonts.jetBrainsMono(color: MeshColors.textMuted, fontSize: 10),
                          ),
                        ],
                      ),
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          '${isSent ? "-" : "+"}${tx['valor'] ?? '0'} MESH',
                          style: GoogleFonts.jetBrainsMono(
                            color: isSent ? MeshColors.neonRed : MeshColors.neonGreen,
                            fontWeight: FontWeight.w700,
                            fontSize: 13,
                          ),
                        ),
                        StatusDot(
                          color: MeshColors.neonGold,
                          label: 'Mesh P2P',
                        ),
                      ],
                    ),
                  ],
                ),
              );
            }),
        ],
      ),
    );
  }
}
