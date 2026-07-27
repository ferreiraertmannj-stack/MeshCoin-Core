# 15 SECURITY

## Ataques Possíveis
- **Double Spend e Replay:** Alta probabilidade de sucesso devido à falta aparente de rastreamento de nonces por endereço de envio ou verificação aprofundada de modelo de conta vs. UTXO de maneira forte.
- **Spoofing e Flood:** Faltam filtros de rate-limit severos no `pc_node/network.go`. Um atacante conectando no TCP pode inundar o nó de blocos inválidos.
- **DDoS no Cloud Node:** Qualquer usuário pode usar o endpoint de `/upload` do `node_daemon.go` sem autenticação.
- **Integridade:** `saveLedger()` não salva para um arquivo temporário antes do rename atômico. Falha elétrica corromperá o `ledger.json` irremediavelmente.
