import hashlib
import os
from ecdsa import SigningKey, SECP256k1
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

class HybridCrypto:
    """
    Core Criptográfico Híbrido:
    Combina ECDSA (secp256k1) tradicional com chaves Pós-Quânticas (Dilithium / Kyber).
    (Nota: A parte PQC está sendo mockada neste protótipo em Python puro, mas
    a arquitetura permite injeção via C/JNI no app final).
    """

    @staticmethod
    def generate_keypair():
        """Gera um par de chaves Híbrido (ECDSA + PQC)"""
        # 1. Chave Tradicional
        sk_ecdsa = SigningKey.generate(curve=SECP256k1)
        vk_ecdsa = sk_ecdsa.get_verifying_key()
        
        # 2. Chave Pós-Quântica (Mock - na vida real usaríamos liboqs/Kyber)
        sk_pqc = os.urandom(32) # Seed privada
        vk_pqc = hashlib.sha256(sk_pqc).digest() # Chave Pública Kyber (Mock)
        
        # O Endereço Mesh é o hash das duas chaves combinadas
        address_raw = vk_ecdsa.to_string() + vk_pqc
        address = "MESH" + hashlib.sha256(address_raw).hexdigest()[:36]
        
        return {
            "address": address,
            "private_ecdsa": sk_ecdsa.to_pem().decode('utf-8'),
            "public_ecdsa": vk_ecdsa.to_pem().decode('utf-8'),
            "private_pqc_seed": sk_pqc.hex(),
            "public_pqc": vk_pqc.hex()
        }
        
    @staticmethod
    def sign_transaction(private_ecdsa_pem, private_pqc_hex, tx_data):
        """Assina dados usando as DUAS chaves (Híbrido)"""
        # 1. Assinatura ECDSA
        sk_ecdsa = SigningKey.from_pem(private_ecdsa_pem)
        sig_ecdsa = sk_ecdsa.sign(tx_data.encode('utf-8')).hex()
        
        # 2. Assinatura PQC (Dilithium Mock)
        # Mock de assinatura: HMAC-like usando o PQC seed
        sig_pqc = hashlib.sha512(bytes.fromhex(private_pqc_hex) + tx_data.encode('utf-8')).hexdigest()
        
        return {
            "ecdsa": sig_ecdsa,
            "pqc": sig_pqc
        }

    @staticmethod
    def encrypt_chat_message(shared_secret, message_text):
        """Usa AES-256-GCM para criptografia de ponta a ponta (E2EE) nas mensagens"""
        # O shared_secret viria de um Key Exchange PQC (Kyber)
        aesgcm = AESGCM(shared_secret[:32])
        nonce = os.urandom(12)
        ciphertext = aesgcm.encrypt(nonce, message_text.encode('utf-8'), None)
        return nonce.hex() + ciphertext.hex()

    @staticmethod
    def decrypt_chat_message(shared_secret, encrypted_hex):
        aesgcm = AESGCM(shared_secret[:32])
        encrypted_bytes = bytes.fromhex(encrypted_hex)
        nonce = encrypted_bytes[:12]
        ciphertext = encrypted_bytes[12:]
        return aesgcm.decrypt(nonce, ciphertext, None).decode('utf-8')

if __name__ == "__main__":
    print("--- Teste de Criptografia Híbrida MeshCoin ---")
    carteira = HybridCrypto.generate_keypair()
    print(f"🌍 Endereço Híbrido: {carteira['address']}")
    
    tx = "Transferência de 10 MESH para Joao"
    assinaturas = HybridCrypto.sign_transaction(carteira['private_ecdsa'], carteira['private_pqc_seed'], tx)
    print(f"✍️ Assinatura ECDSA: {assinaturas['ecdsa'][:40]}...")
    print(f"✍️ Assinatura PQC (Dilithium Mock): {assinaturas['pqc'][:40]}...")
    
    # Teste de chat E2EE
    segredo_compartilhado = os.urandom(32) # Kyber KEM result
    cifrado = HybridCrypto.encrypt_chat_message(segredo_compartilhado, "Olá P2P Off-Grid!")
    print(f"🔒 Chat Cifrado (AES-GCM): {cifrado[:50]}...")
    decifrado = HybridCrypto.decrypt_chat_message(segredo_compartilhado, cifrado)
    print(f"🔓 Chat Decifrado: {decifrado}")
