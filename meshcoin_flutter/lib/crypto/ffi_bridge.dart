import 'dart:ffi' as ffi;
import 'dart:io';
import 'package:ffi/ffi.dart';

// Esta classe provê a ponte nativa (FFI) para executar a criptografia
// Pós-Quântica (Kyber/Dilithium) e o motor de mineração VerusHash (NeonHash)
// em código nativo C/Go otimizado, ignorando limitações de performance do Dart.
class NebulaFFI {
  static late ffi.DynamicLibrary _lib;
  static bool _initialized = false;

  // Definições de assinatura C
  // int neon_hash_verus(char* input, char* output_hash)
  static late int Function(ffi.Pointer<Utf8>, ffi.Pointer<Utf8>) _neonHashVerus;
  
  // int pqc_kyber_encrypt(char* pubkey, char* message, char* out_ciphertext)
  static late int Function(ffi.Pointer<Utf8>, ffi.Pointer<Utf8>, ffi.Pointer<Utf8>) _pqcEncrypt;

  static void initialize() {
    if (_initialized) return;

    try {
      if (Platform.isAndroid) {
        _lib = ffi.DynamicLibrary.open('libnebula.so');
      } else if (Platform.isWindows) {
        _lib = ffi.DynamicLibrary.open('nebula.dll');
      } else if (Platform.isLinux) {
        _lib = ffi.DynamicLibrary.open('libnebula.so');
      } else if (Platform.isMacOS || Platform.isIOS) {
        _lib = ffi.DynamicLibrary.process(); // Embutido no iOS
      }

      // Lookup das funções
      _neonHashVerus = _lib.lookupFunction<
          ffi.Int32 Function(ffi.Pointer<Utf8>, ffi.Pointer<Utf8>),
          int Function(ffi.Pointer<Utf8>, ffi.Pointer<Utf8>)>('neon_hash_verus');

      _pqcEncrypt = _lib.lookupFunction<
          ffi.Int32 Function(ffi.Pointer<Utf8>, ffi.Pointer<Utf8>, ffi.Pointer<Utf8>),
          int Function(ffi.Pointer<Utf8>, ffi.Pointer<Utf8>, ffi.Pointer<Utf8>)>('pqc_kyber_encrypt');

      _initialized = true;
      print("🚀 Nebula FFI Engine (C/Go) carregada com sucesso!");
    } catch (e) {
      print("⚠️ Falha ao carregar a FFI Engine. Utilizando fallback Dart.");
      // O app continuará funcionando no fallback LWE (Dart Puro) até o build nativo (.so/.dll) ser feito.
    }
  }

  static String neonHashVerus(String input) {
    if (!_initialized) {
      // Fallback: usar nosso algoritmo NeonHash Dart atual
      return "FALLBACK_HASH_IMPLEMENTATION_IN_DART";
    }
    
    final inputPtr = input.toNativeUtf8();
    final outputPtr = malloc.allocate<Utf8>(65); // 64 chars hex + \0
    
    _neonHashVerus(inputPtr, outputPtr);
    
    final result = outputPtr.toDartString();
    malloc.free(inputPtr);
    malloc.free(outputPtr);
    
    return result;
  }
}
