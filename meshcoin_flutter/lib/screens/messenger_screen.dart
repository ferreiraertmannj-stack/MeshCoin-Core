import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../mesh_node.dart';
import 'chat_detail_screen.dart';

class MessengerScreen extends StatefulWidget {
  final MeshNode meshNode;
  final String myAddress;
  final List<Map<String, dynamic>> messages;
  final Function(String, String) onSendMessage;

  const MessengerScreen({
    super.key,
    required this.meshNode,
    required this.myAddress,
    required this.messages,
    required this.onSendMessage,
  });

  @override
  State<MessengerScreen> createState() => _MessengerScreenState();
}

class _MessengerScreenState extends State<MessengerScreen> {
  // Agrupa mensagens por Contato
  Map<String, List<Map<String, dynamic>>> get _conversations {
    Map<String, List<Map<String, dynamic>>> grouped = {};
    for (var msg in widget.messages) {
      String dest = msg['destinatario'] ?? 'Desconhecido';
      String contact;
      if (dest == 'BROADCAST') {
        contact = 'BROADCAST';
      } else {
        contact = msg['isMine'] == true ? dest : (msg['remetente'] ?? 'Desconhecido');
      }
      
      if (!grouped.containsKey(contact)) {
        grouped[contact] = [];
      }
      grouped[contact]!.add(msg);
    }
    return grouped;
  }

  void _openChat(String contactId) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ChatDetailScreen(
          meshNode: widget.meshNode,
          myAddress: widget.myAddress,
          contactId: contactId,
          initialMessages: _conversations[contactId] ?? [],
          onSendMessage: widget.onSendMessage,
        ),
      ),
    );
  }

  void _showNewChatDialog() {
    final TextEditingController contactController = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: MeshColors.surface,
        title: Text('Nova Conversa', style: TextStyle(color: MeshColors.textPrimary)),
        content: TextField(
          controller: contactController,
          style: TextStyle(color: MeshColors.textPrimary),
          decoration: InputDecoration(
            hintText: 'Endereço ou @username',
            hintStyle: TextStyle(color: MeshColors.textMuted),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text('Cancelar', style: TextStyle(color: MeshColors.textMuted))),
          NeonButton(
            label: 'Iniciar Chat',
            isSmall: true,
            onPressed: () {
              final contact = contactController.text.trim();
              if (contact.isNotEmpty) {
                Navigator.pop(ctx);
                _openChat(contact);
              }
            },
          ),
        ],
      ),
    );
  }

  @override
  void initState() {
    super.initState();
    widget.meshNode.onMessage(_handleNetworkMessage);
  }

  void _handleNetworkMessage(Map<String, dynamic> packet) {
    if (mounted) {
      setState(() {});
    }
  }

  @override
  void dispose() {
    widget.meshNode.removeMessage(_handleNetworkMessage);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final convs = _conversations.entries.toList();
    // Ordena pela última mensagem mais recente
    convs.sort((a, b) {
      int tA = a.value.last['timestamp'] ?? 0;
      int tB = b.value.last['timestamp'] ?? 0;
      return tB.compareTo(tA);
    });

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: convs.isEmpty
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.chat_bubble_outline, size: 80, color: MeshColors.textMuted),
                  const SizedBox(height: 20),
                  Text('Nenhuma conversa iniciada.', style: GoogleFonts.inter(color: MeshColors.textMuted)),
                ],
              ),
            )
          : ListView.builder(
              physics: const BouncingScrollPhysics(),
              itemCount: convs.length,
              itemBuilder: (context, index) {
                String contact = convs[index].key;
                var lastMsg = convs[index].value.last;
                String snippet = lastMsg['texto'] ?? '';
                if (snippet.startsWith('[IMG]:')) snippet = '📷 Imagem';

                return ListTile(
                  leading: CircleAvatar(
                    backgroundColor: MeshColors.surfaceLight,
                    child: Text(contact.isNotEmpty ? contact[0].toUpperCase() : '?', style: TextStyle(color: MeshColors.neonCyan)),
                  ),
                  title: Text(contact, style: GoogleFonts.inter(color: MeshColors.textPrimary, fontWeight: FontWeight.bold), maxLines: 1, overflow: TextOverflow.ellipsis),
                  subtitle: Text(snippet, style: GoogleFonts.inter(color: MeshColors.textSecondary), maxLines: 1, overflow: TextOverflow.ellipsis),
                  onTap: () => _openChat(contact),
                );
              },
            ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showNewChatDialog,
        backgroundColor: MeshColors.neonCyan,
        child: const Icon(Icons.edit, color: MeshColors.background),
      ),
    );
  }
}
