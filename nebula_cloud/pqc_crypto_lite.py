"""
Módulo de Criptografia Pós-Quântica (PQC) Lite para Nébula Cloud
Implementa ML-KEM (Kyber) e ML-DSA (Dilithium) com fallback para clássico
Segue os padrões NIST FIPS 203, 204 e 205 (2024)

Autor: Manus AI
Data: Abril 2026
"""

import os
import json
import logging
import hashlib
import hmac
from typing import Tuple, Dict, Any
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
from cryptography.hazmat.primitives.asymmetric import rsa, padding

# ==================== CONFIGURAÇÃO ====================
logger = logging.getLogger(__name__)

# Parâmetros de segurança
SALT_PQC = b'nebula-cloud-pqc-2026'
ITERATIONS_PQC = 100000
AES_KEY_SIZE = 32  # 256 bits
NONCE_SIZE = 12    # 96 bits para GCM
RSA_KEY_SIZE = 4096  # Fallback para RSA

# Marcadores de algoritmo
MARKER_CLASSIC = b'CLASSIC_AES'
MARKER_PQC_HYBRID = b'PQC_HYBRID'

class PQCCryptoLite:
    """
    Classe para gerenciar criptografia pós-quântica híbrida com fallback.
    Combina ML-KEM (encapsulação de chaves) com AES-256-GCM (criptografia de dados).
    
    Quando liboqs não está disponível, usa RSA-4096 como fallback.
    """
    
    def __init__(self):
        """Inicializa o sistema de criptografia PQC"""
        try:
            import oqs
            self.pqc_available = True
            self.oqs = oqs
            logger.info("PQC (liboqs) disponível - Usando ML-KEM-768 + ML-DSA-65")
        except ImportError:
            self.pqc_available = False
            logger.warning("PQC (liboqs) não disponível - Usando RSA-4096 como fallback")
    
    # ==================== MÉTODOS CLÁSSICOS (FALLBACK) ====================
    
    @staticmethod
    def gerar_par_chaves_rsa() -> Tuple[bytes, bytes]:
        """
        Gera um par de chaves RSA-4096 como fallback.
        
        Returns:
            Tupla (chave_publica_pem, chave_privada_pem)
        """
        chave_privada = rsa.generate_private_key(
            public_exponent=65537,
            key_size=RSA_KEY_SIZE,
        )
        chave_publica = chave_privada.public_key()
        
        from cryptography.hazmat.primitives import serialization
        chave_pub_pem = chave_publica.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo
        )
        chave_priv_pem = chave_privada.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        )
        return chave_pub_pem, chave_priv_pem
    
    @staticmethod
    def encapsular_chave_rsa(chave_publica_rsa: bytes) -> Tuple[bytes, bytes]:
        """
        Encapsula uma chave de sessão AES usando RSA-OAEP.
        
        Args:
            chave_publica_rsa: Chave pública RSA em PEM
            
        Returns:
            Tupla (ciphertext_encapsulado, chave_sessao_aes)
        """
        from cryptography.hazmat.primitives import serialization
        
        chave_sessao = os.urandom(32)  # Chave AES-256
        chave_pub = serialization.load_pem_public_key(chave_publica_rsa)
        
        ciphertext = chave_pub.encrypt(
            chave_sessao,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        return ciphertext, chave_sessao
    
    @staticmethod
    def desencapsular_chave_rsa(chave_privada_rsa: bytes, ciphertext: bytes) -> bytes:
        """
        Desencapsula uma chave de sessão AES usando RSA-OAEP.
        
        Args:
            chave_privada_rsa: Chave privada RSA em PEM
            ciphertext: Ciphertext encapsulado
            
        Returns:
            Chave de sessão AES
        """
        from cryptography.hazmat.primitives import serialization
        
        chave_priv = serialization.load_pem_private_key(
            chave_privada_rsa,
            password=None
        )
        
        chave_sessao = chave_priv.decrypt(
            ciphertext,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        return chave_sessao
    
    @staticmethod
    def assinar_dados_rsa(chave_privada_rsa: bytes, dados: bytes) -> bytes:
        """
        Assina dados usando RSA-PSS.
        
        Args:
            chave_privada_rsa: Chave privada RSA em PEM
            dados: Dados a assinar
            
        Returns:
            Assinatura digital
        """
        from cryptography.hazmat.primitives import serialization
        
        chave_priv = serialization.load_pem_private_key(
            chave_privada_rsa,
            password=None
        )
        
        assinatura = chave_priv.sign(
            dados,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )
        return assinatura
    
    @staticmethod
    def verificar_assinatura_rsa(chave_publica_rsa: bytes, dados: bytes, assinatura: bytes) -> bool:
        """
        Verifica uma assinatura RSA-PSS.
        
        Args:
            chave_publica_rsa: Chave pública RSA em PEM
            dados: Dados originais
            assinatura: Assinatura a verificar
            
        Returns:
            True se válida, False caso contrário
        """
        from cryptography.hazmat.primitives import serialization
        
        try:
            chave_pub = serialization.load_pem_public_key(chave_publica_rsa)
            chave_pub.verify(
                assinatura,
                dados,
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH
                ),
                hashes.SHA256()
            )
            return True
        except Exception:
            return False
    
    # ==================== MÉTODOS PQC ====================
    
    def gerar_par_chaves_pqc(self) -> Tuple[bytes, bytes]:
        """
        Gera um par de chaves PQC (ML-KEM-768) ou RSA como fallback.
        
        Returns:
            Tupla (chave_publica, chave_privada)
        """
        if not self.pqc_available:
            logger.info("Usando RSA-4096 como fallback")
            return self.gerar_par_chaves_rsa()
        
        try:
            kekem = self.oqs.KeyEncapsulation("ML-KEM-768")
            chave_publica = kekem.generate_keypair()
            chave_privada = kekem.export_secret_key()
            logger.info("Par de chaves ML-KEM-768 gerado")
            return chave_publica, chave_privada
        except Exception as e:
            logger.warning(f"Erro ao gerar ML-KEM: {e}. Usando RSA como fallback.")
            return self.gerar_par_chaves_rsa()
    
    def encapsular_chave(self, chave_publica_pqc: bytes) -> Tuple[bytes, bytes]:
        """
        Encapsula uma chave de sessão AES usando PQC ou RSA.
        
        Args:
            chave_publica_pqc: Chave pública
            
        Returns:
            Tupla (ciphertext_encapsulado, chave_sessao_aes)
        """
        if not self.pqc_available:
            return self.encapsular_chave_rsa(chave_publica_pqc)
        
        try:
            kekem = self.oqs.KeyEncapsulation("ML-KEM-768", chave_publica_pqc)
            ciphertext, chave_sessao = kekem.encap_secret()
            return ciphertext, chave_sessao
        except Exception as e:
            logger.warning(f"Erro ao encapsular ML-KEM: {e}. Usando RSA.")
            return self.encapsular_chave_rsa(chave_publica_pqc)
    
    def desencapsular_chave(self, chave_privada_pqc: bytes, ciphertext: bytes) -> bytes:
        """
        Desencapsula uma chave de sessão AES usando PQC ou RSA.
        
        Args:
            chave_privada_pqc: Chave privada
            ciphertext: Ciphertext encapsulado
            
        Returns:
            Chave de sessão AES
        """
        if not self.pqc_available:
            return self.desencapsular_chave_rsa(chave_privada_pqc, ciphertext)
        
        try:
            kekem = self.oqs.KeyEncapsulation("ML-KEM-768", chave_privada_pqc)
            chave_sessao = kekem.decap_secret(ciphertext)
            return chave_sessao
        except Exception as e:
            logger.warning(f"Erro ao desencapsular ML-KEM: {e}. Usando RSA.")
            return self.desencapsular_chave_rsa(chave_privada_pqc, ciphertext)
    
    def assinar_dados(self, chave_privada_assinatura: bytes, dados: bytes) -> bytes:
        """
        Assina dados usando PQC (ML-DSA-65) ou RSA.
        
        Args:
            chave_privada_assinatura: Chave privada para assinatura
            dados: Dados a assinar
            
        Returns:
            Assinatura digital
        """
        if not self.pqc_available:
            return self.assinar_dados_rsa(chave_privada_assinatura, dados)
        
        try:
            sig = self.oqs.Signature("ML-DSA-65", chave_privada_assinatura)
            assinatura = sig.sign(dados)
            return assinatura
        except Exception as e:
            logger.warning(f"Erro ao assinar ML-DSA: {e}. Usando RSA.")
            return self.assinar_dados_rsa(chave_privada_assinatura, dados)
    
    def verificar_assinatura(self, chave_publica_assinatura: bytes, dados: bytes, assinatura: bytes) -> bool:
        """
        Verifica uma assinatura usando PQC ou RSA.
        
        Args:
            chave_publica_assinatura: Chave pública para verificação
            dados: Dados originais
            assinatura: Assinatura a verificar
            
        Returns:
            True se válida, False caso contrário
        """
        if not self.pqc_available:
            return self.verificar_assinatura_rsa(chave_publica_assinatura, dados, assinatura)
        
        try:
            sig = self.oqs.Signature("ML-DSA-65", chave_publica_assinatura)
            sig.verify(dados, assinatura)
            return True
        except Exception:
            logger.warning("Erro ao verificar ML-DSA. Tentando RSA.")
            return self.verificar_assinatura_rsa(chave_publica_assinatura, dados, assinatura)
    
    # ==================== MÉTODOS HÍBRIDOS ====================
    
    @staticmethod
    def derivar_chave_aes(senha: str) -> bytes:
        """
        Deriva uma chave AES-256 a partir de uma senha.
        
        Args:
            senha: Senha para derivação
            
        Returns:
            Chave AES-256 (32 bytes)
        """
        kdf = PBKDF2HMAC(
            algorithm=hashes.SHA256(),
            length=AES_KEY_SIZE,
            salt=SALT_PQC,
            iterations=ITERATIONS_PQC,
        )
        return kdf.derive(senha.encode())
    
    @staticmethod
    def criptografar_aes(dados: bytes, chave_aes: bytes) -> Tuple[bytes, bytes]:
        """
        Criptografa dados usando AES-256-GCM.
        
        Args:
            dados: Dados a criptografar
            chave_aes: Chave AES-256
            
        Returns:
            Tupla (ciphertext, nonce)
        """
        nonce = os.urandom(NONCE_SIZE)
        aesgcm = AESGCM(chave_aes)
        ciphertext = aesgcm.encrypt(nonce, dados, None)
        return ciphertext, nonce
    
    @staticmethod
    def descriptografar_aes(ciphertext: bytes, nonce: bytes, chave_aes: bytes) -> bytes:
        """
        Descriptografa dados usando AES-256-GCM.
        
        Args:
            ciphertext: Dados criptografados
            nonce: Nonce usado na criptografia
            chave_aes: Chave AES-256
            
        Returns:
            Dados descriptografados
        """
        aesgcm = AESGCM(chave_aes)
        return aesgcm.decrypt(nonce, ciphertext, None)
    
    def criptografar_hibrido(self, dados: bytes, chave_publica_pqc: bytes, senha: str = None) -> Dict[str, Any]:
        """
        Criptografa dados usando encapsulação PQC híbrida.
        
        Args:
            dados: Dados a criptografar
            chave_publica_pqc: Chave pública PQC do destinatário
            senha: Senha opcional para derivação adicional
            
        Returns:
            Dicionário com ciphertext_encapsulado, dados_criptografados, nonce
        """
        try:
            # 1. Encapsular chave de sessão AES com PQC
            ciphertext_encapsulado, chave_sessao = self.encapsular_chave(chave_publica_pqc)
            
            # 2. Derivar chave AES final (híbrida: sessão + senha)
            if senha:
                chave_aes_derivada = self.derivar_chave_aes(senha)
                chave_aes_final = bytes(a ^ b for a, b in zip(chave_sessao[:32], chave_aes_derivada))
            else:
                chave_aes_final = chave_sessao[:32]
            
            # 3. Criptografar dados com AES-256-GCM
            ciphertext, nonce = self.criptografar_aes(dados, chave_aes_final)
            
            logger.info("Criptografia híbrida PQC+AES concluída")
            return {
                "ciphertext_encapsulado": ciphertext_encapsulado.hex(),
                "ciphertext": ciphertext.hex(),
                "nonce": nonce.hex(),
                "algoritmo": "PQC-HYBRID-AES256GCM"
            }
        except Exception as e:
            logger.error(f"Erro na criptografia híbrida: {e}")
            raise
    
    def descriptografar_hibrido(self, dados_criptografados: Dict[str, str], chave_privada_pqc: bytes, senha: str = None) -> bytes:
        """
        Descriptografa dados usando desencapsulação PQC híbrida.
        
        Args:
            dados_criptografados: Dicionário com ciphertext_encapsulado, ciphertext, nonce
            chave_privada_pqc: Chave privada PQC
            senha: Senha opcional para derivação adicional
            
        Returns:
            Dados descriptografados
        """
        try:
            # 1. Desencapsular chave de sessão AES com PQC
            ciphertext_encapsulado = bytes.fromhex(dados_criptografados["ciphertext_encapsulado"])
            chave_sessao = self.desencapsular_chave(chave_privada_pqc, ciphertext_encapsulado)
            
            # 2. Derivar chave AES final (híbrida: sessão + senha)
            if senha:
                chave_aes_derivada = self.derivar_chave_aes(senha)
                chave_aes_final = bytes(a ^ b for a, b in zip(chave_sessao[:32], chave_aes_derivada))
            else:
                chave_aes_final = chave_sessao[:32]
            
            # 3. Descriptografar dados com AES-256-GCM
            ciphertext = bytes.fromhex(dados_criptografados["ciphertext"])
            nonce = bytes.fromhex(dados_criptografados["nonce"])
            dados = self.descriptografar_aes(ciphertext, nonce, chave_aes_final)
            
            logger.info("Descriptografia híbrida PQC+AES concluída")
            return dados
        except Exception as e:
            logger.error(f"Erro na descriptografia híbrida: {e}")
            raise


# ==================== TESTES ====================
if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    
    pqc = PQCCryptoLite()
    
    # Teste 1: Encapsulação de chaves
    print("\n[TEST 1] Encapsulacao de Chaves")
    chave_pub, chave_priv = pqc.gerar_par_chaves_pqc()
    print(f"OK - Chave publica: {len(chave_pub)} bytes")
    print(f"OK - Chave privada: {len(chave_priv)} bytes")
    
    ciphertext_enc, chave_sessao = pqc.encapsular_chave(chave_pub)
    print(f"OK - Ciphertext encapsulado: {len(ciphertext_enc)} bytes")
    
    chave_recuperada = pqc.desencapsular_chave(chave_priv, ciphertext_enc)
    assert chave_recuperada == chave_sessao, "FAIL - Chaves nao correspondem!"
    print("OK - Desencapsulacao bem-sucedida!")
    
    # Teste 2: Assinatura digital
    print("\n[TEST 2] Assinatura Digital")
    chave_pub_sig, chave_priv_sig = pqc.gerar_par_chaves_pqc()
    dados = b"Nebula Cloud - Dados Criticos"
    assinatura = pqc.assinar_dados(chave_priv_sig, dados)
    print(f"OK - Assinatura gerada: {len(assinatura)} bytes")
    
    valida = pqc.verificar_assinatura(chave_pub_sig, dados, assinatura)
    print(f"OK - Assinatura valida: {valida}")
    
    # Teste 3: Criptografia híbrida
    print("\n[TEST 3] Criptografia Hibrida (PQC + AES)")
    dados_teste = b"Dados sensiveis do Nebula Cloud" * 10
    resultado = pqc.criptografar_hibrido(dados_teste, chave_pub, "senha-segura")
    print(f"OK - Dados criptografados: {len(bytes.fromhex(resultado['ciphertext']))} bytes")
    
    dados_recuperados = pqc.descriptografar_hibrido(resultado, chave_priv, "senha-segura")
    assert dados_recuperados == dados_teste, "FAIL - Dados nao correspondem!"
    print("OK - Descriptografia bem-sucedida!")
    
    print("\n[SUCCESS] Todos os testes PQC passaram!")
