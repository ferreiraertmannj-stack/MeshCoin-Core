import time
import threading
import sys
from rede_p2p import MeshNode

def iniciar_chat():
    print("==========================================")
    print("      📡  MESHCOIN P2P CHAT v0.2  📡      ")
    print("==========================================")
    
    meu_nome = input("Digite seu nome (Ex: Loja, Cliente): ").strip()
    print(f"\n✅ Conectado como: {meu_nome}")
    
    # Inicializa o nó P2P
    no_p2p = MeshNode(node_name=meu_nome)
    no_p2p.start()

    def ao_receber_mensagem(pacote):
        if pacote.get("tipo") == "CHAT":
            remetente = pacote.get("remetente")
            texto = pacote.get("texto")
            if remetente != meu_nome:
                # Limpa a linha atual, imprime a mensagem e restaura o prompt
                sys.stdout.write('\r' + ' ' * 50 + '\r')
                print(f"📨 [{remetente}] diz: {texto}")
                sys.stdout.write(f"👉 {meu_nome}: ")
                sys.stdout.flush()

    no_p2p.on_message(ao_receber_mensagem)

    print("Procurando outros nós na rede local (UDP)...")
    time.sleep(2)
    print("Comece a digitar (ou 'sair' para fechar)...\n")

    # O loop principal (A "boca")
    while True:
        mensagem = input(f"👉 {meu_nome}: ")
        
        if mensagem.lower() == 'sair':
            break
            
        if mensagem:
            # Envia para a rede P2P real
            pacote = {
                "tipo": "CHAT",
                "remetente": meu_nome,
                "texto": mensagem
            }
            no_p2p.broadcast_data(pacote)

if __name__ == "__main__":
    iniciar_chat()