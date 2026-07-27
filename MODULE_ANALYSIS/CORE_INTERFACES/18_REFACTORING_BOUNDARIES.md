# 18 REFACTORING BOUNDARIES

## Limites Seguros
- **Isolados:** O `Crypto` e `Consensus (Hash Alg)` podem ser alterados isoladamente, pois recebem um Block e devolvem um booleano.
- **Acoplados:** `Storage` e `Ledger` precisam ser alterados JUNTOS. Mudar de JSON para LevelDB obriga alterar todo o State Machine do Ledger.
- **Risco Maior:** Trocar TCP iterativo por Canais (Actor Model) no `Network`. Mudará todo o ciclo de vida dos nós.
