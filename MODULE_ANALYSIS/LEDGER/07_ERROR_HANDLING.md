# 07 ERROR HANDLING

### Falhas de Parsing / Corrupção
- O `initLedger()` intercepta qualquer erro de Unmarshal (seja corrupção, seja arquivo em branco) e aplica o bypass: sobrescreve com um array contendo apenas o Bloco 0 (Gênesis). É uma estratégia destrutiva de fail-open onde falhar custa toda a base de dados em vez de um `panic`.

### Falhas de Assinatura
- `VerifyTransaction()` faz um bail-out seguro em chaves ruins ou strings menores que 128 chars. Apenas descarta a Tx. Não penaliza (ban) a conexão TCP chamadora porque essa função é pura, isolada da rede.

### Falhas de Hash
- `VerifyNeonHash()` também pura. Falha silenciosamente logando o erro sem quebrar o laço de execução principal.
