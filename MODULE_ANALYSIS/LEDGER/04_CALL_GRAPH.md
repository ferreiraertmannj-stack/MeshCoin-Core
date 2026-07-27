# 04 CALL GRAPH

## Grafos Textuais de Chamada

**Fluxo de Injeção de Bloco:**
```text
(Rede/TCP) -> handleNewBlock()
                 │
                 ├─> VerifyNeonHash()
                 │       └─> calculateHash()
                 │
                 ├─> VerifyTransaction() [Para cada Tx]
                 │       ├─> formatDartDouble()
                 │       └─> secp256k1.ParsePubKey()
                 │       └─> ecdsa.Verify()
                 │
                 ├─> (Mempool Cleanup)
                 │
                 ├─> go uploadToNebulaCloud() [A cada 10]
                 │
                 └─> go saveLedger() [I/O Assíncrono]
```

**Dependências Entrantes:**
- `pc_node/network.go` -> Chama `handleNewBlock`, `HandleNewTransaction` e `getLedgerJSON`.
- `pc_node/node_daemon.go` (Cloud) -> Lê blocos para upload.
