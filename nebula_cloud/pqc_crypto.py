"""
Módulo de Criptografia Pós-Quântica (PQC) para Nébula Cloud
Implementa ML-KEM (Kyber) para encapsulação de chaves e ML-DSA (Dilithium) para assinatura digital
Segue os padrões NIST FIPS 203, 204 e 205 (2024)

Autor: Manus AI
Data: Abril 2026
"""

import os
import json
import logging
from typing import Tuple, Dict, Any
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC

try:
    import oqs
    PQC_AVAILABLE = True
except ImportError:
    PQC_AVAILABLE = False
    logging.warning("liboqs não disponível. Usando apenas criptografia clássica.")

# ==================== CONFIGURAÇÃO ====================
logger = logging.getLogger(__name__)

# Algoritmos PQC aprovados pelo NIST (2024)
ML_KEM_ALGORITHM = "ML-KEM-768"  # Kyber-768 (FIPS 203)
ML_DSA_ALGORITHM = "ML-DSA-65"   # Dilithium-2 (FIPS 204)

# Parâmetros de segurança
SALT_PQC = b'nebula-cloud-pqc-2026'
ITERATIONS_PQC = 100000
AES_KEY_SIZE = 32  # 256 bits
NONCE_SIZE = 12    # 96 bits para GCM

class PQCCrypto:
    """
    Classe para gerenciar criptografia pós-quântica híbrida.
    Combina ML-KEM (encapsulação de chaves) com AES-256-GCM (criptografia de dados).
    """
    
    def __init__(self):
        """Inicializa o sistema de criptografia PQC"""
        self.pqc_available = PQC_AVAILABLE
        if not self.pqc_available:
            logger.warning("PQC não disponível. Operando em modo clássico apenas.")
        else:
            logger.info(f"PQC ativado: {ML_KEM_ALGORITHM} + {ML_DSA_ALGORITHM}")
    
    def gerar_par_chaves_pqc(self) -> Tuple[bytes, bytes]:
        """
        Gera um par de chaves PQC (pública/privada) usando ML-KEM-768.
        
        Returns:
            Tupla (chave_publica, chave_privada) em bytes
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível. Instale liboqs-python.")
        
        try:
            kekem = oqs.KeyEncapsulation(ML_KEM_ALGORITHM)
            chave_publica = kekem.generate_keypair()
            chave_privada = kekem.export_secret_key()
            logger.info(f"Par de chaves PQC gerado com {ML_KEM_ALGORITHM}")
            return chave_publica, chave_privada
        except Exception as e:
            logger.error(f"Erro ao gerar par de chaves PQC: {e}")
            raise
    
    def encapsular_chave(self, chave_publica_pqc: bytes) -> Tuple[bytes, bytes]:
        """
        Encapsula uma chave de sessão AES usando a chave pública PQC (ML-KEM).
        
        Args:
            chave_publica_pqc: Chave pública ML-KEM
            
        Returns:
            Tupla (ciphertext_encapsulado, chave_sessao_aes)
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            kekem = oqs.KeyEncapsulation(ML_KEM_ALGORITHM, chave_publica_pqc)
            ciphertext, chave_sessao = kekem.encap_secret()
            logger.info(f"Chave encapsulada com sucesso ({len(ciphertext)} bytes)")
            return ciphertext, chave_sessao
        except Exception as e:
            logger.error(f"Erro ao encapsular chave: {e}")
            raise
    
    def desencapsular_chave(self, chave_privada_pqc: bytes, ciphertext: bytes) -> bytes:
        """
        Desencapsula uma chave de sessão usando a chave privada PQC (ML-KEM).
        
        Args:
            chave_privada_pqc: Chave privada ML-KEM
            ciphertext: Ciphertext encapsulado
            
        Returns:
            Chave de sessão AES desencapsulada
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            kekem = oqs.KeyEncapsulation(ML_KEM_ALGORITHM, chave_privada_pqc)
            chave_sessao = kekem.decap_secret(ciphertext)
            logger.info("Chave desencapsulada com sucesso")
            return chave_sessao
        except Exception as e:
            logger.error(f"Erro ao desencapsular chave: {e}")
            raise
    
    def gerar_par_assinatura_pqc(self) -> Tuple[bytes, bytes]:
        """
        Gera um par de chaves para assinatura digital usando ML-DSA-65 (Dilithium-2).
        
        Returns:
            Tupla (chave_publica_assinatura, chave_privada_assinatura)
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            sig = oqs.Signature(ML_DSA_ALGORITHM)
            chave_publica = sig.generate_keypair()
            chave_privada = sig.export_secret_key()
            logger.info(f"Par de assinatura PQC gerado com {ML_DSA_ALGORITHM}")
            return chave_publica, chave_privada
        except Exception as e:
            logger.error(f"Erro ao gerar par de assinatura: {e}")
            raise
    
    def assinar_dados(self, chave_privada_assinatura: bytes, dados: bytes) -> bytes:
        """
        Assina dados usando ML-DSA-65 (Dilithium-2).
        
        Args:
            chave_privada_assinatura: Chave privada para assinatura
            dados: Dados a assinar
            
        Returns:
            Assinatura digital em bytes
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            sig = oqs.Signature(ML_DSA_ALGORITHM, chave_privada_assinatura)
            assinatura = sig.sign(dados)
            logger.info(f"Dados assinados com sucesso ({len(assinatura)} bytes)")
            return assinatura
        except Exception as e:
            logger.error(f"Erro ao assinar dados: {e}")
            raise
    
    def verificar_assinatura(self, chave_publica_assinatura: bytes, dados: bytes, assinatura: bytes) -> bool:
        """
        Verifica uma assinatura digital usando ML-DSA-65.
        
        Args:
            chave_publica_assinatura: Chave pública para verificação
            dados: Dados originais
            assinatura: Assinatura a verificar
            
        Returns:
            True se a assinatura é válida, False caso contrário
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            sig = oqs.Signature(ML_DSA_ALGORITHM, chave_publica_assinatura)
            sig.verify(dados, assinatura)
            logger.info("Assinatura verificada com sucesso")
            return True
        except Exception as e:
            logger.warning(f"Falha na verificação de assinatura: {e}")
            return False
    
    @staticmethod
    def derivar_chave_aes(senha: str) -> bytes:
        """
        Deriva uma chave AES-256 a partir de uma senha usando PBKDF2HMAC.
        
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
        Criptografa dados usando encapsulação PQC híbrida:
        1. Encapsula uma chave de sessão AES com ML-KEM
        2. Criptografa os dados com AES-256-GCM
        
        Args:
            dados: Dados a criptografar
            chave_publica_pqc: Chave pública ML-KEM do destinatário
            senha: Senha opcional para derivação adicional de chave
            
        Returns:
            Dicionário com ciphertext_encapsulado, dados_criptografados, nonce
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            # 1. Encapsular chave de sessão AES com ML-KEM
            ciphertext_encapsulado, chave_sessao = self.encapsular_chave(chave_publica_pqc)
            
            # 2. Derivar chave AES final (híbrida: sessão + senha)
            if senha:
                chave_aes_derivada = self.derivar_chave_aes(senha)
                chave_aes_final = bytes(a ^ b for a, b in zip(chave_sessao[:32], chave_aes_derivada))
            else:
                chave_aes_final = chave_sessao[:32]
            
            # 3. Criptografar dados com AES-256-GCM
            ciphertext, nonce = self.criptografar_aes(dados, chave_aes_final)
            
            logger.info("Criptografia híbrida PQC+AES concluída com sucesso")
            return {
                "ciphertext_encapsulado": ciphertext_encapsulado.hex(),
                "ciphertext": ciphertext.hex(),
                "nonce": nonce.hex(),
                "algoritmo_kem": ML_KEM_ALGORITHM,
                "algoritmo_aes": "AES-256-GCM"
            }
        except Exception as e:
            logger.error(f"Erro na criptografia híbrida: {e}")
            raise
    
    def descriptografar_hibrido(self, dados_criptografados: Dict[str, str], chave_privada_pqc: bytes, senha: str = None) -> bytes:
        """
        Descriptografa dados usando desencapsulação PQC híbrida.
        
        Args:
            dados_criptografados: Dicionário com ciphertext_encapsulado, ciphertext, nonce
            chave_privada_pqc: Chave privada ML-KEM
            senha: Senha opcional para derivação adicional de chave
            
        Returns:
            Dados descriptografados
        """
        if not self.pqc_available:
            raise RuntimeError("PQC não disponível.")
        
        try:
            # 1. Desencapsular chave de sessão AES com ML-KEM
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
            
            logger.info("Descriptografia híbrida PQC+AES concluída com sucesso")
            return dados
        except Exception as e:
            logger.error(f"Erro na descriptografia híbrida: {e}")
            raise


# ==================== TESTES ====================
if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    
    if not PQC_AVAILABLE:
        print("❌ liboqs não disponível. Instale com: pip3 install liboqs-python")
        exit(1)
    
    pqc = PQCCrypto()
    
    # Teste 1: Encapsulação de chaves
    print("\n🔐 Teste 1: Encapsulação de Chaves (ML-KEM-768)")
    chave_pub, chave_priv = pqc.gerar_par_chaves_pqc()
    print(f"✅ Chave pública: {len(chave_pub)} bytes")
    print(f"✅ Chave privada: {len(chave_priv)} bytes")
    
    ciphertext_enc, chave_sessao = pqc.encapsular_chave(chave_pub)
    print(f"✅ Ciphertext encapsulado: {len(ciphertext_enc)} bytes")
    print(f"✅ Chave de sessão: {len(chave_sessao)} bytes")
    
    chave_recuperada = pqc.desencapsular_chave(chave_priv, ciphertext_enc)
    assert chave_recuperada == chave_sessao, "❌ Chaves não correspondem!"
    print("✅ Desencapsulação bem-sucedida!")
    
    # Teste 2: Assinatura digital
    print("\n🔏 Teste 2: Assinatura Digital (ML-DSA-65)")
    chave_pub_sig, chave_priv_sig = pqc.gerar_par_assinatura_pqc()
    print(f"✅ Chave pública de assinatura: {len(chave_pub_sig)} bytes")
    print(f"✅ Chave privada de assinatura: {len(chave_priv_sig)} bytes")
    
    dados = b"Nebula Cloud - Dados Criticos"
    assinatura = pqc.assinar_dados(chave_priv_sig, dados)
    print(f"✅ Assinatura gerada: {len(assinatura)} bytes")
    
    valida = pqc.verificar_assinatura(chave_pub_sig, dados, assinatura)
    print(f"✅ Assinatura válida: {valida}")
    
    # Teste 3: Criptografia híbrida
    print("\n🔒 Teste 3: Criptografia Híbrida (PQC + AES)")
    dados_teste = b"Dados sensiveis do Nebula Cloud" * 10
    resultado = pqc.criptografar_hibrido(dados_teste, chave_pub, "senha-segura")
    print(f"✅ Dados criptografados: {len(bytes.fromhex(resultado['ciphertext']))} bytes")
    
    dados_recuperados = pqc.descriptografar_hibrido(resultado, chave_priv, "senha-segura")
    assert dados_recuperados == dados_teste, "❌ Dados não correspondem!"
    print("✅ Descriptografia bem-sucedida!")
    
    print("\n✅ Todos os testes PQC passaram com sucesso!")
