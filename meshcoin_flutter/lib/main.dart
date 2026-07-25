import 'dart:math';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'theme/app_theme.dart';
import 'mesh_node.dart';
import 'crypto/wallet_crypto.dart';
import 'tunneling/mesh_tunnel.dart';
import 'blockchain/ledger.dart';
import 'blockchain/transaction.dart';
import 'blockchain/block.dart';
import 'screens/home_screen.dart';
import 'screens/wallet_screen.dart';
import 'screens/messenger_screen.dart';
import 'screens/mining_screen.dart';
import 'screens/network_screen.dart';
import 'screens/nebula_screen.dart';
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
    if (Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
      // Modo Desktop Sidecar: Bypass nas permissões Android e não inicia MeshNode nativo
      debugPrint("🖥️ MODO DESKTOP ATIVADO: O nó Go Sidecar gerencia a rede.");
      return;
    }
    
    await _requestPermissions();
    await meshNode.start();
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
    String tipo = packet['tipo'] ?? '';

    if (tipo == 'CHAT') {
      setState(() {
        chatMessages.add({
          ...packet,
          'isMine': false,
        });
      });
    } else if (tipo == 'NEW_TRANSACTION') {
      try {
        var tx = Transaction.fromJson(packet['tx']);
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

  void _onSendPayment(String dest, String valor) {
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
      // Broadcast para os nós
      Map<String, dynamic> packet = {
        'tipo': 'NEW_TRANSACTION',
        'tx': tx.toJson(),
      };
      meshNode.sendRoutedData('BROADCAST', packet);

      setState(() {
        transactions.add({
          'remetente': tx.senderAddress,
          'destinatario': tx.receiverAddress,
          'valor': tx.amount,
        });
      });
    }
  }

  void _onSendMessage(String text) {
    Map<String, dynamic> msg = {
      'tipo': 'CHAT',
      'remetente': _address.length > 12 ? _address.substring(0, 12) : _address,
      'texto': text,
    };
    meshNode.sendRoutedData('BROADCAST', msg);

    setState(() {
      chatMessages.add({
        ...msg,
        'isMine': true,
      });
    });
  }

  String get _appBarTitle {
    switch (_currentIndex) {
      case 0: return '🌌 Nebula';
      case 1: return '💰 Carteira';
      case 2: return '💬 Nebula Chat';
      case 3: return '⛏️ Mineração';
      case 4: return '🛰️ Rede';
      case 5: return '☁️ Nebula Cloud';
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
        child: IndexedStack(
          index: _currentIndex,
          children: [
            HomeScreen(
              meshNode: meshNode,
              address: _address,
              balance: _balance,
              tunnelQueue: tunnelQueue,
              recentTransactions: transactions,
            ),
            WalletScreen(
              meshNode: meshNode,
              address: _address,
              balance: _balance,
              onWalletGenerated: _onWalletGenerated,
              onSendPayment: _onSendPayment,
              transactions: transactions,
            ),
            MessengerScreen(
              meshNode: meshNode,
              myAddress: _address,
              messages: chatMessages,
              onSendMessage: _onSendMessage,
            ),
            MiningScreen(
              ledger: ledger,
              minerAddress: _address,
              meshNode: meshNode,
            ),
            NetworkScreen(
              meshNode: meshNode,
              tunnelQueue: tunnelQueue,
            ),
            const NebulaScreen(),
          ],
        ),
      ),
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
              icon: Icon(Icons.dashboard),
              activeIcon: Icon(Icons.dashboard, size: 28),
              label: 'Home',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.account_balance_wallet_outlined),
              activeIcon: Icon(Icons.account_balance_wallet, size: 28),
              label: 'Carteira',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.forum_outlined),
              activeIcon: Icon(Icons.forum, size: 28),
              label: 'Chat',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.memory),
              activeIcon: Icon(Icons.memory, size: 28),
              label: 'Minerar',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.hub_outlined),
              activeIcon: Icon(Icons.hub, size: 28),
              label: 'Rede',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.cloud_outlined),
              activeIcon: Icon(Icons.cloud, size: 28),
              label: 'Nebula',
            ),
          ],
        ),
      ),
    );
  }
}
