import time
import json
import os

# Configurações
ARQUIVO_REDE = "rede_mesh_publica.txt"

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
        # Extrai os dados (Gambiarra inteligente para achar as linhas certas)
        priv_key = [l for l in linhas if "Private Key" in l][0].split(": ")[1].strip()
        pub_addr = [l for l in linhas if "Public Address" in l][0].split(": ")[1].strip()
        
    return priv_key, pub_addr

def realizar_pagamento():
    print("--- 💸 MESHCOIN TRANSFER (OFFLINE) 💸 ---")
    
    # 1. Autenticação
    chave_privada, meu_endereco = carregar_minha_carteira()
    if not chave_privada: return

    print(f"👤 De (Você): {meu_endereco}")
    
    # 2. Dados do Pagamento
    destinatario = input("👉 Para (Endereço ou Nome): ")
    valor = input("💰 Valor (MESH): ")
    
    # 3. Criando o Pacote da Transação (JSON)
    transacao = {
        "tipo": "TRANSACAO",
        "remetente": meu_endereco,
        "destinatario": destinatario,
        "valor": valor,
        "timestamp": time.time(),
        "assinatura": f"ASSINADO_POR_{chave_privada[:10]}..." # Simulação de criptografia
    }
    
    # Converte para texto
    pacote_dados = json.dumps(transacao)
    
    # 4. Broadcast (Jogar na rede)
    print("\n📡 Enviando para a Rede Mesh...")
    time.sleep(1)
    
    with open(ARQUIVO_REDE, "a", encoding="utf-8") as f:
        f.write(f"[SISTEMA] NOVA TRANSAÇÃO: {pacote_dados}\n")
        
    print("✅ SUCESSO! Pagamento registrado na rede.")
    print("Todos os nós conectados agora podem ver sua transferência.")

if __name__ == "__main__":
    realizar_pagamento()