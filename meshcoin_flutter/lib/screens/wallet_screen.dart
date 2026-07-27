import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import '../theme/app_theme.dart';
import '../crypto/wallet_crypto.dart';
import '../mesh_node.dart';
import '../blockchain/ledger.dart';
import '../blockchain/transaction.dart';
import 'package:path_provider/path_provider.dart';
import 'dart:io';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'mining_screen.dart';

class WalletScreen extends StatefulWidget {
  final MeshNode meshNode;
  final Ledger ledger;
  final String address;
  final String publicKey;
  final String privateKey;
  final double balance;
  final Function(Map<String, String>) onWalletGenerated;
  final Function(String, String) onSendPayment;
  final List<Map<String, dynamic>> transactions;

  const WalletScreen({
    super.key,
    required this.meshNode,
    required this.ledger,
    required this.address,
    required this.publicKey,
    required this.privateKey,
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
    await Future.delayed(const Duration(milliseconds: 500)); // UX delay
    
    try {
      String mnemonic = WalletCrypto.generateMnemonic();
      Map<String, String> wallet = WalletCrypto.generateKeypairFromMnemonic(mnemonic);
      await WalletCrypto.saveWallet(wallet);
      widget.onWalletGenerated(wallet);
      
      if (mounted) {
        _showMnemonicDialog(mnemonic);
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

  void _showMnemonicDialog(String mnemonic) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) {
        return AlertDialog(
          backgroundColor: MeshColors.surface,
          title: Text('Carteira Gerada', style: GoogleFonts.inter(color: MeshColors.textPrimary)),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Guarde estas 24 palavras em local seguro. Elas são a ÚNICA forma de recuperar seu saldo.',
                style: GoogleFonts.inter(color: MeshColors.neonRed, fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: MeshColors.background,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  mnemonic,
                  style: GoogleFonts.jetBrainsMono(color: MeshColors.neonCyan, fontSize: 13),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: mnemonic));
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Copiado para a área de transferência!')));
              },
              child: Text('Copiar', style: GoogleFonts.inter(color: MeshColors.textMuted)),
            ),
            NeonButton(
              label: 'Salvar TXT',
              isSmall: true,
              onPressed: () async {
                try {
                  final directory = await getExternalStorageDirectory();
                  if (directory != null) {
                    final file = File('${directory.path}/nebula_wallet_backup.txt');
                    await file.writeAsString('Nebula Network - Backup de Recuperação\n\nFrase (24 palavras):\n$mnemonic\n\nNUNCA COMPARTILHE ISSO COM NINGUÉM.');
                    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Salvo em: ${file.path}')));
                  }
                } catch (e) {
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Erro ao salvar arquivo.')));
                }
              },
            ),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text('OK, Eu Salvei', style: GoogleFonts.inter(color: MeshColors.neonCyan)),
            ),
          ],
        );
      },
    );
  }

  Future<void> _recoverWallet() async {
    TextEditingController mnemonicController = TextEditingController();
    
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          backgroundColor: MeshColors.surface,
          title: Text('Recuperar Carteira', style: GoogleFonts.inter(color: MeshColors.textPrimary)),
          content: TextField(
            controller: mnemonicController,
            maxLines: 4,
            style: GoogleFonts.inter(color: MeshColors.textPrimary),
            decoration: InputDecoration(
              hintText: 'Cole aqui suas 24 palavras separadas por espaço',
              hintStyle: GoogleFonts.inter(color: MeshColors.textMuted),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text('Cancelar', style: GoogleFonts.inter(color: MeshColors.textMuted)),
            ),
            NeonButton(
              label: 'Recuperar',
              isSmall: true,
              onPressed: () async {
                Navigator.pop(context);
                setState(() => _isGenerating = true);
                await Future.delayed(const Duration(milliseconds: 500));
                
                try {
                  String mnemonic = mnemonicController.text.trim();
                  Map<String, String> wallet = WalletCrypto.generateKeypairFromMnemonic(mnemonic);
                  await WalletCrypto.saveWallet(wallet);
                  widget.onWalletGenerated(wallet);
                  
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Carteira recuperada com sucesso!'), backgroundColor: MeshColors.neonGreen),
                  );
                } catch (e) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Erro: $e'), backgroundColor: MeshColors.neonRed),
                  );
                }
                setState(() => _isGenerating = false);
              },
            ),
          ],
        );
      },
    );
  }

  Future<void> _syncWithPC() async {
    TextEditingController ipController = TextEditingController();
    
    // Tenta sugerir o IP da rede local se já tiver descoberto
    if (widget.meshNode.directPeers.isNotEmpty) {
      String suggested = widget.meshNode.directPeers.first.split(':').first;
      ipController.text = suggested;
    }

    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          backgroundColor: MeshColors.surface,
          title: Text('Sincronizar com PC Node', style: GoogleFonts.inter(color: MeshColors.textPrimary)),
          content: TextField(
            controller: ipController,
            style: GoogleFonts.inter(color: MeshColors.textPrimary),
            decoration: InputDecoration(
              hintText: 'Digite o IP do PC (ex: 192.168.1.15)',
              hintStyle: GoogleFonts.inter(color: MeshColors.textMuted),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text('Cancelar', style: GoogleFonts.inter(color: MeshColors.textMuted)),
            ),
            NeonButton(
              label: 'Baixar Ledger',
              isSmall: true,
              onPressed: () async {
                Navigator.pop(context);
                String ip = ipController.text.trim();
                if (ip.isNotEmpty) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Sincronizando com PC Node...')),
                  );
                  bool success = await widget.ledger.syncWithPCNode(ip);
                  if (success) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Sincronizado! Saldo atualizado.'), backgroundColor: MeshColors.neonGreen),
                    );
                    // Atualiza o saldo usando a função que iterará a chain baixada
                    // Note: No design atual, se a chain mudar, a Wallet precisa ser recarregada
                  } else {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Falha ao sincronizar. PC não encontrado.'), backgroundColor: MeshColors.neonRed),
                    );
                  }
                }
              },
            ),
          ],
        );
      },
    );
  }

  Future<void> _sendPayment() async {
    String dest = _destController.text.trim();
    String valStr = _valorController.text.trim();
    double amount = double.tryParse(valStr) ?? 0.0;

    if (dest.isEmpty || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Endereço ou valor inválido.', style: TextStyle(color: Colors.white)), backgroundColor: Colors.red));
      return;
    }

    try {
      // Create and sign tx
      Transaction tx = Transaction.create(
        senderPubKey: widget.publicKey,
        senderAddress: widget.address,
        privateKeyHex: widget.privateKey,
        receiverAddress: dest,
        amount: amount,
      );

      // Adiciona à Mempool Local do Smartphone para que a mineração possa incluir esta transação no próximo bloco!
      bool addedLocal = widget.ledger.addTransaction(tx);
      if (!addedLocal) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Erro: Saldo insuficiente ou transação duplicada localmente.', style: TextStyle(color: Colors.white)), backgroundColor: Colors.red),
        );
        return;
      }

      // Envia para a Rede Mesh P2P
      widget.meshNode.sendRoutedData('BROADCAST', {
        'tipo': 'NEW_TRANSACTION',
        'tx': tx.toJson(),
      });

      final prefs = await SharedPreferences.getInstance();
      String bridgeIp = prefs.getString('pc_node_ip') ?? '';
      if (bridgeIp.contains(':')) bridgeIp = bridgeIp.split(':')[0];

      if (bridgeIp.isNotEmpty && bridgeIp != '127.0.0.1') {
        Socket socket = await Socket.connect(bridgeIp, 5556, timeout: const Duration(seconds: 5));
        
        Map<String, dynamic> tcpPacket = {
          'tipo': 'NEW_TRANSACTION',
          'transaction': tx.toJson(),
        };
        
        // CRITICAL: The \n at the end is required by Go's json.NewDecoder
        socket.write('${json.encode(tcpPacket)}\n');
        await socket.flush();
        
        // Aguarda resposta
        String responseData = "";
        try {
          List<int> data = await socket.first.timeout(const Duration(seconds: 5));
          responseData = utf8.decode(data);
        } catch(e) {
          print("Erro aguardando resposta: $e");
        }
        socket.destroy();
        
        if (responseData.contains('"error"')) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Erro no PC Node: Saldo insuficiente ou transação rejeitada.', style: TextStyle(color: Colors.white)), backgroundColor: Colors.red));
          return;
        }
        
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Transação adicionada à Mempool! Minere um bloco para confirmar.', style: TextStyle(color: Colors.white)), backgroundColor: Colors.green));
        _valorController.clear();
        _destController.clear();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Transação salva na Mempool Local! Configure o IP do PC Node nos Ajustes para sincronizar.', style: TextStyle(color: Colors.white)), backgroundColor: Colors.orange));
        _valorController.clear();
        _destController.clear();
      }
    } catch (e) {
      print("Transaction error: $e");
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Falha na rede: $e', style: const TextStyle(color: Colors.white)), backgroundColor: Colors.red));
    }
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
                    '${widget.ledger.getBalanceOfAddress(widget.address).toStringAsFixed(2)} NBL',
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
                : Column(
                    children: [
                      SizedBox(
                        width: double.infinity,
                        child: NeonButton(
                          label: hasWallet ? 'Nova Carteira' : 'Criar Carteira',
                          icon: hasWallet ? Icons.refresh : Icons.add,
                          isSmall: true,
                          onPressed: _generateWallet,
                          gradient: hasWallet
                              ? const LinearGradient(colors: [MeshColors.neonViolet, MeshColors.neonPink])
                              : null,
                        ),
                      ),
                      const SizedBox(height: 8),
                      SizedBox(
                        width: double.infinity,
                        child: NeonButton(
                          label: 'Recuperar',
                          icon: Icons.settings_backup_restore,
                          isSmall: true,
                          onPressed: _recoverWallet,
                          gradient: const LinearGradient(colors: [MeshColors.surfaceLight, MeshColors.textMuted]),
                        ),
                      ),
                      const SizedBox(height: 8),
                      SizedBox(
                        width: double.infinity,
                        child: NeonButton(
                          label: 'Sync PC Node',
                          icon: Icons.sync,
                          isSmall: true,
                          onPressed: _syncWithPC,
                          gradient: const LinearGradient(colors: [MeshColors.neonCyan, MeshColors.neonViolet]),
                        ),
                      ),
                    ],
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
            'Escaneie para receber NBL',
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
                'Enviar NBL',
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
                onPressed: () async {
                  final result = await Navigator.push(
                    context,
                    MaterialPageRoute(builder: (_) => const QRScannerScreen()),
                  );
                  if (result != null && result is String) {
                    _destController.text = result;
                  }
                },
              ),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _valorController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            style: GoogleFonts.jetBrainsMono(color: MeshColors.textPrimary, fontSize: 13),
            decoration: const InputDecoration(
              labelText: 'Valor (NBL)',
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
                          '${isSent ? "-" : "+"}${tx['valor'] ?? '0'} NBL',
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

class QRScannerScreen extends StatefulWidget {
  const QRScannerScreen({Key? key}) : super(key: key);

  @override
  State<QRScannerScreen> createState() => _QRScannerScreenState();
}

class _QRScannerScreenState extends State<QRScannerScreen> {
  late MobileScannerController controller;
  bool _isProcessing = false;

  @override
  void initState() {
    super.initState();
    controller = MobileScannerController(detectionSpeed: DetectionSpeed.noDuplicates);
  }

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(title: const Text('Escanear Endereço NBL')),
      body: MobileScanner(
        controller: controller,
        onDetect: (capture) {
          if (_isProcessing) return;
          final List<Barcode> barcodes = capture.barcodes;
          if (barcodes.isNotEmpty && barcodes.first.rawValue != null) {
            _isProcessing = true; // Lock to prevent multiple pops
            final String code = barcodes.first.rawValue!;
            // Stop camera before closing view to prevent thread lock
            controller.stop().then((_) {
              if (mounted) Navigator.pop(context, code);
            });
          }
        },
      ),
    );
  }
}
