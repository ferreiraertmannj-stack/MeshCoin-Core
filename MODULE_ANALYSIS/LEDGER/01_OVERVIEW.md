# 01 OVERVIEW

## Ledger Module (`pc_node/ledger.go`)

### Tipos e Estruturas
1. **Transaction**: Representa uma transferência na rede. Possui chaves, assinaturas (ECDSA `secp256k1` e stub PQC `Dilithium`), timestamps, taxas e valores.
2. **Block**: Container indexado de transações, carimbado temporalmente, possuindo referência ao bloco anterior (`PreviousHash`), raiz de Merkle (não implementado o preenchimento, apenas campo existente), Nonce, dados da Cloud (`MinerStorage`, `StorageType`) e seu próprio Hash final (NeonHash).
3. **Ledger**: Agrupador principal contendo a cadeia `Chain []Block` e um bloqueio global de Leitura/Escrita `sync.RWMutex`.

### Variáveis Globais
- `ledger`: Instância singleton persistindo o estado primário.
- `PendingTransactions`: Mempool array em RAM.
- `mempoolMutex`: Lock isolado para acesso à Mempool.
- `ledgerFile`: Path constante `ledger.json`.

### Dependências (Imports)
- Nativos Go: `crypto/sha256`, `encoding/hex`, `encoding/json`, `fmt`, `io/ioutil`, `log`, `strings`, `sync`.
- Externos (Consensus/Wallet): `github.com/decred/dcrd/dcrec/secp256k1/v4` (ECDSA parser e verifier).
