import time
import os
from crypto_core import HybridCrypto

def gerar_carteira_hibrida():
    print("--- 🔐 GERADOR DE CARTEIRA MESHCOIN V3 (Hybrid Crypto) ---")
    
    # Gera o par de chaves usando o novo Core Híbrido
    carteira = HybridCrypto.generate_keypair()
    
    address = carteira['address']
    print(f"\n✅ Carteira Híbrida Gerada com Sucesso!")
    print(f"🌍 Endereço (Público): {address}")
    
    # Salva no Arquivo
    nome_arquivo = f"carteira_{address[:8]}.txt"
    
    try:
        with open(nome_arquivo, "w") as f:
            f.write("--- MESHCOIN HYBRID WALLET ---\n")
            f.write(f"Data de Criação: {time.ctime()}\n")
            f.write("-" * 30 + "\n")
            f.write(f"Endereço Público: {address}\n")
            f.write(f"Private Key ECDSA: \n{carteira['private_ecdsa']}\n")
            f.write(f"Private Seed PQC: {carteira['private_pqc_seed']}\n")
            f.write("-" * 30 + "\n")
            f.write("GUARDE ESTE ARQUIVO EM LOCAL SEGURO!")
            
        print(f"\n💾 SUCESSO! Dados salvos no arquivo: {nome_arquivo}")
        print(f"📂 Verifique a pasta: {os.getcwd()}")
        
    except Exception as e:
        print(f"\n❌ Erro ao salvar arquivo: {e}")

if __name__ == "__main__":
    gerar_carteira_hibrida()