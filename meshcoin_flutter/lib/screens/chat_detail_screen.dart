import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import 'package:image_picker/image_picker.dart';
import '../mesh_node.dart';
import '../crypto/pqc.dart';
// import 'package:record/record.dart'; // Audio temporariamente abstraído para foco na UI

class ChatDetailScreen extends StatefulWidget {
  final MeshNode meshNode;
  final String myAddress;
  final String contactId; // Address ou Username
  final List<Map<String, dynamic>> initialMessages;
  final Function(String, String) onSendMessage;

  const ChatDetailScreen({
    super.key,
    required this.meshNode,
    required this.myAddress,
    required this.contactId,
    required this.initialMessages,
    required this.onSendMessage,
  });

  @override
  State<ChatDetailScreen> createState() => _ChatDetailScreenState();
}

class _ChatDetailScreenState extends State<ChatDetailScreen> {
  final TextEditingController _msgController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  late List<Map<String, dynamic>> _messages;

  @override
  void initState() {
    super.initState();
    _messages = List.from(widget.initialMessages);
    widget.meshNode.onMessage(_handleIncomingMessage);
    WidgetsBinding.instance.addPostFrameCallback((_) => _scrollToBottom());
  }

  void _handleIncomingMessage(Map<String, dynamic> packet) {
    if (packet['tipo'] == 'DATA_ROUTE' && packet.containsKey('payload')) {
      packet = packet['payload'] as Map<String, dynamic>;
    }
    
    if (packet['tipo'] == 'CHAT') {
      String remetente = packet['remetente'] ?? '';
      String destinatario = packet['destinatario'] ?? '';
      String destAddr = packet['destinatarioAddress'] ?? '';
      
      bool isBroadcast = (widget.contactId == 'BROADCAST' && destinatario == 'BROADCAST');
      bool isDirectMessage = (remetente == widget.contactId) || 
                             (destinatario == widget.contactId) ||
                             (destAddr == widget.myAddress && remetente == widget.contactId) ||
                             (destinatario == widget.myAddress);

      bool isForThisChat = isBroadcast || isDirectMessage;
      
      if (isForThisChat) {
        String payload = packet['texto'] ?? '';
        
        if (payload.startsWith("NBL_PQC_V1::")) {
          String dummyPqcKey = base64Encode(utf8.encode(widget.myAddress.padRight(128, '0').substring(0, 128)));
          payload = NebulaPQC.decryptMessage(payload, dummyPqcKey);
        }
        
        if (mounted) {
          setState(() {
            _messages.add({
              ...packet,
              'texto': payload,
              'isMine': false,
            });
          });
          _scrollToBottom();
        }
      }
    }
  }

  @override
  void dispose() {
    widget.meshNode.removeMessage(_handleIncomingMessage);
    _msgController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  // Permite injetar novas mensagens que chegaram da tela pai
  void updateMessages(List<Map<String, dynamic>> updatedMessages) {
    setState(() {
      _messages = List.from(updatedMessages);
    });
    _scrollToBottom();
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent + 100,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  void _sendMessage() {
    final text = _msgController.text.trim();
    if (text.isEmpty) return;
    
    widget.onSendMessage(widget.contactId, text);
    
    setState(() {
      _messages.add({
        'remetente': 'eu',
        'destinatario': widget.contactId,
        'texto': text,
        'isMine': true,
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      });
      _msgController.clear();
    });
    _scrollToBottom();
  }

  Future<void> _pickAndSendImage() async {
    final ImagePicker picker = ImagePicker();
    // Limita qualidade para manter leve na P2P
    final XFile? image = await picker.pickImage(source: ImageSource.gallery, imageQuality: 30);
    
    if (image != null) {
      final bytes = await image.readAsBytes();
      String base64Image = base64Encode(bytes);
      
      // Enviando marcador prefixado de imagem
      final imagePayload = '[IMG]:$base64Image';
      widget.onSendMessage(widget.contactId, imagePayload);
      
      setState(() {
        _messages.add({
          'remetente': 'eu',
          'destinatario': widget.contactId,
          'texto': imagePayload,
          'isMine': true,
          'timestamp': DateTime.now().millisecondsSinceEpoch,
        });
      });
      _scrollToBottom();
    }
  }

  Widget _buildMessageBubble(Map<String, dynamic> msg) {
    bool isMine = msg['isMine'] == true;
    String texto = msg['texto'] ?? '';
    
    Widget content;
    if (texto.startsWith('[IMG]:')) {
      String base64str = texto.substring(6);
      content = Image.memory(base64Decode(base64str), fit: BoxFit.cover);
    } else {
      content = Text(
        texto,
        style: GoogleFonts.inter(color: isMine ? MeshColors.background : MeshColors.textPrimary),
      );
    }

    return Align(
      alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 12),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isMine ? MeshColors.neonCyan : MeshColors.surfaceLight,
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(16),
            topRight: const Radius.circular(16),
            bottomLeft: Radius.circular(isMine ? 16 : 0),
            bottomRight: Radius.circular(isMine ? 0 : 16),
          ),
        ),
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
        child: content,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: MeshColors.background,
      appBar: AppBar(
        title: Text(widget.contactId, style: GoogleFonts.inter(fontSize: 18, fontWeight: FontWeight.bold)),
        backgroundColor: MeshColors.surface,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: MeshColors.neonCyan),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.symmetric(vertical: 16),
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                return _buildMessageBubble(_messages[index]);
              },
            ),
          ),
          
          // Input Box
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
            decoration: BoxDecoration(
              color: MeshColors.surface,
              border: Border(top: BorderSide(color: MeshColors.neonCyan.withOpacity(0.2))),
            ),
            child: Row(
              children: [
                IconButton(
                  icon: const Icon(Icons.attach_file, color: MeshColors.textMuted),
                  onPressed: _pickAndSendImage,
                ),
                Expanded(
                  child: TextField(
                    controller: _msgController,
                    style: TextStyle(color: MeshColors.textPrimary),
                    decoration: InputDecoration(
                      hintText: 'Mensagem (P2P)...',
                      hintStyle: TextStyle(color: MeshColors.textMuted),
                      border: InputBorder.none,
                      contentPadding: const EdgeInsets.symmetric(horizontal: 16),
                    ),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.mic, color: MeshColors.neonCyan),
                  onPressed: () {
                    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Áudio em desenvolvimento...')));
                  },
                ),
                IconButton(
                  icon: const Icon(Icons.send, color: MeshColors.neonCyan),
                  onPressed: _sendMessage,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
