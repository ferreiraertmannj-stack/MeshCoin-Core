# 15 SHARED STRUCTURES

## Structs Centrais
1. **Block**: Transita entre `Network` (Socket), `Consensus` (Miner), `Ledger` (Array) e `Storage` (JSON). É o artefato de mais alto acoplamento do sistema.
2. **Transaction**: Transita entre `Wallet` (Assinatura UI), `Network` (Payload UDP/TCP), `Ledger` (Mempool), `Consensus` (Validação individual).
