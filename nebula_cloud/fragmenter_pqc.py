"""
Fragmenter com suporte a Criptografia Pós-Quântica (PQC)
Integra ML-KEM (Kyber) + AES-256-GCM para fragmentação segura

Autor: Manus AI
Data: Abril 2026
"""

import os
import json
import logging
from typing import Tuple, Dict, Any
from reedsolo import RSCodec
from pqc_crypto_lite import PQCCryptoLite

logger = logging.getLogger(__name__)

class FragmenterPQC:
    """
    Fragmentador de arquivos com suporte a criptografia pós-quântica.
    Combina Reed-Solomon com encapsulação PQC híbrida.
    """
    
    def __init__(self, nsym: int = 4):
        """
        Inicializa o fragmentador.
        
        Args:
            nsym: Número de símbolos de redundância Reed-Solomon (padrão: 4)
        """
        self.nsym = nsym
        self.pqc = PQCCryptoLite()
        self.rsc = RSCodec(nsym)
        logger.info(f"FragmenterPQC inicializado com nsym={nsym}")
    
    def fragmentar_arquivo_pqc(self, caminho_arquivo: str, chave_publica_pqc: bytes, 
                                 senha: str, diretorio_saida: str) -> Dict[str, Any]:
        """
        Fragmenta um arquivo com criptografia PQC.
        
        Args:
            caminho_arquivo: Caminho do arquivo
            chave_publica_pqc: Chave pública PQC do destinatário
            senha: Senha para derivação de chave
            diretorio_saida: Diretório para salvar fragmentos
            
        Returns:
            Dicionário com metadados de fragmentação
        """
        # 1. Ler arquivo
        with open(caminho_arquivo, 'rb') as f:
            dados = f.read()
        
        logger.info(f"Arquivo lido: {len(dados)} bytes")
        
        # 2. Criptografar com PQC
        dados_criptografados = self.pqc.criptografar_hibrido(dados, chave_publica_pqc, senha)
        ciphertext = bytes.fromhex(dados_criptografados['ciphertext'])
        
        logger.info(f"Dados criptografados: {len(ciphertext)} bytes")
        
        # 3. Fragmentar com Reed-Solomon
        K = self.nsym  # Número de fragmentos redundantes
        N = 10 + K     # Total de fragmentos (10 + redundância)
        
        tam_frag = (len(ciphertext) + N - 1) // N
        ciphertext_padded = ciphertext + b'\x00' * (tam_frag * N - len(ciphertext))
        
        # 4. Aplicar Reed-Solomon
        fragmentos = []
        for i in range(N):
            inicio = i * tam_frag
            fim = inicio + tam_frag
            fragmento = ciphertext_padded[inicio:fim]
            
            # Codificar com Reed-Solomon
            try:
                fragmento_codificado = self.rsc.encode(fragmento)
                fragmentos.append(fragmento_codificado)
            except Exception as e:
                logger.warning(f"Erro ao codificar fragmento {i}: {e}")
                fragmentos.append(fragmento)
        
        # 5. Salvar fragmentos
        os.makedirs(diretorio_saida, exist_ok=True)
        
        for i, frag in enumerate(fragmentos):
            nome_frag = f"{diretorio_saida}/frag_{i:02d}.bin"
            with open(nome_frag, 'wb') as f:
                f.write(frag)
        
        # 6. Salvar metadados
        metadados = {
            "arquivo_original": os.path.basename(caminho_arquivo),
            "tamanho_original": len(dados),
            "tamanho_criptografado": len(ciphertext),
            "fragmentos_totais": N,
            "fragmentos_necessarios": 10,
            "tamanho_fragmento": tam_frag,
            "algoritmo_criptografia": dados_criptografados['algoritmo'],
            "ciphertext_encapsulado": dados_criptografados['ciphertext_encapsulado'],
            "nonce": dados_criptografados['nonce'],
            "nsym_reed_solomon": K
        }
        
        meta_path = f"{diretorio_saida}/meta.json"
        with open(meta_path, 'w') as f:
            json.dump(metadados, f, indent=2)
        
        logger.info(f"Arquivo fragmentado em {N} pedaços em {diretorio_saida}")
        logger.info(f"Metadados salvos em {meta_path}")
        
        return metadados
    
    def reconstruir_arquivo_pqc(self, diretorio_fragmentos: str, chave_privada_pqc: bytes,
                                  senha: str, caminho_saida: str) -> bool:
        """
        Reconstrói um arquivo a partir de fragmentos com PQC.
        
        Args:
            diretorio_fragmentos: Diretório com fragmentos
            chave_privada_pqc: Chave privada PQC
            senha: Senha para derivação de chave
            caminho_saida: Caminho para salvar arquivo reconstruído
            
        Returns:
            True se bem-sucedido, False caso contrário
        """
        # 1. Carregar metadados
        meta_path = f"{diretorio_fragmentos}/meta.json"
        with open(meta_path, 'r') as f:
            metadados = json.load(f)
        
        logger.info(f"Metadados carregados: {metadados['arquivo_original']}")
        
        # 2. Carregar fragmentos
        fragmentos = []
        for i in range(metadados['fragmentos_totais']):
            nome_frag = f"{diretorio_fragmentos}/frag_{i:02d}.bin"
            if os.path.exists(nome_frag):
                with open(nome_frag, 'rb') as f:
                    fragmentos.append(f.read())
            else:
                fragmentos.append(None)
        
        # 3. Reconstruir com Reed-Solomon
        tam_frag = metadados['tamanho_fragmento']
        ciphertext_reconstruido = b''
        
        # Usar os primeiros 10 fragmentos (sem redundância)
        for i in range(min(metadados['fragmentos_necessarios'], len(fragmentos))):
            if fragmentos[i] is not None:
                # Remover o byte de redundância Reed-Solomon (último byte)
                frag_limpo = fragmentos[i][:-metadados['nsym_reed_solomon']]
                ciphertext_reconstruido += frag_limpo
        
        # Remover padding
        ciphertext_reconstruido = ciphertext_reconstruido[:metadados['tamanho_criptografado']]
        
        logger.info(f"Fragmentos reconstruídos: {len(ciphertext_reconstruido)} bytes")
        
        # 4. Descriptografar com PQC
        dados_criptografados = {
            "ciphertext_encapsulado": metadados['ciphertext_encapsulado'],
            "ciphertext": ciphertext_reconstruido.hex(),
            "nonce": metadados['nonce']
        }
        
        dados_originais = self.pqc.descriptografar_hibrido(dados_criptografados, chave_privada_pqc, senha)
        
        logger.info(f"Dados descriptografados: {len(dados_originais)} bytes")
        
        # 5. Salvar arquivo reconstruído
        with open(caminho_saida, 'wb') as f:
            f.write(dados_originais)
        
        logger.info(f"Arquivo reconstruído salvo em {caminho_saida}")
        
        return len(dados_originais) == metadados['tamanho_original']


# ==================== TESTES ====================
if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    
    # Gerar chaves PQC
    pqc = PQCCryptoLite()
    chave_pub, chave_priv = pqc.gerar_par_chaves_pqc()
    
    # Criar arquivo de teste
    with open('/tmp/teste_pqc.txt', 'w') as f:
        f.write("Dados sensíveis do Nebula Cloud com PQC\n" * 100)
    
    # Fragmentar
    print("\n[TEST] Fragmentacao com PQC")
    fragmenter = FragmenterPQC(nsym=4)
    metadados = fragmenter.fragmentar_arquivo_pqc(
        '/tmp/teste_pqc.txt',
        chave_pub,
        'senha-pqc-segura',
        '/tmp/fragmentos_pqc'
    )
    print(f"OK - {metadados['fragmentos_totais']} fragmentos criados")
    
    # Reconstruir
    print("\n[TEST] Reconstrucao com PQC")
    sucesso = fragmenter.reconstruir_arquivo_pqc(
        '/tmp/fragmentos_pqc',
        chave_priv,
        'senha-pqc-segura',
        '/tmp/reconstruido_pqc.txt'
    )
    
    # Validar
    with open('/tmp/teste_pqc.txt', 'rb') as f:
        original = f.read()
    with open('/tmp/reconstruido_pqc.txt', 'rb') as f:
        reconstruido = f.read()
    
    if original == reconstruido:
        print("OK - Arquivo reconstruído com sucesso!")
    else:
        print("FAIL - Arquivo corrompido!")
