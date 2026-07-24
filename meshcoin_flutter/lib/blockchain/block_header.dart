import 'dart:convert';
import 'package:crypto/crypto.dart';

/// SPV (Simplified Payment Verification) / Light Node Block Header
/// 
/// Usado por celulares para economizar RAM e armazenamento.
/// Eles não guardam o histórico completo das transações, apenas o cabeçalho
/// que serve como prova matemática de que o bloco é autêntico.
class BlockHeader {
  final int index;
  final int timestamp;
  final String previousHash;
  final String merkleRoot;
  final int nonce;
  final String hash;

  BlockHeader({
    required this.index,
    required this.timestamp,
    required this.previousHash,
    required this.merkleRoot,
    required this.nonce,
    required this.hash,
  });

  Map<String, dynamic> toJson() {
    return {
      'index': index,
      'timestamp': timestamp,
      'previousHash': previousHash,
      'merkleRoot': merkleRoot,
      'nonce': nonce,
      'hash': hash,
    };
  }

  factory BlockHeader.fromJson(Map<String, dynamic> json) {
    return BlockHeader(
      index: json['index'],
      timestamp: json['timestamp'],
      previousHash: json['previousHash'],
      merkleRoot: json['merkleRoot'],
      nonce: json['nonce'],
      hash: json['hash'],
    );
  }

  /// Verifica o Proof of Work deste header
  bool isValid(int difficulty) {
    String blockData = '$index$timestamp$previousHash$merkleRoot$nonce';
    var bytes = utf8.encode(blockData);
    var computedHash = sha256.convert(bytes).toString();

    if (computedHash != hash) return false;
    
    String target = '0' * difficulty;
    return hash.startsWith(target);
  }
}
