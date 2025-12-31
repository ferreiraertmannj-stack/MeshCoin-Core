import time
import threading
import os

ARQUIVO_REDE = "rede_mesh_publica.txt"

def escutar_rede(meu_nome):
    """
    Esta função roda em segundo plano (Thread).
    Ela fica vigiando o arquivo de texto para ver se alguém escreveu algo novo.
    """
    if not os.path.exists(ARQUIVO_REDE):
        with open(ARQUIVO_REDE, "w", encoding="utf-8") as f:
            f.write("--- INÍCIO DA REDE MESH SIMULADA ---\n")

    # Abre o arquivo e vai para o final dele
    with open(ARQUIVO_REDE, "r", encoding="utf-8") as f:
        # Pula para o final do arquivo
        f.seek(0, 2)
        
        while True:
            linha = f.readline()
            if linha:
                # Se a linha não foi escrita por mim, eu mostro na tela
                if not linha.startswith(f"[{meu_nome}]"):
                    print(f"\n📨 {linha.strip()}")
                    print(f"👉 {meu_nome}: ", end="", flush=True) # Restaura o prompt
            else:
                time.sleep(0.5) # Espera um pouco antes de checar de novo

def iniciar_chat():
    print("==========================================")
    print("      📡  MESHCOIN P2P CHAT v0.1  📡      ")
    print("==========================================")
    
    meu_nome = input("Digite seu nome (Ex: Loja, Cliente): ").strip()
    print(f"\n✅ Conectado como: {meu_nome}")
    print("Comece a digitar (ou 'sair' para fechar)...\n")

    # Inicia a "orelha" em paralelo
    thread_escuta = threading.Thread(target=escutar_rede, args=(meu_nome,))
    thread_escuta.daemon = True # Morre quando o programa principal fechar
    thread_escuta.start()

    # O loop principal (A "boca")
    while True:
        mensagem = input(f"👉 {meu_nome}: ")
        
        if mensagem.lower() == 'sair':
            break
            
        if mensagem:
            # Escreve a mensagem no "quadro de avisos" (Arquivo)
            with open(ARQUIVO_REDE, "a", encoding="utf-8") as f:
                f.write(f"[{meu_nome}] diz: {mensagem}\n")

if __name__ == "__main__":
    iniciar_chat()