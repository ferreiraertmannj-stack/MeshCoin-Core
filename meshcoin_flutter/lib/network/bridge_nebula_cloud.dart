import 'dart:convert';
import 'package:http/http.dart' as http;

/// Integração teórica (Ponte) entre a Blockchain MeshCoin e a Nébula Cloud.
/// 
/// A Nebula Cloud é uma nuvem de armazenamento descentralizado (Reed-Solomon).
/// O MeshCoin (quando operando como Full Node em um PC/Servidor) usará esta
/// classe para enviar o Ledger compactado para a nuvem.
/// Celulares (Light Nodes) não fazem upload de blocos, mas podem consultar o Tracker.
class NebulaCloudBridge {
  // Endereço do Tracker da Nebula Cloud (orquestrador)
  final String trackerUrl;

  NebulaCloudBridge({this.trackerUrl = "http://127.0.0.1:8001"});

  /// Envia a lista de blocos (Ledger) para ser fragmentada e armazenada nos nós da Nebula
  Future<bool> backupLedgerToCloud(String ledgerJson, String password) async {
    try {
      var request = http.MultipartRequest('POST', Uri.parse('\$trackerUrl/upload'));
      request.fields['senha'] = password;
      request.files.add(
        http.MultipartFile.fromString('arquivo', ledgerJson, filename: 'meshcoin_ledger.json'),
      );

      var response = await request.send();
      if (response.statusCode == 200) {
        print("✅ Ledger particionado e salvo com sucesso na Nebula Cloud!");
        return true;
      } else {
        print("❌ Falha ao salvar Ledger na Nébula Cloud. Status: \${response.statusCode}");
        return false;
      }
    } catch (e) {
      print("Erro de conexão com Nébula Cloud Tracker: \$e");
      return false;
    }
  }

  /// Recupera o Ledger completo a partir dos fragmentos da Nebula Cloud
  Future<String?> downloadLedgerFromCloud(String password) async {
    try {
      var response = await http.post(
        Uri.parse('\$trackerUrl/download'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'nome_arquivo': 'meshcoin_ledger.json',
          'senha': password
        }),
      );

      if (response.statusCode == 200) {
        print("✅ Ledger reconstruído via Nebula Cloud (Reed-Solomon)!");
        return response.body; // Retorna o JSON do Ledger
      } else {
        print("❌ Erro ao reconstruir Ledger: \${response.statusCode}");
        return null;
      }
    } catch (e) {
      print("Erro ao tentar recuperar da Nébula Cloud: \$e");
      return null;
    }
  }
}
