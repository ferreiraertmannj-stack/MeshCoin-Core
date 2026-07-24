import 'package:flutter/material.dart';
import 'mesh_node.dart';

void main() {
  runApp(MeshCoinApp());
}

class MeshCoinApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'MeshCoin',
      theme: ThemeData.dark().copyWith(
        primaryColor: Colors.blueAccent,
        scaffoldBackgroundColor: Color(0xFF121212),
      ),
      home: MeshCoinHome(),
    );
  }
}

class MeshCoinHome extends StatefulWidget {
  @override
  _MeshCoinHomeState createState() => _MeshCoinHomeState();
}

class _MeshCoinHomeState extends State<MeshCoinHome> {
  int _currentIndex = 0;
  String myAddress = "Não gerada";
  
  late MeshNode meshNode;
  List<String> chatMessages = [];
  
  TextEditingController _chatController = TextEditingController();
  TextEditingController _destController = TextEditingController();
  TextEditingController _valorController = TextEditingController();

  @override
  void initState() {
    super.initState();
    meshNode = MeshNode("AndroidUser_\${Random().nextInt(1000)}");
    meshNode.onMessage((packet) {
      if (packet['tipo'] == 'CHAT') {
        setState(() {
          chatMessages.add("[\${packet['remetente']}]: \${packet['texto']}");
        });
      }
    });
    meshNode.start();
  }

  void gerarCarteira() {
    setState(() {
      myAddress = "MESH\${Random().nextInt(999999999)}ABCD";
    });
  }

  void enviarPagamento() {
    if (myAddress == "Não gerada") return;
    
    Map<String, dynamic> tx = {
      "tipo": "TRANSACAO",
      "remetente": myAddress,
      "destinatario": _destController.text,
      "valor": _valorController.text,
      "timestamp": DateTime.now().millisecondsSinceEpoch,
    };
    
    meshNode.sendRoutedData("BROADCAST", tx);
    
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Transação enviada para \${meshNode.directPeers.length} nós!'))
    );
  }

  void enviarMensagem() {
    if (_chatController.text.isEmpty) return;
    
    Map<String, dynamic> msg = {
      "tipo": "CHAT",
      "remetente": "AndroidUser",
      "texto": _chatController.text
    };
    
    meshNode.sendRoutedData("BROADCAST", msg);
    setState(() {
      chatMessages.add("[Você]: \${_chatController.text}");
      _chatController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('🌐 MeshCoin P2P')),
      body: _currentIndex == 0 ? _buildCarteira() : _buildChat(),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) => setState(() => _currentIndex = index),
        items: [
          BottomNavigationBarItem(icon: Icon(Icons.account_balance_wallet), label: 'Carteira'),
          BottomNavigationBarItem(icon: Icon(Icons.chat), label: 'Chat P2P'),
        ],
      ),
    );
  }

  Widget _buildCarteira() {
    return Padding(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text("Minha Carteira", style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
          SizedBox(height: 10),
          Text(myAddress, style: TextStyle(color: Colors.greenAccent)),
          SizedBox(height: 10),
          ElevatedButton(onPressed: gerarCarteira, child: Text("Gerar Nova Carteira")),
          Divider(height: 40),
          Text("Enviar Pagamento", style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
          TextField(controller: _destController, decoration: InputDecoration(labelText: "Destinatário")),
          TextField(controller: _valorController, decoration: InputDecoration(labelText: "Valor (MESH)"), keyboardType: TextInputType.number),
          SizedBox(height: 10),
          ElevatedButton(onPressed: enviarPagamento, child: Text("Enviar Transação")),
        ],
      ),
    );
  }

  Widget _buildChat() {
    return Column(
      children: [
        Expanded(
          child: ListView.builder(
            itemCount: chatMessages.length,
            itemBuilder: (context, index) => ListTile(title: Text(chatMessages[index])),
          ),
        ),
        Padding(
          padding: EdgeInsets.all(8),
          child: Row(
            children: [
              Expanded(child: TextField(controller: _chatController, decoration: InputDecoration(hintText: "Mensagem..."))),
              IconButton(icon: Icon(Icons.send), onPressed: enviarMensagem),
            ],
          ),
        )
      ],
    );
  }
}
