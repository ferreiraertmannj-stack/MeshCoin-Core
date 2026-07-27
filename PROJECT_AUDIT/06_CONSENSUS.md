# 06 CONSENSUS

## Fluxo Completo
O consenso se baseia majoritariamente na Prova de Trabalho (PoW - NeonHash). Ao receber um bloco via TCP (em `handleNewBlockPacket`), o nó valida se `PreviousHash` corresponde ao último bloco, e recalcula o PoW.

## Forks e Rollback
**INEXISTENTE.** Atualmente o código apenas rejeita blocos (retornando `false`) se o índice for `<=` ao atual ou se o `PreviousHash` for inválido. Não há implementação de reestruturação de cadeia (reorg/rollback) se uma cadeia paralela mais pesada for detectada.

## Mempool
A mempool (`PendingTransactions`) é um array global em memória protegido por um Mutex (`mempoolMutex`). Quando um bloco é validado, as transações nele contidas são removidas da mempool. Transações são transmitidas via P2P.

## Estado Atual
Altamente simplificado. É um consenso frágil para redes distribuídas instáveis, sendo mais semelhante a uma cadeia de logs anexáveis com verificação de hash.
