# 01 TRANSACTION FLOW

## Mapeamento de Execução (Tx)
1. **Quem chama:** `handleConnection` (network.go) ou `handleWebSocket` (main.go).
2. **Quem é chamado:** `handleNewTransactionPacket` -> `HandleNewTransaction` -> `VerifyTransaction`.
3. **Ordem de execução:** 
   - Parse JSON do payload TCP/WS para struct.
   - `HandleNewTransaction(tx)` é acionado.
   - `VerifyTransaction(tx)` verifica ECDSA e Hash (sem Lock).
   - `ledger.mu.RLock()` acionado.
   - Varredura O(N) no `ledger.Chain` para contabilidade de saldo local (cálculo iterativo).
   - `ledger.mu.RUnlock()` liberado.
   - `mempoolMutex.Lock()` acionado.
   - Abate saldo de Txs pendentes.
   - Rejeita se saldo < Tx.Amount + Tx.Fee.
   - Anexa à `PendingTransactions`.
   - `mempoolMutex.Unlock()` liberado.
   - Retorno booleano em cascata.
4. **Locks:** `ledger.mu.RLock` e `mempoolMutex.Lock`.
5. **Estruturas:** `Transaction`, `PendingTransactions`.
6. **I/O:** Nenhum (puramente em memória).
7. **Rede:** Após sucesso, `broadcastTCP` faz broadcast para vizinhos e WebSockets.
