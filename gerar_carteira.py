import hashlib
import secrets
import time
import os

def gerar_carteira_v2():
    print("--- 🔐 GERADOR DE CARTEIRA MESHCOIN V2 (Windows Compatible) ---")
    
    # 1. Gera a Chave Privada
    private_key = secrets.token_hex(32)
    
    # 2. Gera a Chave Pública
    public_key_raw = f"Pub_Key_From_{private_key}"
    public_key = hashlib.sha256(public_key_raw.encode()).hexdigest()
    
    # 3. Gera o Endereço (Usando SHA256 para garantir compatibilidade no Windows)
    # Pegamos os primeiros 40 caracteres para ficar parecido com um endereço real
    address_hash = hashlib.sha256(public_key.encode()).hexdigest()
    wallet_address = f"MESH{address_hash[:40]}"
    
    print("\n✅ Carteira Gerada com Sucesso!")
    print(f"🌍 Endereço: {wallet_address}")
    
    # 4. Salva no Arquivo (Agora garantido!)
    nome_arquivo = f"carteira_{wallet_address[:8]}.txt"
    
    try:
        with open(nome_arquivo, "w") as f:
            f.write("--- MESHCOIN PAPER WALLET ---\n")
            f.write(f"Data de Criação: {time.ctime()}\n")
            f.write("-" * 30 + "\n")
            f.write(f"Private Key (SEGREDO): {private_key}\n")
            f.write(f"Public Address (COMPARTILHE): {wallet_address}\n")
            f.write("-" * 30 + "\n")
            f.write("GUARDE ESTE ARQUIVO EM LOCAL SEGURO!")
            
        print(f"\n💾 SUCESSO! Dados salvos no arquivo: {nome_arquivo}")
        print(f"📂 Verifique a pasta: {os.getcwd()}")
        
    except Exception as e:
        print(f"\n❌ Erro ao salvar arquivo: {e}")

if __name__ == "__main__":
    gerar_carteira_v2()
    input("\nPressione Enter para sair...")