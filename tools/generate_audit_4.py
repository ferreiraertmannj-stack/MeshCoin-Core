import os

out_dir = "PROJECT_AUDIT"
os.makedirs(out_dir, exist_ok=True)

files = {
    "16_PERFORMANCE.md": """# 16 PERFORMANCE

## Gargalos Identificados
- `saveLedger()` (Go): Usa JSON `MarshalIndent` carregando todo o histórico na RAM para salvar no disco a cada bloco confirmado.
- Tratamento de conexões TCP em `handleConnection` gerando uma nova goroutine (desempenho normal em Go, mas perigoso com floods P2P devido à ausência de socket limits).
- Broadcasts UDP agressivos (cada 5 seg) em rede local, impactante para smartphones (`battery drain`).
- Vetor do `NeonHash` em memória gasta ~4KB. Na escala de múltiplas threads do minerador, é extremamente amigável ao cache L1 da CPU, mas a verificação síncrona bloqueia a validação.
""",

    "17_CODE_QUALITY.md": """# 17 CODE QUALITY

## Problemas de Qualidade
- **TODOs:** 5 pendências gerais.
- **FIXMEs:** 1 fixme crítico de acordo com auditoria.
- **Complexidade:** Maioria dos arquivos contêm funções aglomeradas. O `ledger.go` mescla lógica de persistência de disco, validação matemática e regras de recompensa. 
- **Duplicações:** Conceitos de validação dispersos entre a versão Python de teste (`crypto_core.py`) e a versão oficial em Go.
- Falta separação de pacotes no código Go, tudo está em `main` no diretório `pc_node`.
""",

    "18_KNOWN_BUGS.md": """# 18 KNOWN BUGS

1. **Bug: Data Loss no Ledger Corrompido**
   - **Local:** `pc_node/ledger.go` (initLedger)
   - **Descrição:** Se o unmarshal falhar, a variável `ledger` assume cadeia limpa.
   - **Impacto:** Destruição completa da blockchain local com reescrita ao receber novo bloco.
   - **Severidade:** CRÍTICA.

2. **Bug: Upload Cloud sem Autenticação**
   - **Local:** `nebula_cloud/node_daemon.go`
   - **Descrição:** A API `/upload` aceita 100MB de qualquer IP indefinidamente.
   - **Impacto:** Esgotamento de disco do nó em questão de minutos (DDoS).
   - **Severidade:** CRÍTICA.

3. **Bug: Broadcast TCP Loop**
   - **Local:** `pc_node/network.go` (broadcastTCP)
   - **Descrição:** Falta validação de ciclo de pacote P2P.
   - **Impacto:** Amplificação infinita do pacote se houver ciclos de roteamento.
   - **Severidade:** ALTA.

(Nenhum código foi alterado nesta auditoria, os bugs permanecem na base.)
""",

    "19_TECHNICAL_DEBT.md": """# 19 TECHNICAL DEBT

## Arquitetural
- Armazenamento em arquivos planos JSON, exigindo refatoração imediata para LevelDB.
- Sistema P2P sem algoritmo real Kademlia/DHT para descoberta em WAN, dependendo de Broadcasts 255.255.255.255 (só LAN).

## Código
- Dependência de variáveis globais e Mutexes únicos (`ledger.mu`, `mempoolMutex`) em vez de canais (channels) ou DB locks em Go.

## Documentação
- A documentação afirma roteamento mesh B.A.T.M.A.N. inteligente, mas o código oficial é TCP estático.
""",

    "20_RECOMMENDATIONS.md": """# 20 RECOMMENDATIONS

Nenhuma destas recomendações deve ser executada sem planejamento futuro.

1. **CRÍTICO:** Trocar imediatamente o modelo de dados de arquivo `ledger.json` inteiro por um banco Key-Value DB, gravando blocos por chaves binárias.
2. **CRÍTICO:** Implementar salvamento atômico (`file.tmp` -> rename `file.json`) na função `saveLedger()` antes da troca de DB.
3. **CRÍTICO:** Adicionar rate limits severos no socket TCP e checagem de assinatura *antes* de deserializar pacotes de 100MB de transações forjadas.
4. **ALTO:** Proteger o Cloud Upload. Um nó da Cloud deve requerer prova de posse de chave privada válida com saldo, ou PoW Hashcash antes de alocar 100MB de seu disco.
5. **MÉDIO:** Refatorar pacote `main` do `pc_node` em submódulos isolados: `network`, `core`, `storage`.
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
