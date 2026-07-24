"""
node_log_bridge.py — Nébula Cloud
Módulo de integração entre o Node Daemon e o Nébula Core.

Inclua este arquivo no diretório nebula-cloud e importe-o no node_daemon.py:
    from node_log_bridge import NebulaBridge

Uso:
    bridge = NebulaBridge()
    bridge.log("no_dell_01", "upload", "Fragmento frag_00.bin recebido", "info")
"""

import os
import json
import time
import threading
import logging
from typing import Literal, Optional
from queue import Queue, Empty

try:
    import requests
    REQUESTS_AVAILABLE = True
except ImportError:
    REQUESTS_AVAILABLE = False
    logging.warning("[NebulaBridge] 'requests' não instalado. Logs não serão enviados ao Nébula Core.")

# ==================== CONFIGURAÇÃO ====================

NEBULA_CORE_URL = os.environ.get("NEBULA_CORE_URL", "http://192.168.0.108:3000")
NEBULA_API_KEY  = os.environ.get("NEBULA_API_KEY",  "nebula-node-secret-2026")
LOG_TIMEOUT_SEC = int(os.environ.get("NEBULA_LOG_TIMEOUT", "3"))
LOG_RETRY_MAX   = int(os.environ.get("NEBULA_LOG_RETRY", "2"))
LOG_QUEUE_SIZE  = int(os.environ.get("NEBULA_LOG_QUEUE", "200"))

EventType = Literal["upload", "download", "delete", "node_online", "node_offline",
                    "health_check", "error", "warning", "info", "auth"]
Severity  = Literal["info", "warning", "error", "critical"]


# ==================== BRIDGE ====================

class NebulaBridge:
    """
    Envia logs do Node Daemon para o Nébula Core de forma assíncrona.
    Os logs são enfileirados e enviados em background para não bloquear o daemon.
    """

    def __init__(self, core_url: str = NEBULA_CORE_URL, api_key: str = NEBULA_API_KEY):
        self.core_url = core_url.rstrip("/")
        self.api_key  = api_key
        self._queue: Queue = Queue(maxsize=LOG_QUEUE_SIZE)
        self._running = True
        self._thread  = threading.Thread(target=self._worker, daemon=True, name="NebulaBridgeWorker")
        self._thread.start()
        logging.info(f"[NebulaBridge] Iniciado → {self.core_url}")

    # ------------------------------------------------------------------
    # API pública
    # ------------------------------------------------------------------

    def log(
        self,
        node_id: str,
        event_type: EventType,
        message: str,
        severity: Severity = "info",
        metadata: Optional[dict] = None,
    ) -> None:
        """Enfileira um evento de log para envio assíncrono ao Nébula Core."""
        payload = {
            "apiKey":    self.api_key,
            "nodeId":    node_id,
            "eventType": event_type,
            "message":   message,
            "severity":  severity,
        }
        if metadata:
            payload["metadata"] = json.dumps(metadata)

        try:
            self._queue.put_nowait(payload)
        except Exception:
            # Fila cheia — descartar silenciosamente para não bloquear o daemon
            pass

    def shutdown(self, timeout: float = 5.0) -> None:
        """Para o worker de forma limpa, aguardando o envio dos logs pendentes."""
        self._running = False
        self._thread.join(timeout=timeout)

    # ------------------------------------------------------------------
    # Worker interno
    # ------------------------------------------------------------------

    def _worker(self) -> None:
        while self._running or not self._queue.empty():
            try:
                payload = self._queue.get(timeout=1.0)
            except Empty:
                continue

            self._send_with_retry(payload)
            self._queue.task_done()

    def _send_with_retry(self, payload: dict) -> bool:
        if not REQUESTS_AVAILABLE:
            return False

        url = f"{self.core_url}/api/trpc/logs.create"
        body = {"0": {"json": payload}}

        for attempt in range(1, LOG_RETRY_MAX + 1):
            try:
                resp = requests.post(url, json=body, timeout=LOG_TIMEOUT_SEC)
                if resp.status_code in (200, 201):
                    return True
                logging.debug(f"[NebulaBridge] HTTP {resp.status_code} na tentativa {attempt}")
            except requests.exceptions.ConnectionError:
                logging.debug(f"[NebulaBridge] Conexão recusada (tentativa {attempt})")
            except requests.exceptions.Timeout:
                logging.debug(f"[NebulaBridge] Timeout (tentativa {attempt})")
            except Exception as exc:
                logging.debug(f"[NebulaBridge] Erro inesperado: {exc}")

            if attempt < LOG_RETRY_MAX:
                time.sleep(0.5 * attempt)

        return False


# ==================== INSTÂNCIA GLOBAL ====================

# Instância singleton pronta para importar no node_daemon.py
bridge = NebulaBridge()


# ==================== FUNÇÕES DE CONVENIÊNCIA ====================

def log_upload(node_id: str, fragment_name: str, size_bytes: int = 0) -> None:
    bridge.log(node_id, "upload",
               f"Fragmento recebido: {fragment_name} ({size_bytes} bytes)",
               "info", {"fragment": fragment_name, "size": size_bytes})

def log_download(node_id: str, fragment_name: str) -> None:
    bridge.log(node_id, "download",
               f"Fragmento servido: {fragment_name}", "info",
               {"fragment": fragment_name})

def log_delete(node_id: str, fragment_name: str) -> None:
    bridge.log(node_id, "delete",
               f"Fragmento removido: {fragment_name}", "warning",
               {"fragment": fragment_name})

def log_node_online(node_id: str) -> None:
    bridge.log(node_id, "node_online", f"Nó {node_id} iniciado e online", "info")

def log_node_offline(node_id: str, reason: str = "") -> None:
    bridge.log(node_id, "node_offline",
               f"Nó {node_id} offline" + (f": {reason}" if reason else ""),
               "warning")

def log_health_check(node_id: str, latency_ms: int, fragments: int) -> None:
    bridge.log(node_id, "health_check",
               f"Health check: {latency_ms}ms | {fragments} fragmentos", "info",
               {"latency_ms": latency_ms, "fragments": fragments})

def log_error(node_id: str, message: str) -> None:
    bridge.log(node_id, "error", message, "error")


# ==================== TESTE STANDALONE ====================

if __name__ == "__main__":
    import sys
    logging.basicConfig(level=logging.DEBUG)

    node = sys.argv[1] if len(sys.argv) > 1 else "no_test_01"
    print(f"[Teste] Enviando logs de teste para {NEBULA_CORE_URL} como nó '{node}'...")

    log_node_online(node)
    time.sleep(0.2)
    log_upload(node, "frag_00.bin", 1024)
    time.sleep(0.2)
    log_download(node, "frag_00.bin")
    time.sleep(0.2)
    log_health_check(node, 12, 10)
    time.sleep(0.2)

    bridge.shutdown(timeout=5)
    print("[Teste] Concluído! Verifique os logs no painel /admin/logs")
