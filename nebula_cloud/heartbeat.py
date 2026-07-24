import time
import requests
import json
import os
import logging
from datetime import datetime

# ==================== CONFIGURAÇÃO DE LOGGING ====================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# ==================== CONFIGURAÇÃO ====================
NOS = [
    {"id": "no_01", "ip": "127.0.0.1", "porta": 8000, "tier": "platinum"},
    # Exemplo futuro: {"id": "no_02", "ip": "192.168.0.109", "porta": 8000, "tier": "platinum"},
]

STATUS_FILE = "tracker_db/nos_status.json"
HISTORICO_FILE = "tracker_db/nos_historico.json"
os.makedirs("tracker_db", exist_ok=True)

# ==================== FUNÇÕES ====================
def verificar_no(no):
    """Verifica se o nó está online e coleta métricas"""
    try:
        inicio = time.time()
        r = requests.get(
            f"http://{no['ip']}:{no['porta']}/status",
            timeout=3
        )
        latencia = round((time.time() - inicio) * 1000, 2)

        if r.status_code == 200:
            dados = r.json()
            return {
                "id": no["id"],
                "ip": no["ip"],
                "tier": no["tier"],
                "status": "online",
                "latencia_ms": latencia,
                "fragmentos": dados.get("fragmentos_armazenados", 0),
                "tamanho_total_bytes": dados.get("tamanho_total_bytes", 0),
                "ultimo_ping": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            }
    except requests.exceptions.Timeout:
        logger.warning(f"Timeout ao conectar em {no['id']}")
    except requests.exceptions.ConnectionError:
        logger.warning(f"Erro de conexão com {no['id']}")
    except Exception as e:
        logger.error(f"Erro ao verificar {no['id']}: {e}")

    return {
        "id": no["id"],
        "ip": no["ip"],
        "tier": no["tier"],
        "status": "offline",
        "latencia_ms": None,
        "fragmentos": 0,
        "tamanho_total_bytes": 0,
        "ultimo_ping": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    }

def salvar_status(resultados):
    """Salva status atual dos nós"""
    try:
        with open(STATUS_FILE, 'w') as f:
            json.dump(resultados, f, indent=2)
    except Exception as e:
        logger.error(f"Erro ao salvar status: {e}")

def salvar_historico(resultados):
    """Salva histórico de status dos nós"""
    try:
        historico = []
        if os.path.exists(HISTORICO_FILE):
            with open(HISTORICO_FILE, 'r') as f:
                historico = json.load(f)
        
        # Manter apenas últimas 1000 entradas
        historico.append({
            "timestamp": datetime.now().isoformat(),
            "status": resultados
        })
        historico = historico[-1000:]
        
        with open(HISTORICO_FILE, 'w') as f:
            json.dump(historico, f, indent=2)
    except Exception as e:
        logger.error(f"Erro ao salvar histórico: {e}")

def exibir_status(resultados):
    """Exibe status dos nós no terminal"""
    print(f"\n{'='*70}")
    print(f"🌐 NÉBULA CLOUD — HEARTBEAT v0.2")
    print(f"⏰ {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"{'='*70}")

    online = sum(1 for no in resultados if no["status"] == "online")
    tamanho_total = sum(no.get("tamanho_total_bytes", 0) for no in resultados if no["status"] == "online")
    fragmentos_total = sum(no.get("fragmentos", 0) for no in resultados if no["status"] == "online")
    
    for no in resultados:
        if no["status"] == "online":
            tamanho_gb = no.get("tamanho_total_bytes", 0) / (1024**3)
            print(f"✅ {no['id']} [{no['tier'].upper()}]")
            print(f"   IP: {no['ip']} | Latência: {no['latencia_ms']}ms")
            print(f"   Fragmentos: {no['fragmentos']} | Tamanho: {tamanho_gb:.2f} GB")
        else:
            print(f"❌ {no['id']} [{no['tier'].upper()}] — OFFLINE")

    tamanho_total_gb = tamanho_total / (1024**3)
    print(f"{'─'*70}")
    print(f"📊 Nós online: {online}/{len(resultados)}")
    print(f"📦 Fragmentos totais: {fragmentos_total}")
    print(f"💾 Capacidade total: {tamanho_total_gb:.2f} GB")
    print(f"{'='*70}\n")

def rodar_heartbeat(intervalo=10):
    """Executa heartbeat contínuo"""
    logger.info(f"🚀 Heartbeat iniciado — verificando nós a cada {intervalo} segundos.")
    logger.info("Pressione Ctrl+C para parar.\n")

    try:
        while True:
            resultados = [verificar_no(no) for no in NOS]
            exibir_status(resultados)
            salvar_status(resultados)
            salvar_historico(resultados)
            time.sleep(intervalo)
    except KeyboardInterrupt:
        logger.info("\n⏹️ Heartbeat parado pelo usuário.")
    except Exception as e:
        logger.error(f"Erro fatal no heartbeat: {e}")

# ==================== MAIN ====================
if __name__ == "__main__":
    rodar_heartbeat(intervalo=10)
