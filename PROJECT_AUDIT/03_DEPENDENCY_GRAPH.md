# 03 DEPENDENCY GRAPH

## Dependências de Linguagem
- **Go Modules (`go.mod`)**: Dependências em `github.com/decred/dcrd/dcrec/secp256k1/v4` para curva elíptica e ECDSA.
- **Python (`requirements.txt`)**: Dependências para protótipos de criptografia e rede.
- **Flutter (`pubspec.yaml`)**: Plugins de UI, armazenamento local, criptografia.

## Acoplamentos e Riscos
- **Acoplamento forte** entre `pc_node` e o sistema de arquivos local (`ledger.json`). Não utiliza um banco de dados KV puro (como LevelDB ou RocksDB).
- **Dependência Cíclica / Forte** na propagação de blocos e transações entre `network.go` e `ledger.go` usando mutex globais (`ledger.mu`, `mempoolMutex`), o que pode causar gargalos.
