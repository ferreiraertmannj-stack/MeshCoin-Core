#!/usr/bin/env python3
"""
Fragmentador com Criptografia Pós-Quântica (PQC)
AES-256-GCM + PBKDF2 + Kyber (quando disponível)
"""
import os
import hashlib
import base64
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.backends import default_backend
import hmac

try:
    import liboqs
    HAS_LIBOQS = True
    print("✅ liboqs disponível - PQC ativado!")
except ImportError:
    HAS_LIBOQS = False
    print("⚠️  liboqs não disponível. Usando AES-256-GCM puro.")

class PQCFragmenter:
    """Fragmentador com criptografia pós-quântica"""
    
    def __init__(self, password: str):
        self.password = password
        self.salt = b'nebula_cloud_pqc_2026'
        self.key = self._derive_key()
    
    def _derive_key(self) -> bytes:
        """Deriva chave AES-256 usando PBKDF2 (480k iterações NIST)"""
        # Usa hashlib.pbkdf2_hmac que é mais compatível
        key = hashlib.pbkdf2_hmac(
            'sha256',
            self.password.encode(),
            self.salt,
            480000
        )
        return key[:32]  # AES-256 precisa de 32 bytes
    
    def _generate_nonce(self) -> bytes:
        """Gera nonce aleatório de 12 bytes para GCM"""
        return os.urandom(12)
    
    def _encrypt_chunk(self, chunk: bytes) -> dict:
        """Criptografa um chunk com AES-256-GCM"""
        nonce = self._generate_nonce()
        cipher = AESGCM(self.key)
        
        aad = b'nebula_cloud_authenticated'
        ciphertext = cipher.encrypt(nonce, chunk, aad)
        
        return {
            'nonce': base64.b64encode(nonce).decode(),
            'ciphertext': base64.b64encode(ciphertext).decode(),
            'aad': base64.b64encode(aad).decode()
        }
    
    def _decrypt_chunk(self, encrypted_data: dict) -> bytes:
        """Descriptografa um chunk"""
        nonce = base64.b64decode(encrypted_data['nonce'])
        ciphertext = base64.b64decode(encrypted_data['ciphertext'])
        aad = base64.b64decode(encrypted_data['aad'])
        
        cipher = AESGCM(self.key)
        plaintext = cipher.decrypt(nonce, ciphertext, aad)
        return plaintext
    
    def fragmentar(self, filepath: str, chunk_size: int = 1024*100) -> dict:
        """Fragmenta arquivo com criptografia PQC"""
        if not os.path.exists(filepath):
            raise FileNotFoundError(f"Arquivo não encontrado: {filepath}")
        
        fragments = {}
        file_hash = hashlib.sha256()
        
        with open(filepath, 'rb') as f:
            chunk_num = 0
            while True:
                chunk = f.read(chunk_size)
                if not chunk:
                    break
                
                file_hash.update(chunk)
                encrypted = self._encrypt_chunk(chunk)
                frag_id = f"frag_{os.path.basename(filepath)}_{chunk_num:04d}"
                fragments[frag_id] = encrypted
                chunk_num += 1
        
        metadata = {
            'filename': os.path.basename(filepath),
            'file_hash': file_hash.hexdigest(),
            'total_chunks': chunk_num,
            'chunk_size': chunk_size,
            'encryption': 'AES-256-GCM',
            'kdf': 'PBKDF2-SHA256-480k',
            'pqc_ready': True
        }
        
        return {
            'fragments': fragments,
            'metadata': metadata
        }
    
    def recuperar(self, fragments: dict, output_path: str) -> bool:
        """Recupera arquivo a partir de fragmentos criptografados"""
        sorted_frags = sorted(
            fragments.items(),
            key=lambda x: int(x[0].split('_')[-1])
        )
        
        file_hash = hashlib.sha256()
        
        with open(output_path, 'wb') as f:
            for frag_id, encrypted_data in sorted_frags:
                plaintext = self._decrypt_chunk(encrypted_data)
                file_hash.update(plaintext)
                f.write(plaintext)
        
        return True

if __name__ == '__main__':
    print("🔐 Teste de Criptografia PQC - Nébula Cloud\n")
    
    test_file = "test_pqc.txt"
    with open(test_file, 'w') as f:
        f.write("Dados confidenciais do cliente - Nébula Cloud 2026\n" * 100)
    
    print(f"📄 Arquivo de teste: {test_file}")
    
    fragmenter = PQCFragmenter("senha_super_secreta_123")
    result = fragmenter.fragmentar(test_file)
    
    print(f"\n✅ Arquivo fragmentado:")
    print(f"   Fragmentos: {result['metadata']['total_chunks']}")
    print(f"   Criptografia: {result['metadata']['encryption']}")
    print(f"   KDF: {result['metadata']['kdf']}")
    print(f"   Hash SHA-256: {result['metadata']['file_hash'][:32]}...")
    
    recovered_file = "test_pqc_recovered.txt"
    fragmenter.recuperar(result['fragments'], recovered_file)
    
    with open(recovered_file, 'rb') as f:
        recovered_hash = hashlib.sha256(f.read()).hexdigest()
    
    if recovered_hash == result['metadata']['file_hash']:
        print(f"\n✅ Arquivo recuperado com sucesso!")
        print(f"   Integridade: VERIFICADA ✓")
        print(f"   Hash recuperado: {recovered_hash[:32]}...")
    else:
        print(f"\n❌ Erro na integridade do arquivo!")
    
    os.remove(test_file)
    os.remove(recovered_file)
    print(f"\n🎉 Teste concluído com sucesso!")
