import 'dart:convert';
import 'dart:isolate' as isolate;
import 'dart:math';
import 'dart:typed_data';
import 'package:crypto/crypto.dart';
import 'package:pointycastle/export.dart' as pc;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:bip39/bip39.dart' as bip39;
import 'package:pointycastle/random/fortuna_random.dart';

/// ═══════════════════════════════════════════════════════════════
/// Nebula Wallet Crypto — BIP39 + Híbrida (ECDSA secp256k1 + PQC-Ready)
/// ═══════════════════════════════════════════════════════════════

class WalletCrypto {
  static final pc.ECDomainParameters _params = pc.ECDomainParameters('secp256k1');

  /// Gera 24 palavras Mnemônicas (BIP39) usando 256 bits de entropia
  static String generateMnemonic() {
    return bip39.generateMnemonic(strength: 256); // 24 words
  }

  /// Valida as 24 palavras
  static bool validateMnemonic(String mnemonic) {
    return bip39.validateMnemonic(mnemonic);
  }

  /// Gera a chave Híbrida em background (Isolate) para evitar travamento da UI
  static Future<Map<String, String>> generateKeypairFromMnemonicAsync(String mnemonic) async {
    return await isolate.Isolate.run(() => generateKeypairFromMnemonic(mnemonic));
  }

  /// Gera a chave Híbrida (ECDSA + Preparado para PQC) a partir do Mnemonic (Síncrono)
  static Map<String, String> generateKeypairFromMnemonic(String mnemonic) {
    if (!validateMnemonic(mnemonic)) {
      throw Exception('Frase de recuperação inválida.');
    }

    // O Seed BIP39 possui 64 bytes (512 bits)
    Uint8List seed = bip39.mnemonicToSeed(mnemonic);

    // Para ECDSA secp256k1, usamos o SHA-256 da Seed (32 bytes)
    var ecdsaPrivateBytes = sha256.convert(seed).bytes;
    String privateHex = _bytesToHex(Uint8List.fromList(ecdsaPrivateBytes));

    // Derivar chave pública ECDSA
    BigInt privateKeyInt = BigInt.parse(privateHex, radix: 16);
    pc.ECPrivateKey privateKey = pc.ECPrivateKey(privateKeyInt, _params);
    pc.ECPoint? Q = _params.G * privateKey.d;
    if (Q == null) throw Exception('Falha ao derivar chave pública.');
    pc.ECPublicKey publicKey = pc.ECPublicKey(Q, _params);

    String publicHex = _compressPublicKey(publicKey.Q!);

    // O Seed será usado no futuro para gerar a chave Dilithium (PQC) diretamente:
    // Exemplo estrutural (não-funcional na v1):
    // String pqcPublicHex = PQC.generateDilithiumKey(seed.sublist(32));

    String address = _generateAddress(publicHex);

    return {
      'mnemonic': mnemonic,
      'privateKey': privateHex,
      'publicKey': publicHex,
      'address': address,
    };
  }

  /// Gera o endereço MeshCoin a partir da chave pública
  static String _generateAddress(String publicKeyHex) {
    Uint8List pubKeyBytes = _hexToBytes(publicKeyHex);

    // SHA-256
    var sha256Hash = sha256.convert(pubKeyBytes).bytes;

    // RIPEMD-160
    final ripemd160 = pc.Digest('RIPEMD-160');
    Uint8List ripemdHash = ripemd160.process(Uint8List.fromList(sha256Hash));

    // Adiciona version byte (0x4E = 'N' para Nebula)
    Uint8List versionedPayload = Uint8List(1 + ripemdHash.length);
    versionedPayload[0] = 0x4E; // Version byte Nebula
    versionedPayload.setRange(1, versionedPayload.length, ripemdHash);

    // Checksum: primeiros 4 bytes de SHA256(SHA256(versionedPayload))
    var firstHash = sha256.convert(versionedPayload).bytes;
    var secondHash = sha256.convert(firstHash).bytes;
    Uint8List checksum = Uint8List.fromList(secondHash.sublist(0, 4));

    // Payload final
    Uint8List addressBytes = Uint8List(versionedPayload.length + checksum.length);
    addressBytes.setRange(0, versionedPayload.length, versionedPayload);
    addressBytes.setRange(versionedPayload.length, addressBytes.length, checksum);

    // Base58 encode
    String base58Address = _base58Encode(addressBytes);

    return 'NBL$base58Address';
  }

  /// Comprime a chave pública EC (formato 33 bytes)
  static String _compressPublicKey(pc.ECPoint point) {
    BigInt x = point.x!.toBigInteger()!;
    BigInt y = point.y!.toBigInteger()!;
    int prefix = y.isEven ? 0x02 : 0x03;
    String xHex = x.toRadixString(16).padLeft(64, '0');
    return prefix.toRadixString(16).padLeft(2, '0') + xHex;
  }

  /// Assina uma transação com a chave privada ECDSA
  static String signTransaction(String privateKeyHex, String txData) {
    BigInt privateKeyInt = BigInt.parse(privateKeyHex, radix: 16);
    pc.ECPrivateKey privateKey = pc.ECPrivateKey(privateKeyInt, _params);

    // Hash da transação
    var txHash = sha256.convert(utf8.encode(txData)).bytes;

    // Assinar com ECDSA usando pre-computed hash (digest null)
    final signer = pc.ECDSASigner(null);
    
    // Configurar SecureRandom para evitar RegistryFactoryException
    final secureRandom = pc.FortunaRandom();
    final seedSource = Random.secure();
    final seeds = <int>[];
    for (int i = 0; i < 32; i++) {
      seeds.add(seedSource.nextInt(256));
    }
    secureRandom.seed(pc.KeyParameter(Uint8List.fromList(seeds)));

    signer.init(true, pc.ParametersWithRandom(pc.PrivateKeyParameter<pc.ECPrivateKey>(privateKey), secureRandom));

    pc.ECSignature signature = signer.generateSignature(Uint8List.fromList(txHash)) as pc.ECSignature;

    // Forçar Low-S para compatibilidade com o secp256k1 do Go Node
    BigInt s = signature.s;
    BigInt halfN = _params.n >> 1;
    if (s.compareTo(halfN) > 0) {
      s = _params.n - s;
    }

    String rStr = signature.r.toRadixString(16).padLeft(64, '0');
    String sStr = s.toRadixString(16).padLeft(64, '0');

    return '$rStr$sStr';
  }

  /// Verifica uma assinatura ECDSA
  static bool verifySignature(String publicKeyHex, String txData, String signatureHex) {
    try {
      Uint8List pubKeyBytes = _hexToBytes(publicKeyHex);
      pc.ECPoint? point = _params.curve.decodePoint(pubKeyBytes);
      if (point == null) return false;

      pc.ECPublicKey publicKey = pc.ECPublicKey(point, _params);
      var txHash = sha256.convert(utf8.encode(txData)).bytes;

      String rHex = signatureHex.substring(0, 64);
      String sHex = signatureHex.substring(64);
      pc.ECSignature sig = pc.ECSignature(
        BigInt.parse(rHex, radix: 16),
        BigInt.parse(sHex, radix: 16),
      );

      final verifier = pc.ECDSASigner(null);
      verifier.init(false, pc.PublicKeyParameter<pc.ECPublicKey>(publicKey));

      return verifier.verifySignature(Uint8List.fromList(txHash), sig);
    } catch (e) {
      return false;
    }
  }

  // ─── Persistência Local ───

  static Future<void> saveWallet(Map<String, String> wallet) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('nebula_privateKey', wallet['privateKey']!);
    await prefs.setString('nebula_publicKey', wallet['publicKey']!);
    await prefs.setString('nebula_address', wallet['address']!);
    if (wallet.containsKey('mnemonic')) {
      await prefs.setString('nebula_mnemonic', wallet['mnemonic']!);
    }
  }

  static Future<Map<String, String>?> loadWallet() async {
    final prefs = await SharedPreferences.getInstance();
    String? privateKey = prefs.getString('nebula_privateKey');
    String? publicKey = prefs.getString('nebula_publicKey');
    String? address = prefs.getString('nebula_address');

    if (privateKey != null && publicKey != null && address != null) {
      return {
        'privateKey': privateKey,
        'publicKey': publicKey,
        'address': address,
      };
    }
    return null;
  }

  // ─── Utilitários ───

  static pc.SecureRandom _getSecureRandom() {
    final secureRandom = pc.FortunaRandom();
    final random = Random.secure();
    final seeds = List<int>.generate(32, (_) => random.nextInt(256));
    secureRandom.seed(pc.KeyParameter(Uint8List.fromList(seeds)));
    return secureRandom;
  }

  static Uint8List _hexToBytes(String hex) {
    List<int> bytes = [];
    for (int i = 0; i < hex.length; i += 2) {
      bytes.add(int.parse(hex.substring(i, i + 2), radix: 16));
    }
    return Uint8List.fromList(bytes);
  }

  static String _bytesToHex(Uint8List bytes) {
    return bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
  }

  static const String _base58Alphabet =
      '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';

  static String _base58Encode(Uint8List bytes) {
    BigInt intData = BigInt.zero;
    for (int byte in bytes) {
      intData = (intData << 8) | BigInt.from(byte);
    }

    String result = '';
    while (intData > BigInt.zero) {
      BigInt remainder = intData % BigInt.from(58);
      intData = intData ~/ BigInt.from(58);
      result = _base58Alphabet[remainder.toInt()] + result;
    }

    // Adiciona '1' para cada byte zero no início
    for (int byte in bytes) {
      if (byte == 0) {
        result = '1$result';
      } else {
        break;
      }
    }

    return result;
  }
}
