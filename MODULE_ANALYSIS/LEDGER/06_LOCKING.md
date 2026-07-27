# 06 LOCKING

### Mutexes Ativos
1. **`ledger.mu` (sync.RWMutex)**:
   - Bloqueia *toda* a escrita do Ledger.
   - Chamado em: `saveLedger()`, `handleNewBlock()`.
   - Read-lock (`RLock`) chamado em `HandleNewTransaction()` e `getLedgerJSON()`.
   
2. **`mempoolMutex` (sync.Mutex)**:
   - Protege apenas o array global `PendingTransactions`.
   - Chamado em: `handleNewBlock()` (durante a limpeza), e `HandleNewTransaction()`.

### Riscos de Deadlock e Concorrência
- `saveLedger` (chamado via goroutine) dá Lock. Se ocorrerem múltiplos `NEW_BLOCK` num burst da rede, a fila do Mutex do Go forçará a serialização dessas chamadas, paralisando a rede (TCP handler esperando `handleNewBlock`).
