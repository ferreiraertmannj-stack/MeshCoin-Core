# 04 LOCK ANALYSIS

## Regiões Críticas e Locks
1. **`ledger.mu` (RWMutex)**
   - `Lock()`: Em `handleNewBlock` (duração longa: processa hashes e Txs); Em `saveLedger` (duração longa: MarshalIndent da chain toda).
   - `RLock()`: Em `getLedgerJSON` e em `HandleNewTransaction` (varre blocos).
   - **Risco de Contenção:** Muito Alto. A validação matemática (PoW) do bloco é feita *dentro* do Lock. Sincronização do disco em `saveLedger` tenta obter o mesmo Lock gerando engasgos severos na recepção.
2. **`mempoolMutex` (Mutex)**
   - `Lock()`: Em `HandleNewTransaction` (para bater salto e inserir) e `handleNewBlock` (para purgar). Duração curta.
3. **`tcpMutex` (Mutex)**
   - `Lock()`: Em `handleConnection` (add/remove) e `broadcastTCP` (loop de write).
   - **Risco de Starvation:** `broadcastTCP` faz `.Write` em socket de rede DENTRO do lock de mapa, o que é um anti-pattern letal para alta carga de rede, pois um cliente TCP lento travará o envio a todos os outros.
