# 07 CRITICAL SECTIONS

As três seções de maior perigo identificadas:
1. **Loop de Envio TCP (`network.go`)**: O `broadcastTCP` itera sobre `activeTCPClients` segurando o `tcpMutex`. Se um cliente estiver travado (timeout), nenhum outro cliente recebe mensagens.
2. **Loop de Envio WS (`main.go`)**: Map `clients` sendo modificado (`clients[ws] = true`) na thread de conexão sem Mutex, e iterado na thread `handleMessages()`, violando as regras do Go map (Data Race e fatal panic).
3. **Validação de Bloco (`ledger.go`)**: `VerifyNeonHash` roda pesado na CPU (algoritmo Memory-Hard com loop). Atualmente roda dentro de `ledger.mu.Lock()`! O nó inteiro para e não aceita mais transações até a validação criptográfica acabar.
