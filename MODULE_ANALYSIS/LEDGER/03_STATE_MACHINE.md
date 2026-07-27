# 03 STATE MACHINE

## Máquina de Estado do Bloco Mestre (Ledger)

**Estado Inicial: DISCONNECTED/UNINITIALIZED**
- Dispara `initLedger()`.
- Lê arquivo. Se Válido -> **Estado: SYNCED (RAM)**.
- Se Inválido/Inexistente -> Subscreve Gênesis -> `saveLedger()` -> **Estado: SYNCED (RAM)**.

**Estado: SYNCED (RAM)**
- *Evento:* Recebe `NEW_BLOCK`
- *Transição:* Vai para **VALIDATING_BLOCK**.

**Estado: VALIDATING_BLOCK**
- *Locks Ativos:* `ledger.mu.Lock()`
- Se Sucesso: Adiciona ao Array, limpa Mempool, chama IO Assíncrono -> Volta para **SYNCED**.
- Se Falha (Fork, Bad PoW, Bad Tx): Desbloqueia, descarta bloco -> Volta para **SYNCED**.
