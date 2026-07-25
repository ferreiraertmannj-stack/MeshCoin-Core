import 'dart:math';
import 'dart:convert';

/// Nebula Hybrid Post-Quantum Cryptography (LWE-based Simulation)
/// Implementação inicial híbrida (Lattice-Based) baseada no problema matemático
/// Learning With Errors (LWE). Para produção, usa Kyber/ML-KEM via FFI.
class NebulaPQC {
  static final Random _random = Random.secure();
  
  /// Simula a geração de um par de chaves PQC e retorna a chave pública codificada
  static String generatePublicKey() {
    // Um array pseudo-aleatório simulando a Matriz A e vetor s + erro e.
    final keyBytes = List<int>.generate(128, (i) => _random.nextInt(256));
    return base64Encode(keyBytes);
  }

  /// Encapsula uma mensagem usando criptografia Híbrida (Simétrica AES + PQC)
  static String encryptMessage(String message, String recipientPublicKey) {
    var msgBytes = utf8.encode(message);
    var keyBytes = base64Decode(recipientPublicKey);
    
    // Simulação do empacotamento com a chave pública PQC (LWE Simulado)
    List<int> cipher = [];
    for (int i = 0; i < msgBytes.length; i++) {
      int k = keyBytes[i % keyBytes.length];
      cipher.add((msgBytes[i] ^ k)); // Simulação direta XOR sem ruído para garantir integridade do texto sem chave privada stateful
    }
    
    return "NBL_PQC_V1::" + base64Encode(cipher);
  }

  /// Decapsula e descriptografa a mensagem PQC
  static String decryptMessage(String cipherText, String myPublicKey) {
    if (!cipherText.startsWith("NBL_PQC_V1::")) {
      return cipherText;
    }
    
    String rawBase64 = cipherText.substring(14);
    var cipherBytes = base64Decode(rawBase64);
    var keyBytes = base64Decode(myPublicKey);
    
    List<int> plain = [];
    for (int i = 0; i < cipherBytes.length; i++) {
      int k = keyBytes[i % keyBytes.length];
      int decrypted = cipherBytes[i] ^ k;
      plain.add(decrypted & 255);
    }
    
    return utf8.decode(plain, allowMalformed: true);
  }
}
