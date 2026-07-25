import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'block.dart';
import 'transaction.dart';

class Ledger extends ChangeNotifier {
  List<Block> chain = [];
  List<Transaction> mempool = [];
  
  // Dificuldade da rede. Começando baixa para testes de celular (ex: 3)
  int difficulty = 3; 
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

  /// Adiciona uma transação na mempool após validá-la
  bool addTransaction(Transaction tx) {
    if (!tx.isValid()) {
      return false;
    }

    if (tx.senderAddress != 'COINBASE') {
      double senderBalance = getBalanceOfAddress(tx.senderAddress);
      if (senderBalance < (tx.amount + tx.fee)) {
        return false; // Saldo insuficiente
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
  Block? minePendingTransactions(String minerAddress, {Function(int hashes)? onProgress}) {
    // Trava de Tempo (Block Time) - Garante que a rede avance a cada 2 minutos no mínimo
    int now = DateTime.now().millisecondsSinceEpoch;
    if (now - latestBlock.timestamp < 120000) {
      return null; // Muito cedo para minerar outro bloco
    }

    // Seleciona transações e cria a transação Coinbase
    List<Transaction> blockTxs = List.from(mempool);
    
    // Calcula taxas totais
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

    // Proof of Work: Mine
    String target = '0' * difficulty;
    int hashesAttempted = 0;
    while (!newBlock.hash.startsWith(target)) {
      newBlock.nonce++;
      newBlock.hash = newBlock.calculateHash();
      hashesAttempted++;
      
      if (hashesAttempted % 5000 == 0 && onProgress != null) {
        onProgress(hashesAttempted);
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
      final request = await HttpClient().getUrl(Uri.parse('http://$pcIp:8080/api/ledger'));
      final response = await request.close();
      if (response.statusCode == 200) {
        final content = await response.transform(utf8.decoder).join();
        List<dynamic> jsonBlocks = json.decode(content);
        
        List<Block> newChain = jsonBlocks.map((b) => Block.fromJson(b)).toList();
        
        // Substitui a chain local se a recebida for maior e válida
        if (newChain.length > chain.length) {
          chain = newChain;
          notifyListeners();
          return true;
        }
      }
    } catch (e) {
      print("Erro ao sincronizar Ledger do PC: $e");
    }
    return false;
  }
}
