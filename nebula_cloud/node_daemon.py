import os
import logging
import uvicorn
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import FileResponse
from datetime import datetime

# Integração com Nébula Core (opcional — funciona sem o painel rodando)
try:
    from node_log_bridge import log_upload, log_download, log_delete, log_node_online, log_error
    BRIDGE_ENABLED = True
except ImportError:
    BRIDGE_ENABLED = False
    def log_upload(*a, **kw): pass
    def log_download(*a, **kw): pass
    def log_delete(*a, **kw): pass
    def log_node_online(*a, **kw): pass
    def log_error(*a, **kw): pass

NODE_ID = os.environ.get("NODE_ID", "nebula-node-01")

# ==================== CONFIGURAÇÃO DE LOGGING ====================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# ==================== INICIALIZAÇÃO ====================
app = FastAPI(title="Nébula Node Daemon", version="0.2")

STORAGE_DIR = "node_storage"
os.makedirs(STORAGE_DIR, exist_ok=True)

logger.info(f"Node Daemon iniciado. Diretório de armazenamento: {STORAGE_DIR}")

# ==================== ROTAS ====================
@app.get("/status")
def status():
    """Retorna status do nó e fragmentos armazenados"""
    try:
        fragmentos = os.listdir(STORAGE_DIR)
        tamanho_total = sum(os.path.getsize(f"{STORAGE_DIR}/{f}") for f in fragmentos if os.path.isfile(f"{STORAGE_DIR}/{f}"))
        
        return {
            "status": "online",
            "node": NODE_ID,
            "timestamp": datetime.now().isoformat(),
            "fragmentos_armazenados": len(fragmentos),
            "tamanho_total_bytes": tamanho_total,
            "arquivos": fragmentos
        }
    except Exception as e:
        logger.error(f"Erro ao obter status: {e}")
        raise HTTPException(status_code=500, detail=f"Erro ao obter status: {e}")

@app.post("/fragmento/{nome}")
async def receber_fragmento(nome: str, file: UploadFile = File(...)):
    """Recebe e armazena um fragmento"""
    try:
        # Validar nome do arquivo
        if ".." in nome or "/" in nome or "\\" in nome:
            raise HTTPException(status_code=400, detail="Nome de arquivo inválido")
        
        caminho = f"{STORAGE_DIR}/{nome}"
        conteudo = await file.read()
        
        if len(conteudo) == 0:
            raise HTTPException(status_code=400, detail="Arquivo vazio")
        
        with open(caminho, 'wb') as f:
            f.write(conteudo)
        
        logger.info(f"✅ Fragmento recebido: {nome} ({len(conteudo)} bytes)")
        log_upload(NODE_ID, nome, len(conteudo))
        return {
            "status": "salvo",
            "fragmento": nome,
            "tamanho_bytes": len(conteudo),
            "timestamp": datetime.now().isoformat()
        }
    except Exception as e:
        logger.error(f"❌ Erro ao receber fragmento {nome}: {e}")
        log_error(NODE_ID, f"Erro ao receber fragmento {nome}: {e}")
        raise HTTPException(status_code=500, detail=f"Erro ao salvar fragmento: {e}")

@app.get("/fragmento/{nome}")
def enviar_fragmento(nome: str):
    """Envia um fragmento armazenado"""
    try:
        # Validar nome do arquivo
        if ".." in nome or "/" in nome or "\\" in nome:
            raise HTTPException(status_code=400, detail="Nome de arquivo inválido")
        
        caminho = f"{STORAGE_DIR}/{nome}"
        
        if not os.path.exists(caminho):
            logger.warning(f"Fragmento não encontrado: {nome}")
            raise HTTPException(status_code=404, detail="Fragmento não encontrado")
        
        logger.info(f"📥 Fragmento enviado: {nome}")
        log_download(NODE_ID, nome)
        return FileResponse(caminho)
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"❌ Erro ao enviar fragmento {nome}: {e}")
        raise HTTPException(status_code=500, detail=f"Erro ao enviar fragmento: {e}")

@app.delete("/fragmento/{nome}")
def deletar_fragmento(nome: str):
    """Deleta um fragmento armazenado"""
    try:
        # Validar nome do arquivo
        if ".." in nome or "/" in nome or "\\" in nome:
            raise HTTPException(status_code=400, detail="Nome de arquivo inválido")
        
        caminho = f"{STORAGE_DIR}/{nome}"
        
        if not os.path.exists(caminho):
            raise HTTPException(status_code=404, detail="Fragmento não encontrado")
        
        os.remove(caminho)
        logger.info(f"🗑️ Fragmento deletado: {nome}")
        log_delete(NODE_ID, nome)
        return {"status": "deletado", "fragmento": nome}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"❌ Erro ao deletar fragmento {nome}: {e}")
        raise HTTPException(status_code=500, detail=f"Erro ao deletar fragmento: {e}")

@app.get("/health")
def health_check():
    """Health check simples"""
    return {"status": "healthy", "timestamp": datetime.now().isoformat()}

# ==================== INICIALIZAÇÃO ====================
if __name__ == "__main__":
    port = int(os.environ.get("NODE_PORT", 8000))
    logger.info(f"🚀 Iniciando Node Daemon '{NODE_ID}' na porta {port}...")
    log_node_online(NODE_ID)
    uvicorn.run(app, host="0.0.0.0", port=port, log_level="info")
