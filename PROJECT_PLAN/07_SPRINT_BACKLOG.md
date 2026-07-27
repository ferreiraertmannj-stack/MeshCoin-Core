# 07 SPRINT BACKLOG

## Tarefa: Migração para LevelDB
- **Descrição:** Reescrever `saveLedger` e a estrutura base do array `Chain` para leitura/gravação indexada via goleveldb.
- **Dependências:** Nenhuma.
- **Estimativa:** 13 pontos (Alta)
- **Criticidade:** Crítica
- **Critério de aceite:** O nó inicializa lendo do DB. O salvamento não sobrecarrega a RAM, testes passam.

## Tarefa: Atomic Fallback temporário
- **Descrição:** Enquanto M1 não termina, alterar `ioutil.WriteFile` para gravar em `tmp` e dar rename, protegendo JSON.
- **Dependências:** Nenhuma.
- **Estimativa:** 2 pontos (Baixa)
- **Criticidade:** Crítica
- **Critério de aceite:** Ao simular queda de energia no save, o arquivo anterior não é corrompido.

## Tarefa: Autenticação na API Cloud
- **Descrição:** Modificar handlers HTTP do `node_daemon.go` para validar assinatura ECDSA.
- **Dependências:** Assinatura na Wallet já funcional (Dart e Go).
- **Estimativa:** 5 pontos (Média)
- **Criticidade:** Crítica
- **Critério de aceite:** HTTP 401 para uploads forjados ou maiores que 100MB; HTTP 200 para uploads corretos.

## Tarefa: Remover Mutex Global
- **Descrição:** Utilizar Channels em `pc_node/ledger.go` para enfileirar requests de `NEW_BLOCK` num único state-manager thread.
- **Dependências:** Tarefa N1 LevelDB concluída (para não criar fila infinita bloqueada por IO lento).
- **Estimativa:** 8 pontos (Alta)
- **Criticidade:** Alta
- **Critério de aceite:** Testes de stress de blocos concorrentes viajam na fila sem causar Data Race.

## Tarefa: Roteamento Mesh via Kademlia
- **Descrição:** Integrar ou construir abstração DHT em substituição ao Broadcast de UDP e map de TCP fixo em `network.go`.
- **Dependências:** Mutex Removido.
- **Estimativa:** 21 pontos (GG)
- **Criticidade:** Alta
- **Critério de aceite:** Nó PC em SP descobre Nó PC na Alemanha via DHT bootstrap.
