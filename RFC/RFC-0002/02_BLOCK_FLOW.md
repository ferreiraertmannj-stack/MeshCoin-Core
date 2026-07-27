# 02 BLOCK FLOW

## Mapeamento de Execução (Block)
1. **Quem chama:** `handleConnection` (via TCP NEW_BLOCK) ou `handleWebSocket`.
2. **Quem é chamado:** `handleNewBlockPacket` -> `handleNewBlock` -> `VerifyNeonHash` -> `VerifyTransaction`.
3. **Ordem de execução:** 
   - Decode do JSON para struct `Block`.
   - `handleNewBlock` acionado.
   - **Gargalo:** `ledger.mu.Lock()` congela TODO o ledger (Escrita exclusiva).
   - Validação estrutural: Index e PreviousHash.
   - Validação PoW: `VerifyNeonHash` roda matemática com 4KB vector array.
   - Validação Interna: `VerifyTransaction` repetido para cada Tx interna.
   - Acatado: `ledger.Chain = append(ledger.Chain, block)`.
   - Remoção de Txs confirmadas do Mempool via `mempoolMutex.Lock()`.
   - `ledger.mu.Unlock()` liberado via defer!
   - `go saveLedger()` disparado paralelamente (que tentará re-adquirir o `ledger.mu.Lock()`!).
4. **Locks:** `ledger.mu.Lock()` mantido *durante toda validação matemática do bloco*.
5. **Estruturas:** `Block`, `Ledger`, `PendingTransactions`.
6. **I/O:** `go saveLedger()` aciona persistência em disco JSON.
7. **Rede:** Respondido ACK de sucesso ou rejeição no socket.
