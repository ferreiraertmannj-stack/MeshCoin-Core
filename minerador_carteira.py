import hashlib
import time
import random

def cabecalho():
    print("==========================================")
    print("      ⛏️  MESHCOIN MINER V1.0  ⛏️      ")
    print("     (Conectado à Carteira Local)         ")
    print("==========================================")

def minerar_para_carteira(meu_endereco):
    bloco_atual = 1
    dificuldade = "0000" # Dificuldade simulada
    recompensa = 50 # Recompensa por bloco (50 MESH)
    saldo_acumulado = 0

    print(f"\n🌍 Iniciando Nó de Mineração para: {meu_endereco}")
    print("📡 Conectando à rede Mesh (Simulado)... Conectado!\n")
    time.sleep(1)

    while True:
        print(f"--- 🔨 Tentando resolver o Bloco {bloco_atual} ---")
        nonce = 0
        inicio = time.time()
        
        while True:
            # O bloco contém: Número + Quem recebe (VOCÊ) + Nonce
            conteudo_bloco = f"{bloco_atual}{meu_endereco}{nonce}".encode()
            hash_resultado = hashlib.sha256(conteudo_bloco).hexdigest()

            # Mostra o esforço a cada 20.000 tentativas
            if nonce % 50000 == 0:
                print(f"   [Hash Rate: {random.randint(2000, 4000)} H/s] Tentando Nonce: {nonce}...")

            # Achou!
            if hash_resultado.startswith(dificuldade):
                fim = time.time()
                tempo = round(fim - inicio, 2)
                saldo_acumulado += recompensa
                
                print(f"\n✅ BLOCO {bloco_atual} MINERADO!")
                print(f"🔑 Hash: {hash_resultado}")
                print(f"⏱  Tempo: {tempo}s")
                print(f"💰 Recompensa: +{recompensa} MESH enviadas para sua carteira.")
                print(f"🏦 SALDO TOTAL NA SESSÃO: {saldo_acumulado} MESH")
                print("-" * 40 + "\n")
                
                time.sleep(2) # Pausa para respirar
                bloco_atual += 1
                break # Vai para o próximo bloco
            
            nonce += 1

if __name__ == "__main__":
    cabecalho()
    # O programa vai pedir para você colar o endereço que gerou
    carteira = input("👉 Cole aqui o seu Endereço Público (Começa com MESH): ").strip()
    
    if carteira.startswith("MESH"):
        minerar_para_carteira(carteira)
    else:
        print("❌ Erro: Endereço inválido! Tem certeza que copiou do arquivo .txt?")