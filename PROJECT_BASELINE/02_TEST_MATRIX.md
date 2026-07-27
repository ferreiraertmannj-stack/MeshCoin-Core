# 02 TEST MATRIX

## Blockchain
- `test_append_block`: Valida se o bloco adicionado à chain incrementa índice e atualiza hash.
- `test_corrupt_ledger_init`: Garante panic/recuperação ao invés de sobrescrever genesis.

## Consensus
- `test_neonhash_valid`: Verifica corretude da prova (nonce, hash, difficulty).
- `test_reject_invalid_pow`: Garante que pacotes de rede com hashes falsos sejam barrados na porta sem lock.
- `test_fork_resolution`: Força submissão de cadeia paralela maior.

## Wallet
- `test_key_generation`: Valida curva secp256k1 e geração de endereço Base58Check.
- `test_transaction_sign`: Valida assinatura ECDSA da transação.
- `test_invalid_signature`: Transação forjada deve ser descartada.

## Mesh
- `test_broadcast_loop`: Simula topologia anelar; garante que o pacote TCP morre por TTL/ID cache.
- `test_dht_discovery`: Validar encontro de peers WAN.

## Mining
- `test_halving_reward`: Teste matemático da fórmula com storage bonuses variados.

## Networking
- `test_tcp_flood`: Conexões massivas para testar OOM.
- `test_rate_limit`: Bloqueia envio de +10 blocos/seg do mesmo peer.

## Nebula Cloud
- `test_cloud_upload_auth`: Falha intencional de POST sem signature header.
- `test_file_chunking`: Teste do algoritmo Reed-Solomon.

## Storage
- `test_leveldb_write`: Mock de migração para DB nativo, valida ACID (Atomic).

## Synchronization
- `test_fast_sync`: Nó zera e demanda cadeia completa aos peers em batch.

## Crypto
- `test_pqc_stub`: Testar futura assinatura Dilithium simulada.
