# 13 CALL GRAPH

## Principais Chamadas entre Módulos
- **Network -> Ledger:** Quando um peer envia um TCP payload contendo transação, a função `HandleNewTransaction` é acionada (cruzamento do pacote de socket para a Mempool).
- **Consensus -> Crypto:** Durante a mineração ou validação (`VerifyNeonHash`), o consenso delega a geração do `sha256` e a checagem `ecdsa.Verify` à infraestrutura criptográfica em Go.
- **Ledger -> Storage:** Ao acatar um bloco em RAM, o Ledger invoca `saveLedger`, cruzando para a interface de File System.
