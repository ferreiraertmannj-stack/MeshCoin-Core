import 'dart:convert';
import 'package:crypto/crypto.dart';
import '../crypto/wallet_crypto.dart';

class Transaction {
  final String id;
  final String senderPubKey;
  final String senderAddress;
  final String receiverAddress;
  final double amount;
  final double fee;
  final int timestamp;
  final String signature;

  Transaction({
    required this.id,
    required this.senderPubKey,
    required this.senderAddress,
    required this.receiverAddress,
    required this.amount,
    required this.fee,
    required this.timestamp,
    required this.signature,
  });

  /// Serializa os dados centrais da transação para geração de Hash e Assinatura
  String get txData => '$senderPubKey:$senderAddress:$receiverAddress:$amount:$fee:$timestamp';

  /// Verifica se a transação possui assinaturas e hashes válidos
  bool isValid() {
    // 1. Transação de Base (Recompensa de Mineração) não tem remetente
    if (senderAddress == 'COINBASE') {
      return true; // Blocos recém minerados validam a coinbase separadamente
    }

    // 2. Verificar se o ID bate com os dados reais
    var computedHash = sha256.convert(utf8.encode(txData)).toString();
    if (computedHash != id) return false;

    // 3. Verificar assinatura criptográfica real (ECDSA secp256k1)
    if (!WalletCrypto.verifySignature(senderPubKey, txData, signature)) {
      return false;
    }

    // 4. Regras de negócio
    if (amount <= 0 || fee < 0) return false;

    return true;
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'senderPubKey': senderPubKey,
      'senderAddress': senderAddress,
      'receiverAddress': receiverAddress,
      'amount': amount,
      'fee': fee,
      'timestamp': timestamp,
      'signature': signature,
    };
  }

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'],
      senderPubKey: json['senderPubKey'],
      senderAddress: json['senderAddress'],
      receiverAddress: json['receiverAddress'],
      amount: (json['amount'] is int) ? (json['amount'] as int).toDouble() : json['amount'],
      fee: (json['fee'] is int) ? (json['fee'] as int).toDouble() : json['fee'],
      timestamp: json['timestamp'],
      signature: json['signature'],
    );
  }

  /// Constrói uma nova transação assinada localmente
  static Transaction create({
    required String senderPubKey,
    required String senderAddress,
    required String privateKeyHex,
    required String receiverAddress,
    required double amount,
    double fee = 0.0,
  }) {
    int timestamp = DateTime.now().millisecondsSinceEpoch;
    String txData = '$senderPubKey:$senderAddress:$receiverAddress:$amount:$fee:$timestamp';
    
    String id = sha256.convert(utf8.encode(txData)).toString();
    String signature = WalletCrypto.signTransaction(privateKeyHex, txData);

    return Transaction(
      id: id,
      senderPubKey: senderPubKey,
      senderAddress: senderAddress,
      receiverAddress: receiverAddress,
      amount: amount,
      fee: fee,
      timestamp: timestamp,
      signature: signature,
    );
  }

  /// Constrói uma Coinbase Transaction (Recompensa de Bloco)
  static Transaction createCoinbase(String minerAddress, double reward) {
    int timestamp = DateTime.now().millisecondsSinceEpoch;
    String txData = 'COINBASE:COINBASE:$minerAddress:$reward:0.0:$timestamp';
    String id = sha256.convert(utf8.encode(txData)).toString();

    return Transaction(
      id: id,
      senderPubKey: 'COINBASE',
      senderAddress: 'COINBASE',
      receiverAddress: minerAddress,
      amount: reward,
      fee: 0.0,
      timestamp: timestamp,
      signature: 'COINBASE_SIG',
    );
  }
}
