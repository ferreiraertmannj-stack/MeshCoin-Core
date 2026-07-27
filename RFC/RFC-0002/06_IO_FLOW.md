# 06 IO FLOW

## Mapeamento de I/O em Disco
1. **Operação Inicial:** `initLedger()` (leitura do disco para RAM).
2. **Operação Contínua:** `saveLedger()` chamado via `go saveLedger()` ao receber um bloco.
   - Obtém `ledger.mu.Lock()`.
   - Roda `json.MarshalIndent`.
   - Roda `os.CreateTemp`, `Write`, `Sync`, `os.Rename`.
3. **Operação de Cloud:** `uploadToNebulaCloud` (mock no código, envia para cloud externa).
4. **Problema I/O:** `go saveLedger` bloqueia as threads TCP porque disputa o mesmo Mutex com `handleNewBlock`. O dump atinge O(N) do tamanho da Chain.
