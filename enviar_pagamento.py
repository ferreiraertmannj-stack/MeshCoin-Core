import time
import json
import os
from rede_p2p import MeshNode

def carregar_minha_carteira():
    # Procura arquivos que começam com "carteira_" na pasta
    arquivos = [f for f in os.listdir() if f.startswith("carteira_") and f.endswith(".txt")]
    
    if not arquivos:
        print("❌ Nenhuma carteira encontrada! Rode o 'gerar_carteira.py' primeiro.")
        return None, None
    
    # Pega a primeira carteira que achar
    nome_arquivo = arquivos[0]
    with open(nome_arquivo, "r") as f:
        linhas = f.readlines()
        # Extrai os dados
        priv_key = [l for l in linhas if "Private Key" in l][0].split(": ")[1].strip()
        pub_addr = [l for l in linhas if "Public Address" in l][0].split(": ")[1].strip()
        
    return priv_key, pub_addr

def realizar_pagamento():
    print("--- 💸 MESHCOIN TRANSFER (P2P) 💸 ---")
    
    # 1. Autenticação
    chave_privada, meu_endereco = carregar_minha_carteira()
    if not chave_privada: return

    print(f"👤 De (Você): {meu_endereco}")
    
    # 2. Dados do Pagamento
    destinatario = input("👉 Para (Endereço ou Nome): ")
    valor = input("💰 Valor (MESH): ")
    
    # Inicia o nó P2P rapidamente
    no_p2p = MeshNode(node_name=meu_endereco[:8])
    no_p2p.start()
    
    print("\n📡 Conectando à rede local... (Aguarde 3s)")
    time.sleep(3) # Tempo para descobrir peers UDP
    
    # 3. Criando o Pacote da Transação
    transacao = {
        "tipo": "TRANSACAO",
        "remetente": meu_endereco,
        "destinatario": destinatario,
        "valor": valor,
        "timestamp": time.time(),
        "assinatura": f"ASSINADO_POR_{chave_privada[:10]}..."
    }
    
    # 4. Broadcast (Jogar na rede)
    print("📡 Transmitindo para a Rede Mesh P2P...")
    no_p2p.broadcast_data(transacao)
    
    print("✅ SUCESSO! Pagamento enviado para a rede.")
    print(f"Alcançou {len(no_p2p.peers)} nós visíveis.")
    time.sleep(1)

if __name__ == "__main__":
    realizar_pagamento()