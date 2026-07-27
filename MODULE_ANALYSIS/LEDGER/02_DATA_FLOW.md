# 02 DATA FLOW

## Recebimento de Transação
1. **Origem:** Cliente (via `TCP NEW_TRANSACTION`).
2. **Entrada:** `HandleNewTransaction(tx)`.
3. **Validação Cryptográfica:** `VerifyTransaction()` recalcula hash da string Dart e verifica ECDSA.
4. **Validação Contábil:** Varre O(N) do `ledger` em Lock Read, e O(M) da Mempool em Lock Mutex para somar/subtrair balance.
5. **Atualização de Estado (RAM):** `PendingTransactions = append(...)`.

## Recebimento de Bloco
1. **Origem:** Rede / Mineiro Local (`TCP NEW_BLOCK`).
2. **Entrada:** `handleNewBlock(block)`.
3. **Bloqueio:** `ledger.mu.Lock()` congela o sistema.
4. **Validação Estrutural:** Index > LastIndex; PreviousHash == LastHash.
5. **Validação PoW:** `VerifyNeonHash()` recalcula a prova O(1).
6. **Validação de Txs Internas:** Re-verifica signatures.
7. **Atualização de Estado (RAM):** `ledger.Chain = append(...)`.
8. **Limpeza:** Remove Txs mineradas do `PendingTransactions`.
9. **Hooks:** Dispara Goroutine para Cloud (a cada 10 blocos).
10. **Persistência:** Dispara Goroutine para `saveLedger()`.
