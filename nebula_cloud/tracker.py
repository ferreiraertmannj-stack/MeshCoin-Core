import os
import json
import requests
from fragmenter import fragmentar_arquivo, reconstruir_arquivo

# ==================== CONFIGURAÇÃO ====================
NOS = [
    {"id": "no_01", "ip": "127.0.0.1", "porta": 8000, "tier": "platinum"},
    # Adicione mais nós aqui (ex: Infinix Hot 50)
    # {"id": "no_02", "ip": "192.168.1.105", "porta": 8000, "tier": "platinum"},
]

TRACKER_DIR = "tracker_db"
os.makedirs(TRACKER_DIR, exist_ok=True)

# ==================== FUNÇÕES ====================
def verificar_nos():
    """Verifica quais nós estão online"""
    nos_online = []
    for no in NOS:
        try:
            r = requests.get(f"http://{no['ip']}:{no['porta']}/status", timeout=3)
            if r.status_code == 200:
                dados = r.json()
                no["status"] = "online"
                no["fragmentos"] = dados.get("fragmentos_armazenados", 0)
                nos_online.append(no)
                print(f"✅ Nó {no['id']} online — {dados.get('fragmentos_armazenados', 0)} fragmentos")
        except Exception as e:
            no["status"] = "offline"
            print(f"❌ Nó {no['id']} offline")
    return nos_online

def distribuir_arquivo(caminho: str, senha: str):
    """Fragmenta e distribui arquivo pelos nós disponíveis"""
    nos_online = verificar_nos()
    
    if len(nos_online) == 0:
        print("❌ Nenhum nó disponível!")
        return

    # Fragmenta o arquivo
    pasta_frags = f"{TRACKER_DIR}/temp_frags"
    fragmentar_arquivo(caminho, senha, pasta_frags)

    # Lê metadados
    with open(f"{pasta_frags}/meta.json") as f:
        meta = json.load(f)

    # Distribui fragmentos (round-robin)
    mapa = {}
    fragmentos = [f for f in os.listdir(pasta_frags) if f.endswith('.bin')]

    for i, frag in enumerate(sorted(fragmentos)):
        no = nos_online[i % len(nos_online)]
        caminho_frag = f"{pasta_frags}/{frag}"
        
        with open(caminho_frag, 'rb') as f:
            r = requests.post(
                f"http://{no['ip']}:{no['porta']}/fragmento/{frag}",
                files={"file": f}
            )
        
        if r.status_code == 200:
            mapa[frag] = {"no_id": no["id"], "ip": no["ip"], "porta": no["porta"]}
            print(f"📤 {frag} → Nó {no['id']}")
        else:
            print(f"❌ Erro ao enviar {frag}")

    # Salva mapa no tracker
    nome_arquivo = os.path.basename(caminho)
    registro = {
        "arquivo": nome_arquivo,
        "senha_hint": senha[:3] + "***",
        "meta": meta,
        "mapa": mapa
    }
    with open(f"{TRACKER_DIR}/{nome_arquivo}.map.json", 'w') as f:
        json.dump(registro, f, indent=2)

    print(f"\n✅ Arquivo '{nome_arquivo}' distribuído com sucesso!")
    print(f"📋 Mapa salvo em: {TRACKER_DIR}/{nome_arquivo}.map.json")

def recuperar_arquivo(nome_arquivo: str, senha: str, saida: str):
    """Recupera fragmentos e reconstrói o arquivo"""
    mapa_path = f"{TRACKER_DIR}/{nome_arquivo}.map.json"
    
    if not os.path.exists(mapa_path):
        print(f"❌ Arquivo '{nome_arquivo}' não encontrado no tracker")
        return

    with open(mapa_path) as f:
        registro = json.load(f)

    pasta_recuperado = f"{TRACKER_DIR}/recuperando"
    os.makedirs(pasta_recuperado, exist_ok=True)

    print("🔄 Recuperando fragmentos dos nós...")
    for frag, info in registro["mapa"].items():
        try:
            r = requests.get(f"http://{info['ip']}:{info['porta']}/fragmento/{frag}")
            if r.status_code == 200:
                with open(f"{pasta_recuperado}/{frag}", 'wb') as f:
                    f.write(r.content)
                print(f"✅ {frag} recuperado")
        except Exception as e:
            print(f"❌ Falha ao recuperar {frag}: {e}")

    # Salvar metadados para reconstrução
    with open(f"{pasta_recuperado}/meta.json", 'w') as f:
        json.dump(registro["meta"], f, indent=2)

    # Reconstrói
    try:
        reconstruir_arquivo(pasta_recuperado, senha, saida)
        print(f"✅ Arquivo reconstruído com sucesso: {saida}")
    except Exception as e:
        print(f"❌ Erro ao reconstruir arquivo: {e}")

# ==================== MENU SIMPLES ====================
if __name__ == "__main__":
    print("\n🚀 NÉBULA CLOUD TRACKER v0.1")
    print("1 - Distribuir arquivo")
    print("2 - Recuperar arquivo")
    print("3 - Ver nós online")
    
    opcao = input("\nEscolha uma opção: ")
    
    if opcao == "1":
        caminho = input("Caminho do arquivo: ")
        senha = input("Senha de criptografia: ")
        distribuir_arquivo(caminho, senha)
    elif opcao == "2":
        nome = input("Nome do arquivo (ex: teste.txt): ")
        senha = input("Senha: ")
        saida = input("Nome do arquivo de saída: ")
        recuperar_arquivo(nome, senha, saida)
    elif opcao == "3":
        verificar_nos()
