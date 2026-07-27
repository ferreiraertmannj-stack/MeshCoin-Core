import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../mesh_node.dart';
import '../blockchain/ledger.dart';
import 'dart:convert';

class MarketplaceScreen extends StatefulWidget {
  final MeshNode meshNode;
  final Ledger ledger;
  final String address;
  
  const MarketplaceScreen({
    super.key,
    required this.meshNode,
    required this.ledger,
    required this.address,
  });

  @override
  State<MarketplaceScreen> createState() => _MarketplaceScreenState();
}

class _MarketplaceScreenState extends State<MarketplaceScreen> {
  final List<Map<String, dynamic>> _posts = [];

  @override
  void initState() {
    super.initState();
    // Escuta novas mensagens da rede Mesh (tipo MARKET_POST)
    widget.meshNode.onMessageCallbacks.add(_handleIncomingMessage);
  }

  void _handleIncomingMessage(Map<String, dynamic> data) {
    if (data['tipo'] == 'MARKET_POST') {
      if (mounted) {
        setState(() {
          _posts.insert(0, data);
        });
      }
    }
  }

  void _showCreatePostDialog() {
    final TextEditingController titleController = TextEditingController();
    final TextEditingController descController = TextEditingController();
    final TextEditingController priceController = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          backgroundColor: MeshColors.surface,
          title: Text('Nova Oferta P2P', style: TextStyle(color: MeshColors.textPrimary)),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: titleController,
                  style: TextStyle(color: MeshColors.textPrimary),
                  decoration: InputDecoration(hintText: 'Título (Ex: Troco 100 NBL por XMR)', hintStyle: TextStyle(color: MeshColors.textMuted)),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: descController,
                  maxLines: 3,
                  style: TextStyle(color: MeshColors.textPrimary),
                  decoration: InputDecoration(hintText: 'Descrição detalhada', hintStyle: TextStyle(color: MeshColors.textMuted)),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: priceController,
                  keyboardType: TextInputType.number,
                  style: TextStyle(color: MeshColors.neonGreen),
                  decoration: InputDecoration(hintText: 'Preço em NBL (Opcional)', hintStyle: TextStyle(color: MeshColors.textMuted)),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: Text('Cancelar', style: TextStyle(color: MeshColors.textMuted))),
            NeonButton(
              label: 'Publicar na Rede',
              isSmall: true,
              onPressed: () {
                if (titleController.text.trim().isEmpty) return;
                
                final post = {
                  'tipo': 'MARKET_POST',
                  'author': widget.address,
                  'title': titleController.text.trim(),
                  'desc': descController.text.trim(),
                  'price': priceController.text.trim(),
                  'timestamp': DateTime.now().millisecondsSinceEpoch,
                };
                
                widget.meshNode.sendRoutedData('BROADCAST', post);
                
                // Add localmente
                setState(() {
                  _posts.insert(0, post);
                });
                
                Navigator.pop(ctx);
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Oferta propagada via Mesh!')));
              },
            ),
          ],
        );
      },
    );
  }

  void _contactSeller(String sellerAddress) {
    if (sellerAddress == widget.address) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Esta é sua própria oferta!')));
      return;
    }
    
    // Mostra um aviso informando para buscar esse endereço no Chat
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: MeshColors.surface,
        title: Text('Contatar Vendedor', style: TextStyle(color: MeshColors.neonCyan)),
        content: Text('Para negociar com o vendedor, inicie um Chat (aba Chat) com o seguinte endereço:\n\n$sellerAddress', style: TextStyle(color: MeshColors.textPrimary)),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text('Entendi', style: TextStyle(color: MeshColors.neonCyan))),
        ],
      )
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: _posts.isEmpty
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.storefront, size: 80, color: MeshColors.textMuted),
                  const SizedBox(height: 20),
                  Text('Nenhuma oferta P2P na sua região.', style: GoogleFonts.inter(color: MeshColors.textMuted)),
                ],
              ),
            )
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              physics: const BouncingScrollPhysics(),
              itemCount: _posts.length,
              itemBuilder: (context, index) {
                final post = _posts[index];
                return GlassCard(
                  borderColor: MeshColors.neonCyan.withOpacity(0.3),
                  child: Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(post['title'] ?? '', style: GoogleFonts.inter(color: MeshColors.textPrimary, fontSize: 18, fontWeight: FontWeight.bold)),
                        const SizedBox(height: 8),
                        Text(post['desc'] ?? '', style: GoogleFonts.inter(color: MeshColors.textSecondary, fontSize: 14)),
                        const SizedBox(height: 12),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              post['price'] != null && post['price'].toString().isNotEmpty ? '${post['price']} NBL' : 'Negociável',
                              style: GoogleFonts.inter(color: MeshColors.neonGreen, fontWeight: FontWeight.bold),
                            ),
                            ElevatedButton.icon(
                              icon: const Icon(Icons.chat_bubble_outline, size: 16),
                              label: const Text('Negociar'),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: MeshColors.surfaceLight,
                                foregroundColor: MeshColors.neonCyan,
                              ),
                              onPressed: () => _contactSeller(post['author'] ?? ''),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text('Vendedor: ${post['author']}', style: TextStyle(color: MeshColors.textMuted, fontSize: 10)),
                      ],
                    ),
                  ),
                );
              },
            ),
      floatingActionButton: widget.address != 'Não gerada'
        ? FloatingActionButton.extended(
            onPressed: _showCreatePostDialog,
            backgroundColor: MeshColors.neonCyan,
            icon: const Icon(Icons.add, color: MeshColors.background),
            label: Text('Criar Oferta', style: GoogleFonts.inter(color: MeshColors.background, fontWeight: FontWeight.bold)),
          )
        : null,
    );
  }
}
