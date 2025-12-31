import hashlib
import time
import random

# Simulação do Cabeçalho do Bloco MeshCoin
def minerar_bloco(numero_bloco, transacoes):
    print(f"--- Iniciando Mineração do Bloco {numero_bloco} ---")
    print(f"Dados: {transacoes}")
    
    prefixo_alvo = "0000" # Dificuldade (4 zeros)
    nonce = 0
    inicio = time.time()
    
    while True:
        # Cria o conteúdo do bloco: Bloco + Nonce + Dados
        conteudo = f"{numero_bloco}{nonce}{transacoes}".encode()
        
        # Gera o Hash (Impressão digital)
        hash_resultado = hashlib.sha256(conteudo).hexdigest()
        
        # Efeito visual "Matrix" (opcional, mostra tentativa a cada 1000 nonces)
        if nonce % 50000 == 0:
            print(f"Tentando: {nonce} | Hash: {hash_resultado}...")
        
        # Verifica se achou o Hash premiado (começa com 0000)
        if hash_resultado.startswith(prefixo_alvo):
            fim = time.time()
            tempo_gasto = round(fim - inicio, 2)
            print(f"\n✅ SUCESSO! Bloco Encontrado!")
            print(f"🔨 Nonce: {nonce}")
            print(f"🔑 Hash: {hash_resultado}")
            print(f"⏱ Tempo: {tempo_gasto} segundos")
            print("-" * 30 + "\n")
            return hash_resultado
        
        nonce += 1

# O loop principal
if __name__ == "__main__":
    print("🔋 INICIANDO SISTEMA MESHCOIN V0.1 (SIMULATION) 🔋")
    bloco_atual = 1
    
    while True:
        # Simula transações aleatórias acontecendo na rede
        txs = f"Tx_User_{random.randint(1,100)}_send_5_MESH"
        
        # Começa a minerar
        minerar_bloco(bloco_atual, txs)
        
        bloco_atual += 1
        time.sleep(2) # Pausa dramática entre blocos