# 07 TEST DATASETS

Conjuntos necessários de simulação estruturada:

1. **GENESIS_DATASET:** O bloco Gênesis padrão em múltiplos formatos e cadeias derivadas de 1 a 10 blocos "limpos" para testes de reorg/long-chain.
2. **BAD_ACTORS:** Arquivos JSON com blocos forjados (assinatura falsa, hash manipulado, índice errado, transação double-spend, saldo fantasma).
3. **NETWORK_FLOOD_PAYLOADS:** Pacotes brutos de TCP malformados e estourando 100MB de limite de upload de cloud.
4. **MEMPOOL_DUMP:** 5.000 transações válidas pré-computadas para teste de engasgo (throughput testing) e lock contention no Validator Node.
5. **CORRUPTED_LEDGER:** Simulações de arquivo `ledger.json` cortados pela metade (Null byte injetado) para teste do "Safe Init" e "Atomic Save".
