# 08 CALL GRAPH

## Fluxo Textual Real (Recepção de Bloco)

```text
[Network Socket TCP] ou [WebSocket]
↓
handleConnection / handleWebSocket
↓
handleNewBlockPacket (faz json decode)
↓
handleNewBlock
|
├── [LOCK: ledger.mu.Lock()]
|
├── VerifyNeonHash (pesado na CPU)
|
├── VerifyTransaction (loop p/ cada tx)
|
├── append(ledger.Chain)
|
├── [LOCK: mempoolMutex.Lock()]
|     └── Purgar mempool
|     └── [UNLOCK]
|
├── [UNLOCK: ledger.mu.Unlock()]
↓
go saveLedger() (Assíncrono)
|
├── [LOCK: ledger.mu.Lock()] (Aguarda liberação)
|
└── I/O Atômico em Disco (Cria, Write, Sync, Rename)
      └── [UNLOCK: ledger.mu.Unlock()]
```
