# 18 KNOWN BUGS

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
