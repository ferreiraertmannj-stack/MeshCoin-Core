import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'transaction.dart';

class Block {
  final int index;
  final int timestamp;
  final String previousHash;
  final String merkleRoot;
  final List<Transaction> transactions;
  int nonce;
  String hash;

  Block({
    required this.index,
    required this.timestamp,
    required this.previousHash,
    required this.merkleRoot,
    required this.transactions,
    this.nonce = 0,
    this.hash = '',
  }) {
    if (hash.isEmpty) {
      hash = calculateHash();
    }
  }

  /// Calcula o Hash do bloco (Algoritmo NeonHash simplificado em Dart)
  String calculateHash() {
    String blockData;
    int mStorage = 0;
    String sType = '';
    
    try {
      // Usaremos o toJson para ver se essas chaves foram injetadas pós-criação ou no Node PC
      var json = this.toJson();
      if (json.containsKey('minerStorage')) mStorage = json['minerStorage'] ?? 0;
      if (json.containsKey('storageType')) sType = json['storageType'] ?? '';
    } catch(e) {}
    
    if (mStorage > 0 || (sType.isNotEmpty && sType != 'Mobile')) {
      blockData = '$index$timestamp$previousHash$merkleRoot$nonce$mStorage$sType';
    } else {
      blockData = '$index$timestamp$previousHash$merkleRoot$nonce';
    }
    
    var bytes = utf8.encode(blockData);
    var digest = sha256.convert(bytes);
    return digest.toString();
  }

  /// Gera a Árvore de Merkle das transações do bloco
  static String calculateMerkleRoot(List<Transaction> txs) {
    if (txs.isEmpty) return '';
    List<String> tree = txs.map((tx) => tx.id).toList();

    while (tree.length > 1) {
      List<String> nextLevel = [];
      for (int i = 0; i < tree.length; i += 2) {
        String left = tree[i];
        String right = (i + 1 < tree.length) ? tree[i + 1] : left; // Duplicate last if odd
        String combined = sha256.convert(utf8.encode(left + right)).toString();
        nextLevel.add(combined);
      }
      tree = nextLevel;
    }
    return tree.first;
  }

  /// Valida o bloco
  bool isValid(int difficulty) {
    if (hash != calculateHash()) return false;
    if (merkleRoot != calculateMerkleRoot(transactions)) return false;
    
    // Verifica a dificuldade (Proof of Work)
    String target = '0' * difficulty;
    if (!hash.startsWith(target)) return false;

    // Valida cada transação
    for (var tx in transactions) {
      if (!tx.isValid()) return false;
    }

    return true;
  }

  Map<String, dynamic> toJson() {
    return {
      'index': index,
      'timestamp': timestamp,
      'previousHash': previousHash,
      'merkleRoot': merkleRoot,
      'nonce': nonce,
      'hash': hash,
      'minerStorage': 0, // Smartphone padrão é 0
      'storageType': 'Mobile',
      'transactions': transactions.map((t) => t.toJson()).toList(),
    };
  }

  factory Block.fromJson(Map<String, dynamic> json) {
    var txList = json['transactions'] as List;
    List<Transaction> transactions = txList.map((t) => Transaction.fromJson(t)).toList();

    return Block(
      index: json['index'],
      timestamp: json['timestamp'],
      previousHash: json['previousHash'],
      merkleRoot: json['merkleRoot'],
      transactions: transactions,
      nonce: json['nonce'],
      hash: json['hash'],
    );
  }

  /// Gera o Bloco Gênesis (Início da rede)
  static Block genesis() {
    return Block(
      index: 0,
      timestamp: 1672531200000, // Exemplo fixo
      previousHash: '0' * 64,
      merkleRoot: '',
      transactions: [],
      nonce: 0,
      hash: '000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f',
    );
  }
}
