import 'dart:convert';
import 'dart:io';
import 'dart:isolate';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:path_provider/path_provider.dart';
import 'block.dart';
import 'transaction.dart';

class Ledger extends ChangeNotifier {
  List<Block> chain = [];
  List<Transaction> mempool = [];
  
  // Dificuldade da rede restaurada para produção (Proof of Work real)
  int difficulty = 6; 
  final double miningReward = 50.0; // 50 NBL

  Ledger() {
    chain.add(Block.genesis());
  }

  Block get latestBlock => chain.last;

  /// Retorna o saldo real de um endereço iterando sobre toda a blockchain
  double getBalanceOfAddress(String address) {
    if (address == 'Não gerada') return 0.0;

    double balance = 0;

    for (var block in chain) {
      for (var trans in block.transactions) {
        if (trans.senderAddress == address) {
          balance -= trans.amount;
          balance -= trans.fee;
        }
        if (trans.receiverAddress == address) {
          balance += trans.amount;
        }
      }
    }
    return balance;
  }

  /// Mapeia todos os usernames registrados e seus donos (Endereço)
  Map<String, String> getRegisteredUsernames() {
    Map<String, String> usernames = {};
    for (var block in chain) {
      for (var trans in block.transactions) {
        if (trans.isUsernameRegistration()) {
          // Apenas o primeiro registro ganha (imutável)
          if (!usernames.containsKey(trans.receiverAddress)) {
            usernames[trans.receiverAddress] = trans.senderAddress;
          }
        }
      }
    }
    return usernames;
  }

  /// Resolve um nome (ex: @joao) para um endereço NBL...
  String? resolveUsername(String username) {
    return getRegisteredUsernames()[username];
  }

  /// Adiciona uma transação na mempool após validá-la
  bool addTransaction(Transaction tx) {
    if (!tx.isValid()) {
      return false;
    }

    if (tx.senderAddress != 'COINBASE') {
      double senderBalance = getBalanceOfAddress(tx.senderAddress);
      double pendingOut = mempool
          .where((t) => t.senderAddress == tx.senderAddress)
          .fold(0.0, (sum, t) => sum + t.amount + t.fee);

      if ((senderBalance - pendingOut) < (tx.amount + tx.fee)) {
        return false; // Saldo insuficiente
      }
    }

    // Regra de unicidade de Username
    if (tx.isUsernameRegistration()) {
      Map<String, String> registered = getRegisteredUsernames();
      if (registered.containsKey(tx.receiverAddress)) {
        return false; // Username já registrado
      }
      // Checar se já não está pendente na mempool
      if (mempool.any((t) => t.isUsernameRegistration() && t.receiverAddress == tx.receiverAddress)) {
        return false;
      }
    }

    // Verifica se já existe na mempool
    if (mempool.any((t) => t.id == tx.id)) {
      return false;
    }

    mempool.add(tx);
    notifyListeners();
    return true;
  }

  /// Minera um novo bloco incluindo as transações da mempool (permite blocos vazios na Nebula Network)
  Future<Block?> minePendingTransactions(String minerAddress, {Function(int hashes)? onProgress}) async {
    List<Transaction> blockTxs = List.from(mempool);
    double totalFees = blockTxs.fold(0.0, (sum, tx) => sum + tx.fee);
    
    Transaction coinbaseTx = Transaction.createCoinbase(minerAddress, miningReward + totalFees);
    blockTxs.insert(0, coinbaseTx);

    String merkleRoot = Block.calculateMerkleRoot(blockTxs);

    Block newBlock = Block(
      index: latestBlock.index + 1,
      timestamp: DateTime.now().millisecondsSinceEpoch,
      previousHash: latestBlock.hash,
      merkleRoot: merkleRoot,
      transactions: blockTxs,
      nonce: 0,
      hash: '',
    );

    // Usando Isolate.spawn para Proof of Work assíncrono real + FFI Bridge seguro
    ReceivePort receivePort = ReceivePort();
    await Isolate.spawn(_miningWorker, {
      'sendPort': receivePort.sendPort,
      'block': newBlock,
      'difficulty': difficulty,
    });

    await for (var msg in receivePort) {
      if (msg is int) {
        if (onProgress != null) onProgress(msg);
      } else if (msg is Map) {
        newBlock.nonce = msg['nonce'];
        newBlock.hash = msg['hash'];
        receivePort.close();
        break;
      }
    }

    // Bloco minerado com sucesso
    chain.add(newBlock);
    mempool.clear();
    notifyListeners();

    return newBlock;
  }

  /// Adiciona um bloco recebido da rede P2P
  bool receiveBlock(Block newBlock) {
    // 1. Verifica se já temos o bloco
    if (newBlock.index <= latestBlock.index) return false;

    // 2. Se for um bloco futuro distante, precisa fazer Sync (ignora por agora)
    if (newBlock.index > latestBlock.index + 1) return false;

    // 3. Valida se o bloco bate com nossa chain local
    if (newBlock.previousHash != latestBlock.hash) return false;

    // 4. Valida hashes e assinaturas
    if (!newBlock.isValid(difficulty)) return false;

    chain.add(newBlock);
    
    // Limpa da mempool as transações que vieram neste bloco
    Set<String> blockTxIds = newBlock.transactions.map((t) => t.id).toSet();
    mempool.removeWhere((tx) => blockTxIds.contains(tx.id));

    notifyListeners();
    return true;
  }

  bool isChainValid() {
    for (int i = 1; i < chain.length; i++) {
      Block currentBlock = chain[i];
      Block previousBlock = chain[i - 1];

      if (!currentBlock.isValid(difficulty)) return false;
      if (currentBlock.previousHash != previousBlock.hash) return false;
    }
    return true;
  }

  Future<bool> syncWithPCNode(String pcIp) async {
    try {
      // Strip any port if the user accidentally included it
      String cleanIp = pcIp.contains(':') ? pcIp.split(':')[0] : pcIp;
      
      print("Fazendo GET HTTP para: http://$cleanIp:8080/api/ledger");
      final response = await http.get(Uri.parse('http://$cleanIp:8080/api/ledger')).timeout(const Duration(seconds: 5));
      
      if (response.statusCode == 200) {
        List<dynamic> jsonBlocks = json.decode(response.body);
        
        List<Block> newChain = jsonBlocks.map((b) => Block.fromJson(b)).toList();
        
        // O PC Node é o Ledger Mestre. Sempre sincronizamos com ele para evitar forks isolados (Desync de Saldo).
        if (newChain.isNotEmpty) { 
          chain = newChain;
          
          // Gravação Local: sobrescrever o ledger.json local do smartphone
          if (!kIsWeb) {
            final directory = await getApplicationDocumentsDirectory();
            final file = File('${directory.path}/ledger.json');
            await file.writeAsString(response.body);
          }
          
          notifyListeners();
          return true;
        }
      } else {
        print("Erro HTTP: \${response.statusCode}");
      }
    } catch (e) {
      print("Erro ao sincronizar Ledger via HTTP: $e");
    }
    return false;
  }
}

// Top-Level function isolada para rodar o NeonHash sem travar a UI
void _miningWorker(Map<String, dynamic> args) {
  SendPort sendPort = args['sendPort'];
  Block block = args['block'];
  int difficulty = args['difficulty'];

  String target = '0' * difficulty;
  int hashesAttempted = 0;
  
  while (!block.hash.startsWith(target)) {
    block.nonce++;
    block.hash = block.calculateHash();
    hashesAttempted++;
    
    if (hashesAttempted % 2000 == 0) {
      sendPort.send(hashesAttempted);
    }
  }
  
  sendPort.send({
    'nonce': block.nonce,
    'hash': block.hash,
  });
}
