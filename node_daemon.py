import time
import json
import os
from rede_p2p import MeshNode

ARQUIVO_REDE = "rede_mesh_publica.txt"

def iniciar_daemon():
    print("==========================================")
    print("      🌐 MESHCOIN CORE DAEMON V1.0 🌐     ")
    print("==========================================")
    
    no_p2p = MeshNode(node_name="Servidor_Core")
    no_p2p.start()
    
    if not os.path.exists(ARQUIVO_REDE):
        with open(ARQUIVO_REDE, "w", encoding="utf-8") as f:
            f.write("--- INÍCIO DA REDE MESH PERSISTENTE ---\n")

    def ao_receber_mensagem(pacote):
        tipo = pacote.get("tipo")
        if tipo == "TRANSACAO":
            print(f"📦 [DAEMON] Nova Transação Recebida de {pacote.get('remetente')[:10]}...")
            with open(ARQUIVO_REDE, "a", encoding="utf-8") as f:
                f.write(f"[SISTEMA] NOVA TRANSAÇÃO: {json.dumps(pacote)}\n")
                
        elif tipo == "CHAT":
            remetente = pacote.get("remetente")
            texto = pacote.get("texto")
            with open(ARQUIVO_REDE, "a", encoding="utf-8") as f:
                f.write(f"[{remetente}] diz: {texto}\n")
                
        elif tipo == "BLOCO":
            print(f"🧱 [DAEMON] Novo Bloco Minerado: {pacote.get('hash')}")
            with open(ARQUIVO_REDE, "a", encoding="utf-8") as f:
                f.write(f"[SISTEMA] NOVO BLOCO: {json.dumps(pacote)}\n")
                
    no_p2p.on_message(ao_receber_mensagem)
    
    print("✅ Daemon rodando e registrando eventos na blockchain local (rede_mesh_publica.txt)")
    print("Pressione Ctrl+C para sair.\n")
    
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nDesligando Daemon...")

if __name__ == "__main__":
    iniciar_daemon()
