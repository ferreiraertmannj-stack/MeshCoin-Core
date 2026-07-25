import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../mesh_node.dart';

class MessengerScreen extends StatefulWidget {
  final MeshNode meshNode;
  final String myAddress;
  final List<Map<String, dynamic>> messages;
  final Function(String recipient, String text) onSendMessage;

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
  final TextEditingController _recipientController = TextEditingController();
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();

  void _send() {
    if (_messageController.text.trim().isEmpty) return;
    String recipient = _recipientController.text.trim();
    if (recipient.isEmpty) {
      recipient = "BROADCAST";
    }
    
    // O texto é passado, e a main vai criptografar se for um destinatário específico
    widget.onSendMessage(recipient, _messageController.text.trim());
    _messageController.clear();
    
    // Auto-scroll para o final
    Future.delayed(const Duration(milliseconds: 100), () {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Header E2EE
        _buildE2EEHeader(),
        // Campo de Destinatário P2P
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
          child: TextField(
            controller: _recipientController,
            style: GoogleFonts.inter(color: MeshColors.textPrimary, fontSize: 13),
            decoration: InputDecoration(
              hintText: 'Endereço destino (deixe vazio p/ Todos)',
              hintStyle: GoogleFonts.inter(color: MeshColors.textSecondary, fontSize: 13),
              filled: true,
              fillColor: MeshColors.surfaceLight,
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
              prefixIcon: const Icon(Icons.person, color: MeshColors.neonCyan, size: 18),
            ),
          ),
        ),
        // Lista de mensagens
        Expanded(
          child: widget.messages.isEmpty
              ? _buildEmptyState()
              : ListView.builder(
                  controller: _scrollController,
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  itemCount: widget.messages.length,
                  itemBuilder: (context, index) {
                    return _buildMessageBubble(widget.messages[index]);
                  },
                ),
        ),
        // Input bar
        _buildInputBar(),
      ],
    );
  }

  Widget _buildE2EEHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: MeshColors.surface,
        border: Border(bottom: BorderSide(color: MeshColors.surfaceLight.withOpacity(0.5))),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.lock, color: MeshColors.neonGreen, size: 14),
              const SizedBox(width: 6),
              Text(
                'Criptografia Ponta-a-Ponta · AES-256-GCM',
                style: GoogleFonts.inter(
                  color: MeshColors.neonGreen,
                  fontSize: 11,
                  fontWeight: FontWeight.w500,
                ),
              ),
              const SizedBox(width: 6),
              Icon(Icons.lock, color: MeshColors.neonGreen, size: 14),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            'Mensagens expiram em 48h (Renovar c/ Gas)',
            style: GoogleFonts.inter(
              color: MeshColors.neonCyan,
              fontSize: 10,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: MeshColors.neonViolet.withOpacity(0.1),
                border: Border.all(color: MeshColors.neonViolet.withOpacity(0.3)),
              ),
              child: const Icon(Icons.forum, color: MeshColors.neonViolet, size: 36),
            ),
            const SizedBox(height: 20),
            Text(
              'Nebula Chat',
              style: GoogleFonts.inter(
                color: MeshColors.textPrimary,
                fontSize: 20,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Envie mensagens E2EE fragmentadas via rede Nebula.\nDoe 5GB de espaço para a Nebula Cloud e ganhe boost na mineração.',
              textAlign: TextAlign.center,
              style: GoogleFonts.inter(
                color: MeshColors.textMuted,
                fontSize: 13,
                height: 1.5,
              ),
            ),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                _buildFeatureChip(Icons.wifi_off, 'Off-Grid'),
                const SizedBox(width: 8),
                _buildFeatureChip(Icons.shield, 'E2EE'),
                const SizedBox(width: 8),
                _buildFeatureChip(Icons.hub, 'Multi-Hop'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFeatureChip(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: MeshColors.surfaceLight,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: MeshColors.neonCyan.withOpacity(0.2)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: MeshColors.neonCyan, size: 14),
          const SizedBox(width: 4),
          Text(label, style: GoogleFonts.inter(color: MeshColors.textSecondary, fontSize: 11)),
        ],
      ),
    );
  }

  Widget _buildMessageBubble(Map<String, dynamic> msg) {
    bool isMine = msg['isMine'] == true;
    String text = msg['texto'] ?? '';
    String sender = msg['remetente'] ?? 'Anônimo';

    return Align(
      alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          gradient: isMine
              ? const LinearGradient(
                  colors: [MeshColors.neonCyan, MeshColors.neonViolet],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                )
              : null,
          color: isMine ? null : MeshColors.surfaceLight,
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(16),
            topRight: const Radius.circular(16),
            bottomLeft: Radius.circular(isMine ? 16 : 4),
            bottomRight: Radius.circular(isMine ? 4 : 16),
          ),
          boxShadow: isMine
              ? [BoxShadow(color: MeshColors.neonCyan.withOpacity(0.2), blurRadius: 8, offset: const Offset(0, 2))]
              : null,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (!isMine)
              Padding(
                padding: const EdgeInsets.only(bottom: 4),
                child: Text(
                  sender,
                  style: GoogleFonts.inter(
                    color: MeshColors.neonCyan,
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            Text(
              text,
              style: GoogleFonts.inter(
                color: isMine ? MeshColors.background : MeshColors.textPrimary,
                fontSize: 14,
                fontWeight: isMine ? FontWeight.w500 : FontWeight.w400,
              ),
            ),
            const SizedBox(height: 4),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  Icons.lock,
                  size: 10,
                  color: isMine ? MeshColors.background.withOpacity(0.5) : MeshColors.textMuted,
                ),
                const SizedBox(width: 4),
                Text(
                  'Via Mesh',
                  style: GoogleFonts.inter(
                    fontSize: 9,
                    color: isMine ? MeshColors.background.withOpacity(0.5) : MeshColors.textMuted,
                  ),
                ),
                if (!isMine) ...[
                  const SizedBox(width: 12),
                  GestureDetector(
                    onTap: () {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text('Transação de Gas (0.01 NBL) enviada para fixar mensagem na Cloud.', style: GoogleFonts.inter()),
                          backgroundColor: MeshColors.neonViolet,
                        ),
                      );
                    },
                    child: Row(
                      children: [
                        Icon(Icons.push_pin, size: 10, color: MeshColors.neonCyan),
                        const SizedBox(width: 4),
                        Text(
                          'Fixar (Gas)',
                          style: GoogleFonts.inter(fontSize: 9, color: MeshColors.neonCyan),
                        ),
                      ],
                    ),
                  ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInputBar() {
    return Container(
      padding: EdgeInsets.only(
        left: 12,
        right: 8,
        top: 8,
        bottom: MediaQuery.of(context).padding.bottom + 8,
      ),
      decoration: BoxDecoration(
        color: MeshColors.surface,
        border: Border(top: BorderSide(color: MeshColors.surfaceLight.withOpacity(0.5))),
      ),
      child: Row(
        children: [
          IconButton(
            icon: const Icon(Icons.attach_file, color: MeshColors.textMuted),
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('Requisito de 5GB atingido! Envio de mídia fragmentada (PQC) em rollout Beta.', style: GoogleFonts.inter()),
                  backgroundColor: MeshColors.neonViolet,
                  behavior: SnackBarBehavior.floating,
                ),
              );
            },
          ),
          Expanded(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              decoration: BoxDecoration(
                color: MeshColors.surfaceLight,
                borderRadius: BorderRadius.circular(24),
              ),
              child: TextField(
                controller: _messageController,
                style: GoogleFonts.inter(color: MeshColors.textPrimary, fontSize: 14),
                decoration: InputDecoration(
                  hintText: 'Mensagem criptografada...',
                  hintStyle: GoogleFonts.inter(color: MeshColors.textMuted, fontSize: 14),
                  border: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(vertical: 10),
                ),
                onSubmitted: (_) => _send(),
              ),
            ),
          ),
          const SizedBox(width: 8),
          GestureDetector(
            onTap: _send,
            child: Container(
              width: 42,
              height: 42,
              decoration: BoxDecoration(
                gradient: MeshColors.neonGradient,
                shape: BoxShape.circle,
                boxShadow: [
                  BoxShadow(
                    color: MeshColors.neonCyan.withOpacity(0.3),
                    blurRadius: 8,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: const Icon(Icons.send, color: MeshColors.background, size: 20),
            ),
          ),
        ],
      ),
    );
  }
}
