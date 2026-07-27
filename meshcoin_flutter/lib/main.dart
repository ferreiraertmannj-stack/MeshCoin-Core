import 'dart:math';
import 'dart:io';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'theme/app_theme.dart';
import 'screens/messenger_screen.dart';
import 'screens/nebula_screen.dart';
import 'screens/marketplace_screen.dart';
import 'screens/wallet_screen.dart';
import 'screens/mining_screen.dart';
import 'screens/settings_screen.dart';
import 'mesh_node.dart';
import 'screens/mining_screen.dart';
import 'crypto/wallet_crypto.dart';
import 'crypto/pqc.dart';
import 'tunneling/mesh_tunnel.dart';
import 'blockchain/ledger.dart';
import 'blockchain/transaction.dart';
import 'blockchain/block.dart';
import 'package:permission_handler/permission_handler.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const MeshCoinApp());
}

class MeshCoinApp extends StatelessWidget {
  const MeshCoinApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Nebula Network',
      debugShowCheckedModeBanner: false,
      theme: MeshTheme.darkTheme,
      home: const SplashScreen(),
    );
  }
}

/// ═══════════════════════════════════════════════════════════════
/// Splash Screen — Loading animado com logo
/// ═══════════════════════════════════════════════════════════════
class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _fadeIn;
  late Animation<double> _scaleUp;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 2000),
      vsync: this,
    );
    _fadeIn = Tween<double>(begin: 0, end: 1).animate(
      CurvedAnimation(parent: _controller, curve: const Interval(0, 0.6, curve: Curves.easeIn)),
    );
    _scaleUp = Tween<double>(begin: 0.5, end: 1.0).animate(
      CurvedAnimation(parent: _controller, curve: const Interval(0, 0.6, curve: Curves.elasticOut)),
    );
    _controller.forward();

    Future.delayed(const Duration(milliseconds: 2500), () {
      if (mounted) {
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(builder: (_) => const MeshCoinShell()),
        );
      }
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: MeshColors.background,
      body: Stack(
        fit: StackFit.expand,
        children: [
          // Fundo cósmico da Nebulosa
          Opacity(
            opacity: 0.6,
            child: Image.asset(
              'assets/images/splash_bg.png',
              fit: BoxFit.cover,
              errorBuilder: (context, error, stack) => Container(color: MeshColors.background),
            ),
          ),
          // Gradiente sobre a imagem
          Container(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  MeshColors.background.withOpacity(0.3),
                  MeshColors.background.withOpacity(0.85),
                  MeshColors.background,
                ],
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                stops: const [0.0, 0.6, 1.0],
              ),
            ),
          ),
          // Conteúdo centralizado
          Center(
            child: AnimatedBuilder(
              animation: _controller,
              builder: (context, child) {
                return Opacity(
                  opacity: _fadeIn.value,
                  child: Transform.scale(
                    scale: _scaleUp.value,
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        // Logo com glow
                        Container(
                          width: 120,
                          height: 120,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            boxShadow: [
                              BoxShadow(
                                color: MeshColors.neonCyan.withOpacity(0.4),
                                blurRadius: 40,
                                spreadRadius: 10,
                              ),
                              BoxShadow(
                                color: MeshColors.neonViolet.withOpacity(0.2),
                                blurRadius: 60,
                                spreadRadius: 20,
                              ),
                            ],
                          ),
                          child: ClipOval(
                            child: Image.asset(
                              'assets/images/logo.png',
                              width: 120,
                              height: 120,
                              fit: BoxFit.cover,
                              errorBuilder: (context, error, stack) => Container(
                                decoration: const BoxDecoration(
                                  shape: BoxShape.circle,
                                  gradient: MeshColors.neonGradient,
                                ),
                                child: const Icon(Icons.blur_on, size: 60, color: MeshColors.background),
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(height: 24),
                        Text(
                          'Nebula',
                          style: GoogleFonts.inter(
                            color: MeshColors.textPrimary,
                            fontSize: 32,
                            fontWeight: FontWeight.w800,
                            letterSpacing: 1,
                          ),
                        ),
                        const SizedBox(height: 6),
                        ShaderMask(
                          shaderCallback: (bounds) => MeshColors.neonGradient.createShader(bounds),
                          child: Text(
                            'A rede está lá. Você só não vê.',
                            style: GoogleFonts.inter(
                              color: Colors.white,
                              fontSize: 13,
                              fontWeight: FontWeight.w500,
                              letterSpacing: 1.5,
                            ),
                          ),
                        ),
                        const SizedBox(height: 40),
                        SizedBox(
                          width: 24,
                          height: 24,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation(MeshColors.neonCyan.withOpacity(0.6)),
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

/// ═══════════════════════════════════════════════════════════════
/// App Shell — 5 Tabs Navigation
/// ═══════════════════════════════════════════════════════════════
class MeshCoinShell extends StatefulWidget {
  const MeshCoinShell({super.key});

  @override
  State<MeshCoinShell> createState() => _MeshCoinShellState();
}

class _MeshCoinShellState extends State<MeshCoinShell> {
  int _currentIndex = 0;
  String _address = 'Não gerada';
  String _publicKey = '';
  String _privateKey = '';
  
  late MeshNode meshNode;
  final MeshTunnelQueue tunnelQueue = MeshTunnelQueue();
  List<Map<String, dynamic>> transactions = [];
  List<Map<String, dynamic>> chatMessages = [];
  
  late Ledger ledger;

  @override
  void initState() {
    super.initState();
    ledger = Ledger();
    ledger.addListener(() {
      if (mounted) setState(() {});
    });

    meshNode = MeshNode('MeshNode_${Random().nextInt(9999)}');
    meshNode.onMessage(_handleIncomingPacket);
    
    _initializeNetwork();
    _loadSavedWallet();
  }

  Future<void> _initializeNetwork() async {
    if (!Platform.isWindows && !Platform.isLinux && !Platform.isMacOS) {
      await _requestPermissions();
    }
    
    // Inicia o MeshNode (No desktop ele vai ligar apenas o WebSocket Bridge)
    await meshNode.start();
    
    // Tenta baixar a Blockchain real do Go Node (Sidecar) se existir
    _syncLedgerFromGoNode();
  }

  Future<void> _syncLedgerFromGoNode() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      String bridgeIp = prefs.getString('pc_node_ip') ?? '';
      
      // Remove a porta se o usuário salvou com a porta acidentalmente
      if (bridgeIp.contains(':')) {
        bridgeIp = bridgeIp.split(':')[0];
      }

      if ((bridgeIp.isEmpty || bridgeIp == '127.0.0.1') && meshNode.directPeers.isNotEmpty) {
        bridgeIp = meshNode.directPeers.first.split(':')[0];
      }

      if (bridgeIp.isNotEmpty && bridgeIp != '127.0.0.1') {
        print("🔗 Tentando Sync Automático com PC Node: $bridgeIp");
        bool success = await ledger.syncWithPCNode(bridgeIp);
        if (success) {
          if (mounted) setState(() {});
          print("✅ Ledger Sincronizado automaticamente no startup!");
        }
      } else {
        // Tenta novamente em 3 segundos caso o UDP broadcast chegue um pouco depois
        Future.delayed(const Duration(seconds: 3), () {
          if (mounted) _syncLedgerFromGoNode();
        });
      }
    } catch (e) {
      print("⚠️ Falha ao baixar ledger do Go Node no startup: $e");
    }
  }

  Future<void> _requestPermissions() async {
    // Solicita as permissões cruciais para o Wi-Fi Direct e Bluetooth funcionarem no Android 12+
    await [
      Permission.bluetooth,
      Permission.bluetoothScan,
      Permission.bluetoothConnect,
      Permission.bluetoothAdvertise,
      Permission.location,
      Permission.locationAlways,
      Permission.nearbyWifiDevices,
    ].request();
  }

  double get _balance => ledger.getBalanceOfAddress(_address);

  Future<void> _loadSavedWallet() async {
    Map<String, String>? saved = await WalletCrypto.loadWallet();
    if (saved != null) {
      setState(() {
        _address = saved['address']!;
        _publicKey = saved['publicKey']!;
        _privateKey = saved['privateKey']!;
      });
    }
  }

  void _handleIncomingPacket(Map<String, dynamic> packet) {
    if (!mounted) return;
    
    // Desempacota pacotes roteados pelo PC Node (DATA_ROUTE)
    if (packet['tipo'] == 'DATA_ROUTE' && packet.containsKey('payload')) {
      packet = packet['payload'] as Map<String, dynamic>;
    }
    
    String tipo = packet['tipo'] ?? '';

    if (tipo == 'CHAT') {
      String destAddress = packet['destinatarioAddress'] ?? packet['destinatario'] ?? '';
      
      // Se não for broadcast e não for pra mim, ignora (já que agora usamos BROADCAST no roteamento B.A.T.M.A.N.)
      if (destAddress != 'BROADCAST' && destAddress != _address) {
        return;
      }

      String payload = packet['texto'] ?? '';
      
      if (payload.startsWith("NBL_PQC_V1::")) {
        // Tenta decriptar com a "chave PQC" simulada a partir do nosso endereço
        String dummyPqcKey = base64Encode(utf8.encode(_address.padRight(128, '0').substring(0, 128)));
        payload = NebulaPQC.decryptMessage(payload, dummyPqcKey);
      }
      
      setState(() {
        chatMessages.add({
          ...packet,
          'texto': payload,
          'isMine': false,
        });
      });
    } else if (tipo == 'NEW_TRANSACTION') {
      try {
        var tx = Transaction.fromJson(packet['transaction'] ?? packet['tx']);
        bool added = ledger.addTransaction(tx);
        if (added) {
          transactions.add({
            'remetente': tx.senderAddress,
            'destinatario': tx.receiverAddress,
            'valor': tx.amount,
          });
        }
      } catch (e) {
        print("Erro ao processar transação: $e");
      }
    } else if (tipo == 'NEW_BLOCK') {
      try {
        var block = Block.fromJson(packet['block']);
        ledger.receiveBlock(block);
      } catch (e) {
        print("Erro ao processar bloco: $e");
      }
    }
  }

  void _onWalletGenerated(Map<String, String> wallet) {
    setState(() {
      _address = wallet['address']!;
      _publicKey = wallet['publicKey']!;
      _privateKey = wallet['privateKey']!;
    });
  }

  void _onSendPayment(String dest, String valor) async {
    if (_address == 'Não gerada') return;

    double amount = double.tryParse(valor) ?? 0.0;
    if (amount <= 0) return;

    // Criar e assinar a transação
    Transaction tx = Transaction.create(
      senderPubKey: _publicKey,
      senderAddress: _address,
      privateKeyHex: _privateKey,
      receiverAddress: dest,
      amount: amount,
    );

    // Tenta adicionar localmente primeiro
    bool added = ledger.addTransaction(tx);
    
    if (added) {
      // 1. Enviar para a Rede Mesh Offline
      Map<String, dynamic> packet = {
        'tipo': 'NEW_TRANSACTION',
        'tx': tx.toJson(),
      };
      meshNode.sendRoutedData('BROADCAST', packet);

      // 2. Enviar diretamente via TCP para o PC Node (Mempool Global)
      try {
        final prefs = await SharedPreferences.getInstance();
        String bridgeIp = prefs.getString('pc_node_ip') ?? '';
        if (bridgeIp.contains(':')) bridgeIp = bridgeIp.split(':')[0];

        if (bridgeIp.isNotEmpty && bridgeIp != '127.0.0.1') {
          print("📡 Enviando NEW_TRANSACTION para Mempool do PC Node: $bridgeIp:5556");
          Socket socket = await Socket.connect(bridgeIp, 5556, timeout: const Duration(seconds: 3));
          
          Map<String, dynamic> tcpPacket = {
            'tipo': 'NEW_TRANSACTION',
            'transaction': tx.toJson(),
          };
          
          socket.write('${json.encode(tcpPacket)}\n');
          await socket.flush();
          socket.close();
        }
      } catch (e) {
        print("⚠️ PC Node indisponível para receber transação TCP: $e");
      }

      setState(() {
        transactions.add({
          'remetente': tx.senderAddress,
          'destinatario': tx.receiverAddress,
          'valor': tx.amount,
        });
      });
    }
  }

  void _onSendMessage(String recipient, String text) {
    String payload = text;
    String targetAddress = recipient;

    // Resolve username para endereço se aplicável
    if (recipient.startsWith('@')) {
      String? resolved = ledger.resolveUsername(recipient);
      if (resolved != null && resolved.isNotEmpty) {
        targetAddress = resolved;
      }
    }
    
    if (recipient != "BROADCAST") {
      String dummyPqcKey = base64Encode(utf8.encode(targetAddress.padRight(128, '0').substring(0, 128)));
      payload = NebulaPQC.encryptMessage(text, dummyPqcKey);
    }
    
    Map<String, dynamic> msg = {
      'tipo': 'CHAT',
      'remetente': _address,
      'destinatario': recipient,
      'destinatarioAddress': targetAddress,
      'texto': payload,
      'timestamp': DateTime.now().millisecondsSinceEpoch,
    };
    
    meshNode.sendRoutedData('BROADCAST', msg);

    setState(() {
      chatMessages.add({
        ...msg,
        'texto': text,
        'isMine': true,
      });
    });
  }

  String get _appBarTitle {
    switch (_currentIndex) {
      case 0: return '💬 Nebula Chat';
      case 1: return '☁️ Nebula Cloud';
      case 2: return '🛒 Marketplace';
      case 3: return '⛏️ Mineração';
      case 4: return '💼 Carteira & PoW';
      case 5: return '⚙️ Configurações';
      default: return 'Nebula';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBody: true,
      appBar: AppBar(
        title: Text(_appBarTitle),
        flexibleSpace: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: [MeshColors.background, MeshColors.surface],
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
            ),
          ),
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: StatusDot(
              color: meshNode.directPeers.isEmpty ? MeshColors.neonRed : MeshColors.neonGreen,
              label: meshNode.directPeers.isEmpty ? 'Offline' : '${meshNode.directPeers.length} nós',
            ),
          ),
        ],
      ),
      body: Container(
        decoration: const BoxDecoration(
          gradient: MeshColors.darkGradient,
        ),
        child: SafeArea(
          bottom: false,
          child: Padding(
            padding: const EdgeInsets.only(bottom: 80.0), // Espaço para o BottomNav
            child: IndexedStack(
              index: _currentIndex,
              children: [
                MessengerScreen(
              meshNode: meshNode,
              myAddress: _address,
              messages: chatMessages,
              onSendMessage: _onSendMessage,
            ),
            const NebulaScreen(),
            MarketplaceScreen(
              meshNode: meshNode,
              ledger: ledger,
              address: _address,
            ),
            MiningScreen(
              ledger: ledger,
              minerAddress: _address,
              meshNode: meshNode,
            ),
            WalletScreen(
              meshNode: meshNode,
              ledger: ledger,
              address: _address,
              publicKey: _publicKey,
              privateKey: _privateKey,
              balance: ledger.getBalanceOfAddress(_address),
              onWalletGenerated: _onWalletGenerated,
              onSendPayment: _onSendPayment,
              transactions: transactions,
            ),
            const SettingsScreen(),
          ],
        ),
          ), // Padding
        ), // SafeArea
      ), // Container

      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: MeshColors.surface,
          border: Border(
            top: BorderSide(color: MeshColors.neonCyan.withOpacity(0.1), width: 1),
          ),
          boxShadow: [
            BoxShadow(
              color: MeshColors.neonCyan.withOpacity(0.05),
              blurRadius: 20,
              offset: const Offset(0, -5),
            ),
          ],
        ),
        child: BottomNavigationBar(
          currentIndex: _currentIndex,
          onTap: (index) => setState(() => _currentIndex = index),
          type: BottomNavigationBarType.fixed,
          backgroundColor: MeshColors.surface,
          selectedItemColor: MeshColors.neonCyan,
          unselectedItemColor: MeshColors.textMuted,
          items: const [
            BottomNavigationBarItem(
              icon: Icon(Icons.chat_bubble_outline),
              activeIcon: Icon(Icons.chat_bubble),
              label: 'Chat',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.cloud_queue),
              activeIcon: Icon(Icons.cloud),
              label: 'Cloud',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.shopping_cart_outlined),
              activeIcon: Icon(Icons.shopping_cart),
              label: 'Mercado',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.memory_outlined),
              activeIcon: Icon(Icons.memory),
              label: 'Minerar',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.account_balance_wallet_outlined),
              activeIcon: Icon(Icons.account_balance_wallet),
              label: 'Carteira',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.settings_outlined),
              activeIcon: Icon(Icons.settings),
              label: 'Ajustes',
            ),
          ],
        ),
      ),
    );
  }
}
