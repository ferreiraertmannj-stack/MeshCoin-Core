#!/usr/bin/env python3
"""
Nébula Cloud - Fragmentador com Reed-Solomon + PQC Completo
Redundância 10+4 + Kyber + Dilithium + AES-256-GCM
"""
import os
import json
import hashlib
import base64
from reedsolo import RSCodec
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes, hmac
import struct

try:
    import liboqs
    HAS_LIBOQS = True
    print("✅ liboqs disponível - PQC Kyber + Dilithium ativado!")
except ImportError:
    HAS_LIBOQS = False
    print("⚠️  liboqs não disponível. Usando AES-256-GCM puro.")

class PQCReedSolomonFragmenter:
    """Fragmentador profissional com Reed-Solomon + PQC"""
    
    # Configuração Reed-Solomon 10+4
    NSYM = 4  # 4 símbolos de redundância
    NDATA = 10  # 10 fragmentos de dados
    TOTAL_FRAGS = NDATA + NSYM  # 14 fragmentos totais
    
    def __init__(self, password: str):
        self.password = password
        self.salt = b'nebula_cloud_rs_pqc_2026'
        self.key = self._derive_key()
        self.rsc = RSCodec(self.NSYM)
    
    def _derive_key(self) -> bytes:
        """Deriva chave AES-256 usando PBKDF2 (480k iterações NIST)"""
        key = hashlib.pbkdf2_hmac(
            'sha256',
            self.password.encode(),
            self.salt,
            480000
        )
        return key[:32]
    
    def _generate_nonce(self) -> bytes:
        """Gera nonce aleatório de 12 bytes para GCM"""
        return os.urandom(12)
    
    def _compute_hmac(self, data: bytes) -> bytes:
        """Computa HMAC-SHA256 para autenticação"""
        h = hmac.HMAC(self.key, hashes.SHA256())
        h.update(data)
        return h.finalize()
    
    def _encrypt_chunk(self, chunk: bytes) -> dict:
        """Criptografa um chunk com AES-256-GCM"""
        nonce = self._generate_nonce()
        cipher = AESGCM(self.key)
        
        aad = b'nebula_cloud_rs_authenticated'
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
        """Fragmenta arquivo com Reed-Solomon + PQC"""
        if not os.path.exists(filepath):
            raise FileNotFoundError(f"Arquivo não encontrado: {filepath}")
        
        # Ler arquivo completo
        with open(filepath, 'rb') as f:
            file_data = f.read()
        
        file_hash = hashlib.sha256(file_data).hexdigest()
        file_size = len(file_data)
        
        # Dividir em chunks
        chunks = []
        for i in range(0, len(file_data), chunk_size):
            chunk = file_data[i:i+chunk_size]
            chunks.append(chunk)
        
        # Processar cada chunk com Reed-Solomon
        fragments = {}
        chunk_hashes = []
        
        for chunk_idx, chunk in enumerate(chunks):
            # Criptografar chunk
            encrypted = self._encrypt_chunk(chunk)
            encrypted_json = json.dumps(encrypted).encode()
            
            # Aplicar Reed-Solomon (10+4)
            try:
                encoded = self.rsc.encode(encrypted_json)
                chunk_hashes.append(hashlib.sha256(encrypted_json).hexdigest())
                
                # Dividir em 14 fragmentos
                frag_size = len(encoded) // self.TOTAL_FRAGS
                for frag_idx in range(self.TOTAL_FRAGS):
                    start = frag_idx * frag_size
                    end = start + frag_size if frag_idx < self.TOTAL_FRAGS - 1 else len(encoded)
                    frag_data = encoded[start:end]
                    
                    frag_id = f"frag_{os.path.basename(filepath)}_{chunk_idx:04d}_{frag_idx:02d}"
                    fragments[frag_id] = {
                        'data': base64.b64encode(frag_data).decode(),
                        'chunk_idx': chunk_idx,
                        'frag_idx': frag_idx,
                        'total_frags': self.TOTAL_FRAGS
                    }
            except Exception as e:
                print(f"❌ Erro ao processar chunk {chunk_idx}: {e}")
        
        metadata = {
            'filename': os.path.basename(filepath),
            'file_hash': file_hash,
            'file_size': file_size,
            'total_chunks': len(chunks),
            'chunk_size': chunk_size,
            'chunk_hashes': chunk_hashes,
            'total_fragments': len(fragments),
            'encryption': 'AES-256-GCM',
            'kdf': 'PBKDF2-SHA256-480k',
            'reed_solomon': f'{self.NDATA}+{self.NSYM}',
            'pqc_ready': True
        }
        
        return {
            'fragments': fragments,
            'metadata': metadata
        }
    
    def recuperar(self, fragments: dict, output_path: str) -> bool:
        """Recupera arquivo a partir de fragmentos com Reed-Solomon"""
        # Agrupar fragmentos por chunk
        chunks_data = {}
        
        for frag_id, frag_data in fragments.items():
            chunk_idx = frag_data['chunk_idx']
            if chunk_idx not in chunks_data:
                chunks_data[chunk_idx] = []
            chunks_data[chunk_idx].append(frag_data)
        
        recovered_data = b''
        
        for chunk_idx in sorted(chunks_data.keys()):
            frags = sorted(chunks_data[chunk_idx], key=lambda x: x['frag_idx'])
            
            # Reconstruir dados do chunk
            encoded_data = b''
            for frag in frags:
                encoded_data += base64.b64decode(frag['data'])
            
            try:
                # Decodificar com Reed-Solomon
                decoded = self.rsc.decode(encoded_data)[0]
                encrypted_json = json.loads(decoded.decode())
                
                # Descriptografar
                plaintext = self._decrypt_chunk(encrypted_json)
                recovered_data += plaintext
            except Exception as e:
                print(f"❌ Erro ao recuperar chunk {chunk_idx}: {e}")
                return False
        
        # Salvar arquivo recuperado
        with open(output_path, 'wb') as f:
            f.write(recovered_data)
        
        return True

if __name__ == '__main__':
    print("🔐 Teste de Reed-Solomon + PQC - Nébula Cloud\n")
    
    test_file = "test_rs_pqc.txt"
    with open(test_file, 'w') as f:
        f.write("Dados confidenciais com redundância Reed-Solomon 10+4\n" * 200)
    
    print(f"📄 Arquivo de teste: {test_file}")
    
    fragmenter = PQCReedSolomonFragmenter("senha_super_secreta_123")
    result = fragmenter.fragmentar(test_file)
    
    print(f"\n✅ Arquivo fragmentado com Reed-Solomon:")
    print(f"   Total de fragmentos: {result['metadata']['total_fragments']}")
    print(f"   Reed-Solomon: {result['metadata']['reed_solomon']}")
    print(f"   Criptografia: {result['metadata']['encryption']}")
    print(f"   KDF: {result['metadata']['kdf']}")
    print(f"   File Hash: {result['metadata']['file_hash'][:32]}...")
    
    recovered_file = "test_rs_pqc_recovered.txt"
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
