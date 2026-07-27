import 'dart:io';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../theme/app_theme.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final TextEditingController _ipController = TextEditingController();
  String _myLocalIp = '';

  bool _enableP2POffline = true;
  bool _enableBridgeOnline = true;
  String _language = 'pt-BR';

  @override
  void initState() {
    super.initState();
    _loadSettings();
  }

  Future<void> _fetchLocalIp() async {
    String myLocalIp = '';
    try {
      if (Platform.isWindows || Platform.isLinux || Platform.isMacOS) {
        final interfaces = await NetworkInterface.list();
        for (var interface in interfaces) {
          String name = interface.name.toLowerCase();
          
          // Rejeita ativamente interfaces virtuais comuns e loopbacks
          if (name.contains('virtual') || 
              name.contains('wsl') || 
              name.contains('vmware') || 
              name.contains('vbox') || 
              name.contains('vethernet') ||
              name.contains('loopback') ||
              name.contains('hyper')) {
            continue;
          }

          // Busca primeiro as principais (Wi-Fi e Ethernet real)
          if (name.contains('wi-fi') || name.contains('ethernet') || name.startsWith('eth') || name.startsWith('wlan') || name.startsWith('en')) {
            for (var addr in interface.addresses) {
              if (addr.type == InternetAddressType.IPv4 && !addr.isLoopback) {
                // Filtro para garantir sub-rede local comum (evita IPs bizarros)
                if (addr.address.startsWith('192.168.') || addr.address.startsWith('10.') || addr.address.startsWith('172.')) {
                  myLocalIp = addr.address;
                  break;
                }
              }
            }
          }
          if (myLocalIp.isNotEmpty) break;
        }
        
        // Fallback genérico caso o sistema esteja em outro idioma ou naming convention
        if (myLocalIp.isEmpty) {
          for (var interface in interfaces) {
            String name = interface.name.toLowerCase();
            if (name.contains('virtual') || name.contains('wsl') || name.contains('vethernet')) continue;
            
            for (var addr in interface.addresses) {
              if (addr.type == InternetAddressType.IPv4 && !addr.isLoopback) {
                if (addr.address.startsWith('192.168.') || addr.address.startsWith('10.') || addr.address.startsWith('172.')) {
                  myLocalIp = addr.address;
                  break;
                }
              }
            }
            if (myLocalIp.isNotEmpty) break;
          }
        }
      }
    } catch (e) {
      myLocalIp = 'Erro ao ler IP: $e';
    }
    if (mounted) {
      setState(() {
        _myLocalIp = myLocalIp;
      });
    }
  }

  Future<void> _loadSettings() async {
    final prefs = await SharedPreferences.getInstance();
    
    await _fetchLocalIp();

    if (mounted) {
      setState(() {
        _ipController.text = prefs.getString('pc_node_ip') ?? '';
        _enableP2POffline = prefs.getBool('enable_p2p_offline') ?? true;
        _enableBridgeOnline = prefs.getBool('enable_bridge_online') ?? true;
        _language = prefs.getString('app_language') ?? 'pt-BR';
      });
    }
  }

  Future<void> _saveSettings() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('pc_node_ip', _ipController.text.trim());
    await prefs.setBool('enable_p2p_offline', _enableP2POffline);
    await prefs.setBool('enable_bridge_online', _enableBridgeOnline);
    await prefs.setString('app_language', _language);
    
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Configurações salvas! Reinicie o app.'), backgroundColor: MeshColors.neonGreen),
      );
    }
  }

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
        
        if (_myLocalIp.isNotEmpty)
          Container(
            padding: const EdgeInsets.all(12),
            margin: const EdgeInsets.only(bottom: 16),
            decoration: BoxDecoration(
              color: MeshColors.neonCyan.withOpacity(0.1),
              border: Border.all(color: MeshColors.neonCyan),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('PC Node / Bridge Local', style: GoogleFonts.inter(color: MeshColors.neonCyan, fontWeight: FontWeight.bold)),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('IP Local da sua máquina: $_myLocalIp', style: GoogleFonts.inter(color: MeshColors.textPrimary)),
                    IconButton(
                      icon: const Icon(Icons.refresh, color: MeshColors.neonCyan, size: 20),
                      onPressed: _fetchLocalIp,
                      tooltip: 'Atualizar IP',
                    ),
                  ],
                ),
                Text('Sufixo Padrão: :8080 (Copie e conecte os smartphones)', style: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 12)),
              ],
            ),
          ),
        
        // P2P Offline Toggle
        SwitchListTile(
          activeColor: MeshColors.neonGreen,
          title: Text('Modo P2P Offline', style: TextStyle(color: MeshColors.textPrimary)),
          subtitle: Text('Ativa o rádio sem fio (Wi-Fi Direct / BLE) para sincronizar sem internet.', style: TextStyle(color: MeshColors.textMuted, fontSize: 12)),
          value: _enableP2POffline,
          onChanged: (val) {
            setState(() => _enableP2POffline = val);
            _saveSettings();
          },
        ),

        // Bridge Online Toggle
        SwitchListTile(
          activeColor: MeshColors.neonCyan,
          title: Text('Modo Bridge Online', style: TextStyle(color: MeshColors.textPrimary)),
          subtitle: Text('Habilita a ponte via IP/WebSockets para longo alcance na rede local.', style: TextStyle(color: MeshColors.textMuted, fontSize: 12)),
          value: _enableBridgeOnline,
          onChanged: (val) {
            setState(() => _enableBridgeOnline = val);
            _saveSettings();
          },
        ),
        
        if (_enableBridgeOnline)
          // PC Node Bridge IP
          ListTile(
            leading: const Icon(Icons.router, color: MeshColors.neonCyan),
            title: Text('PC Node Bridge (IP)', style: TextStyle(color: MeshColors.textPrimary)),
            subtitle: TextField(
              controller: _ipController,
              style: GoogleFonts.inter(color: MeshColors.textPrimary),
              decoration: InputDecoration(
                hintText: 'Ex: 192.168.0.15:8080',
                hintStyle: GoogleFonts.inter(color: MeshColors.textMuted),
                border: InputBorder.none,
              ),
              onSubmitted: (_) => _saveSettings(),
            ),
            trailing: IconButton(
              icon: const Icon(Icons.save, color: MeshColors.neonCyan),
              onPressed: _saveSettings,
            ),
          ),
        
        const Divider(color: MeshColors.surfaceLight),
        
        // Language Selection
        ListTile(
          leading: const Icon(Icons.language, color: MeshColors.neonCyan),
          title: Text('Idioma (Language)', style: TextStyle(color: MeshColors.textPrimary)),
          trailing: DropdownButton<String>(
            dropdownColor: MeshColors.surfaceLight,
            value: _language,
            style: const TextStyle(color: MeshColors.neonCyan),
            underline: Container(),
            items: const [
              DropdownMenuItem(value: 'pt-BR', child: Text('Português (BR)')),
              DropdownMenuItem(value: 'en', child: Text('English')),
              DropdownMenuItem(value: 'es', child: Text('Español')),
            ],
            onChanged: (val) {
              if (val != null) {
                setState(() => _language = val);
                _saveSettings();
              }
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
