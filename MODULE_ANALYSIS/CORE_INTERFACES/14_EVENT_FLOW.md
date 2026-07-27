# 14 EVENT FLOW

## Eventos Importantes Documentados
1. **Nova Transação:** Dispara validação local -> Adiciona em RAM (Mempool) -> Faz Broadcast UDP/TCP para vizinhos.
2. **Novo Bloco:** Dispara hash match -> Verifica assinaturas internas -> Acata bloco na Chain -> Invalida Txs duplicadas na Mempool -> Salva no disco.
3. **Novo Peer:** Dispara `OGM` (Originator Message) no estilo B.A.T.M.A.N -> Adiciona no map `activeTCPClients`.
4. **Sincronização:** (Flow falho atual) Nó aguarda novo bloco e insere à frente, ignorando o passado vazio.
