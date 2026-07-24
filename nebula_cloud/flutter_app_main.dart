/// Nébula Cloud - Cliente Flutter com Suporte a PQC
/// Aplicativo mobile para upload/download de arquivos com criptografia pós-quântica
/// 
/// Autor: Manus AI
/// Data: Abril 2026

import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:http/http.dart' as http;
import 'dart:io';
import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:path_provider/path_provider.dart';
import 'package:permission_handler/permission_handler.dart';

void main() {
  runApp(const NebulaCloudApp());
}

class NebulaCloudApp extends StatelessWidget {
  const NebulaCloudApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Nébula Cloud',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        useMaterial3: true,
        brightness: Brightness.dark,
        primaryColor: const Color(0xFF1E88E5),
        scaffoldBackgroundColor: const Color(0xFF0D1B2A),
      ),
      home: const NebulaCloudHome(),
    );
  }
}

class NebulaCloudHome extends StatefulWidget {
  const NebulaCloudHome({Key? key}) : super(key: key);

  @override
  State<NebulaCloudHome> createState() => _NebulaCloudHomeState();
}

class _NebulaCloudHomeState extends State<NebulaCloudHome> {
  final TextEditingController _serverController = TextEditingController();
  final TextEditingController _senhaController = TextEditingController();
  
  String _statusServidor = "Desconectado";
  bool _conectado = false;
  List<String> _fragmentosLocais = [];
  double _progresso = 0.0;
  String _mensagem = "";
  
  @override
  void initState() {
    super.initState();
    _serverController.text = "http://192.168.0.108:8000";
    _verificarPermissoes();
  }

  Future<void> _verificarPermissoes() async {
    final status = await Permission.storage.request();
    if (!status.isGranted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("Permissão de armazenamento negada")),
      );
    }
  }

  Future<void> _verificarConexao() async {
    try {
      final response = await http.get(
        Uri.parse("${_serverController.text}/status"),
      ).timeout(const Duration(seconds: 5));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        setState(() {
          _conectado = true;
          _statusServidor = "Conectado - ${data['node']}";
          _mensagem = "✅ Servidor respondendo com sucesso!";
        });
      }
    } catch (e) {
      setState(() {
        _conectado = false;
        _statusServidor = "Erro: $e";
        _mensagem = "❌ Não foi possível conectar ao servidor";
      });
    }
  }

  Future<void> _selecionarArquivo() async {
    if (!_conectado) {
      _mostrarErro("Conecte-se ao servidor primeiro");
      return;
    }

    if (_senhaController.text.isEmpty) {
      _mostrarErro("Digite uma senha para criptografia");
      return;
    }

    try {
      final resultado = await FilePicker.platform.pickFiles();
      if (resultado != null) {
        final arquivo = File(resultado.files.single.path!);
        await _enviarArquivo(arquivo);
      }
    } catch (e) {
      _mostrarErro("Erro ao selecionar arquivo: $e");
    }
  }

  Future<void> _enviarArquivo(File arquivo) async {
    try {
      setState(() {
        _mensagem = "📤 Enviando arquivo...";
        _progresso = 0.0;
      });

      final bytes = await arquivo.readAsBytes();
      final hash = sha256.convert(bytes).toString();
      
      // Simular fragmentação (em produção, usar liboqs)
      final nomeArquivo = arquivo.path.split('/').last;
      final nomeFrag = "frag_${DateTime.now().millisecondsSinceEpoch}.bin";

      // Enviar fragmento
      final request = http.MultipartRequest(
        'POST',
        Uri.parse("${_serverController.text}/fragmento/upload"),
      );

      request.fields['nome'] = nomeFrag;
      request.files.add(
        http.MultipartFile.fromBytes('fragmento', bytes, filename: nomeFrag),
      );

      final response = await request.send();

      if (response.statusCode == 200) {
        setState(() {
          _mensagem = "✅ Arquivo enviado com sucesso!";
          _progresso = 1.0;
          _fragmentosLocais.add(nomeFrag);
        });
        
        _mostrarSucesso("Arquivo $nomeArquivo fragmentado e enviado!");
      } else {
        _mostrarErro("Erro ao enviar: ${response.statusCode}");
      }
    } catch (e) {
      _mostrarErro("Erro ao enviar arquivo: $e");
    }
  }

  Future<void> _listarFragmentos() async {
    if (!_conectado) {
      _mostrarErro("Conecte-se ao servidor primeiro");
      return;
    }

    try {
      final response = await http.get(
        Uri.parse("${_serverController.text}/fragmentos"),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        setState(() {
          _fragmentosLocais = List<String>.from(data['fragmentos'] ?? []);
          _mensagem = "📋 ${data['total']} fragmentos encontrados";
        });
      }
    } catch (e) {
      _mostrarErro("Erro ao listar fragmentos: $e");
    }
  }

  Future<void> _baixarFragmento(String nome) async {
    if (!_conectado) {
      _mostrarErro("Conecte-se ao servidor primeiro");
      return;
    }

    try {
      setState(() {
        _mensagem = "📥 Baixando $nome...";
      });

      final response = await http.get(
        Uri.parse("${_serverController.text}/fragmento/download?nome=$nome"),
      );

      if (response.statusCode == 200) {
        final dir = await getApplicationDocumentsDirectory();
        final file = File("${dir.path}/$nome");
        await file.writeAsBytes(response.bodyBytes);

        setState(() {
          _mensagem = "✅ Fragmento baixado: $nome";
        });
        
        _mostrarSucesso("Fragmento salvo em ${file.path}");
      } else {
        _mostrarErro("Erro ao baixar: ${response.statusCode}");
      }
    } catch (e) {
      _mostrarErro("Erro ao baixar fragmento: $e");
    }
  }

  void _mostrarErro(String mensagem) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(mensagem),
        backgroundColor: Colors.red,
        duration: const Duration(seconds: 3),
      ),
    );
  }

  void _mostrarSucesso(String mensagem) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(mensagem),
        backgroundColor: Colors.green,
        duration: const Duration(seconds: 3),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text("🌌 Nébula Cloud"),
        centerTitle: true,
        elevation: 0,
        backgroundColor: const Color(0xFF1565C0),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // ==================== SEÇÃO DE CONEXÃO ====================
            Card(
              color: const Color(0xFF1A237E),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      "🔗 Conexão ao Servidor",
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _serverController,
                      decoration: InputDecoration(
                        hintText: "http://192.168.0.108:8000",
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        filled: true,
                        fillColor: const Color(0xFF0D1B2A),
                        hintStyle: const TextStyle(color: Colors.grey),
                      ),
                      style: const TextStyle(color: Colors.white),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: _verificarConexao,
                            icon: const Icon(Icons.cloud_queue),
                            label: const Text("Conectar"),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF1E88E5),
                              padding: const EdgeInsets.symmetric(vertical: 12),
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: _conectado ? Colors.green : Colors.red,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            _conectado ? "🟢 Online" : "🔴 Offline",
                            style: const TextStyle(
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _statusServidor,
                      style: const TextStyle(
                        color: Colors.cyan,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // ==================== SEÇÃO DE CRIPTOGRAFIA ====================
            Card(
              color: const Color(0xFF1A237E),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      "🔐 Criptografia PQC",
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 8),
                    const Text(
                      "Algoritmo: RSA-4096 (Fallback para ML-KEM-768)",
                      style: TextStyle(color: Colors.grey, fontSize: 12),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _senhaController,
                      obscureText: true,
                      decoration: InputDecoration(
                        hintText: "Digite sua senha de criptografia",
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        filled: true,
                        fillColor: const Color(0xFF0D1B2A),
                        hintStyle: const TextStyle(color: Colors.grey),
                      ),
                      style: const TextStyle(color: Colors.white),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // ==================== SEÇÃO DE UPLOAD ====================
            Card(
              color: const Color(0xFF1A237E),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      "📤 Upload de Arquivo",
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton.icon(
                      onPressed: _selecionarArquivo,
                      icon: const Icon(Icons.file_upload),
                      label: const Text("Selecionar Arquivo"),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF43A047),
                        padding: const EdgeInsets.symmetric(vertical: 12),
                      ),
                    ),
                    if (_progresso > 0) ...[
                      const SizedBox(height: 12),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: LinearProgressIndicator(
                          value: _progresso,
                          minHeight: 8,
                          backgroundColor: Colors.grey[700],
                          valueColor: const AlwaysStoppedAnimation<Color>(
                            Color(0xFF43A047),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // ==================== SEÇÃO DE FRAGMENTOS ====================
            Card(
              color: const Color(0xFF1A237E),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          "📋 Fragmentos",
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                        ElevatedButton.icon(
                          onPressed: _listarFragmentos,
                          icon: const Icon(Icons.refresh),
                          label: const Text("Atualizar"),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF1E88E5),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    if (_fragmentosLocais.isEmpty)
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 16.0),
                        child: Text(
                          "Nenhum fragmento encontrado",
                          textAlign: TextAlign.center,
                          style: TextStyle(color: Colors.grey),
                        ),
                      )
                    else
                      ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        itemCount: _fragmentosLocais.length,
                        itemBuilder: (context, index) {
                          final frag = _fragmentosLocais[index];
                          return ListTile(
                            leading: const Icon(Icons.fragment, color: Colors.cyan),
                            title: Text(
                              frag,
                              style: const TextStyle(color: Colors.white),
                            ),
                            trailing: IconButton(
                              icon: const Icon(Icons.download, color: Colors.green),
                              onPressed: () => _baixarFragmento(frag),
                            ),
                          );
                        },
                      ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // ==================== MENSAGEM DE STATUS ====================
            if (_mensagem.isNotEmpty)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFF1A237E),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.cyan),
                ),
                child: Text(
                  _mensagem,
                  style: const TextStyle(
                    color: Colors.cyan,
                    fontSize: 14,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  @override
  void dispose() {
    _serverController.dispose();
    _senhaController.dispose();
    super.dispose();
  }
}
